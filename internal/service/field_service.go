package service

import (
	"errors"
	"sort"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"

	"gorm.io/gorm"
)

const (
	IMAGE_URL = "https://t3.storageapi.dev/stocked-pocket-jm-kiclnxm/"
)

type FieldService struct {
	db             *gorm.DB
	repo           *repository.FieldRepository
	userRepo       *repository.UserRepository
	storageService *StorageService
}

func NewFieldService(db *gorm.DB, repo *repository.FieldRepository, userRepo *repository.UserRepository, storageService *StorageService) *FieldService {
	return &FieldService{
		db:             db,
		repo:           repo,
		userRepo:       userRepo,
		storageService: storageService,
	}
}

func (s *FieldService) CreateField(req *model.CreateFieldRequest) (*model.Field, error) {
	var field *model.Field

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Validate owner exists
		err := s.userRepo.GetDB(tx).Where("id = ?", req.OwnerID).First(&model.User{}).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("owner not found")
			}
			return err
		}

		// 2. Map request to model
		field = &model.Field{
			OwnerID:      req.OwnerID,
			Name:         req.Name,
			SportType:    req.SportType,
			PricePerHour: float64(req.PricePerHour),
			OpenTime:     req.OpenTime,
			CloseTime:    req.CloseTime,
			Province:     req.Province,
			District:     req.District,
			AddressLine:  req.AddressLine,
			Description:  req.Description,
			Status:       "active",
		}

		// 3. Save Field
		if err := s.repo.CreateField(tx, field); err != nil {
			return err
		}

		// 4. Save Images if any
		if len(req.Images) > 0 {
			fieldImages := make([]model.FieldImage, len(req.Images))
			for i, imgReq := range req.Images {
				fieldImages[i] = model.FieldImage{
					FieldID:   field.ID,
					ObjectKey: imgReq.ObjectKey,
					SortOrder: imgReq.SortOrder,
					ImageUrl:  IMAGE_URL,
				}
			}
			if err := s.repo.CreateFieldImages(tx, fieldImages); err != nil {
				return err
			}
			field.Images = fieldImages
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return field, nil
}

func (s *FieldService) UpdateField(id string, req *model.UpdateFieldRequest) (*model.Field, error) {
	var field *model.Field

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Find existing field
		var err error
		field, err = s.repo.FindFieldByID(tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("field not found")
			}
			return err
		}

		// 2. Validate owner
		if field.OwnerID != req.OwnerID {
			return errors.New("unauthorized: you do not own this field")
		}

		// 3. Update field fields
		field.Name = req.Name
		field.SportType = req.SportType
		field.PricePerHour = float64(req.PricePerHour)
		field.OpenTime = req.OpenTime
		field.CloseTime = req.CloseTime
		field.Province = req.Province
		field.District = req.District
		field.AddressLine = req.AddressLine
		field.Description = req.Description

		if err := s.repo.UpdateField(tx, field); err != nil {
			return err
		}

		// 4. Update Images (Delete all and re-add)
		if err := s.repo.DeleteFieldImages(tx, id); err != nil {
			return err
		}

		if len(req.Images) > 0 {
			fieldImages := make([]model.FieldImage, len(req.Images))
			for i, imgReq := range req.Images {
				fieldImages[i] = model.FieldImage{
					FieldID:   field.ID,
					ObjectKey: imgReq.ObjectKey,
					SortOrder: imgReq.SortOrder,
					ImageUrl:  IMAGE_URL,
				}
			}
			if err := s.repo.CreateFieldImages(tx, fieldImages); err != nil {
				return err
			}
			field.Images = fieldImages
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return field, nil
}

func (s *FieldService) GetFieldsByOwnerID(ownerID string) ([]model.Field, error) {
	// 1. Fetch fields
	fields, err := s.repo.FindFieldsByOwnerID(nil, ownerID)
	if err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		return fields, nil
	}

	// 2. Collect field IDs
	fieldIDs := make([]string, len(fields))
	for i, f := range fields {
		fieldIDs[i] = f.ID.String()
	}

	// 3. Fetch images for those field IDs
	images, err := s.repo.FindImagesByFieldIDs(nil, fieldIDs)
	if err != nil {
		return nil, err
	}

	// 4. Map images to fields
	imageMap := make(map[string][]model.FieldImage)

	for _, img := range images {
		fID := img.FieldID.String()

		// 🔥 Generate presigned GET URL จาก object_key
		imageURL, err := s.storageService.GeneratePresignedGetURL(img.ObjectKey)
		if err != nil {
			return nil, err
		}

		img.ImageUrl = imageURL

		imageMap[fID] = append(imageMap[fID], img)
	}

	// 5. Attach images + thumbnail
	for i := range fields {
		fID := fields[i].ID.String()

		if imgs, ok := imageMap[fID]; ok {

			// sort images ตาม sort_order
			sort.Slice(imgs, func(a, b int) bool {
				return imgs[a].SortOrder < imgs[b].SortOrder
			})

			fields[i].Images = imgs

			// ตั้ง thumbnail จากรูปแรก
			if len(imgs) > 0 {
				fields[i].ThumbnailUrl = imgs[0].ImageUrl
			}

		} else {
			fields[i].Images = []model.FieldImage{}
			fields[i].ThumbnailUrl = ""
		}
	}

	return fields, nil
}

func (s *FieldService) UpdateFieldStatus(req *model.UpdateFieldStatusRequest) error {
	// 1. Find existing field
	field, err := s.repo.FindFieldByID(nil, req.FieldID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("field not found")
		}
		return err
	}

	// 2. Validate owner
	if field.OwnerID != req.OwnerID {
		return errors.New("unauthorized: you do not own this field")
	}

	// 3. Update status
	return s.repo.UpdateFieldStatus(nil, req.FieldID.String(), req.Status)
}

func (s *FieldService) GetFieldsBySection(section, province string, lat, lng float64, limit, offset int) ([]model.Field, error) {
	var fields []model.Field
	var err error

	//GET /v1/fields?section=all&limit=10&offset=0
	//GET /v1/fields?section=popular&limit=10&offset=0
	//GET /v1/fields?section=nearby&lat=13.7563&lng=100.5018&limit=10
	//GET /v1/fields?section=province&province=กรุงเทพมหานคร&limit=10

	// 1. Fetch fields based on section
	switch section {
	case "province":
		fields, err = s.repo.FindFieldsWithPagination(nil, province, limit, offset)
	case "nearby":
		// For nearby, typically we fetch a larger set then sort or use spatial query.
		// For simplicity, we'll fetch all and sort by distance in memory for now.
		fields, err = s.repo.FindAllFields(nil)
	case "popular":
		// Placeholder: sort by created_at or id for now
		fields, err = s.repo.FindFieldsWithPagination(nil, "", limit, offset)
	default: // "all" or any other
		fields, err = s.repo.FindFieldsWithPagination(nil, "", limit, offset)
	}

	if err != nil {
		return nil, err
	}

	if len(fields) == 0 {
		return fields, nil
	}

	// 2. Fetch images and generate presigned URLs
	fieldIDs := make([]string, len(fields))
	for i, f := range fields {
		fieldIDs[i] = f.ID.String()
	}

	images, err := s.repo.FindImagesByFieldIDs(nil, fieldIDs)
	if err != nil {
		return nil, err
	}

	imageMap := make(map[string][]model.FieldImage)
	for _, img := range images {
		fID := img.FieldID.String()
		imageURL, err := s.storageService.GeneratePresignedGetURL(img.ObjectKey)
		if err == nil {
			img.ImageUrl = imageURL
		}
		imageMap[fID] = append(imageMap[fID], img)
	}

	// 3. Attach images and calculate distance for nearby
	for i := range fields {
		fID := fields[i].ID.String()
		if imgs, ok := imageMap[fID]; ok {
			sort.Slice(imgs, func(a, b int) bool {
				return imgs[a].SortOrder < imgs[b].SortOrder
			})
			fields[i].Images = imgs
			if len(imgs) > 0 {
				fields[i].ThumbnailUrl = imgs[0].ImageUrl
			}
		} else {
			fields[i].Images = []model.FieldImage{}
		}
	}

	// 4. Special sorting for nearby
	if section == "nearby" && lat != 0 && lng != 0 {
		sort.Slice(fields, func(i, j int) bool {
			distI := (fields[i].Latitude-lat)*(fields[i].Latitude-lat) + (fields[i].Longitude-lng)*(fields[i].Longitude-lng)
			distJ := (fields[j].Latitude-lat)*(fields[j].Latitude-lat) + (fields[j].Longitude-lng)*(fields[j].Longitude-lng)
			return distI < distJ
		})
		// Apply pagination after sorting
		start := offset
		if start > len(fields) {
			return []model.Field{}, nil
		}
		end := offset + limit
		if end > len(fields) {
			end = len(fields)
		}
		fields = fields[start:end]
	}

	return fields, nil
}
