package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/services"
)

// RequireFeature acts as a bouncer, blocking requests if the Admin has disabled the feature
func RequireFeature(configService services.ConfigService, featureName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Fetch the latest configuration from the database
		config, err := configService.GetConfig()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "System configuration error",
			})
		}

		// 2. Check which feature we are verifying
		isEnabled := false
		switch featureName {
		case "apply_loan":
			isEnabled = config.ApplyLoanEnabled
		case "apply_job":
			isEnabled = config.ApplyJobEnabled
		case "register":
			isEnabled = config.RegisterEnabled
		case "login":
			isEnabled = config.LoginEnabled
		case "profile_update":
			isEnabled = config.ProfileUpdateEnabled
		case "feedback":
			isEnabled = config.FeedbackEnabled
		case "loan_history":
			isEnabled = config.LoanHistoryEnabled
		case "repay":
			isEnabled = config.RepayEnabled
		case "auto_pay":
			isEnabled = config.AutoPayEnabled
		case "internal_score":
			isEnabled = config.InternalScoreEnabled
		case "cibil_score":
			isEnabled = config.CibilScoreEnabled
		case "blog":
			isEnabled = config.BlogEnabled
		case "chat_support":
			isEnabled = config.ChatSupportEnabled
		case "free_consultation":
			isEnabled = config.FreeConsultationEnabled
		}

		// 3. Block the request if the feature is turned off
		if !isEnabled {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "This feature is currently temporarily disabled by the administrator.",
			})
		}

		// 4. If enabled, let the request continue to the Controller
		return c.Next()
	}
}