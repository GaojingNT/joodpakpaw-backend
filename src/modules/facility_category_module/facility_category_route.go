package facility_category_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	hosts := v1.Group("hosts")
	facilityCategories := hosts.Group("facility-categories")

	// GET /api/v1/hosts/facility-categories/all - Get all facility categories
	facilityCategories.Get("/all", func(c *fiber.Ctx) error { return m.Controller().GetAllFacilityCategories(c) })

	// GET /api/v1/hosts/facility-categories/:id - Get facility category by ID
	facilityCategories.Get("/:id", func(c *fiber.Ctx) error { return m.Controller().GetFacilityCategoryByID(c) })
}
