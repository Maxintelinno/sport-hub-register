package model

import (
	"time"

	"github.com/google/uuid"
)

type BookingCredit struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;unique"`
	Balance      float64   `json:"balance" gorm:"type:numeric(12,2);not null;default:0"`
	TotalEarned  float64   `json:"total_earned" gorm:"type:numeric(12,2);not null;default:0"`
	TotalUsed    float64   `json:"total_used" gorm:"type:numeric(12,2);not null;default:0"`
	TotalExpired float64   `json:"total_expired" gorm:"type:numeric(12,2);not null;default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"not null;default:now()"`
}

func (BookingCredit) TableName() string {
	return "booking_credits"
}

type CheckoutCreditPreviewRequest struct {
	TotalAmount float64 `json:"total_amount" validate:"required"`
}

type CheckoutCreditPreviewResponse struct {
	TotalAmount      float64 `json:"total_amount"`
	CreditBalance    float64 `json:"credit_balance"`
	CreditToUse      float64 `json:"credit_to_use"`
	RemainingToPay   float64 `json:"remaining_to_pay"`
}
