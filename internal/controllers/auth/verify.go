package auth

import (
	"context"
	"fmt"
	"lendogo-backen/internal/database"
	"lendogo-backen/internal/models"

	"github.com/gofiber/fiber/v2"
)

// VerifyOTP checks Redis and activates the user account in PostgreSQL
func VerifyOTP(c *fiber.Ctx) error {
	var req models.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid verification data format",
		})
	}

	redisKey := fmt.Sprintf("otp:%s", req.Email)
	cachedOTP, err := database.RedisClient.Get(context.Background(), redisKey).Result()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Verification code has expired or is invalid. Please request a new one.",
		})
	}

	if cachedOTP != req.OTP {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Incorrect verification code",
		})
	}

	database.RedisClient.Del(context.Background(), redisKey)

	err = database.DB.Model(&models.User{}).Where("email = ?", req.Email).Update("is_email_verified", true).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user verification status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email verified successfully! Your account is now active.",
	})
}