package repository

import (
	"sport-hub-register/internal/model"

	"gorm.io/gorm"
)

type CourtRepository struct {
	db *gorm.DB
}

func NewCourtRepository(db *gorm.DB) *CourtRepository {
	return &CourtRepository{db: db}
}

func (r *CourtRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *CourtRepository) CreateCourt(tx *gorm.DB, court *model.FieldCourt) error {
	return r.getDB(tx).Create(court).Error
}

func (r *CourtRepository) CreateCourts(tx *gorm.DB, courts []model.FieldCourt) error {
	if len(courts) == 0 {
		return nil
	}
	return r.getDB(tx).Create(&courts).Error
}

func (r *CourtRepository) FindCourtsByFieldID(tx *gorm.DB, fieldID string) ([]model.FieldCourt, error) {
	var courts []model.FieldCourt
	err := r.getDB(tx).Where("field_id = ?", fieldID).Find(&courts).Error
	return courts, err
}

func (r *CourtRepository) FindCourtByID(tx *gorm.DB, id string) (*model.FieldCourt, error) {
	var court model.FieldCourt
	err := r.getDB(tx).Where("id = ?", id).First(&court).Error
	if err != nil {
		return nil, err
	}
	return &court, nil
}
