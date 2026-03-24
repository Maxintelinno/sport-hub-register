package model

import (
	"time"

	"github.com/google/uuid"
)

type Field struct {
	ID           uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OwnerID      uuid.UUID    `json:"owner_id" gorm:"type:uuid;not null"`
	Name         string       `json:"name" gorm:"size:150;not null"`
	SportType    string       `json:"sport_type" gorm:"size:50;not null"` 
	OpenTime     string       `json:"open_time" gorm:"type:time;not null"`
	CloseTime    string       `json:"close_time" gorm:"type:time;not null"`
	Province     string       `json:"province" gorm:"size:100;not null"`
	District     string       `json:"district" gorm:"size:100;not null"`
	AddressLine  string       `json:"address_line" gorm:"type:text;not null"`
	Description  string       `json:"description" gorm:"type:text"`
	Status       string       `json:"status" gorm:"size:20;not null;default:'pending_review'"`
	ThumbnailUrl string       `json:"thumbnail_url" gorm:"size:255;null"`
	Latitude     float64      `json:"latitude" gorm:"type:numeric(10,7);null"`
	Longitude    float64      `json:"longitude" gorm:"type:numeric(10,7);null"`
	CreatedAt    time.Time    `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"not null;default:now()"`
	Images       []FieldImage `json:"images" gorm:"foreignKey:FieldID"`
}

type FieldImage struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FieldID   uuid.UUID `json:"field_id" gorm:"type:uuid;not null"`
	ObjectKey string    `json:"object_key" gorm:"type:text;not null"`
	ImageUrl  string    `json:"image_url" `
	SortOrder int       `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;default:now()"`
}

func (Field) TableName() string {
	return "fields"
}

func (FieldImage) TableName() string {
	return "field_images"
}

type CreateFieldRequest struct {
	OwnerID      uuid.UUID           `json:"owner_id" validate:"required"`
	Name         string              `json:"name" validate:"required"`
	SportType    string              `json:"sport_type" validate:"required"`
	OpenTime     string              `json:"open_time" validate:"required"`
	CloseTime    string              `json:"close_time" validate:"required"`
	Province     string              `json:"province" validate:"required"`
	District     string              `json:"district" validate:"required"`
	AddressLine  string              `json:"address_line" validate:"required"`
	Description  string              `json:"description"`
	Latitude     float64             `json:"latitude"`
	Longitude    float64             `json:"longitude"`
	Images       []FieldImageRequest `json:"images"`
}

type UpdateFieldRequest struct {
	OwnerID      uuid.UUID           `json:"owner_id" validate:"required"`
	Name         string              `json:"name" validate:"required"`
	SportType    string              `json:"sport_type" validate:"required"`
	OpenTime     string              `json:"open_time" validate:"required"`
	CloseTime    string              `json:"close_time" validate:"required"`
	Province     string              `json:"province" validate:"required"`
	District     string              `json:"district" validate:"required"`
	AddressLine  string              `json:"address_line" validate:"required"`
	Description  string              `json:"description"`
	Latitude     float64             `json:"latitude"`
	Longitude    float64             `json:"longitude"`
	Images       []FieldImageRequest `json:"images"`
}

type FieldImageRequest struct {
	ObjectKey string `json:"object_key" validate:"required"`
	SortOrder int    `json:"sort_order"`
}

type UpdateFieldStatusRequest struct {
	OwnerID uuid.UUID `json:"owner_id" validate:"required"`
	FieldID uuid.UUID `json:"field_id" validate:"required"`
	Status  string    `json:"status" validate:"required,oneof=active inactive"`
}
