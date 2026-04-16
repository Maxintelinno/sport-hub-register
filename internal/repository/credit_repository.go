package repository

import (
	"sport-hub-register/internal/model"

	"gorm.io/gorm"
)

type CreditRepository struct {
	db *gorm.DB
}

func NewCreditRepository(db *gorm.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) GetByUserID(tx *gorm.DB, userID string) (*model.BookingCredit, error) {
	var credit model.BookingCredit
	db := r.db
	if tx != nil {
		db = tx
	}
	err := db.Where("user_id = ?", userID).First(&credit).Error
	if err != nil {
		return nil, err
	}
	return &credit, nil
}
