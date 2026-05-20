package auth

import (
	"fmt"
	"lendogo-backen/internal/database"
	"lendogo-backen/internal/models"
	authService "lendogo-backen/internal/services/auth" // 1. Import your service layer with an alias!

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Signup handles registering a fresh user securely
func Signup(c *fiber.Ctx) error {
	var req models.SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "All fields (username, email, password) are required",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process security credentials",
		})
	}

	newUser := models.User{
		Username:        req.Username,
		Email:           req.Email,
		Password:        string(hashedPassword),
		IsEmailVerified: false,
	}

	// Persist to PostgreSQL
	if err := database.DB.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username or email already exists",
		})
	}

	// 2. Generate and save OTP to Redis cache!
	otpCode, err := authService.GenerateAndSaveOTP(newUser.Email)
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"error": "User created, but failed to generate verification token",
		})
	}

	// 🚀 3. NEW SMTP INTEGRATION: Dispatch the real email!
	// If the email is fake, it will catch the error and log it to your terminal,
	// but the user will still be signed up and receive the debug token.
	err = authService.SendOTPEmail(newUser.Email, otpCode)
	if err != nil {
		fmt.Printf("⚠️ SMTP Warning: Could not deliver email to %s: %v\n", newUser.Email, err)
	}

	// 4. We still leave the debug_otp here so your fake email testing NEVER breaks!
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "User registered successfully! Check your inbox or use the debug token.",
		"user_id":   newUser.ID,
		"debug_otp": otpCode, // High-speed fallback for testing fake emails!
	})
}