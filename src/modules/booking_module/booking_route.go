package booking_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")

	// Guest booking routes
	booking := v1.Group("bookings")
	booking.Get("/paginate", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_bookings"}), func(c *fiber.Ctx) error { return m.Controller().GetAllBookingsPaginated(c) })    // Guest (paginated)
	booking.Get("/all", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_bookings"}), func(c *fiber.Ctx) error { return m.Controller().GetAllBookings(c) }) // Guest
	booking.Get("/:id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_bookings"}), func(c *fiber.Ctx) error { return m.Controller().GetBooking(c) }) // Guest
	booking.Post("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"create_bookings"}), func(c *fiber.Ctx) error { return m.Controller().CreateBooking(c) }) // Guest
	booking.Patch("/:id/confirm-payment", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"confirm_booking_payments"}), func(c *fiber.Ctx) error { return m.Controller().ConfirmBookingPayment(c) }) // Guest
	booking.Patch("/:id/cancel", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"cancel_bookings"}), func(c *fiber.Ctx) error { return m.Controller().CancelBooking(c) }) // Guest/Host
	booking.Post("/:id/review", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"create_booking_reviews"}), func(c *fiber.Ctx) error { return m.Controller().CreateBookingReview(c) }) // Guest

	// Host routes
	host := v1.Group("hosts")
	hostBookings := host.Group("bookings")
	hostBookings.Get("/paginate", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_booking_for_hosts"}), func(c *fiber.Ctx) error { return m.Controller().GetAllBookingsForHostPaginated(c) }) // Host (paginated)
	hostBookings.Get("/all", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_booking_for_hosts"}), func(c *fiber.Ctx) error { return m.Controller().GetAllBookingsForHost(c) })             // Host
	hostBookings.Get("/:id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_booking_for_hosts"}), func(c *fiber.Ctx) error { return m.Controller().GetBookingForHost(c) }) // Host
 	hostBookings.Patch("/:id/success", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"success_bookings"}), func(c *fiber.Ctx) error { return m.Controller().SuccessBooking(c) }) // Host
	// Host property routes
	hostProperties := host.Group("properties")
	propertyBookings := hostProperties.Group(":property_id/bookings")
	propertyBookings.Get("/paginate", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_booking_for_hosts"}), func(c *fiber.Ctx) error { return m.Controller().GetAllBookingsForPropertyPaginated(c) }) // Host (paginated)
	propertyBookings.Get("/all", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_booking_for_hosts"}), func(c *fiber.Ctx) error { return m.Controller().GetAllBookingsForProperty(c) })             // Host
	propertyBookings.Get("/:booking_id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_booking_for_hosts"}), func(c *fiber.Ctx) error { return m.Controller().GetBookingForProperty(c) }) // Host

}
