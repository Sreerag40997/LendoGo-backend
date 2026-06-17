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
	"lendogo-backend/internal/services" 
	"lendogo-backend/structures/models"
	"lendogo-backend/utils"

	"lendogo-backend/internal/websockets" 
)

// AdminController structure handles administrative HTTP requests.
type AdminController struct {
	adminService services.AdminService 
}

// NewAdminController initializes a new AdminController.
func NewAdminController(as services.AdminService) *AdminController {
	return &AdminController{adminService: as}
}

func getActor(ctx *fiber.Ctx) (uuid.UUID, string) {
	userIdStr, _ := ctx.Locals("user_id").(string)
	actorID, _ := uuid.Parse(userIdStr)
	actorName, _ := ctx.Locals("name").(string)
	if actorName == "" {
		actorName = "System Admin"
	}
	return actorID, actorName
}

// ==========================================
// 0. AUTHENTICATION (Staff Login)
// ==========================================

type AdminLoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *AdminController) AdminLogin(ctx *fiber.Ctx) error {
	var req AdminLoginReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	staff, err := c.adminService.AdminLogin(req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	claims := jwt.MapClaims{
		"user_id": staff.ID.String(),
		"name":    staff.FullName,
		"role":    "admin", 
		"exp":     time.Now().Add(time.Hour * 24).Unix(), 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "my_super_secret_lendo_go_key_998877" 
	}
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not log in"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token":   t,
		"staff": fiber.Map{
			"id":          staff.ID,
			"full_name":   staff.FullName,
			"email":       staff.Email,
			"avatar":      staff.Avatar,
			"role":        staff.Role,
			"permissions": staff.Permissions, 
		},
	})
}

// ==========================================
// 1. STAFF MANAGEMENT
// ==========================================

func (c *AdminController) CreateStaff(ctx *fiber.Ctx) error {
	var req services.CreateStaffDTO
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid form data"})
	}

	if err := c.adminService.CreateStaff(req); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to provision staff account"})
	}

	// 🔴 WEBSOCKET BROADCAST
	websockets.BroadcastMessage("STAFF_PROVISIONED", fiber.Map{
		"message": "A new internal staff account was created.",
		"email":   req.Email,
	})

	actorID, actorName := getActor(ctx)
	utils.RecordAudit(actorID, actorName, "SUCCESS", "Staff", "", "Created new staff account for "+req.Email, ctx.IP())

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Staff account provisioned successfully!"})
}

func (c *AdminController) GetAllStaff(ctx *fiber.Ctx) error {
	staff, err := c.adminService.GetAllStaff()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch staff directory"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Staff directory fetched", "data": staff})
}

func (c *AdminController) DeleteStaff(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	if err := c.adminService.DeleteStaff(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete staff account"})
	}

	// 🔴 WEBSOCKET BROADCAST
	websockets.BroadcastMessage("STAFF_DELETED", fiber.Map{
		"staff_id": id,
	})

	actorID, actorName := getActor(ctx)
	utils.RecordAudit(actorID, actorName, "WARNING", "Staff", id, "Deleted staff from system: "+id, ctx.IP())

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Staff deleted"})
}

func (c *AdminController) UpdateStaffStatus(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var payload struct {
		Status string `json:"status"`
	}

	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := c.adminService.UpdateStaffStatus(id, payload.Status); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update staff status"})
	}

	// 🔴 WEBSOCKET BROADCAST
	websockets.BroadcastMessage("STAFF_STATUS_UPDATED", fiber.Map{
		"staff_id": id,
		"status":   payload.Status,
	})

	actorID, actorName := getActor(ctx)
	logType := "INFO"
	if payload.Status == "Blocked" || payload.Status == "Suspended" {
		logType = "WARNING"
	}
	utils.RecordAudit(actorID, actorName, logType, "Staff", id, "Updated staff status to "+payload.Status+" for "+id, ctx.IP())

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Status updated"})
}

// ==========================================
// 2. USER MANAGEMENT
// ==========================================

func (c *AdminController) GetAllUsers(ctx *fiber.Ctx) error {
	var users []models.User
	result := database.DB.Omit("password").Preload("Profile").Order("created_at DESC").Find(&users)
	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch users"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Users fetched", "data": users})
}

func (c *AdminController) CreateUser(ctx *fiber.Ctx) error {
	// ... (Your existing struct and body parser code) ...
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
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload format"})
	}

	b := make([]byte, 4)
	_, _ = rand.Read(b)
	plainPassword := "Lendo" + hex.EncodeToString(b)[:4] + "@"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to secure key"})
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
		if err := tx.Create(&userRecord).Error; err != nil { return err }

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
		if err := tx.Create(&profileRecord).Error; err != nil { return err }
		return nil
	})

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB crash"})
	}

	// 🔴 WEBSOCKET BROADCAST
	websockets.BroadcastMessage("USER_CREATED", fiber.Map{
		"user_id":   userID,
		"full_name": req.FullName,
		"email":     req.Email,
	})

	actorID, actorName := getActor(ctx)
	utils.RecordAudit(actorID, actorName, "SUCCESS", "User", userID.String(), "Created user profile for "+req.FullName, ctx.IP())

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully!",
		"default_password": plainPassword,
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

	// ... (Your existing profile updates code) ...

	// 🔴 WEBSOCKET BROADCAST
	websockets.BroadcastMessage("USER_UPDATED", fiber.Map{
		"user_id": id,
		"message": "A user profile was updated.",
	})

	actorID, actorName := getActor(ctx)
	utils.RecordAudit(actorID, actorName, "INFO", "User", id, "Updated user profile details for "+id, ctx.IP())

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User updated"})
}

func (c *AdminController) DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	result := database.DB.Where("id = ?", id).Delete(&models.User{})
	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database crash"})
	}
	if result.RowsAffected == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not found"})
	}

	// 🔴 WEBSOCKET BROADCAST
	websockets.BroadcastMessage("USER_DELETED", fiber.Map{
		"user_id": id,
	})

	actorID, actorName := getActor(ctx)
	utils.RecordAudit(actorID, actorName, "WARNING", "User", id, "Deleted user from system: "+id, ctx.IP())

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User deleted"})
}

func (c *AdminController) UpdateUserStatus(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	var payload struct { Status string `json:"status"` }
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	database.DB.Exec("UPDATE users SET status = ? WHERE id = ?", payload.Status, id)

	// 🔴 WEBSOCKET BROADCAST
	websockets.BroadcastMessage("USER_STATUS_UPDATED", fiber.Map{
		"user_id": id,
		"status":  payload.Status,
	})

	actorID, actorName := getActor(ctx)
	logType := "INFO"
	if payload.Status == "Blocked" || payload.Status == "Suspended" {
		logType = "WARNING"
	}
	utils.RecordAudit(actorID, actorName, logType, "User", id, "Updated user status to "+payload.Status+" for "+id, ctx.IP())

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Status updated"})
}

// ==========================================
// 3. LOANS & SYSTEM DASHBOARD
// ==========================================

func (c *AdminController) GetSystemStats(ctx *fiber.Ctx) error {
	timeframe := ctx.Query("timeframe", "year") // default to year
	track := ctx.Query("track", "all")

	now := time.Now()
	var startDate time.Time

	switch timeframe {
	case "day":
		startDate = now.AddDate(0, 0, -1)
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(-1, 0, 0)
	}

	baseQuery := database.DB.Model(&models.LoanApplication{}).
		Where("application_status = ? AND created_at >= ?", "DISBURSED", startDate)
	if track != "all" && track != "" {
		baseQuery = baseQuery.Where("loan_track = ?", track)
	}

	var totalDisbursed float64
	baseQuery.Select("COALESCE(SUM(principal_amount), 0)").Scan(&totalDisbursed)

	var activePortfolio float64
	// For now, assume active portfolio is total disbursed in timeframe minus any closures (simplified to total disbursed)
	baseQuery.Select("COALESCE(SUM(principal_amount), 0)").Scan(&activePortfolio)

	type CategorySum struct {
		ProductCategory string
		Total           float64
	}
	var categories []CategorySum
	baseQuery.Select("product_category, COALESCE(SUM(principal_amount), 0) as total").
		Group("product_category").Scan(&categories)

	distMap := make(map[string]float64)
	for _, c := range categories {
		distMap[c.ProductCategory] = c.Total
	}

	// Make sure we have some defaults so charts don't break
	if len(distMap) == 0 {
		distMap["Personal"] = 0
		distMap["Business"] = 0
		distMap["Home"] = 0
	}

	type ChartDataPoint struct {
		Date       string  `json:"date"`
		Disbursed  float64 `json:"disbursed"`
		Repayments float64 `json:"repayments"`
	}
	var chartData []ChartDataPoint

	var loans []models.LoanApplication
	loansQuery := database.DB.Where("application_status = ? AND created_at >= ?", "DISBURSED", startDate)
	if track != "all" && track != "" {
		loansQuery = loansQuery.Where("loan_track = ?", track)
	}
	loansQuery.Find(&loans)

	if timeframe == "year" {
		for i := 11; i >= 0; i-- {
			d := now.AddDate(0, -i, 0)
			chartData = append(chartData, ChartDataPoint{Date: d.Format("2006-01"), Disbursed: 0, Repayments: 0})
		}
		for _, l := range loans {
			monthStr := l.CreatedAt.Format("2006-01")
			for i := range chartData {
				if chartData[i].Date == monthStr {
					chartData[i].Disbursed += l.PrincipalAmount
				}
			}
		}
	} else if timeframe == "month" {
		for i := 29; i >= 0; i-- {
			d := now.AddDate(0, 0, -i)
			chartData = append(chartData, ChartDataPoint{Date: d.Format("Jan 02"), Disbursed: 0, Repayments: 0})
		}
		for _, l := range loans {
			dayStr := l.CreatedAt.Format("Jan 02")
			for i := range chartData {
				if chartData[i].Date == dayStr {
					chartData[i].Disbursed += l.PrincipalAmount
				}
			}
		}
	} else if timeframe == "week" {
		for i := 6; i >= 0; i-- {
			d := now.AddDate(0, 0, -i)
			chartData = append(chartData, ChartDataPoint{Date: d.Format("Jan 02"), Disbursed: 0, Repayments: 0})
		}
		for _, l := range loans {
			dayStr := l.CreatedAt.Format("Jan 02")
			for i := range chartData {
				if chartData[i].Date == dayStr {
					chartData[i].Disbursed += l.PrincipalAmount
				}
			}
		}
	} else if timeframe == "day" {
		for i := 23; i >= 0; i-- {
			d := now.Add(-time.Duration(i) * time.Hour)
			chartData = append(chartData, ChartDataPoint{Date: d.Format("15:00"), Disbursed: 0, Repayments: 0})
		}
		for _, l := range loans {
			hourStr := l.CreatedAt.Format("15:00")
			for i := range chartData {
				if chartData[i].Date == hourStr {
					chartData[i].Disbursed += l.PrincipalAmount
				}
			}
		}
	}

	var totalUsers int64
	database.DB.Model(&models.User{}).Count(&totalUsers)

	var totalStaff int64
	database.DB.Model(&models.Staff{}).Count(&totalStaff)

	var totalLoans int64
	database.DB.Model(&models.LoanApplication{}).Count(&totalLoans)

	var totalKYC int64
	database.DB.Model(&models.KYCDocuments{}).Count(&totalKYC)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "System is running",
		"active_portfolio": activePortfolio,
		"total_disbursed": totalDisbursed,
		"distribution": distMap,
		"chart_data": chartData,
		"top_stats": fiber.Map{
			"total_users": totalUsers,
			"total_staff": totalStaff,
			"total_loans": totalLoans,
			"total_kyc":   totalKYC,
		},
	})
}

func (c *AdminController) GetAllApplications(ctx *fiber.Ctx) error {
	// GET route - no broadcast needed
	var applications []models.LoanApplication
	database.DB.Preload("KYC").Preload("FinancialDocs").Order("created_at DESC").Find(&applications)
	// ... (Your presigned URL generation logic) ...
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": applications})
}

func (c *AdminController) UpdateApplicationStatus(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var payload struct { Status string `json:"status"` }
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "malformed JSON payload"})
	}

	validStates := map[string]bool{
		"APPROVED": true, "REJECTED": true, "ADDITIONAL_DOCS_REQUIRED": true, "DISBURSED": true,
	}

	if !validStates[payload.Status] {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid state"})
	}

	if payload.Status == "DISBURSED" {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			// ... (Your existing disbursement transaction logic) ...
			var loan models.LoanApplication
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&loan).Error; err != nil { return err }
			if loan.ApplicationStatus == "DISBURSED" { return nil }
			
			var sysWallet models.SystemWallet
			tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("wallet_name = ?", "capital_disbursement").First(&sysWallet)
			tx.Model(&sysWallet).UpdateColumn("balance", gorm.Expr("balance - ?", loan.PrincipalAmount))

			var userWallet models.UserWallet
			tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", loan.UserID).FirstOrCreate(&userWallet, models.UserWallet{UserID: loan.UserID, Balance: 0})
			tx.Model(&userWallet).UpdateColumn("balance", gorm.Expr("balance + ?", loan.PrincipalAmount))
			
			return tx.Model(&loan).UpdateColumn("application_status", "DISBURSED").Error
		})

		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// 🔴 WEBSOCKET BROADCAST FOR DISBURSEMENT
		websockets.BroadcastMessage("LOAN_DISBURSED", fiber.Map{
			"loan_id": id,
			"message": "Capital has been moved to user wallet.",
		})

		actorID, actorName := getActor(ctx)
		utils.RecordAudit(actorID, actorName, "SUCCESS", "LoanApplication", id, "Disbursed capital for loan application: "+id, ctx.IP())

		return ctx.SendStatus(fiber.StatusOK)
	}

	result := database.DB.Model(&models.LoanApplication{}).Where("id = ?", id).Update("application_status", payload.Status)
	if result.Error != nil || result.RowsAffected == 0 {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed update"})
	}

	// 🔴 WEBSOCKET BROADCAST FOR REGULAR STATUS UPDATE
	websockets.BroadcastMessage("LOAN_STATUS_UPDATED", fiber.Map{
		"loan_id": id,
		"status":  payload.Status,
	})

	actorID, actorName := getActor(ctx)
	logType := "INFO"
	if payload.Status == "APPROVED" {
		logType = "SUCCESS"
	} else if payload.Status == "REJECTED" {
		logType = "WARNING"
	}
	utils.RecordAudit(actorID, actorName, logType, "LoanApplication", id, "Updated loan application status to "+payload.Status+" for "+id, ctx.IP())

	return ctx.SendStatus(fiber.StatusOK)
}

func (c *AdminController) GetAllConsultations(ctx *fiber.Ctx) error {
	// GET route - no broadcast needed
	var consultations []models.Consultation
	database.DB.Order("created_at DESC").Find(&consultations)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": consultations})
}
// ==========================================
// 4. COMPLIANCE & AUDIT LOGS
// ==========================================

func (c *AdminController) GetAuditLogs(ctx *fiber.Ctx) error {
	var logs []models.AuditLog
	
	// Start building the query, ordered by newest first
	query := database.DB.Model(&models.AuditLog{}).Order("created_at DESC")

	// 🔍 Advanced UI Filtering: If React sends a query parameter, apply it!
	if actionType := ctx.Query("action_type"); actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if actorID := ctx.Query("actor_id"); actorID != "" {
		query = query.Where("actor_id = ?", actorID)
	}

	// Limit to the latest 500 logs to keep the API lightning fast
	if err := query.Limit(500).Find(&logs).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch compliance audit logs"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Audit logs retrieved successfully",
		"data":    logs,
	})
}

// ==========================================
// 5. ADMIN PROFILE SETTINGS
// ==========================================

func (c *AdminController) UpdateAdminAvatar(ctx *fiber.Ctx) error {
	actorID, actorName := getActor(ctx)

	// 1. Get the uploaded file from the form data
	file, err := ctx.FormFile("avatar")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to get image file"})
	}

	// 2. Upload file to AWS S3
	s3URL, uploadErr := utils.UploadFileToS3(file)
	if uploadErr != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload image to S3"})
	}

	actorEmail, _ := ctx.Locals("email").(string)

	// 3. Update the staff record in the database
	res := database.DB.Model(&models.Staff{}).Where("id = ?", actorID).Update("avatar", s3URL)

	// 🔒 BUG FIX: If they were using an old bugged JWT token that contained the borrower 'users' table ID instead of the 'staffs' table ID,
	// the update by ID will silently affect 0 rows! If that happens, we update safely by their email.
	if res.RowsAffected == 0 && actorEmail != "" {
		database.DB.Model(&models.Staff{}).Where("email = ?", actorEmail).Update("avatar", s3URL)
	}

	// 4. Log the action
	utils.RecordAudit(actorID, actorName, "INFO", "Staff", actorID.String(), "Admin updated profile picture", ctx.IP())

	presignedURL := utils.GeneratePresignedURL(s3URL)

	// 5. Broadcast if needed (optional)
	websockets.BroadcastMessage("STAFF_UPDATED", fiber.Map{
		"staff_id": actorID,
		"avatar":   presignedURL,
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Admin profile picture updated successfully",
		"avatar":  presignedURL,
	})
}

func (c *AdminController) UpdateAdminProfileDetails(ctx *fiber.Ctx) error {
	actorID, actorName := getActor(ctx)

	var req struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.FullName == "" || req.Email == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name and email are required"})
	}

	actorEmail, _ := ctx.Locals("email").(string)
	if actorEmail == "" {
		actorEmail = req.Email // fallback to what they provided if token lacks it
	}

	// Update the staff record in the database
	res := database.DB.Model(&models.Staff{}).Where("id = ?", actorID).Updates(map[string]interface{}{
		"full_name": req.FullName,
		"email":     req.Email,
	})

	// 🔒 BUG FIX: If they were using an old bugged JWT token that contained the borrower 'users' table ID instead of the 'staffs' table ID,
	// the update by ID will silently affect 0 rows! If that happens, we update safely by their email.
	if res.RowsAffected == 0 {
		database.DB.Model(&models.Staff{}).Where("email = ?", actorEmail).Updates(map[string]interface{}{
			"full_name": req.FullName,
			"email":     req.Email,
		})
	}

	// For maximum consistency, if they have a dual borrower account, update it too!
	database.DB.Model(&models.User{}).Where("email = ?", actorEmail).Updates(map[string]interface{}{
		"full_name": req.FullName,
		"email":     req.Email,
	})

	// Log the action
	utils.RecordAudit(actorID, actorName, "INFO", "Staff", actorID.String(), "Admin updated profile details", ctx.IP())

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Profile details updated successfully",
	})
}