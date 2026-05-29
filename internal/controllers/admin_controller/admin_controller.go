package admin_controller

import (
	"github.com/gofiber/fiber/v2"

	"lendogo-backend/database"
	"lendogo-backend/structures/models"
)

// AdminController structure handles administrative HTTP requests.
type AdminController struct {
	// TODO: Transition from global database state (database.DB) to dependency injected repositories.
}

// NewAdminController initializes a new AdminController.
func NewAdminController() *AdminController {
	return &AdminController{}
}

// GetAllUsers retrieves all system users (Admin restricted).
func (c *AdminController) GetAllUsers(ctx *fiber.Ctx) error {
	// TODO: Replace mock with actual DB query
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Success! Here is the list of all users. (Admin Eyes Only)",
	})
}

// GetSystemStats retrieves aggregated system metrics.
func (c *AdminController) GetSystemStats(ctx *fiber.Ctx) error {
	// TODO: Replace mock with actual DB aggregations
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "System is running perfectly.",
		"active_loans": 42,
	})
}

// GetAllApplications fetches all loan records, preloading 1:1 KYC and Financial associations.
func (c *AdminController) GetAllApplications(ctx *fiber.Ctx) error {
	var applications []models.LoanApplication

	// Preload executes LEFT JOINs (or concurrent IN queries) to resolve nested associations efficiently, preventing N+1 issues.
	result := database.DB.
		Preload("KYC").
		Preload("FinancialDocs").
		Order("created_at DESC").
		Find(&applications)

	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to execute data aggregation for applications",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Applications fetched successfully",
		"data":    applications,
	})
}

// UpdateApplicationStatus mutates the status of a specific loan application via a state machine.
func (c *AdminController) UpdateApplicationStatus(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	
	var payload struct {
		Status string `json:"status"`
	}

	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "malformed JSON payload"})
	}

	// Enforce strict state transitions. Do not allow arbitrary strings.
	validStates := map[string]bool{
		"APPROVED":                 true, 
		"REJECTED":                 true, 
		"ADDITIONAL_DOCS_REQUIRED": true,
	}
	
	if !validStates[payload.Status] {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid state transition requested"})
	}

	// Execute the partial update strictly on the core table based on UUID
	result := database.DB.Model(&models.LoanApplication{}).
		Where("id = ?", id).
		Update("status", payload.Status)

	if result.Error != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to commit state change"})
	}

	if result.RowsAffected == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "loan application reference not found"})
	}

	return ctx.SendStatus(fiber.StatusOK)
}