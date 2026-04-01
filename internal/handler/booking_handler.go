package handler

import (
	"log"
	"net/http"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/service"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BookingHandler struct {
	service *service.BookingService
}

func NewBookingHandler(service *service.BookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

func (h *BookingHandler) CreateCourt(c echo.Context) error {
	var req model.CreateCourtsBulkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "Invalid input",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	courts, err := h.service.CreateCourts(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "Courts created successfully",
		Data:    courts,
	})
}

func (h *BookingHandler) GetCourts(c echo.Context) error {
	fieldID := c.QueryParam("field_id")
	if fieldID == "" {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "field_id is required",
		})
	}

	courts, err := h.service.GetCourtsByFieldID(fieldID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Courts retrieved successfully",
		Data:    courts,
	})
}

func (h *BookingHandler) UpdateCourt(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "id is required",
		})
	}

	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, StandardResponse{
			Status:  "error",
			Message: "unauthorized",
		})
	}

	req := new(model.UpdateCourtRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "Invalid input",
		})
	}

	log.Println("UpdateCourt", req)

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	court, err := h.service.UpdateCourt(id, userID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "court not found" {
			status = http.StatusNotFound
		}
		return c.JSON(status, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Court updated successfully",
		Data:    court,
	})
}

func (h *BookingHandler) CreateBooking(c echo.Context) error {
	req := new(model.CreateBookingRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "Invalid request format",
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	booking, err := h.service.CreateBooking(req.UserID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "Booking created successfully",
		Data:    booking,
	})
}

func (h *BookingHandler) GetMyBookings(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "user_id is required",
		})
	}

	bookings, err := h.service.GetUserBookings(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Bookings retrieved successfully",
		Data:    bookings,
	})
}

func (h *BookingHandler) GetAvailability(c echo.Context) error {
	fieldID := c.QueryParam("field_id")
	if fieldID == "" {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "field_id is required",
		})
	}

	date := c.QueryParam("date")
	if date == "" {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "date is required (format: YYYY-MM-DD)",
		})
	}

	availability, err := h.service.GetFieldAvailability(fieldID, date)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Availability retrieved successfully",
		Data:    availability,
	})
}

func (h *BookingHandler) GetOwnerBookings(c echo.Context) error {
	ownerID, ok := c.Get("user_id").(string)
	if !ok || ownerID == "" {
		return c.JSON(http.StatusUnauthorized, StandardResponse{
			Status:  "error",
			Message: "unauthorized",
		})
	}

	fieldID := c.QueryParam("field_id")
	if fieldID == "" {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "field_id is required",
		})
	}

	date := c.QueryParam("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	bookings, err := h.service.GetOwnerBookings(ownerID, fieldID, date)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "unauthorized: you do not own this field" {
			status = http.StatusForbidden
		}
		return c.JSON(status, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Owner bookings retrieved successfully",
		Data:    bookings,
	})
}

func (h *BookingHandler) CreateOfflineBooking(c echo.Context) error {
	ownerIDStr, ok := c.Get("user_id").(string)
	if !ok || ownerIDStr == "" {
		return c.JSON(http.StatusUnauthorized, StandardResponse{
			Status:  "error",
			Message: "unauthorized",
		})
	}
	ownerID, _ := uuid.Parse(ownerIDStr)

	req := new(model.CreateOfflineBookingRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "Invalid request format",
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	booking, err := h.service.CreateOfflineBooking(ownerID, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "Offline booking created successfully",
		Data:    map[string]string{"booking_no": booking.BookingNo},
	})
}
