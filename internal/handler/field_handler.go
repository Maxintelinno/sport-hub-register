package handler

import (
	"net/http"
	"sport-hub-register/internal/model"
	"sport-hub-register/internal/service"
	"strconv"

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

func (h *FieldHandler) GetFieldsBySection(c echo.Context) error {
	section := c.QueryParam("section")
	province := c.QueryParam("province")

	// Pagination defaults
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")
	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Lat/Lng for nearby
	latStr := c.QueryParam("lat")
	lngStr := c.QueryParam("lng")
	var lat, lng float64
	if latStr != "" {
		lat, _ = strconv.ParseFloat(latStr, 64)
	}
	if lngStr != "" {
		lng, _ = strconv.ParseFloat(lngStr, 64)
	}

	fields, err := h.service.GetFieldsBySection(section, province, lat, lng, limit, offset)
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

func (h *FieldHandler) GetFieldByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, StandardResponse{
			Status:  "error",
			Message: "id is required",
		})
	}

	field, err := h.service.GetFieldByID(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "record not found" {
			status = http.StatusNotFound
		}
		return c.JSON(status, StandardResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, StandardResponse{
		Status:  "success",
		Message: "Field retrieved successfully",
		Data:    field,
	})
}
