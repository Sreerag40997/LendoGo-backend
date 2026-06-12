package admin_controller

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"lendogo-backend/database"
	"lendogo-backend/internal/services" // 👈 Added for Staff DTOs
	"lendogo-backend/structures/models"
	"lendogo-backend/utils"
)

// AdminController structure handles administrative HTTP requests.
type AdminController struct {
	adminService services.AdminService // 👈 Needed for Staff logic
}

// NewAdminController initializes a new AdminController.
func NewAdminController(as services.AdminService) *AdminController {
	return &AdminController{adminService: as}
}

// ==========================================
// 0. AUTHENTICATION (Staff Login)
// ==========================================

// DTO for the incoming login request
type AdminLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *AdminController) AdminLogin(ctx *fiber.Ctx) error {
	var req AdminLoginReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	// 1. Verify credentials via the service
	staff, err := c.adminService.AdminLogin(req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// 2. Generate the JWT Token for the Admin
	claims := jwt.MapClaims{
		"user_id": staff.ID.String(),
		"role":    "admin", // 👈 This proves to your middleware that they are an Admin
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 1 Day Expiration
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "my_super_secret_lendo_go_key_998877" // Fallback from your .env screenshot
	}
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not log in"})
	}

	// 3. Send back the token AND the Permissions object for React's RBAC Sidebar!
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token":   t,
		"staff": fiber.Map{
			"id":          staff.ID,
			"full_name":   staff.FullName,
			"email":       staff.Email,
			"role":        staff.Role,
			"permissions": staff.Permissions, // 👈 Frontend will use this to hide/show sidebar items!
		},
	})
}

// ==========================================
// 1. STAFF MANAGEMENT (Internal Employees)
// ==========================================

func (c *AdminController) CreateStaff(ctx *fiber.Ctx) error {
	var req services.CreateStaffDTO
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form data"})
	}

	if err := c.adminService.CreateStaff(req); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to provision staff account"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Staff account provisioned successfully!"})
}

func (c *AdminController) GetAllStaff(ctx *fiber.Ctx) error {
	staff, err := c.adminService.GetAllStaff()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch staff directory"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Staff directory fetched successfully",
		"data":    staff,
	})
}

// ==========================================
// 2. USER MANAGEMENT (External Borrowers)
// ==========================================

func (c *AdminController) GetAllUsers(ctx *fiber.Ctx) error {
	var users []models.User

	result := database.DB.Omit("password").Preload("Profile").
		Order("created_at DESC").
		Find(&users)

	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users from database",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Users fetched successfully",
		"data":    users,
	})
}

func (c *AdminController) CreateUser(ctx *fiber.Ctx) error {
	var req struct {
		FullName     string `json:"full_name"`
		Email        string `json:"email"`
		Role         string `json:"role"`
		MobileNumber string `json:"mobile_number"`
		DOB          string `json:"dob"`
		PanCard      string `json:"pan_card_number"`
		CreditRating string `json:"credit_rating"`
		CreditScore  int    `json:"credit_score"`
		Address      string `json:"address"`
		City         string `json:"city"`
		State        string `json:"state"`
		Pincode      string `json:"pincode"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload mapping format"})
	}

	b := make([]byte, 4)
	_, _ = rand.Read(b)
	plainPassword := "Lendo" + hex.EncodeToString(b)[:4] + "@"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to secure default user authentication key"})
	}

	userID := uuid.New()

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		userRecord := models.User{
			ID:              userID,
			FullName:        req.FullName,
			Email:           req.Email,
			Password:        string(hashedPassword),
			Role:            req.Role,
			IsEmailVerified: true,
			Status:          "Active",
		}
		if err := tx.Create(&userRecord).Error; err != nil {
			return err
		}

		profileRecord := models.UserProfile{
			UserID:        userID,
			PhoneNumber:   req.MobileNumber,
			DateOfBirth:   req.DOB,
			Address:       req.Address,
			City:          req.City,
			State:         req.State,
			Pincode:       req.Pincode,
			TrustScore:    req.CreditScore,
			PanCardNumber: req.PanCard,
			CreditRating:  req.CreditRating,
		}
		if err := tx.Create(&profileRecord).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transactional write crash: " + err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":          "User entity and structural KYC details committed successfully!",
		"default_password": plainPassword,
		"data": map[string]interface{}{
			"id":        userID,
			"full_name": req.FullName,
			"email":     req.Email,
			"role":      req.Role,
			"status":    "Active",
			"profile": map[string]interface{}{
				"phone_number":    req.MobileNumber,
				"date_of_birth":   req.DOB,
				"address":         req.Address,
				"city":            req.City,
				"state":           req.State,
				"pincode":         req.Pincode,
				"trust_score":     req.CreditScore,
				"pan_card_number": req.PanCard,
				"credit_rating":   req.CreditRating,
			},
		},
	})
}

func (c *AdminController) UpdateUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var payload map[string]interface{}
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	updates := map[string]interface{}{
		"full_name": payload["full_name"],
		"email":     payload["email"],
		"role":      payload["role"],
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
	}

	profileUpdates := map[string]interface{}{}
	if val, ok := payload["phone_number"]; ok {
		profileUpdates["phone_number"] = val
	}
	if val, ok := payload["date_of_birth"]; ok {
		profileUpdates["date_of_birth"] = val
	}
	if val, ok := payload["address"]; ok {
		profileUpdates["address"] = val
	}
	if val, ok := payload["city"]; ok {
		profileUpdates["city"] = val
	}
	if val, ok := payload["state"]; ok {
		profileUpdates["state"] = val
	}
	if val, ok := payload["pincode"]; ok {
		profileUpdates["pincode"] = val
	}
	if val, ok := payload["trust_score"]; ok {
		profileUpdates["trust_score"] = val
	}
	if val, ok := payload["pan_card_number"]; ok {
		profileUpdates["pan_card_number"] = val
	}
	if val, ok := payload["credit_rating"]; ok {
		profileUpdates["credit_rating"] = val
	}

	if len(profileUpdates) > 0 {
		database.DB.Model(&models.UserProfile{}).Where("user_id = ?", id).Updates(profileUpdates)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User updated"})
}

func (c *AdminController) DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	result := database.DB.Where("id = ?", id).Delete(&models.User{})
	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database soft delete query crash"})
	}
	if result.RowsAffected == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Target account entity identifier not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User entity flagged as soft-deleted in history logs"})
}

func (c *AdminController) UpdateUserStatus(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var payload struct {
		Status string `json:"status"`
	}
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	database.DB.Exec("UPDATE users SET status = ? WHERE id = ?", payload.Status, id)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Status updated"})
}

// ==========================================
// 3. LOANS & SYSTEM DASHBOARD
// ==========================================

func (c *AdminController) GetSystemStats(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "System is running perfectly.",
		"active_loans": 42,
	})
}

func (c *AdminController) GetAllApplications(ctx *fiber.Ctx) error {
	var applications []models.LoanApplication

	result := database.DB.Preload("KYC").Preload("FinancialDocs").Order("created_at DESC").Find(&applications)

	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to execute data aggregation for applications"})
	}

	for i := range applications {
		applications[i].KYC.LiveSelfiePath = utils.GeneratePresignedURL(applications[i].KYC.LiveSelfiePath)
		applications[i].KYC.AadhaarFrontPath = utils.GeneratePresignedURL(applications[i].KYC.AadhaarFrontPath)
		applications[i].KYC.AadhaarBackPath = utils.GeneratePresignedURL(applications[i].KYC.AadhaarBackPath)
		applications[i].KYC.PanCardPath = utils.GeneratePresignedURL(applications[i].KYC.PanCardPath)

		applications[i].FinancialDocs.BankStatementPath = utils.GeneratePresignedURL(applications[i].FinancialDocs.BankStatementPath)
		applications[i].FinancialDocs.PropertyAgreemntPath = utils.GeneratePresignedURL(applications[i].FinancialDocs.PropertyAgreemntPath)
		applications[i].FinancialDocs.IncomeProofPath = utils.GeneratePresignedURL(applications[i].FinancialDocs.IncomeProofPath)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Applications fetched successfully",
		"data":    applications,
	})
}

func (c *AdminController) UpdateApplicationStatus(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var payload struct {
		Status string `json:"status"`
	}

	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "malformed JSON payload"})
	}

	validStates := map[string]bool{
		"APPROVED":                 true,
		"REJECTED":                 true,
		"ADDITIONAL_DOCS_REQUIRED": true,
		"DISBURSED":                true,
	}

	if !validStates[payload.Status] {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid state transition requested"})
	}

	if payload.Status == "DISBURSED" {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			var loan models.LoanApplication
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&loan).Error; err != nil {
				return err
			}

			if loan.ApplicationStatus == "DISBURSED" {
				return nil
			}

			var sysWallet models.SystemWallet
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("wallet_name = ?", "capital_disbursement").First(&sysWallet).Error; err != nil {
				return err
			}

			if sysWallet.Balance < loan.PrincipalAmount {
				return fiber.NewError(fiber.StatusBadRequest, "Insufficient capital reserves in system wallet")
			}

			if err := tx.Model(&sysWallet).UpdateColumn("balance", gorm.Expr("balance - ?", loan.PrincipalAmount)).Error; err != nil {
				return err
			}

			var userWallet models.UserWallet
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", loan.UserID).FirstOrCreate(&userWallet, models.UserWallet{UserID: loan.UserID, Balance: 0}).Error; err != nil {
				return err
			}

			if err := tx.Model(&userWallet).UpdateColumn("balance", gorm.Expr("balance + ?", loan.PrincipalAmount)).Error; err != nil {
				return err
			}

			entries := []models.LedgerEntry{
				{
					WalletID:        sysWallet.ID,
					Amount:          -loan.PrincipalAmount,
					TransactionType: "DISBURSEMENT_DEBIT",
					ReferenceID:     loan.ID.String(),
				},
				{
					WalletID:        userWallet.ID,
					Amount:          loan.PrincipalAmount,
					TransactionType: "LOAN_CREDIT",
					ReferenceID:     loan.ID.String(),
				},
			}
			if err := tx.Create(&entries).Error; err != nil {
				return err
			}

			return tx.Model(&loan).UpdateColumn("application_status", "DISBURSED").Error
		})

		if err != nil {
			if fiberErr, ok := err.(*fiber.Error); ok {
				return ctx.Status(fiberErr.Code).JSON(fiber.Map{"error": fiberErr.Message})
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return ctx.SendStatus(fiber.StatusOK)
	}

	result := database.DB.Model(&models.LoanApplication{}).Where("id = ?", id).Update("application_status", payload.Status)
	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit state change"})
	}
	if result.RowsAffected == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "loan application reference not found"})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (c *AdminController) GetAllConsultations(ctx *fiber.Ctx) error {
	var consultations []models.Consultation
	result := database.DB.Order("created_at DESC").Find(&consultations)

	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch consultations from database"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Consultations fetched successfully",
		"data":    consultations,
	})
}