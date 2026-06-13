package admin_controller

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/database"
	"lendogo-backend/utils"
	"lendogo-backend/internal/websockets"
)

// GetGlobalPermissions fetches the global RBAC toggles from Redis
func (c *AdminController) GetGlobalPermissions(ctx *fiber.Ctx) error {
	val, err := database.RedisClient.Get(context.Background(), "global_ui_permissions").Result()
	if err != nil {
		// Return default empty if not set
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"permissions": "[]"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"permissions": val})
}

// UpdateGlobalPermissions updates the global RBAC toggles in Redis
func (c *AdminController) UpdateGlobalPermissions(ctx *fiber.Ctx) error {
	var payload struct {
		Permissions []string `json:"permissions"`
	}
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	data, _ := json.Marshal(payload.Permissions)
	err := database.RedisClient.Set(context.Background(), "global_ui_permissions", data, 0).Err()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save global permissions"})
	}

	// 🔴 WEBSOCKET BROADCAST
	websockets.BroadcastMessage("GLOBAL_PERMISSIONS_UPDATED", fiber.Map{
		"permissions": payload.Permissions,
	})

	actorID, actorName := getActor(ctx)
	utils.RecordAudit(actorID, actorName, "WARNING", "System", "global_ui", "Updated Global RBAC Permissions", ctx.IP())

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Global permissions updated successfully"})
}
