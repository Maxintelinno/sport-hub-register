package repository

import (
	"sport-hub-register/internal/model"

	"gorm.io/gorm"
)

type OTPRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *OTPRepository) CreateOTP(tx *gorm.DB, otp *model.OTPRequestRecord) error {
	return r.getDB(tx).Create(otp).Error
}

func (r *OTPRepository) FindLatestByPhone(tx *gorm.DB, phone string) (*model.OTPRequestRecord, error) {
	var otp model.OTPRequestRecord
	err := r.getDB(tx).Where("phone = ?", phone).Order("created_at DESC").First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *OTPRepository) IncrementAttempts(tx *gorm.DB, id string) error {
	return r.getDB(tx).Model(&model.OTPRequestRecord{}).Where("id = ?", id).
		Update("attempts", gorm.Expr("attempts + 1")).Error
}

func (r *OTPRepository) DeleteOTP(tx *gorm.DB, id string) error {
	return r.getDB(tx).Delete(&model.OTPRequestRecord{}, "id = ?", id).Error
}

func (r *OTPRepository) DeleteByPhone(tx *gorm.DB, phone string) error {
	return r.getDB(tx).Delete(&model.OTPRequestRecord{}, "phone = ?", phone).Error
}
