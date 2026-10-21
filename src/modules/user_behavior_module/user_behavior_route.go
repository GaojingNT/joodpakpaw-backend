package user_behavior_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	v1 := route.Group("v1")
	userBehavior := v1.Group("user-behaviors")

	userBehavior.Get("", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error {
		return m.Controller().GetAllUserBehaviors(c)
	})
	userBehavior.Post("/view-property/:propertyId", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error {
		return m.Controller().CreateViewPropertyBehavior(c)
	})
}
