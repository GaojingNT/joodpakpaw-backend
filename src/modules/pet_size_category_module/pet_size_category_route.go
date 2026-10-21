package pet_size_category_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	petSizeCategories := v1.Group("pet-size-categories")

	// GET /api/v1/pet-size-categories - Get all pet size categories
	petSizeCategories.Get("/all", func(c *fiber.Ctx) error { return m.Controller().GetAllPetSizeCategories(c) }) 
}
