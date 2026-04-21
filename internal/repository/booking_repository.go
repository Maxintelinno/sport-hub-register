package repository

import (
	"fmt"
	"sport-hub-register/internal/model"
	"time"

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
	err := r.getDB(tx).
		Preload("Items.Court").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&bookings).Error
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
	err := r.getDB(tx).
		Preload("Items.Court").
		Where("id = ?", id).
		First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *BookingRepository) FindBookedItemsByFieldIDAndDate(tx *gorm.DB, fieldID string, date string) ([]model.BookingItem, error) {
	var items []model.BookingItem
	err := r.getDB(tx).
		Where("field_id = ? AND booking_date = ? AND status IN (?, ?)", fieldID, date, "confirmed", "pending").
		Order("start_at ASC").
		Find(&items).Error
	return items, err
}

func (r *BookingRepository) FindOwnerBookings(tx *gorm.DB, fieldID string, date string) ([]model.BookingItem, error) {
	var items []model.BookingItem
	err := r.getDB(tx).
		Preload("Booking.User").
		Preload("Court").
		Where("field_id = ? AND booking_date = ?", fieldID, date).
		Order("start_at ASC").
		Find(&items).Error
	return items, err
}

func (r *BookingRepository) CancelBookingWithRefund(tx *gorm.DB, bookingID string, cancelReason string, refundAmount float64, paymentStatus string, refundPolicy string) error {
	return r.getDB(tx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 0. Get Booking to find UserID and BookingNo
		var booking model.Booking
		if err := tx.Select("id, user_id, booking_no").Where("id = ?", bookingID).First(&booking).Error; err != nil {
			return err
		}

		// 1. Update Booking
		bookingUpdates := map[string]interface{}{
			"status":         "cancelled",
			"cancelled_at":   &now,
			"cancel_reason":  cancelReason,
			"refund_amount":  refundAmount,
			"payment_status": paymentStatus,
		}
		if err := tx.Model(&model.Booking{}).Where("id = ?", bookingID).Updates(bookingUpdates).Error; err != nil {
			return err
		}

		// 2. Update Owner Settlement
		settlementUpdates := map[string]interface{}{
			"status":        "reversed",
			"reversed_at":   &now,
			"refund_amount": refundAmount,
		}
		// Note: We update if exists, ignoring ErrRecordNotFound if not yet settled
		if err := tx.Model(&model.OwnerSettlement{}).Where("booking_id = ?", bookingID).Updates(settlementUpdates).Error; err != nil {
			return err
		}

		// 3. Update Payment
		paymentUpdates := map[string]interface{}{
			"status":          "refunded",
			"refunded_amount": refundAmount,
			"refund_status":   refundPolicy,
		}
		if err := tx.Model(&model.Payment{}).Where("booking_id = ?", bookingID).Updates(paymentUpdates).Error; err != nil {
			return err
		}

		// 4. Update User Credits if refundAmount > 0
		if refundAmount > 0 {
			var credit model.BookingCredit
			// Use FirstOrCreate to ensure the record exists for the user
			if err := tx.Where(model.BookingCredit{UserID: booking.UserID}).FirstOrCreate(&credit).Error; err != nil {
				return err
			}

			balanceBefore := credit.Balance
			credit.Balance += refundAmount
			credit.TotalEarned += refundAmount
			credit.UpdatedAt = now

			if err := tx.Save(&credit).Error; err != nil {
				return err
			}

			// Record transaction
			transaction := model.BookingCreditTransaction{
				UserID:          booking.UserID,
				BookingCreditID: credit.ID,
				TransactionType: "refund",
				SourceType:      "booking",
				ReferenceID:     &booking.ID,
				Amount:          refundAmount,
				BalanceBefore:   balanceBefore,
				BalanceAfter:    credit.Balance,
				Note:            fmt.Sprintf("Refund from cancelled booking %s", booking.BookingNo),
				CreatedAt:       now,
			}
			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
