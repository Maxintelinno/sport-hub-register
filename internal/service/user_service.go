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

		err = bcrypt.CompareHashAndPassword([]byte(tokenRec.TokenHash), []byte(req.Token))
		if err != nil {
			return errors.New("invalid registration token")
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
			Role:         role,
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
