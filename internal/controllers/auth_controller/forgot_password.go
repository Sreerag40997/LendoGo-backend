package auth_controller

import (
	"lendogo-backend/structures/dto"
	"github.com/gofiber/fiber/v2"
)

// ==========================================
// FORGOT PASSWORD: SEND OTP
// ==========================================
func (c *AuthController) ForgotPasswordSendOTP(ctx *fiber.Ctx) error {
	var req dto.ForgotPasswordReq

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	err := c.authService.SendForgotPasswordOTP(req.Email)
	if err != nil {
		if err.Error() == "please wait 2 minutes before requesting another OTP" {
			return ctx.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If the email exists, an OTP has been sent.",
	})
}

// ==========================================
// FORGOT PASSWORD: RESET PASSWORD
// ==========================================
func (c *AuthController) ResetPassword(ctx *fiber.Ctx) error {
	var req dto.ResetPasswordReq

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if req.Password != req.ConfirmPassword {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Passwords do not match"})
	}

	err := c.authService.ResetPassword(req)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password updated successfully! You can now log in.",
	})
}