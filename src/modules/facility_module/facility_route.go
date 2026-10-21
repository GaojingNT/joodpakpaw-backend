package facility_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	hosts := v1.Group("hosts")
	facilities := hosts.Group("facilities")

	// GET /api/v1/hosts/facilities - Get facilities with pagination
	facilities.Get("", func(c *fiber.Ctx) error { return m.Controller().GetFacilities(c) })

	// GET /api/v1/hosts/facilities/all - Get all facilities
	facilities.Get("/all", func(c *fiber.Ctx) error { return m.Controller().GetAllFacilities(c) })

	// GET /api/v1/hosts/facilities/:id - Get facility by ID
	facilities.Get("/:id", func(c *fiber.Ctx) error { return m.Controller().GetFacilityByID(c) })
}
