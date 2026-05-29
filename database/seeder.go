package database

import (
	"log"
	"os"

	// Make sure this matches your module name in go.mod!
	"lendogo-backend/structures/models"
	"lendogo-backend/utils"
)

func SeedAdmin() {
	var adminCount int64

	// 1. Check if ANY admin already exists in the database
	DB.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)

	if adminCount > 0 {
		// An admin already exists, we don't need to do anything.
		return 
	}

	log.Println("No admin found. Creating default master admin account...")

	// 2. Read the credentials from your .env file
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		log.Println("⚠️ WARNING: ADMIN_EMAIL or ADMIN_PASSWORD is not set in .env. Skipping admin creation.")
		return
	}

	// 3. Hash the password securely using your existing util
	hashedPassword, err := utils.HashPassword(adminPassword)
	if err != nil {
		log.Fatal("Failed to hash default admin password: ", err)
	}

	// 4. Build the Admin User
	adminUser := models.User{
		FullName: "System Administrator",
		Email:    adminEmail,
		Password: hashedPassword,
		Role:     "admin", // This is what gives them special powers!
	}

	// 5. Save to the database
	if err := DB.Create(&adminUser).Error; err != nil {
		log.Fatal("Failed to seed admin user: ", err)
	}

	log.Println("✅ Default Admin account created successfully!")
}