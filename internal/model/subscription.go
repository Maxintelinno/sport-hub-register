package model

import (
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code               string    `json:"code" gorm:"size:50;not null;unique"`
	Name               string    `json:"name" gorm:"size:100;not null"`
	Description        string    `json:"description" gorm:"type:text"`
	Price              float64   `json:"price" gorm:"type:numeric(10,2);not null;default:0"`
	BillingCycle       string    `json:"billing_cycle" gorm:"size:20;not null;default:'monthly'"`
	TrialDays          int       `json:"trial_days" gorm:"type:int;not null;default:0"`
	MaxFields          *int      `json:"max_fields"`
	MaxCourts          *int      `json:"max_courts"`
	HasReports         bool      `json:"has_reports" gorm:"not null;default:false"`
	HasPromotion       bool      `json:"has_promotion" gorm:"not null;default:false"`
	HasHomepageFeature bool      `json:"has_homepage_feature" gorm:"not null;default:false"`
	HasPrioritySupport bool      `json:"has_priority_support" gorm:"not null;default:false"`
	IsActive           bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt          time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"not null;default:now()"`
}

func (Plan) TableName() string {
	return "plans"
}

type Subscription struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	PlanID        uuid.UUID `json:"plan_id" gorm:"type:uuid;not null"`
	Status        string    `json:"status" gorm:"size:30;not null"` // trial, active, expired, cancelled
	StartAt       time.Time `json:"start_at" gorm:"not null"`
	EndAt         time.Time `json:"end_at" gorm:"not null"`
	TrialStartAt  *time.Time `json:"trial_start_at"`
	TrialEndAt    *time.Time `json:"trial_end_at"`
	ActivatedAt   *time.Time `json:"activated_at"`
	ExpiredAt     *time.Time `json:"expired_at"`
	CancelledAt   *time.Time `json:"cancelled_at"`
	CancelReason  string    `json:"cancel_reason" gorm:"type:text"`
	AutoRenew     bool      `json:"auto_renew" gorm:"not null;default:false"`
	CreatedAt     time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"not null;default:now()"`

	User User `json:"-" gorm:"foreignKey:UserID"`
	Plan Plan `json:"plan" gorm:"foreignKey:PlanID"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}
