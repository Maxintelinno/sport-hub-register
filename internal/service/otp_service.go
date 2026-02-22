package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type OTPService struct {
	db        *gorm.DB
	repo      *repository.OTPRepository
	tokenRepo *repository.TokenRepository
}

func NewOTPService(db *gorm.DB, repo *repository.OTPRepository, tokenRepo *repository.TokenRepository) *OTPService {
	return &OTPService{db: db, repo: repo, tokenRepo: tokenRepo}
}

func (s *OTPService) RequestOTP(phone string) (string, error) {
	// 1. Cooldown Check (1 minute)
	lastOTP, err := s.repo.FindLatestByPhone(nil, phone)
	if err == nil && lastOTP != nil {
		if time.Now().Sub(lastOTP.CreatedAt) < 1*time.Minute {
			return "", errors.New("OTP requested too frequently. Please wait 1 minute.")
		}
	}

	// 2. Clean up old OTPs for this phone
	_ = s.repo.DeleteByPhone(nil, phone)

	// 3. Generate 6-digit OTP
	code := s.generateNumericOTP(6)

	// 4. Hash OTP
	hashedCode, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	otpRecord := &model.OTPRequestRecord{
		Phone:     phone,
		OTPHash:   string(hashedCode),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	err = s.repo.CreateOTP(nil, otpRecord)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *OTPService) VerifyOTP(phone, code string) (string, error) {
	var registrationToken string

	err := s.db.Transaction(func(tx *gorm.DB) error {
		otp, err := s.repo.FindLatestByPhone(tx, phone)
		if err != nil {
			return errors.New("OTP not found")
		}

		// 1. Expiry Check
		if time.Now().After(otp.ExpiresAt) {
			return errors.New("OTP expired")
		}

		// 2. Brute Force Protection (Max 5 attempts)
		if otp.Attempts >= 5 {
			return errors.New("too many failed attempts (limit 5)")
		}

		// 3. Match Code
		err = bcrypt.CompareHashAndPassword([]byte(otp.OTPHash), []byte(code))
		if err != nil {
			_ = s.repo.IncrementAttempts(tx, otp.ID.String())
			return errors.New("invalid OTP code")
		}

		// 4. Success -> Delete OTP and Create Registration Token
		if err := s.repo.DeleteOTP(tx, otp.ID.String()); err != nil {
			return err
		}

		// Generate Token
		token := uuid.New().String()
		tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		regToken := &model.RegistrationToken{
			Phone:     phone,
			TokenHash: string(tokenHash),
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		if err := s.tokenRepo.CreateToken(tx, regToken); err != nil {
			return err
		}

		registrationToken = token
		return nil
	})

	if err != nil {
		return "", err
	}

	return registrationToken, nil
}

func (s *OTPService) generateNumericOTP(max int) string {
	b := make([]byte, max)
	_, _ = io.ReadFull(rand.Reader, b)
	var otp string
	for _, v := range b {
		otp += fmt.Sprintf("%d", v%10)
	}
	return otp
}
