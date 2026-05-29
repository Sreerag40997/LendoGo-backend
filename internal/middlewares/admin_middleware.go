package middlewares

import (
	"github.com/gofiber/fiber/v2"
)

// AdminOnly checks if the authenticated user has the 'admin' role
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Grab the role that the Protected() middleware saved
		role := c.Locals("role")

		// 2. If the role is missing or NOT 'admin', kick them out!
		if role != "admin" {
			// 403 Forbidden is the correct HTTP status for "I know who you are, but you don't have permission"
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Admin privileges required.",
			})
		}

		// 3. SUCCESS! They are an admin, let them through to the controller
		return c.Next()
	}
}