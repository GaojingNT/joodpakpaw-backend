package user_recent_view_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	v1 := route.Group("v1")
	recentViews := v1.Group("user-recent-views")

	recentViews.Post("/:propertyId", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error {
		return m.Controller().CreateUserRecentView(c)
	})
}
