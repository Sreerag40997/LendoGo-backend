package auth_controller

import (
	"time" 
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

	// Step 2: Pass to Service layer (DI is perfectly maintained here!)
	// The service knows nothing about HTTP or cookies, just business logic.
	res, err := c.authService.Login(req)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
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

	// Step 4: Attach the cookie to the HTTP response
	ctx.Cookie(cookie)

	// Step 5: Success! Send ONLY the User Data back to React in the JSON body.
	// We no longer send the token in the JSON body because it's safely in the cookie.
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"data":    res.User, 
	})
}