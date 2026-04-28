package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"
	"strings"
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

func (s *OTPService) RequestOTP(phone string) (string, string, error) {
	// 1. Cooldown Check (1 minute) with 5-second grace period for double-clicks
	lastOTP, err := s.repo.FindLatestByPhone(nil, phone)
	if err == nil && lastOTP != nil {
		elapsed := time.Since(lastOTP.CreatedAt)
		if elapsed < 5*time.Second {
			log.Printf("[OTPService] Phone %s requested OTP again within grace period (%v).", phone, elapsed)
		} else if elapsed < 1*time.Minute {
			log.Printf("[OTPService] Phone %s requested OTP too frequently: %v elapsed", phone, elapsed)
			return "", "", errors.New("OTP requested too frequently. Please wait 1 minute.")
		}
	}

	key := os.Getenv("OTP_APP_KEY")
	secret := os.Getenv("OTP_APP_SECRET")

	if key == "" || secret == "" {
		return "", "", errors.New("OTP credentials not configured")
	}

	apiUrl := "https://otp.thaibulksms.com/v2/otp/request"

	data := url.Values{}
	data.Set("key", key)
	data.Set("secret", secret)
	data.Set("msisdn", phone)

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	var otpRes struct {
		Status string `json:"status"`
		Token  string `json:"token"`
		Refno  string `json:"refno"`
		Error  string `json:"error"`
	}

	if err := json.Unmarshal(body, &otpRes); err != nil {
		return "", "", fmt.Errorf("failed to parse OTP response: %v", err)
	}

	if otpRes.Status != "success" {
		return "", "", fmt.Errorf("OTP request failed: %s", string(body))
	}

	// 2. Clean up old OTPs for this phone
	_ = s.repo.DeleteByPhone(nil, phone)

	log.Printf("[OTPService] Generated new external OTP for %s: refno %s", phone, otpRes.Refno)

	otpRecord := &model.OTPRequestRecord{
		Phone:     phone,
		OTPHash:   otpRes.Token, // Store external token here instead of bcrypt hash
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	err = s.repo.CreateOTP(nil, otpRecord)
	if err != nil {
		return "", "", err
	}

	return otpRes.Token, otpRes.Refno, nil
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

		// 3. Match Code with external service
		err = s.verifyOTPExternal(otp.OTPHash, code)
		if err != nil {
			log.Printf("[OTPService] Invalid OTP code for %s (Attempts: %d)", phone, otp.Attempts+1)
			_ = s.repo.IncrementAttempts(tx, otp.ID.String())
			return errors.New("invalid OTP code")
		}

		log.Printf("[OTPService] OTP verified successfully for %s", phone)

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

func (s *OTPService) verifyOTPExternal(token, pin string) error {
	key := os.Getenv("OTP_APP_KEY")
	secret := os.Getenv("OTP_APP_SECRET")

	apiUrl := "https://otp.thaibulksms.com/v2/otp/verify"

	data := url.Values{}
	data.Set("key", key)
	data.Set("secret", secret)
	data.Set("token", token)
	data.Set("pin", pin)

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var verifyRes struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &verifyRes); err != nil {
		return fmt.Errorf("failed to parse verify response: %v", err)
	}

	if verifyRes.Status != "success" {
		return errors.New("invalid OTP code")
	}

	return nil
}
