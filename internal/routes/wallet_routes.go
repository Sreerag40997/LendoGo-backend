package routes

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/controllers/wallet_controller"
	"lendogo-backend/internal/middlewares" // 👈 Added middleware import
)

func SetupWalletRoutes(api fiber.Router, controller *wallet_controller.WalletController) {
	
	// ==========================================
	// 1. ADMIN WALLET ROUTES
	// ==========================================
	// (You can also wrap this in your AdminMiddleware if you haven't globally)
	adminWalletGroup := api.Group("/admin/wallet")

	adminWalletGroup.Get("/balance", controller.GetSystemBalance)
	adminWalletGroup.Post("/create-order", controller.CreateRazorpayOrder)
	adminWalletGroup.Post("/verify-payment", controller.VerifyRazorpayPayment)
	adminWalletGroup.Post("/cheat-fund", controller.DeveloperCheatFund)
	adminWalletGroup.Post("/disburse", controller.DisburseFunds)

	// ==========================================
	// 2. USER WALLET ROUTES
	// ==========================================
	// 👈 NEW: Protected route so only logged-in users can fetch their balance
	userWalletGroup := api.Group("/user/wallet", middlewares.Protected())
	userWalletGroup.Get("/balance", controller.GetMyBalance)
}