package main

import (
	"log"
	"net/http"
	"os"

	"sport-hub-register/internal/database"
	"sport-hub-register/internal/handler"
	"sport-hub-register/internal/middleware"
	"sport-hub-register/internal/pkg/validator"
	"sport-hub-register/internal/model"
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

	// Auto Migration
	db.AutoMigrate(
		&model.User{},
		&model.Field{},
		&model.FieldImage{},
		&model.FieldCourt{},
		&model.Booking{},
		&model.BookingItem{},
		&model.RegistrationToken{},
	)

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

	fieldRepo := repository.NewFieldRepository(db)
	fieldSvc := service.NewFieldService(db, fieldRepo, userRepo, storageSvc)
	fieldHandler := handler.NewFieldHandler(fieldSvc)

	bookingRepo := repository.NewBookingRepository(db)
	courtRepo := repository.NewCourtRepository(db)
	bookingSvc := service.NewBookingService(db, bookingRepo, courtRepo, fieldRepo)
	bookingHandler := handler.NewBookingHandler(bookingSvc)

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
	e.POST("/uploads/presign", uploadHandler.Presign, middleware.Auth)

	// Public Field/Court API
	e.GET("/v1/fields", fieldHandler.GetFieldsBySection)
	e.GET("/v1/courts", bookingHandler.GetCourts)

	// Protected API
	apiV1 := e.Group("/v1")
	apiV1.Use(middleware.Auth)
	apiV1.POST("/fields", fieldHandler.CreateField)
	apiV1.PUT("/fields/:id", fieldHandler.UpdateField)
	apiV1.GET("/fields/:id", fieldHandler.GetFieldByID)
	apiV1.GET("/owner/fields", fieldHandler.GetOwnerFields)
	apiV1.PATCH("/owner/fields/status", fieldHandler.UpdateFieldStatus)

	// Booking Routes
	apiV1.POST("/courts", bookingHandler.CreateCourt)
	apiV1.POST("/bookings", bookingHandler.CreateBooking)
	apiV1.GET("/bookings/my", bookingHandler.GetMyBookings)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
