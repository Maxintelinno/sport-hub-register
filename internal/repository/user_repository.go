package repository

import (
	"sport-hub-register/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *UserRepository) CreateUser(tx *gorm.DB, user *model.User) error {
	return r.getDB(tx).Create(user).Error
}

func (r *UserRepository) FindByPhone(tx *gorm.DB, phone string) (*model.User, error) {
	var user model.User
	err := r.getDB(tx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
