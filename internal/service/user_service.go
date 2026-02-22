package service

import (
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"
	"time"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(req *model.RegisterRequest) (*model.User, error) {
	now := time.Now()
	user := &model.User{
		Username:   req.Username,
		Fullname:   req.Fullname,
		UserType:   req.UserType,
		Password:   req.Password,
		CreateDate: now,
		ModifyDate: now,
	}

	err := s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
