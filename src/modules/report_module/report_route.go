package report_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	v1 := route.Group("v1")

	// Guest & Host routes
	report := v1.Group("reports")
	report.Get("me/paginate", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetMyReports(c) })
	report.Get("me/:id", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetMyReportByID(c) })
	report.Post("review", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().CreateReportReview(c) })
	report.Post("property", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().CreateReportProperty(c) })
	report.Post("unban-property", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().CreateReportUnbanProperty(c) })
	report.Post("unban-review", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().CreateReportUnbanReview(c) })

	// Admin routes
	adminReport := v1.Group("admin/reports")
	adminReport.Get("paginate", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_reports"}), func(c *fiber.Ctx) error { return m.Controller().GetReports(c) })
	adminReport.Get(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"view_reports"}), func(c *fiber.Ctx) error { return m.Controller().GetReportByID(c) })
	adminReport.Patch(":id/action", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"action_reports"}), func(c *fiber.Ctx) error { return m.Controller().ActionReport(c) })
}
