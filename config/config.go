package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println(" No .env file found. Relying on system environment variables.")
	}
}
func GetEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}