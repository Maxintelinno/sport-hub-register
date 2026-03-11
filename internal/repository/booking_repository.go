package repository

import (
	"sport-hub-register/internal/model"

	"gorm.io/gorm"
)

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *BookingRepository) CreateBooking(tx *gorm.DB, booking *model.Booking) error {
	return r.getDB(tx).Create(booking).Error
}

func (r *BookingRepository) CreateBookingItems(tx *gorm.DB, items []model.BookingItem) error {
	return r.getDB(tx).Create(&items).Error
}

func (r *BookingRepository) FindBookingsByUserID(tx *gorm.DB, userID string) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.getDB(tx).Preload("Items").Where("user_id = ?", userID).Order("created_at desc").Find(&bookings).Error
	return bookings, err
}

func (r *BookingRepository) CheckOverlap(tx *gorm.DB, courtID string, startAt, endAt interface{}) (bool, error) {
	var count int64
	// status IN ('confirmed', 'pending')
	err := r.getDB(tx).Model(&model.BookingItem{}).
		Where("court_id = ? AND status IN (?, ?) AND NOT (end_at <= ? OR start_at >= ?)", 
			courtID, "confirmed", "pending", startAt, endAt).
		Count(&count).Error
	return count > 0, err
}

func (r *BookingRepository) GetBookingByID(tx *gorm.DB, id string) (*model.Booking, error) {
	var booking model.Booking
	err := r.getDB(tx).Preload("Items").Where("id = ?", id).First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}
