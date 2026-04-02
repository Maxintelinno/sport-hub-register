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

func (r *UserRepository) GetDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *UserRepository) FindByID(tx *gorm.DB, id string) (*model.User, error) {
	var user model.User
	err := r.GetDB(tx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(tx *gorm.DB, user *model.User) error {
	return r.GetDB(tx).Create(user).Error
}

func (r *UserRepository) FindByPhone(tx *gorm.DB, phone string) (*model.User, error) {
	var user model.User
	err := r.GetDB(tx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(tx *gorm.DB, username string) (*model.User, error) {
	var user model.User
	err := r.GetDB(tx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateOwnerStaff(tx *gorm.DB, staff *model.OwnerStaff) error {
	return r.GetDB(tx).Create(staff).Error
}
