package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID                    uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BookingID             uuid.UUID       `json:"booking_id" gorm:"type:uuid;not null"`
	PaymentNo             string          `json:"payment_no" gorm:"type:varchar(30);not null;unique"`
	Provider              string          `json:"provider" gorm:"type:varchar(50);not null"`
	Method                string          `json:"method" gorm:"type:varchar(30);not null"`
	Amount                float64         `json:"amount" gorm:"type:numeric(10,2);not null"`
	Currency              string          `json:"currency" gorm:"type:varchar(10);not null;default:'THB'"`
	Status                string          `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	ProviderPaymentID     string          `json:"provider_payment_id" gorm:"type:varchar(100)"`
	ProviderTransactionID string          `json:"provider_transaction_id" gorm:"type:varchar(150)"`
	ProviderReference     string          `json:"provider_reference" gorm:"type:varchar(150)"`
	QrPayload             string          `json:"qr_payload" gorm:"type:text"`
	QrImageUrl            string          `json:"qr_image_url" gorm:"type:text"`
	ExpiresAt             time.Time       `json:"expires_at" gorm:"not null"`
	PaidAt                *time.Time      `json:"paid_at"`
	FailedAt              *time.Time      `json:"failed_at"`
	FailureReason         string          `json:"failure_reason" gorm:"type:text"`
	Metadata              json.RawMessage `json:"metadata" gorm:"type:jsonb"`
	CreatedAt             time.Time       `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt             time.Time       `json:"updated_at" gorm:"not null;default:now()"`
	RefundedAmount        float64         `json:"refunded_amount" gorm:"type:numeric(10,2)"`
	RefundStatus          string          `json:"refund_status" gorm:"type:varchar(20)"`
}

func (Payment) TableName() string {
	return "payments"
}
