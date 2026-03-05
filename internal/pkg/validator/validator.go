package validator

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		// Optionally, one could return the error which can be handled by a custom error handler in Echo
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
