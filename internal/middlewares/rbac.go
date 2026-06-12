package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/database"
	"lendogo-backend/structures/models"
)

// RequirePermission is the bouncer that checks specific feature toggles
func RequirePermission(requiredArea string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Master Admin completely bypasses these checks!
		role := c.Locals("role")
		if role == "admin" {
			return c.Next()
		}

		// 2. Get the Staff ID
		staffID := c.Locals("user_id") 
		if staffID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Missing Token Data"})
		}

		// 3. Fetch ONLY the permissions from the database for this specific staff member
		var staff models.Staff
		err := database.DB.Select("permissions").Where("id = ?", staffID).First(&staff).Error
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access Denied: Staff account not verified"})
		}

		// 4. Map snake_case from routes to Title Case UI names from the DB
		mappedArea := requiredArea
		switch requiredArea {
		case "user_management":
			mappedArea = "User Management"
		case "loan_applications":
			mappedArea = "Loan Applications"
		case "customer_care":
			mappedArea = "Customer Care"
		case "dashboard":
			mappedArea = "Dashboard"
		}

		// 5. Check if the required permission toggle is set to TRUE
		if hasAccess, exists := staff.Permissions[mappedArea]; !exists || !hasAccess {
			// 🔒 KICK THEM OUT! They don't have the toggle enabled.
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access Denied: Your role does not have clearance for the '" + mappedArea + "' module.",
			})
		}

		// 4. Access Granted! Send them to the controller.
		return c.Next()
	}
}