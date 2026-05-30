package routes

import (
	walletcontroller "lendogo-backend/internal/controllers/wallet_controller"

	"github.com/gofiber/fiber/v2"
	// Make sure this path matches your module name!
)

// SetupWalletRoutes registers all the admin wallet endpoints
func SetupWalletRoutes(api fiber.Router) {
	// Create a specific group for wallet actions
	walletGroup := api.Group("/admin/wallet")

	// Map the routes to the controller functions
	walletGroup.Get("/balance", walletcontroller.GetSystemBalance)
	walletGroup.Post("/create-order", walletcontroller.CreateRazorpayOrder)
	walletGroup.Post("/verify-payment", walletcontroller.VerifyRazorpayPayment)
}