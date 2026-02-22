package main

import (
	"log"
	"net/http"
	"os"

	"sport-hub-register/internal/database"
	"sport-hub-register/internal/handler"
	"sport-hub-register/internal/repository"
	"sport-hub-register/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize Database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// Initialize Layers
	tokenRepo := repository.NewTokenRepository(db)

	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(db, userRepo, tokenRepo)
	userHandler := handler.NewUserHandler(userSvc)

	otpRepo := repository.NewOTPRepository(db)
	otpSvc := service.NewOTPService(db, otpRepo, tokenRepo)
	otpHandler := handler.NewOTPHandler(otpSvc)

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "API Running")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Register API
	e.POST("/register", userHandler.Register)

	// OTP API
	e.POST("/otp/request", otpHandler.RequestOTP)
	e.POST("/otp/verify", otpHandler.VerifyOTP)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
