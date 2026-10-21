package report_category_module

import (
	"github.com/gofiber/fiber/v2"
	"joodpakpaw/src/middlewares"
)

func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	v1 := route.Group("v1")
	reportCategories := v1.Group("report-categories")

	// GET /api/v1/report-categories - Get all report categories
	reportCategories.Get("", func(c *fiber.Ctx) error { return m.Controller().GetAllReportCategories(c) })
}
