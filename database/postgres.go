package database

import (
	"fmt"
	"log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"lendogo-backend/config" 
	"lendogo-backend/structures/models"
)

var DB *gorm.DB

func Connect() error {
	host := config.GetEnv("DB_HOST", "localhost")
	port := config.GetEnv("DB_PORT", "5432")
	user := config.GetEnv("DB_USER", "postgres")
	dbname := config.GetEnv("DB_NAME", "lendogo")
	password := config.GetEnv("DB_PASSWORD", "")
	if host == "" || port == "" || user == "" || dbname == "" {
		return fmt.Errorf("CRITICAL: Missing database environment variables")
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable TimeZone=Asia/Kolkata",
		host, port, user, dbname, password,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	log.Println("Running Database Migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Consultation{},
		&models.LoanApplication{},
		&models.KYCDocuments{},
		&models.FinancialDetails{},
		&models.SystemWallet{},
		&models.ChatMessage{},
		&models.UserWallet{},
		&models.LedgerEntry{},
		&models.UserProfile{},
		&models.EMISchedule{},
		&models.Staff{},
		&models.AuditLog{},
		&models.CareerOpening{},
		&models.JobApplication{},
		&models.WebConfiguration{},
	)
	if err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}
	DB = db
	log.Println("PostgreSQL connected and all tables migrated successfully!")
	return nil
}