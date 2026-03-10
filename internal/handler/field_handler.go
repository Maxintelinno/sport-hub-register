package handler

import (
	"net/http"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/service"

	"github.com/labstack/echo/v4"
)

type FieldHandler struct {
	service *service.FieldService
}

func NewFieldHandler(service *service.FieldService) *FieldHandler {
	return &FieldHandler{service: service}
}

func (h *FieldHandler) CreateField(c echo.Context) error {
	req := new(model.CreateFieldRequest)
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

	field, err := h.service.CreateField(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "owner not found" {
			status = http.StatusBadRequest
		}
		return c.JSON(status, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, StandardResponse{
		Status:  "success",
		Message: "Stadium created successfully",
		Data:    field,
	})
}

func (h *FieldHandler) UpdateField(c echo.Context) error {
	id := c.Param("id")
	req := new(model.UpdateFieldRequest)
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

	field, err := h.service.UpdateField(id, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "field not found" {
			status = http.StatusNotFound
		} else if err.Error() == "unauthorized: you do not own this field" {
			status = http.StatusForbidden
		}
		return c.JSON(status, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Stadium updated successfully",
		Data:    field,
	})
}

func (h *FieldHandler) GetOwnerFields(c echo.Context) error {
	ownerID := c.QueryParam("owner_id")
	if ownerID == "" {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "owner_id is required",
		})
	}

	fields, err := h.service.GetFieldsByOwnerID(ownerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Fields retrieved successfully",
		Data:    fields,
	})
}

func (h *FieldHandler) UpdateFieldStatus(c echo.Context) error {
	req := new(model.UpdateFieldStatusRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if err := h.service.UpdateFieldStatus(req); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "field not found" || err.Error() == "unauthorized: you do not own this field" {
			status = http.StatusForbidden
		}
		return c.JSON(status, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Field status updated successfully",
	})
}
