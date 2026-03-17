package middleware

import (
	"encoding/base64"
	"encoding/json"
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

		// Attempt to decode the JWT payload to extract user_id
		// A JWT token has 3 parts separated by dots, we need the payload (second part)
		tokenParts := strings.Split(token, ".")
		if len(tokenParts) == 3 {
			// Pad the base64 string if necessary
			payloadBase64 := tokenParts[1]
			if l := len(payloadBase64) % 4; l > 0 {
				payloadBase64 += strings.Repeat("=", 4-l)
			}
			payloadBytes, err := base64.URLEncoding.DecodeString(payloadBase64)
			if err == nil {
				var payloadData map[string]interface{}
				if err := json.Unmarshal(payloadBytes, &payloadData); err == nil {
					if userID, ok := payloadData["userid"].(string); ok {
						c.Set("user_id", userID)
						return next(c)
					}
				}
			}
		}

		// Fallback for non-JWT tokens (e.g. testing)
		c.Set("user_id", token)

		return next(c)
	}
}
