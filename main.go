package main

import (
	"log"
	"net/http"
	"os"

	"sport-hub-register/internal/database"
	"sport-hub-register/internal/handler"
	"sport-hub-register/internal/middleware"
	"sport-hub-register/internal/model"
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

	// Auto Migration
	db.AutoMigrate(
		&model.User{},
		&model.Field{},
		&model.FieldImage{},
		&model.FieldCourt{},
		&model.Booking{},
		&model.BookingItem{},
		&model.RegistrationToken{},
		&model.Plan{},
		&model.Subscription{},
		&model.OwnerStaff{},
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

	planRepo := repository.NewPlanRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)

	fieldRepo := repository.NewFieldRepository(db)
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(db, userRepo, tokenRepo, planRepo, subRepo, fieldRepo)
	userHandler := handler.NewUserHandler(userSvc)

	otpRepo := repository.NewOTPRepository(db)
	otpSvc := service.NewOTPService(db, otpRepo, tokenRepo)
	otpHandler := handler.NewOTPHandler(otpSvc)

	storageSvc := service.NewStorageService()
	uploadHandler := handler.NewUploadHandler(storageSvc)
	fieldSvc := service.NewFieldService(db, fieldRepo, userRepo, storageSvc)
	fieldHandler := handler.NewFieldHandler(fieldSvc)

	bookingRepo := repository.NewBookingRepository(db)
	courtRepo := repository.NewCourtRepository(db)
	bookingSvc := service.NewBookingService(db, bookingRepo, courtRepo, fieldRepo, userRepo)
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
	apiV1.POST("/owner/staffs", userHandler.RegisterStaff)

	// Booking Routes
	apiV1.POST("/courts", bookingHandler.CreateCourt)
	apiV1.PUT("/courts/:id", bookingHandler.UpdateCourt)
	apiV1.POST("/bookings", bookingHandler.CreateBooking)
	apiV1.GET("/bookings/my", bookingHandler.GetMyBookings)
	apiV1.GET("/bookings/:id/detail/cancel", bookingHandler.GetCancelDetail)
	apiV1.POST("/bookings/:id/cancel", bookingHandler.CancelBooking)
	apiV1.GET("/owner/bookings", bookingHandler.GetOwnerBookings)
	apiV1.GET("/availability", bookingHandler.GetAvailability)
	apiV1.POST("/owner/bookings/offline", bookingHandler.CreateOfflineBooking)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
