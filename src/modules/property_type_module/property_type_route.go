package property_type_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	hosts := v1.Group("hosts")
	propertyTypes := hosts.Group("property-types")

	// GET /api/v1/hosts/property-types/all - Get all property types
	propertyTypes.Get("/all", func(c *fiber.Ctx) error { return m.Controller().GetAllPropertyTypes(c) }) // Host

	// GET /api/v1/hosts/property-types/:id - Get property type by ID
	propertyTypes.Get("/:id", func(c *fiber.Ctx) error { return m.Controller().GetPropertyTypeByID(c) }) // Host
}
