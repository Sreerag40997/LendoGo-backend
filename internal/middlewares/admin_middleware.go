package middlewares

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// AdminOnly checks if the authenticated user has the 'admin' role
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Grab the role that the Protected() middleware saved
		role := c.Locals("role")
		fmt.Printf("\n[DEBUG] Token Role is: '%v'\n", role)

		// 2. If the role is missing or 'user', kick them out!
		if role == nil || role == "user" {
			// 403 Forbidden is the correct HTTP status
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Admin or Staff privileges required.",
			})
		}

		// 3. SUCCESS! They are an admin, let them through to the controller
		return c.Next()
	}
}