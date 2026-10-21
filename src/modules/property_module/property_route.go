package property_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	host := v1.Group("hosts")
	propertyForHost := host.Group("properties")
	property := v1.Group("properties")

	// Public routes (with optional auth for favorite status)
	property.Get("", middleware.JwtAuthOptional(), func(c *fiber.Ctx) error { return m.Controller().GetProperties(c) })
	property.Get("/all", middleware.JwtAuthOptional(), func(c *fiber.Ctx) error { return m.Controller().GetAllProperties(c) })
	property.Get("/recommend", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetRecommendedProperties(c) })
	property.Get("/recommend/property/:property_id", middleware.JwtAuthOptional(), func(c *fiber.Ctx) error { return m.Controller().GetRecommendedPropertiesByProperty(c) })
	property.Get("hosts/:host_id/profile", middleware.JwtAuthOptional(), func(c *fiber.Ctx) error { return m.Controller().GetHostProfile(c) })
	property.Get("hosts/:host_id/reviews", func(c *fiber.Ctx) error { return m.Controller().GetHostReviews(c) })
	property.Get(":id", middleware.JwtAuthOptional(), func(c *fiber.Ctx) error { return m.Controller().GetProperty(c) })

	// Protected routes for hosts
	propertyForHost.Get("", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetPropertiesForHostPaginate(c) }) // Host
	propertyForHost.Get("/all", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetPropertiesForHost(c) }) // Host
	propertyForHost.Get(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_properties"}), func(c *fiber.Ctx) error { return m.Controller().GetPropertyForHostByID(c) }) // Host
	propertyForHost.Post("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"create_properties"}), func(c *fiber.Ctx) error { return m.Controller().CreateProperty(c) }) // Host
	propertyForHost.Post(":id/images", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"create_property_images"}), func(c *fiber.Ctx) error { return m.Controller().AddPropertyImages(c) }) // Host
	propertyForHost.Put(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"update_properties"}), func(c *fiber.Ctx) error { return m.Controller().UpdateProperty(c) }) // Host
	propertyForHost.Patch(":id/is-active", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"update_properties"}), func(c *fiber.Ctx) error { return m.Controller().UpdatePropertyIsActive(c) }) // Host
	propertyForHost.Delete(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"delete_properties"}), func(c *fiber.Ctx) error { return m.Controller().DeleteProperty(c) }) // Host

	// Admin routes
	adminProperty := v1.Group("admin/properties")
	adminProperty.Get("paginate", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().AdminGetProperties(c) })
	adminProperty.Get(":id", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().AdminGetPropertyByID(c) })
	adminProperty.Patch(":id/deactivate", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().AdminDeactivateProperty(c) })
	adminProperty.Patch(":id/unban", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().AdminUnbanProperty(c) })
}
