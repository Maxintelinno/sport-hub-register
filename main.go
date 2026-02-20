package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// 🔥 Connect DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(25)                  // max connection พร้อมกัน
	db.SetMaxIdleConns(10)                  // idle pool
	db.SetConnMaxLifetime(30 * time.Minute) // recycle connection
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Connected to PostgreSQL")

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "API Running")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	log.Printf("All ENV: %+v\n", os.Environ())
	log.Println("PORT =", os.Getenv("PORT"))

	port := os.Getenv("PORT")
	fmt.Println("PORT FROM ENV:", port)
	if port == "" {
		port = "8080" // fallback เฉพาะตอน local
	}

	e.Logger.Fatal(e.Start(":" + port))
}
