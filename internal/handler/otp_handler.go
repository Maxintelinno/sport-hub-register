package handler

import (
	"net/http"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/service"

	"github.com/labstack/echo/v4"
)

type OTPHandler struct {
	service *service.OTPService
}

func NewOTPHandler(service *service.OTPService) *OTPHandler {
	return &OTPHandler{service: service}
}

type StandardResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (h *OTPHandler) RequestOTP(c echo.Context) error {
	req := new(model.OTPRequest)
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

	code, err := h.service.RequestOTP(req.Phone)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "OTP sent successfully",
		Data:    map[string]string{"code": code},
	})
}

func (h *OTPHandler) VerifyOTP(c echo.Context) error {
	req := new(model.OTPVerifyRequest)
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

	token, err := h.service.VerifyOTP(req.Phone, req.Code)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "OTP verified successfully",
		Data:    map[string]string{"token": token},
	})
}
