package config_controller

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/services"
	"lendogo-backend/internal/websockets"
	"lendogo-backend/structures/dto"
	"lendogo-backend/structures/models"
)

type ConfigController struct {
	configService services.ConfigService
}

func NewConfigController(cs services.ConfigService) *ConfigController {
	return &ConfigController{configService: cs}
}

// GetPublicConfig is fetched by React on page load to hide/show buttons
func (c *ConfigController) GetPublicConfig(ctx *fiber.Ctx) error {
	config, err := c.configService.GetConfig()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load system config"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": config})
}

// UpdateConfig is used by Admins to toggle features on/off
func (c *ConfigController) UpdateConfig(ctx *fiber.Ctx) error {
	// 1. Parse the request safely into your DTO
	var req dto.UpdateConfigReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON format"})
	}

	// 2. Map the DTO strictly to the Database Model (Security layer)
	configModel := models.WebConfiguration{
		ApplyLoanEnabled:     req.ApplyLoanEnabled,
		LoginEnabled:         req.LoginEnabled,
		RegisterEnabled:      req.RegisterEnabled,
		ApplyJobEnabled:      req.ApplyJobEnabled,
		ProfileUpdateEnabled: req.ProfileUpdateEnabled,
		FeedbackEnabled:      req.FeedbackEnabled,
		LoanHistoryEnabled:   req.LoanHistoryEnabled,
		RepayEnabled:         req.RepayEnabled,
		AutoPayEnabled:       req.AutoPayEnabled,
		InternalScoreEnabled: req.InternalScoreEnabled,
		BlogEnabled:             req.BlogEnabled,
		ChatSupportEnabled:      req.ChatSupportEnabled,
		FreeConsultationEnabled: req.FreeConsultationEnabled,
		MinCreditScore:          req.MinCreditScore,
		BaseInterestRate:        req.BaseInterestRate,
		CibilScoreEnabled:    req.CibilScoreEnabled,
	}

	// 3. Pass the clean, safe model to the Service layer
	updatedConfig, err := c.configService.UpdateConfig(configModel)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update config"})
	}

	// 4. Broadcast the update to all connected clients instantly!
	websockets.BroadcastMessage("SYS_CONFIG_UPDATE", updatedConfig)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "System configuration updated successfully",
		"data":    updatedConfig,
	})
}