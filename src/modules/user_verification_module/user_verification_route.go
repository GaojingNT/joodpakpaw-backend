package user_verification_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	v1 := route.Group("v1")
	admin := v1.Group("admin")
	adminVeri := admin.Group("user-verifications")
	veri := v1.Group("user-verifications")

	// User routes
	veri.Get("me", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetMyVerification(c) })
	veri.Post("", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().SubmitVerification(c) })

	// Admin routes
	adminVeri.Get("paginate", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_user_verifications"}), func(c *fiber.Ctx) error { return m.Controller().GetVerifications(c) })
	adminVeri.Get(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_user_verifications"}), func(c *fiber.Ctx) error { return m.Controller().GetVerificationByID(c) })
	adminVeri.Patch(":id/approve", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"manage_user_verifications"}), func(c *fiber.Ctx) error { return m.Controller().ApproveVerification(c) })
	adminVeri.Patch(":id/reject", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"manage_user_verifications"}), func(c *fiber.Ctx) error { return m.Controller().RejectVerification(c) })
}
