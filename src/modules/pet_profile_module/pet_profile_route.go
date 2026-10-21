package pet_profile_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")

	// Pet profile routes
	petProfiles := v1.Group("profiles/pet-profiles")
	petProfiles.Get("/all", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetAllPetProfiles(c) }) // Guest
	petProfiles.Get("/:id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_pet_profiles"}), func(c *fiber.Ctx) error { return m.Controller().GetPetProfile(c) }) // Guest
	petProfiles.Post("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"create_pet_profiles"}), func(c *fiber.Ctx) error { return m.Controller().CreatePetProfile(c) }) // Guest
	petProfiles.Put("/:id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"update_pet_profiles"}), func(c *fiber.Ctx) error { return m.Controller().UpdatePetProfile(c) }) // Guest
	petProfiles.Delete("/:id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"delete_pet_profiles"}), func(c *fiber.Ctx) error { return m.Controller().DeletePetProfile(c) }) // Guest
}
