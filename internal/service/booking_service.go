package service

import (
	"errors"
	"fmt"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingService struct {
	db          *gorm.DB
	bookingRepo *repository.BookingRepository
	courtRepo   *repository.CourtRepository
	fieldRepo   *repository.FieldRepository
}

func NewBookingService(db *gorm.DB, bookingRepo *repository.BookingRepository, courtRepo *repository.CourtRepository, fieldRepo *repository.FieldRepository) *BookingService {
	return &BookingService{
		db:          db,
		bookingRepo: bookingRepo,
		courtRepo:   courtRepo,
		fieldRepo:   fieldRepo,
	}
}

func (s *BookingService) CreateCourt(req *model.CreateCourtRequest) (*model.FieldCourt, error) {
	court := &model.FieldCourt{
		ID:           uuid.New(),
		FieldID:      req.FieldID,
		Name:         req.Name,
		PricePerHour: req.PricePerHour,
		Capacity:     req.Capacity,
		CourtType:    req.CourtType,
		Status:       "active",
	}

	if err := s.courtRepo.CreateCourt(nil, court); err != nil {
		return nil, err
	}
	return court, nil
}

func (s *BookingService) GetCourtsByFieldID(fieldID string) ([]model.FieldCourt, error) {
	return s.courtRepo.FindCourtsByFieldID(nil, fieldID)
}

func (s *BookingService) CreateBooking(userID uuid.UUID, req *model.CreateBookingRequest) (*model.Booking, error) {
	bookingDate, err := time.Parse("2006-01-02", req.BookingDate)
	if err != nil {
		return nil, errors.New("invalid booking_date format, use YYYY-MM-DD")
	}

	var totalAmount float64
	bookingItems := make([]model.BookingItem, 0)

	// 1. Validate and prepare items
	for _, item := range req.Items {
		court, err := s.courtRepo.FindCourtByID(nil, item.CourtID.String())
		if err != nil {
			return nil, fmt.Errorf("court not found: %s", item.CourtID)
		}

		// Parse times
		startAt, err := time.ParseInLocation("2006-01-02 15:04", req.BookingDate+" "+item.StartTime, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format: %s", item.StartTime)
		}
		endAt, err := time.ParseInLocation("2006-01-02 15:04", req.BookingDate+" "+item.EndTime, time.Local)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format: %s", item.EndTime)
		}

		if !endAt.After(startAt) {
			return nil, fmt.Errorf("end_time must be after start_time for court %s", court.Name)
		}

		// Check overlap
		overlap, err := s.bookingRepo.CheckOverlap(nil, court.ID.String(), startAt, endAt)
		if err != nil {
			return nil, err
		}
		if overlap {
			return nil, fmt.Errorf("court %s is already booked for the selected time", court.Name)
		}

		duration := endAt.Sub(startAt).Hours()
		itemAmount := duration * court.PricePerHour
		totalAmount += itemAmount

		bookingItems = append(bookingItems, model.BookingItem{
			ID:           uuid.New(),
			FieldID:      req.FieldID,
			CourtID:      court.ID,
			BookingDate:  bookingDate,
			StartTime:    item.StartTime,
			EndTime:      item.EndTime,
			StartAt:      startAt,
			EndAt:        endAt,
			PricePerHour: court.PricePerHour,
			TotalAmount:  itemAmount,
			Status:       "confirmed",
		})
	}

	// 2. Create Booking in Transaction
	booking := &model.Booking{
		ID:            uuid.New(),
		BookingNo:     s.generateBookingNo(),
		UserID:        userID,
		FieldID:       req.FieldID,
		BookingDate:   bookingDate,
		TotalAmount:   totalAmount,
		Status:        "pending",
		PaymentStatus: "unpaid",
		Note:          req.Note,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.bookingRepo.CreateBooking(tx, booking); err != nil {
			return err
		}

		for i := range bookingItems {
			bookingItems[i].BookingID = booking.ID
		}

		if err := s.bookingRepo.CreateBookingItems(tx, bookingItems); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	booking.Items = bookingItems
	return booking, nil
}

func (s *BookingService) GetUserBookings(userID string) ([]model.Booking, error) {
	return s.bookingRepo.FindBookingsByUserID(nil, userID)
}

func (s *BookingService) generateBookingNo() string {
	return fmt.Sprintf("BK%d%s", time.Now().Unix(), uuid.New().String()[:4])
}

func (s *BookingService) GetFieldAvailability(fieldID string, date string) (*model.CourtAvailabilityResponse, error) {
	// 1. Fetch Field to get open/close times
	field, err := s.fieldRepo.FindFieldByID(nil, fieldID)
	if err != nil {
		return nil, fmt.Errorf("field not found: %v", err)
	}

	// 2. Fetch all courts for the field
	courts, err := s.courtRepo.FindCourtsByFieldID(nil, fieldID)
	if err != nil {
		return nil, err
	}

	// 3. Fetch all booked items for the date
	bookedItems, err := s.bookingRepo.FindBookedItemsByFieldIDAndDate(nil, fieldID, date)
	if err != nil {
		return nil, err
	}

	// 4. Map booked items by courtID
	bookedSlotsByCourt := make(map[uuid.UUID][]model.TimeSlot)
	for _, item := range bookedItems {
		slot := model.TimeSlot{
			StartTime: item.StartTime,
			EndTime:   item.EndTime,
			Status:    "booked", // items returned are confirmed or pending
		}
		bookedSlotsByCourt[item.CourtID] = append(bookedSlotsByCourt[item.CourtID], slot)
	}

	// 5. Build response
	fieldUUID, _ := uuid.Parse(fieldID)
	response := &model.CourtAvailabilityResponse{
		FieldID:   fieldUUID,
		Date:      date,
		OpenTime:  field.OpenTime,
		CloseTime: field.CloseTime,
		Courts:    make([]model.CourtAvailability, 0),
	}

	for _, court := range courts {
		slots := bookedSlotsByCourt[court.ID]
		if slots == nil {
			slots = []model.TimeSlot{} // Return empty array instead of null
		}
		response.Courts = append(response.Courts, model.CourtAvailability{
			CourtID:      court.ID,
			CourtName:    court.Name,
			PricePerHour: court.PricePerHour,
			BookedSlots:  slots,
		})
	}

	return response, nil
}
