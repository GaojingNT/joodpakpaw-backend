package user_recent_view_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/models"
)

// CreateUserRecentView records a recent property view for the authenticated user
//
//	@Summary		Record recent property view
//	@Description	Record that the authenticated user viewed a specific property
//	@Tags			User Recent View Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			propertyId	path		int	true	"Property ID"
//	@Success		201			{object}	responses.CreateUserRecentViewResponse
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/user-recent-views/{propertyId} [post]
func (c Controller) CreateUserRecentView(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "CreateUserRecentViewController", trace.WithAttributes(
		attribute.String("server", "http"),
		attribute.String("controller", "CreateUserRecentView"),
	))
	defer c.m.tracer.TraceEnd(span)

	// Get authenticated user
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Get property ID from path param
	propertyID, err := f.ParamsInt("propertyId")
	if err != nil || propertyID <= 0 {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}

	response, err := c.userRecentViewService().CreateUserRecentView(ctx, authUser.ID, propertyID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusCreated).JSON(response)
}
