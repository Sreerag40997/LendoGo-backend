package database

import (
	"log"
	"os"
	"lendogo-backend/structures/models"
)
func RunSeeders() {
	log.Println(" Starting database seeders...")
	SeedAdmin()
	seedSystemWallet()
	log.Println("All seeders executed successfully.")
}

//==============================
// Admin seed
// ============================= 

func SeedAdmin() {
	var adminCount int64
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminEmail == "" || adminPassword == "" {
		log.Println("WARNING: ADMIN_EMAIL or ADMIN_PASSWORD is not set in .env. Skipping admin creation.")
		return
	}
	DB.Model(&models.Staff{}).Where("email = ?", adminEmail).Count(&adminCount)
	if adminCount > 0 {
		return
	}
	log.Println("No admin found. Creating default master admin account in staffs table...")
	adminUser := models.Staff{
		FullName: "System Administrator",
		Email:    adminEmail,
		Password: adminPassword, 
		Role:     "Superadmin",  
		Status:   "Active",
		Permissions: map[string]bool{
			"dashboard.view":      true,
			"users.read":          true,
			"users.create":        true,
			"users.update":        true,
			"users.delete":        true,
			"loans.view":          true,
			"loans.update":        true,
			"consultation.view":   true,
		},
	}
	if err := DB.Create(&adminUser).Error; err != nil {
		log.Printf("Warning: Failed to seed admin user: %v\n", err)
	} else {
		log.Println("Default Admin account created successfully in staffs table!")
	}
}
// ======================
// Wallet amount
// =======================
func seedSystemWallet() {
	var wallet models.SystemWallet
	result := DB.FirstOrCreate(&wallet, models.SystemWallet{
		WalletName: "capital_disbursement",
		Balance:    0.0,
	})
	if result.Error != nil {
		log.Printf("Failed to seed System Wallet: %v\n", result.Error)
	} else {
		log.Println("System Wallet verified/seeded.")
	}
}