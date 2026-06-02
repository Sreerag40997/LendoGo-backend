package routes

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/controllers/wallet_controller"
)

func SetupWalletRoutes(api fiber.Router, controller *wallet_controller.WalletController) {
	walletGroup := api.Group("/admin/wallet")

	walletGroup.Get("/balance", controller.GetSystemBalance)
	walletGroup.Post("/create-order", controller.CreateRazorpayOrder)
	walletGroup.Post("/verify-payment", controller.VerifyRazorpayPayment)
	walletGroup.Post("/cheat-fund", controller.DeveloperCheatFund)
}