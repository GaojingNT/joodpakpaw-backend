package chat_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	v1 := route.Group("v1")
	chatRooms := v1.Group("chat-rooms")

	// Protected routes - require authentication
	chatRooms.Get("", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetAllChatRooms(c) })
	chatRooms.Post("", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().CreateChatRoom(c) })
	chatRooms.Get(":id", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetChatRoomById(c) })
	chatRooms.Post(":id/messages", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().SendMessage(c) })
	chatRooms.Patch(":id/messages/read", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().MarkMessagesAsRead(c) })
	chatRooms.Delete(":id", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().DeleteChatRoom(c) })
}
