package service

import (
	"errors"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"

	"gorm.io/gorm"
)

const (
	IMAGE_URL = "https://t3.storageapi.dev/stocked-pocket-jm-kiclnxm/"
)

type FieldService struct {
	db       *gorm.DB
	repo     *repository.FieldRepository
	userRepo *repository.UserRepository
}

func NewFieldService(db *gorm.DB, repo *repository.FieldRepository, userRepo *repository.UserRepository) *FieldService {
	return &FieldService{
		db:       db,
		repo:     repo,
		userRepo: userRepo,
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
			Status:       "pending_review",
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
		imageMap[fID] = append(imageMap[fID], img)
	}

	for i := range fields {
		fID := fields[i].ID.String()
		if imgs, ok := imageMap[fID]; ok {
			fields[i].Images = imgs
		} else {
			fields[i].Images = []model.FieldImage{} // Ensure it's not nil for consistent JSON response
		}
	}

	return fields, nil
}
