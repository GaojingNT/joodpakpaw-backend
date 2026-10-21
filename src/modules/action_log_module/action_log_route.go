package action_log_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	v1 := route.Group("v1")

	adminActionLog := v1.Group("admin/action-logs")
	adminActionLog.Get("paginate", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetActionLogs(c) })
}
