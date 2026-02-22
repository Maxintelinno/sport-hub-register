package repository

import (
	"sport-hub-register/internal/model"

	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *TokenRepository) CreateToken(tx *gorm.DB, token *model.RegistrationToken) error {
	return r.getDB(tx).Create(token).Error
}

func (r *TokenRepository) FindByPhone(tx *gorm.DB, phone string) (*model.RegistrationToken, error) {
	var token model.RegistrationToken
	err := r.getDB(tx).Where("phone = ?", phone).Order("created_at DESC").First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepository) DeleteToken(tx *gorm.DB, id string) error {
	return r.getDB(tx).Delete(&model.RegistrationToken{}, "id = ?", id).Error
}
