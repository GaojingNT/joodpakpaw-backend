package notification_module

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Routes registers all notification routes
func (m module) Routes(router fiber.Router, middleware *middlewares.Middleware) {
	controller := m.Controller()

	// WebSocket endpoint (no version prefix — WS connections are stateful)
	router.Get("/notifications/ws", controller.WebSocketUpgrade, websocket.New(controller.WebSocketHandler()))

	// SSE endpoint — works through HTTP proxies (e.g. bscit.sit.kmutt.ac.th)
	router.Get("/notifications/sse", controller.SSEHandler)

	// REST endpoints
	v1 := router.Group("v1")
	notifications := v1.Group("notifications")
	notifications.Get("", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return controller.GetNotifications(c) })
	notifications.Delete("", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return controller.DeleteAllNotifications(c) })
	notifications.Delete("/:id", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return controller.DeleteNotificationByID(c) })
}
