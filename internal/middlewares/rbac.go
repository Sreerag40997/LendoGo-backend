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

		// 4. Map granular route permissions to top-level UI areas from the DB
		mappedArea := requiredArea
		switch requiredArea {
		case "users.read", "users.create", "users.update", "users.delete", "user_management":
			mappedArea = "User Management"
		case "loans.view", "loans.update", "loan_applications":
			mappedArea = "Loan Applications"
		case "consultation.view", "customer_care":
			mappedArea = "Customer Care"
		case "dashboard.view", "dashboard":
			mappedArea = "Dashboard"
		case "careers.view", "careers.create", "careers.update", "careers.delete":
			mappedArea = "Careers"
		case "kyc.view", "kyc.update":
			mappedArea = "KYC Verifications"
		case "blog.view", "blog.create", "blog.update", "blog.delete":
			mappedArea = "Blog Management"
		}

		// 5. Check if the required permission toggle is set to TRUE
		// First check granular (e.g. "loans.view"), then fallback to top-level UI area
		hasAccess := false
		if val, exists := staff.Permissions[requiredArea]; exists && val {
			hasAccess = true
		} else if val, exists := staff.Permissions[mappedArea]; exists && val {
			hasAccess = true
		}

		if !hasAccess {
			// 🔒 KICK THEM OUT! They don't have the toggle enabled.
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access Denied: Your role does not have clearance for the '" + mappedArea + "' module.",
			})
		}

		// 4. Access Granted! Send them to the controller.
		return c.Next()
	}
}