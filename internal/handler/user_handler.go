package handler

import (
	"net/http"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/service"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Register(c echo.Context) error {
	req := new(model.RegisterRequest)
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

	res, err := h.service.Register(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "User registered successfully",
		Data:    res,
	})
}

func (h *UserHandler) Login(c echo.Context) error {
	req := new(model.LoginRequest)
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

	res, err := h.service.Login(req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Login successful",
		Data:    res,
	})
}

func (h *UserHandler) RegisterStaff(c echo.Context) error {
	ownerIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, StandardResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "Invalid owner user ID",
		})
	}

	req := new(model.RegisterStaffRequest)
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

	res, err := h.service.RegisterStaff(ownerID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "Staff registered successfully",
		Data:    res,
	})
}
