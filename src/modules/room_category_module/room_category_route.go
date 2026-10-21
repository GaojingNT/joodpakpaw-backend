package room_category_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")

	// Host routes (protected with JWT authentication and host role)
	hosts := v1.Group("hosts")
	hostProperties := hosts.Group("properties/:property_id")
	hostRoomCategories := hostProperties.Group("room-categories")

	// GET /api/v1/hosts/properties/:property_id/room-categories - Get all room categories for a property
	hostRoomCategories.Get("/all", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_room_categories"}), func(c *fiber.Ctx) error {return m.Controller().GetAllRoomCategories(c)}) //Host

	// GET /api/v1/hosts/properties/:property_id/room-categories/:id - Get room category by ID under property
	hostRoomCategories.Get("/:id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_room_categories"}), func(c *fiber.Ctx) error {return m.Controller().GetRoomCategoryByID(c)}) //Host

	// POST /api/v1/hosts/properties/:property_id/room-categories - Create room category under property
	hostRoomCategories.Post("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"create_room_categories"}), func(c *fiber.Ctx) error {return m.Controller().CreateRoomCategory(c)}) //Host

	// PUT /api/v1/hosts/properties/:property_id/room-categories/:id - Update room category
	hostRoomCategories.Put("/:id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"update_room_categories"}), func(c *fiber.Ctx) error {return m.Controller().UpdateRoomCategory(c)}) //Host

	// DELETE /api/v1/hosts/properties/:property_id/room-categories/:id - Delete room category
	hostRoomCategories.Delete("/:id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"delete_room_categories"}), func(c *fiber.Ctx) error {return m.Controller().DeleteRoomCategory(c)}) //Host

	// Guest routes (for searching available rooms)
	properties := v1.Group("properties/:property_id")
	roomCategories := properties.Group("room-categories")

	// GET /api/v1/properties/:property_id/room-categories/search - Search available room categories
	roomCategories.Get("/search", func(c *fiber.Ctx) error {return m.Controller().SearchAvailableRoomCategories(c)})
}
