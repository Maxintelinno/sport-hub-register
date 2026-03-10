package repository

import (
	"sport-hub-register/internal/model"

	"gorm.io/gorm"
)

type FieldRepository struct {
	db *gorm.DB
}

func NewFieldRepository(db *gorm.DB) *FieldRepository {
	return &FieldRepository{db: db}
}

func (r *FieldRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *FieldRepository) CreateField(tx *gorm.DB, field *model.Field) error {
	return r.getDB(tx).Create(field).Error
}

func (r *FieldRepository) CreateFieldImages(tx *gorm.DB, images []model.FieldImage) error {
	if len(images) == 0 {
		return nil
	}
	return r.getDB(tx).Create(&images).Error
}

func (r *FieldRepository) UpdateField(tx *gorm.DB, field *model.Field) error {
	return r.getDB(tx).Save(field).Error
}

func (r *FieldRepository) DeleteFieldImages(tx *gorm.DB, fieldID string) error {
	return r.getDB(tx).Where("field_id = ?", fieldID).Delete(&model.FieldImage{}).Error
}

func (r *FieldRepository) FindFieldByID(tx *gorm.DB, id string) (*model.Field, error) {
	var field model.Field
	err := r.getDB(tx).Where("id = ?", id).First(&field).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

func (r *FieldRepository) FindFieldsByOwnerID(tx *gorm.DB, ownerID string) ([]model.Field, error) {
	var fields []model.Field
	err := r.getDB(tx).Where("owner_id = ?", ownerID).Find(&fields).Error
	if err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *FieldRepository) FindImagesByFieldIDs(tx *gorm.DB, fieldIDs []string) ([]model.FieldImage, error) {
	var images []model.FieldImage
	if len(fieldIDs) == 0 {
		return images, nil
	}
	err := r.getDB(tx).Where("field_id IN ?", fieldIDs).Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}
