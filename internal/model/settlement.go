package model

import (
	"time"

	"github.com/google/uuid"
)

type OwnerSettlement struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BookingID      uuid.UUID `json:"booking_id" gorm:"type:uuid;not null;unique"`
	OwnerID        uuid.UUID `json:"owner_id" gorm:"type:uuid;not null"`
	GrossAmount    float64   `json:"gross_amount" gorm:"type:numeric(10,2);not null"`
	PlatformFee    float64   `json:"platform_fee" gorm:"type:numeric(10,2);not null;default:0"`
	DiscountAmount float64   `json:"discount_amount" gorm:"type:numeric(10,2);not null;default:0"`
	NetAmount      float64   `json:"net_amount" gorm:"type:numeric(10,2);not null"`
	Status         string    `json:"status" gorm:"type:varchar(30);not null;default:'pending'"`
	AvailableAt    *time.Time `json:"available_at"`
	PaidAt         *time.Time `json:"paid_at"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"not null;default:now()"`
	RefundAmount   float64   `json:"refund_amount" gorm:"type:numeric(12,2);not null;default:0"`
	NetRevenue     float64   `json:"net_revenue" gorm:"type:numeric(12,2);not null;default:0"`
	OwnerNetAmount float64   `json:"owner_net_amount" gorm:"type:numeric(12,2);not null;default:0"`
	ReversedAt     *time.Time `json:"reversed_at"`
}

func (OwnerSettlement) TableName() string {
	return "owner_settlements"
}
