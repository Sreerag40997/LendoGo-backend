package auth_controller

import (
	"time" 
	"strings"
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/structures/dto"
)

// The Login Handler
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req dto.LoginReq

	// Step 1: Read the JSON from the React frontend
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Trim any accidental spaces from copy/pasting
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	// Step 2: Pass to Service layer (DI is perfectly maintained here!)
	// The service knows nothing about HTTP or cookies, just business logic.
	res, err := c.authService.Login(req)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if res.User.Status == "Blocked" {
        return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Your account has been suspended by an administrator.",
        })
    }

	// 👇 Step 3: Create the HTTP-Only Cookie 👇
	cookie := new(fiber.Cookie)
	cookie.Name = "access_token"
	cookie.Value = res.Token
	cookie.Expires = time.Now().Add(24 * time.Hour) // Should match your JWT expiration
	cookie.HTTPOnly = true  // JavaScript CANNOT read this (Protects against XSS)
	cookie.Secure = false   // IMPORTANT: Keep 'false' for localhost HTTP. Change to 'true' in Production (HTTPS)
	cookie.SameSite = "lax" // Protects against CSRF attacks
	cookie.Path = "/"       // CRITICAL: Ensure the cookie applies to the whole API

	// Step 4: Attach the cookie to the HTTP response
	ctx.Cookie(cookie)

	// Step 5: Success! Send User Data and Token back to React in the JSON body.
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"data":    res.User, 
		"token":   res.Token,
	})
}