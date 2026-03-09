package main

import (
	"log"
	"net/http"
	"os"

	"sport-hub-register/internal/database"
	"sport-hub-register/internal/handler"
	"sport-hub-register/internal/middleware"
	"sport-hub-register/internal/pkg/validator"
	"sport-hub-register/internal/repository"
	"sport-hub-register/internal/service"

	validatorV10 "github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize Database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Echo
	e := echo.New()

	// Validator
	e.Validator = &validator.CustomValidator{Validator: validatorV10.New()}

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())
	e.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(20)))

	// Initialize Layers
	tokenRepo := repository.NewTokenRepository(db)

	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(db, userRepo, tokenRepo)
	userHandler := handler.NewUserHandler(userSvc)

	otpRepo := repository.NewOTPRepository(db)
	otpSvc := service.NewOTPService(db, otpRepo, tokenRepo)
	otpHandler := handler.NewOTPHandler(otpSvc)

	storageSvc := service.NewStorageService()
	uploadHandler := handler.NewUploadHandler(storageSvc)

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "API Running")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Register API
	e.POST("/register", userHandler.Register)
	e.POST("/login", userHandler.Login)

	// OTP API
	e.POST("/otp/request", otpHandler.RequestOTP)
	e.POST("/otp/verify", otpHandler.VerifyOTP)

	// Upload API
	api := e.Group("/api/v1")
	api.Use(middleware.Auth)
	api.POST("/uploads/presign", uploadHandler.Presign)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
