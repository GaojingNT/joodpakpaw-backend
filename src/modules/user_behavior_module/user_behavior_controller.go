package user_behavior_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/models"
)

// GetAllUserBehaviors lists all behaviors for the authenticated user
//
//	@Summary		Get all user behaviors
//	@Description	Get all behavior records for the authenticated user
//	@Tags			User Behavior Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	responses.GetUserBehaviorsResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/user-behaviors [get]
func (c Controller) GetAllUserBehaviors(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetAllUserBehaviorsController", trace.WithAttributes(
		attribute.String("server", "http"),
		attribute.String("controller", "GetAllUserBehaviors"),
	))
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	response, err := c.userBehaviorService().GetAllUserBehaviors(ctx, authUser.ID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(response)
}

// CreateViewPropertyBehavior records a view_property behavior for the authenticated user
//
//	@Summary		Record view property behavior
//	@Description	Record that the authenticated user viewed a property
//	@Tags			User Behavior Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			propertyId	path		int	true	"Property ID"
//	@Success		201			{object}	responses.CreateUserBehaviorResponse
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/user-behaviors/view-property/{propertyId} [post]
func (c Controller) CreateViewPropertyBehavior(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "CreateViewPropertyBehaviorController", trace.WithAttributes(
		attribute.String("server", "http"),
		attribute.String("controller", "CreateViewPropertyBehavior"),
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

	// Create behavior
	response, err := c.userBehaviorService().CreateViewPropertyBehavior(ctx, propertyID, authUser.ID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusCreated).JSON(response)
}
