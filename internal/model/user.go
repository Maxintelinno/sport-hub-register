package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Phone        string    `json:"phone" gorm:"column:phone;unique;not null"`
	Username     string    `json:"username" gorm:"column:username;unique;not null"`
	Fullname     string    `json:"fullname" gorm:"column:fullname;unique;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;not null"`
	Role         string    `json:"role" gorm:"column:role;not null;default:'user'"`
	Province     string    `json:"province" gorm:"column:province;not null"`
	District     string    `json:"district" gorm:"column:district;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;default:now()"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;default:now()"`
}

func (User) TableName() string {
	return "users"
}

type RegisterRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Username string `json:"username" validate:"required"`
	Fullname string `json:"fullname" validate:"required"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type RegisterStaffRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Username string `json:"username" validate:"required"`
	Fullname string `json:"fullname" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type UserSubscriptionResponse struct {
	PlanName     string `json:"plan_name"`
	BillingCycle string `json:"billing_cycle"`
	Status       string `json:"status"`
}

type UserResponse struct {
	User         *User                     `json:"user"`
	Subscription *UserSubscriptionResponse `json:"subscription,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
