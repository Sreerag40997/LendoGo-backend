package auth_controller

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"

	"lendogo-backend/database"
	"lendogo-backend/internal/services"
	"lendogo-backend/structures/dto"
	"lendogo-backend/utils"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(service services.AuthService) *AuthController {
	return &AuthController{authService: service}
}

// ==========================================
// 1. SEND OTP (Generates code, saves to Redis, Emails it)
// ==========================================
func (c *AuthController) SendOTP(ctx *fiber.Ctx) error {
	var req dto.SendOTPReq

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Generate a 6-digit OTP
	rand.Seed(time.Now().UnixNano())
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Save to Redis (Expires in 5 minutes)
	redisKey := "otp:" + req.Email
	err := database.RedisClient.Set(context.Background(), redisKey, otp, 5*time.Minute).Err()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store OTP in cache"})
	}

	// Send Email
	err = utils.SendOTPEmail(req.Email, otp)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send email"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP sent successfully",
		"dev_otp": otp, // Leave this here so you can copy it straight from Postman while testing!
	})
}

// ==========================================
// 2. VERIFY OTP (Checks Redis)
// ==========================================
func (c *AuthController) VerifyOTP(ctx *fiber.Ctx) error {
	var req dto.VerifyOTPReq

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	redisKey := "otp:" + req.Email
	storedOTP, err := database.RedisClient.Get(context.Background(), redisKey).Result()

	if err != nil || storedOTP != req.OTP {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired OTP"})
	}

	// Optional: Delete the OTP from Redis so it can't be reused
	database.RedisClient.Del(context.Background(), redisKey)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP verified successfully. Proceed to set password.",
	})
}

// ==========================================
// 3. SET PASSWORD (Saves to Database)
// ==========================================
func (c *AuthController) SetPassword(ctx *fiber.Ctx) error {
	var req dto.SetPasswordReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if req.Password != req.ConfirmPassword {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Passwords do not match"})
	}

	// Create the register object including the Username
	registerData := dto.RegisterReq{
		FullName: req.FullName, // <--- Add this!
		Email:    req.Email,
		Password: req.Password,
	}

	err := c.authService.Register(registerData)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// 🚀 AUTOMATIC LOGIN ON SUCCESSFUL REGISTRATION 🚀
	loginRes, loginErr := c.authService.Login(dto.LoginReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if loginErr == nil && loginRes != nil {
		cookie := new(fiber.Cookie)
		cookie.Name = "access_token"
		cookie.Value = loginRes.Token
		cookie.Expires = time.Now().Add(24 * time.Hour)
		cookie.HTTPOnly = true
		cookie.Secure = false
		cookie.SameSite = "lax"
		ctx.Cookie(cookie)

		return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Registration and login complete!",
			"data":    loginRes.User,
			"token":   loginRes.Token,
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Registration complete!",
	})
}