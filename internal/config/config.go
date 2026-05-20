package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

// LoadConfig forces godotenv to read the file
func LoadConfig() {
	// This looks for the .env file in the root folder
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ CRITICAL: No .env file found! The server cannot start securely.")
	}
}

// GetEnv grabs the variable, or uses the fallback if it doesn't exist
func GetEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}