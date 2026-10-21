package file_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	files := v1.Group("files")

	// Public route - serve files from storage
	files.Get("/*", func(c *fiber.Ctx) error { return m.ServeFile(c) })
}
