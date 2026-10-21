package favorite_folder_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	favoriteFolder := v1.Group("favorite-folders")
	favoriteProperty := v1.Group("favorite-properties")

	// Protected routes - all require authentication
	favoriteFolder.Get("/all", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetAllFavoriteFolders(c) }) // Guest
	favoriteFolder.Get("", middleware.JwtAuthProtected(),  func(c *fiber.Ctx) error { return m.Controller().GetFavoriteFoldersPaginate(c) }) // Guest
	favoriteFolder.Get(":id", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetFavoriteFolderById(c) }) // Guest
	favoriteFolder.Post("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"create_favorite_folders"}), func(c *fiber.Ctx) error { return m.Controller().AddPropertyToFavorite(c) }) // Guest
	favoriteFolder.Put(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"update_favorite_folders"}), func(c *fiber.Ctx) error { return m.Controller().UpdateFolderName(c) }) // Guest
	favoriteFolder.Delete(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"delete_favorite_folders"}), func(c *fiber.Ctx) error { return m.Controller().DeleteFolder(c) }) // Guest

	// Favorite properties routes - remove property from all folders
	favoriteProperty.Delete(":propertyId", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"delete_favorite_folders"}), func(c *fiber.Ctx) error { return m.Controller().RemovePropertyFromAllFolders(c) }) // Guest
}
