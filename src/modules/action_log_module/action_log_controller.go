package action_log_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
)

// GetActionLogs returns paginated action logs (admin only)
//
//	@Summary		Get action logs
//	@Description	Get paginated action logs
//	@Tags			Action Log Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	false	"page number"
//	@Param			limit	query		int	false	"page size"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/action-logs/paginate [get]
func (c Controller) GetActionLogs(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetActionLogsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetActionLogs")))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	pagination := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 20)),
	)

	result, err := c.actionLogService().GetActionLogs(ctx, pagination)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}
