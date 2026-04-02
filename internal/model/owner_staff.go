package model

import (
	"time"

	"github.com/google/uuid"
)

type OwnerStaff struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OwnerUserID uuid.UUID `json:"owner_user_id" gorm:"type:uuid;not null;uniqueIndex:idx_owner_staff_unique"`
	StaffUserID uuid.UUID `json:"staff_user_id" gorm:"type:uuid;not null;uniqueIndex:idx_owner_staff_unique"`
	RoleCode    string    `json:"role_code" gorm:"column:role_code;type:varchar(30);default:'staff';not null"`
	Status      string    `json:"status" gorm:"column:status;type:varchar(20);default:'active';not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;default:now()"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at;default:now()"`
}

func (OwnerStaff) TableName() string {
	return "owner_staffs"
}
