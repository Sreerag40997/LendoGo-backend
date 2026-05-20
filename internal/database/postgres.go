package database

import (
	"fmt"
	"lendogo-backen/internal/models"
	"log"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Connect initializes the DB and returns an error if anything goes wrong
func Connect() error {
	// 1. Fetch variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")

	// 2. ERROR HANDLING: Check for missing critical variables
	if host == "" || port == "" || user == "" || dbname == "" {
		return fmt.Errorf("CRITICAL: Missing database environment variables in .env")
	}

	// 3. Build connection string safely
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=Asia/Kolkata",
		host, port, user, dbname, password,
	)

	// 4. Open Connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// We return the error wrapped with helpful context, we DON'T use log.Fatal here!
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// 5. Auto Migrate
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	// 6. Success! Assign to global variable and return nil (no error)
	DB = db
	log.Println("✅ PostgreSQL connected and migrated successfully!")
	return nil
}