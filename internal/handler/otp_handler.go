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

func (h *OTPHandler) RequestOTP(c echo.Context) error {
	req := new(model.OTPRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	code, err := h.service.RequestOTP(req.Phone)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "OTP sent successfully",
		"code":    code,
	})
}

func (h *OTPHandler) VerifyOTP(c echo.Context) error {
	req := new(model.OTPVerifyRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}

	token, err := h.service.VerifyOTP(req.Phone, req.Code)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "OTP verified successfully",
		"token":   token,
	})
}
