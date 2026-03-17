package model

import (
	"time"

	"github.com/google/uuid"
)

type FieldCourt struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FieldID      uuid.UUID `json:"field_id" gorm:"type:uuid;not null"`
	Name         string    `json:"name" gorm:"size:100;not null"`
	PricePerHour float64   `json:"price_per_hour" gorm:"type:numeric(10,2);not null"`
	Capacity     int       `json:"capacity" gorm:"type:int"`
	CourtType    string    `json:"court_type" gorm:"size:50"`
	Status       string    `json:"status" gorm:"size:20;not null;default:'active'"` // active, inactive
	CreatedAt    time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"not null;default:now()"`
}

type Booking struct {
	ID            uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BookingNo     string        `json:"booking_no" gorm:"size:30;not null;unique"`
	UserID        uuid.UUID     `json:"user_id" gorm:"type:uuid;not null"`
	FieldID       uuid.UUID     `json:"field_id" gorm:"type:uuid;not null"`
	BookingDate   time.Time     `json:"booking_date" gorm:"type:date;not null"`
	TotalAmount   float64       `json:"total_amount" gorm:"type:numeric(10,2);not null;default:0"`
	Status        string        `json:"status" gorm:"size:20;not null;default:'pending'"`         // pending, confirmed, cancelled, completed, expired
	PaymentStatus string        `json:"payment_status" gorm:"size:20;not null;default:'unpaid'"` // unpaid, paid, refunded
	Note          string        `json:"note" gorm:"type:text"`
	CreatedAt     time.Time     `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt     time.Time     `json:"updated_at" gorm:"not null;default:now()"`
	Items         []BookingItem `json:"items" gorm:"foreignKey:BookingID"`
}

type BookingItem struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BookingID    uuid.UUID `json:"booking_id" gorm:"type:uuid;not null"`
	FieldID      uuid.UUID `json:"field_id" gorm:"type:uuid;not null"`
	CourtID      uuid.UUID `json:"court_id" gorm:"type:uuid;not null"`
	BookingDate  time.Time `json:"booking_date" gorm:"type:date;not null"`
	StartTime    string    `json:"start_time" gorm:"type:time;not null"`
	EndTime      string    `json:"end_time" gorm:"type:time;not null"`
	StartAt      time.Time `json:"start_at" gorm:"not null"`
	EndAt        time.Time `json:"end_at" gorm:"not null"`
	PricePerHour float64   `json:"price_per_hour" gorm:"type:numeric(10,2);not null"`
	TotalAmount  float64   `json:"total_amount" gorm:"type:numeric(10,2);not null"`
	Status       string    `json:"status" gorm:"size:20;not null;default:'confirmed'"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null;default:now()"`
}

// Request Models
type CreateBookingRequest struct {
	UserID      uuid.UUID            `json:"user_id" validate:"required"`
	FieldID     uuid.UUID            `json:"field_id" validate:"required"`
	BookingDate string               `json:"booking_date" validate:"required"` // YYYY-MM-DD
	Note        string               `json:"note"`
	Items       []CreateBookingItem  `json:"items" validate:"required,min=1"`
}

type CreateBookingItem struct {
	CourtID   uuid.UUID `json:"court_id" validate:"required"`
	StartTime string    `json:"start_time" validate:"required"` // HH:mm
	EndTime   string    `json:"end_time" validate:"required"`   // HH:mm
}

type CreateCourtItem struct {
	Name         string  `json:"name" validate:"required"`
	PricePerHour float64 `json:"price_per_hour" validate:"required"`
	Capacity     int     `json:"capacity"`
	CourtType    string  `json:"court_type"`
}

type CreateCourtsBulkRequest struct {
	FieldID uuid.UUID         `json:"field_id" validate:"required"`
	Courts  []CreateCourtItem `json:"courts" validate:"required,dive"`
}

type UpdateCourtRequest struct {
	FieldID      uuid.UUID `json:"field_id" validate:"required"`
	Name         string    `json:"name" validate:"required"`
	PricePerHour float64   `json:"price_per_hour" validate:"required"`
	Capacity     int       `json:"capacity"`
	CourtType    string    `json:"court_type"`
	Status       string    `json:"status" validate:"required,oneof=active inactive"`
}

// Response Models for Availability
type TimeSlot struct {
	StartTime string `json:"start_time"` // HH:mm
	EndTime   string `json:"end_time"`   // HH:mm
	Status    string `json:"status"`     // booked, available (computed by frontend or backend)
}

type CourtAvailability struct {
	CourtID      uuid.UUID  `json:"court_id"`
	CourtName    string     `json:"court_name"`
	PricePerHour float64    `json:"price_per_hour"`
	Capacity     int        `json:"capacity"`
	CourtType    string     `json:"court_type"`
	BookedSlots  []TimeSlot `json:"booked_slots"`
}

type CourtAvailabilityResponse struct {
	FieldID   uuid.UUID           `json:"field_id"`
	Date      string              `json:"date"`
	OpenTime  string              `json:"open_time"`
	CloseTime string              `json:"close_time"`
	Courts    []CourtAvailability `json:"courts"`
}
