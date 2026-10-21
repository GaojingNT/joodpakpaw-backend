package auth_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	auth := v1.Group("auth")
	// Public routes
	auth.Post("/sign-in", func(c *fiber.Ctx) error { return m.Controller().SignIn(c) })
	auth.Post("/sign-up", func(c *fiber.Ctx) error { return m.Controller().SignUp(c) })
	auth.Post("/google", func(c *fiber.Ctx) error { return m.Controller().GoogleSignIn(c) })
	auth.Post("/verify-otp", func(c *fiber.Ctx) error { return m.Controller().VerifyOTP(c) })
	auth.Post("/resend-otp", func(c *fiber.Ctx) error { return m.Controller().ResendOTP(c) })
	auth.Post("/refresh-token", func(c *fiber.Ctx) error { return m.Controller().RefreshToken(c) })
	auth.Post("/forgot-password", func(c *fiber.Ctx) error { return m.Controller().ForgotPassword(c) })
	auth.Post("/reset-password", func(c *fiber.Ctx) error { return m.Controller().ResetPassword(c) })
	auth.Post("/encrypt-password", func(c *fiber.Ctx) error { return m.Controller().EncryptPassword(c) })

	// Protected routes (require authentication)
	auth.Get("/profile", middleware.JwtAuthProtected(),  func(c *fiber.Ctx) error { return m.Controller().GetProfile(c) }) //Guest
}
