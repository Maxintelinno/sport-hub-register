package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Authorization header is missing"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid Authorization format"})
		}

		token := parts[1]
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token is empty"})
		}

		// (Optional) Here we would normally verify the JWT token
		// For now, we'll just allow any non-empty token as a presence check.

		return next(c)
	}
}
