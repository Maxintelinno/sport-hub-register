package model

import (
	"time"

	"github.com/google/uuid"
)

type OTPRequestRecord struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone       string    `json:"phone" gorm:"column:phone;not null"`
	OTPHash     string    `json:"-" gorm:"column:otp_hash;not null"`
	ProviderRef string    `json:"-" gorm:"column:provider_ref"`
	Token       string    `json:"-" gorm:"column:token"`
	Attempts    int       `json:"attempts" gorm:"column:attempts;default:0"`
	ExpiresAt   time.Time `json:"expires_at" gorm:"column:expires_at;not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;default:now()"`
}

func (OTPRequestRecord) TableName() string {
	return "otp_requests"
}

type OTPRequest struct {
	Phone string `json:"phone" validate:"required,numeric,min=10,max=11"`
}

type OTPVerifyRequest struct {
	Phone string `json:"phone" validate:"required,numeric,min=10,max=11"`
	Otp   string `json:"otp" validate:"required,numeric,len=6"`
}

type OtpRes struct {
	Status string `json:"status"`
	Token  string `json:"token"`
	Refno  string `json:"refno"`
	Error  string `json:"error"`
}
