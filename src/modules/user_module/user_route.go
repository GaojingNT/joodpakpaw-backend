package user_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	user := v1.Group("users")
	user.Get("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_users"}), func(c *fiber.Ctx) error { return m.Controller().GetUsers(c) })
	user.Get(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_users"}), func(c *fiber.Ctx) error { return m.Controller().GetUser(c) })                            // Guest
	user.Put(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"update_users"}), func(c *fiber.Ctx) error { return m.Controller().UpdateUser(c) })                       // Guest
	user.Patch(":id/change-password", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"update_users"}), func(c *fiber.Ctx) error { return m.Controller().ChangePassword(c) }) // Guest
	user.Delete(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"delete_users"}), func(c *fiber.Ctx) error { return m.Controller().DeleteUser(c) })
	user.Patch(":id/roles",  func(c *fiber.Ctx) error { return m.Controller().AddRoleToUser(c) })

	adminUser := v1.Group("admin/users")
	adminUser.Get("paginate", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().AdminGetUsers(c) })
	adminUser.Get(":id", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().AdminGetUserByID(c) })
	adminUser.Patch(":id/deactivate-all-properties", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().AdminDeactivateAllProperties(c) })
	adminUser.Patch(":id/unban-all-properties", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().AdminUnbanAllProperties(c) })
}
