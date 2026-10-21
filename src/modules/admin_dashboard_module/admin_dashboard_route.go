package admin_dashboard_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	v1 := route.Group("v1")
	admin := v1.Group("admin")

	admin.Get("dashboard", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_admin_dashboard"}), func(c *fiber.Ctx) error {
		return m.Controller().GetDashboard(c)
	})
}
