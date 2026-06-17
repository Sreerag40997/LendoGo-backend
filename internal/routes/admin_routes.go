package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"lendogo-backend/internal/controllers/admin_controller"
	"lendogo-backend/internal/middlewares"
	"lendogo-backend/internal/websockets"
)

func SetupAdminRoutes(api fiber.Router, adminCtrl *admin_controller.AdminController) {
	adminGroup := api.Group("/admin")
	
	// 👇 PUBLIC ROUTE: Unprotected so staff can actually log in!
	adminGroup.Post("/login", adminCtrl.AdminLogin)

	// WebSocket endpoint for admin live broadcasts
	adminGroup.Get("/ws", websocket.New(func(c *websocket.Conn) {
		websockets.Mutex.Lock()
		websockets.Clients[c] = true
		websockets.Mutex.Unlock()

		defer func() {
			websockets.Mutex.Lock()
			delete(websockets.Clients, c)
			websockets.Mutex.Unlock()
			c.Close()
		}()

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}
	}))

	// ==========================================
	// THE VAULT DOOR: Everything below this line requires a valid Admin JWT
	// ==========================================
	adminGroup.Use(middlewares.Protected(), middlewares.AdminOnly())

	// ==========================================
	// USER MANAGEMENT (Granular UI Toggles)
	// ==========================================
	// Tied to "Read User" toggle
	adminGroup.Get("/all-users", middlewares.RequirePermission("users.read"), adminCtrl.GetAllUsers)
	
	// Tied to "Create User" toggle
	adminGroup.Post("/users", middlewares.RequirePermission("users.create"), adminCtrl.CreateUser)
	
	// Tied to "Update User" toggle
	adminGroup.Put("/users/:id", middlewares.RequirePermission("users.update"), adminCtrl.UpdateUser)
	adminGroup.Patch("/users/:id/status", middlewares.RequirePermission("users.update"), adminCtrl.UpdateUserStatus)
	
	// Tied to "Delete User" toggle
	adminGroup.Delete("/users/:id", middlewares.RequirePermission("users.delete"), adminCtrl.DeleteUser)
	
	// ==========================================
	// LOAN APPLICATIONS
	// ==========================================
	// Tied to "View Applications" toggle
	adminGroup.Get("/applications", middlewares.RequirePermission("loans.view"), adminCtrl.GetAllApplications)
	
	// Tied to "Update Applications" toggle
	adminGroup.Patch("/applications/:id/status", middlewares.RequirePermission("loans.update"), adminCtrl.UpdateApplicationStatus)

	// ==========================================
	// SYSTEM DASHBOARD
	// ==========================================
	// Tied to "View Dashboard" toggle
	adminGroup.Get("/system-stats", middlewares.RequirePermission("dashboard.view"), adminCtrl.GetSystemStats)

	// ==========================================
	// GLOBAL RBAC PERMISSIONS (Sync)
	// ==========================================
	adminGroup.Get("/global-permissions", adminCtrl.GetGlobalPermissions)
	adminGroup.Post("/global-permissions", middlewares.AdminOnly(), adminCtrl.UpdateGlobalPermissions)

	// ==========================================
	// CUSTOMER CARE
	// ==========================================
	// Tied to "View Consultation" toggle
	adminGroup.Get("/consultations", middlewares.RequirePermission("consultation.view"), adminCtrl.GetAllConsultations)

	// ==========================================
	// STAFF MANAGEMENT 
	// ==========================================
	// Tied to User Management toggles (or you can make specific "staff.create" toggles later)
	adminGroup.Post("/staff", middlewares.RequirePermission("users.create"), adminCtrl.CreateStaff)
	adminGroup.Get("/staff", middlewares.RequirePermission("users.read"), adminCtrl.GetAllStaff)
	adminGroup.Delete("/staff/:id", middlewares.RequirePermission("users.delete"), adminCtrl.DeleteStaff)
	adminGroup.Patch("/staff/:id/status", middlewares.RequirePermission("users.update"), adminCtrl.UpdateStaffStatus)

	// ==========================================
	// COMPLIANCE & AUDIT LOGS
	// ==========================================
	// Tied to an "audit.read" toggle in your UI so only authorized staff can see the paper trail
	adminGroup.Get("/audit-logs", middlewares.RequirePermission("audit.read"), adminCtrl.GetAuditLogs)

	// ==========================================
	// CAREER MANAGEMENT
	// ==========================================
	
	// ==========================================
	// ADMIN PROFILE SETTINGS
	// ==========================================
	adminGroup.Post("/profile/avatar", adminCtrl.UpdateAdminAvatar)
	adminGroup.Put("/profile", adminCtrl.UpdateAdminProfileDetails)
}










