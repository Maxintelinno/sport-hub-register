package service

import (
	"errors"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db        *gorm.DB
	repo      *repository.UserRepository
	tokenRepo *repository.TokenRepository
}

func NewUserService(db *gorm.DB, repo *repository.UserRepository, tokenRepo *repository.TokenRepository) *UserService {
	return &UserService{db: db, repo: repo, tokenRepo: tokenRepo}
}

func (s *UserService) Register(req *model.RegisterRequest) (*model.User, error) {
	var user *model.User

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Verify Registration Token
		tokenRec, err := s.tokenRepo.FindByPhone(tx, req.Phone)
		if err != nil {
			return errors.New("registration token not found or expired")
		}

		if time.Now().After(tokenRec.ExpiresAt) {
			return errors.New("registration token expired")
		}

		// 2. Hash Password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// 3. Create User record
		now := time.Now()
		role := req.Role
		if role == "" {
			role = "user"
		}
		user = &model.User{
			Phone:        req.Phone,
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Role:         role + "_" + req.Username + "_" + tokenRec.TokenHash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := s.repo.CreateUser(tx, user); err != nil {
			return err
		}

		// 4. Delete token after successful registration
		if err := s.tokenRepo.DeleteToken(tx, tokenRec.ID.String()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(req *model.LoginRequest) (*model.User, error) {
	// 1. Find User by Username
	user, err := s.repo.FindByUsername(nil, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 2. Compare Password Hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}
