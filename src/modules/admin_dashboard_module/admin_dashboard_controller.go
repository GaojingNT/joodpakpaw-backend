package admin_dashboard_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
)

// GetDashboard returns admin dashboard summary stats
//
//	@Summary		Admin dashboard
//	@Description	Get summary stats: pending reports, total users, active properties, pending verifications
//	@Tags			Admin Dashboard Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	responses.DashboardResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/dashboard [get]
func (c Controller) GetDashboard(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetDashboardController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetDashboard")))
	defer c.m.tracer.TraceEnd(span)

	result, err := c.dashboardService().GetDashboard(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}
