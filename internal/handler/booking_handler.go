package handler

import (
	"net/http"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/service"

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
	req := new(model.CreateCourtRequest)
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

	court, err := h.service.CreateCourt(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "Court created successfully",
		Data:    court,
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

func (h *BookingHandler) CreateBooking(c echo.Context) error {
	userIDStr := c.Get("user_id").(string) // Assumes Auth middleware sets this
	userID, _ := uuid.Parse(userIDStr)

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

	booking, err := h.service.CreateBooking(userID, req)
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
	userID := c.Get("user_id").(string)

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
