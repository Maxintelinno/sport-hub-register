package service

import (
	"errors"
	"fmt"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db        *gorm.DB
	repo      *repository.UserRepository
	tokenRepo *repository.TokenRepository
	planRepo  *repository.PlanRepository
	subRepo   *repository.SubscriptionRepository
	fieldRepo *repository.FieldRepository
}

func NewUserService(db *gorm.DB, repo *repository.UserRepository, tokenRepo *repository.TokenRepository, planRepo *repository.PlanRepository, subRepo *repository.SubscriptionRepository, fieldRepo *repository.FieldRepository) *UserService {
	return &UserService{
		db:        db,
		repo:      repo,
		tokenRepo: tokenRepo,
		planRepo:  planRepo,
		subRepo:   subRepo,
		fieldRepo: fieldRepo,
	}
}

func (s *UserService) Register(req *model.RegisterRequest) (*model.UserResponse, error) {
	var user *model.User
	var sub *model.Subscription

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
			Fullname:     req.Fullname,
			PasswordHash: string(hashedPassword),
			Role:         role + "_" + req.Username + "_" + tokenRec.TokenHash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := s.repo.CreateUser(tx, user); err != nil {
			return err
		}

		// 4. Create Subscription for Owners
		if strings.HasPrefix(user.Role, "owner") {
			plan, err := s.planRepo.FindByCode(tx, "free")
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Seed plan if not exists
					plan = &model.Plan{
						ID:           uuid.New(),
						Code:         "free",
						Name:         "Plans Free",
						Description:  "Free plan for initial registration",
						Price:        0,
						BillingCycle: "lifetime",
						TrialDays:    0,
						IsActive:     true,
					}
					if err := s.planRepo.CreatePlan(tx, plan); err != nil {
						return err
					}
				} else {
					return err
				}
			}

			sub = &model.Subscription{
				ID:      uuid.New(),
				UserID:  user.ID,
				PlanID:  plan.ID,
				Status:  "active",
				StartAt: now,
				EndAt:   now.AddDate(10, 0, 0), // 10 years for free plan
			}

			if err := s.subRepo.CreateSubscription(tx, sub); err != nil {
				return err
			}
			sub.Plan = *plan
		}

		// 5. Delete token after successful registration
		if err := s.tokenRepo.DeleteToken(tx, tokenRec.ID.String()); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	res := &model.UserResponse{User: user}
	if sub != nil {
		res.Subscription = &model.UserSubscriptionResponse{
			PlanName:     sub.Plan.Name,
			BillingCycle: sub.Plan.BillingCycle,
			Status:       sub.Status,
		}
	}

	return res, nil
}

func (s *UserService) Login(req *model.LoginRequest) (*model.UserResponse, error) {
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

	parts := strings.Split(user.Role, "_")
	user.Role = parts[0]

	res := &model.UserResponse{User: user}

	// 3. Fetch Subscription for Owners
	if user.Role == "owner" {
		sub, err := s.subRepo.FindLatestByUserID(nil, user.ID.String())
		if err == nil {
			res.Subscription = &model.UserSubscriptionResponse{
				PlanName:     sub.Plan.Name,
				BillingCycle: sub.Plan.BillingCycle,
				Status:       sub.Status,
			}
		}
	}

	return res, nil
}

func (s *UserService) RegisterStaff(ownerID uuid.UUID, req *model.RegisterStaffRequest) (*model.UserResponse, error) {
	// 1. Check if user already exists
	if _, err := s.repo.FindByPhone(nil, req.Phone); err == nil {
		return nil, errors.New("phone number already registered")
	}
	if _, err := s.repo.FindByUsername(nil, req.Username); err == nil {
		return nil, errors.New("username already taken")
	}

	// 2. Fetch owner's field information for Province and District
	fields, err := s.fieldRepo.FindFieldsByOwnerID(nil, ownerID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch owner fields: %v", err)
	}
	if len(fields) == 0 {
		return nil, errors.New("owner must have at least one field registered before adding staff")
	}

	// For simplicity, take the location from the first field
	province := fields[0].Province
	district := fields[0].District

	var user *model.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 3. Hash Password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("0000000000"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// 4. Create User record (bypass Registration Token)
		now := time.Now()
		user = &model.User{
			Phone:        req.Phone,
			Username:     req.Username,
			Fullname:           req.Fullname,
			PasswordHash:       string(hashedPassword),
			Province:           province,                                  // Automatically set from owner's field location
			District:           district,                                  // Automatically set from owner's field location
			Role:               req.Role + "_" + req.Username + "_direct", // Use "direct" instead of tokenHash as it's registered by owner
			Status:             "active",
			MustChangePassword: true,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		if err := s.repo.CreateUser(tx, user); err != nil {
			return err
		}

		// 5. Save to owner_staffs table
		staff := &model.OwnerStaff{
			OwnerUserID: ownerID,
			StaffUserID: user.ID,
			RoleCode:    req.Role,
			Status:      "active",
		}

		if err := s.repo.CreateOwnerStaff(tx, staff); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &model.UserResponse{User: user}, nil
}
