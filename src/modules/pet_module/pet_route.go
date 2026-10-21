package pet_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	pets := v1.Group("pets")

	// GET /api/v1/pets - Get all pets
	pets.Get("/all", func(c *fiber.Ctx) error { return m.Controller().GetAllPets(c) })
}
