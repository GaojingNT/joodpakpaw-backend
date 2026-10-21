package pet_breed_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	pets := v1.Group("pets")
	petBreeds := pets.Group(":petId/pet-breeds")

	// GET /api/v1/pets/:petId/pet-breeds - Get pet breeds with pagination
	petBreeds.Get("", func(c *fiber.Ctx) error { return m.Controller().GetPetBreeds(c) })

	// GET /api/v1/pets/:petId/pet-breeds/all - Get all pet breeds
	petBreeds.Get("/all", func(c *fiber.Ctx) error { return m.Controller().GetAllPetBreeds(c) })

	// POST /api/v1/pets/:petId/pet-breeds - Create pet breed
	petBreeds.Post("", func(c *fiber.Ctx) error { return m.Controller().CreatePetBreed(c) })

	// GET /api/v1/pets/:petId/pet-breeds/:breedId - Get pet breed by ID
	petBreeds.Get("/:breedId", func(c *fiber.Ctx) error { return m.Controller().GetPetBreedByID(c) })

	// PATCH /api/v1/pets/:petId/pet-breeds/:breedId - Update pet breed
	petBreeds.Patch("/:breedId", func(c *fiber.Ctx) error { return m.Controller().UpdatePetBreed(c) })
}
