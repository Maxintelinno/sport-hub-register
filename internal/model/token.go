package model

import (
	"time"

	"github.com/google/uuid"
)

type RegistrationToken struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone     string    `json:"phone" gorm:"column:phone;not null"`
	TokenHash string    `json:"-" gorm:"column:token_hash;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"column:expires_at;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;default:now()"`
}

func (RegistrationToken) TableName() string {
	return "registration_tokens"
}
