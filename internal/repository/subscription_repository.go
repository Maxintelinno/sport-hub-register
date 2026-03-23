package repository

import (
	"sport-hub-register/internal/model"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) CreateSubscription(tx *gorm.DB, sub *model.Subscription) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(sub).Error
}

func (r *SubscriptionRepository) FindLatestByUserID(tx *gorm.DB, userID string) (*model.Subscription, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var sub model.Subscription
	err := db.Preload("Plan").Where("user_id = ?", userID).Order("created_at desc").First(&sub).Error
	return &sub, err
}

type PlanRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) FindByCode(tx *gorm.DB, code string) (*model.Plan, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	var plan model.Plan
	err := db.Where("code = ?", code).First(&plan).Error
	return &plan, err
}

func (r *PlanRepository) CreatePlan(tx *gorm.DB, plan *model.Plan) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(plan).Error
}
