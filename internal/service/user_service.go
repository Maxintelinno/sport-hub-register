package service

import (
	"errors"
	"encoding/json"
	"bytes"
	"fmt"
	"net/http"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db        *gorm.DB
	repo      *repository.UserRepository
	tokenRepo *repository.TokenRepository
	planRepo  *repository.PlanRepository
	subRepo   *repository.SubscriptionRepository
}

func NewUserService(db *gorm.DB, repo *repository.UserRepository, tokenRepo *repository.TokenRepository, planRepo *repository.PlanRepository, subRepo *repository.SubscriptionRepository) *UserService {
	return &UserService{
		db:        db,
		repo:      repo,
		tokenRepo: tokenRepo,
		planRepo:  planRepo,
		subRepo:   subRepo,
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
	// 1. Prepare external registration request
	registerURL := "https://sport-hub-register-staging.up.railway.app/register"

	regPayload := map[string]string{
		"phone":    req.Phone,
		"username": req.Username,
		"password": "0000000000",
		"fullname": req.Fullname,
		"role":     req.Role,
	}

	payloadBytes, err := json.Marshal(regPayload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(registerURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to call registration API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Message != "" {
			return nil, fmt.Errorf("registration failed: %s", errResp.Message)
		}
		return nil, fmt.Errorf("registration failed with status code: %d", resp.StatusCode)
	}

	// 2. Decode response to get User data
	var apiResponse struct {
		Status  string              `json:"status"`
		Message string              `json:"message"`
		Data    *model.UserResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode registration response: %v", err)
	}

	if apiResponse.Data == nil || apiResponse.Data.User == nil {
		return nil, errors.New("registration response missing user data")
	}

	staffUserID := apiResponse.Data.User.ID

	// 3. Save to owner_staffs table
	err = s.db.Transaction(func(tx *gorm.DB) error {
		staff := &model.OwnerStaff{
			OwnerUserID: ownerID,
			StaffUserID: staffUserID,
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

	return apiResponse.Data, nil
}
