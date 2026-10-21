package routes

import (
	"fmt"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	fiberCors "github.com/gofiber/fiber/v2/middleware/cors"
	fiberEtag "github.com/gofiber/fiber/v2/middleware/etag"
	fiberFavicon "github.com/gofiber/fiber/v2/middleware/favicon"
	fiberLoggerMiddleware "github.com/gofiber/fiber/v2/middleware/logger"
	fiberMonitor "github.com/gofiber/fiber/v2/middleware/monitor"
	fiberPprof "github.com/gofiber/fiber/v2/middleware/pprof"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	fiberRequestID "github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"joodpakpaw/app/http_server"
	"joodpakpaw/config"
	"joodpakpaw/pkg/wshub"
	"joodpakpaw/src/middlewares"
	"joodpakpaw/src/modules/auth_module"
	"joodpakpaw/src/modules/booking_module"
	"joodpakpaw/src/modules/chat_module"
	"joodpakpaw/src/modules/facility_category_module"
	"joodpakpaw/src/modules/facility_module"
	"joodpakpaw/src/modules/favorite_folder_module"
	"joodpakpaw/src/modules/file_module"
	"joodpakpaw/src/modules/notification_module"
	"joodpakpaw/src/modules/user_behavior_module"
	"joodpakpaw/src/modules/user_recent_view_module"
	"joodpakpaw/src/modules/healthcheck_module"
	"joodpakpaw/src/modules/pet_breed_module"
	"joodpakpaw/src/modules/pet_module"
	"joodpakpaw/src/modules/pet_profile_module"
	"joodpakpaw/src/modules/pet_size_category_module"
	"joodpakpaw/src/modules/property_module"
	"joodpakpaw/src/modules/property_type_module"
	"joodpakpaw/src/modules/room_category_module"
	"joodpakpaw/src/modules/user_module"
	"joodpakpaw/src/modules/admin_dashboard_module"
	"joodpakpaw/src/modules/report_category_module"
	"joodpakpaw/src/modules/action_log_module"
	"joodpakpaw/src/modules/report_module"
	"joodpakpaw/src/modules/user_verification_module"

	// swagger handler
	_ "joodpakpaw/docs"
)

func HTTPRootMiddleware(r *http_server.Route) {
	// Default middleware configs
	r.Use(
		fiberRequestID.New(),
		fiberEtag.New(config.Global.Fiber.Middleware.ETag),
		fiberCors.New(config.Global.Fiber.Middleware.Cors),
		fiberLoggerMiddleware.New(config.Global.Fiber.Middleware.Logger),
		fiberFavicon.New(config.Global.Fiber.Middleware.Favicon),
		fiberRecover.New(),
		fiberPprof.New(), // pprof is registered at "/debug/pprof"
	)
}

func HTTPRoutes(s *http_server.HttpServer) {
	// Middlewares
	HTTPRootMiddleware(s.Router)
	middleware := middlewares.NewMiddleware()

	// REST API endpoints ------------------------------------------------------------------

	s.Route().Get("/", func(c *fiber.Ctx) error { return c.Status(fiber.StatusOK).SendString(c.App().Config().AppName) })
	s.Route().Get("/monitor", fiberMonitor.New(fiberMonitor.Config{Title: "App Monitoring"}))
	s.Route().Get("/swagger/*", swagger.HandlerDefault)

	// Cache clear endpoint
	// This endpoint is used to clear the Redis cache, useful for development or debugging purposes.
	s.Route().Get("/cache/clear", func(c *fiber.Ctx) error {
		if err := s.Cacher.Redis.FlushAll(c.Context()); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("ClearCacheFailedError: %+v\n", err))
		}
		return c.SendStatus(fiber.StatusOK)
	})

	// API Grouping
	api := s.Route().Group("api")

	// Mount Socket.IO handler at /api/socket.io
	sio := wshub.CurrentHub().GetSocketIOServer()
	api.All("/socket.io/*", adaptor.HTTPHandler(sio.ServeHandler(nil)))
	api.All("/socket.io", adaptor.HTTPHandler(sio.ServeHandler(nil)))

	// Register routes
	healthcheck_module.NewModule().Routes(s.MainRoute(), middleware)
	auth_module.NewModule().Routes(api, middleware)
	user_module.NewModule().Routes(api, middleware)
	property_module.NewModule().Routes(api, middleware)
	property_type_module.NewModule().Routes(api, middleware)
	facility_module.NewModule().Routes(api, middleware)
	facility_category_module.NewModule().Routes(api, middleware)
	room_category_module.NewModule().Routes(api, middleware)
	pet_size_category_module.NewModule().Routes(api, middleware)
	pet_module.NewModule().Routes(api, middleware)
	pet_breed_module.NewModule().Routes(api, middleware)
	booking_module.NewModule().Routes(api, middleware)
	favorite_folder_module.NewModule().Routes(api, middleware)
	pet_profile_module.NewModule().Routes(api, middleware)
	file_module.NewModule().Routes(api, middleware)
	notification_module.NewModule().Routes(api, middleware)
	user_behavior_module.NewModule().Routes(api, middleware)
	user_recent_view_module.NewModule().Routes(api, middleware)
	chat_module.NewModule().Routes(api, middleware)
	user_verification_module.NewModule().Routes(api, middleware)
	admin_dashboard_module.NewModule().Routes(api, middleware)
	report_category_module.NewModule().Routes(api, middleware)
	action_log_module.NewModule().Routes(api, middleware)
	report_module.NewModule().Routes(api, middleware)
}
