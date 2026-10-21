package facility_category_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/modules/facility_category_module/responses"
)

// GetAllFacilityCategories lists all facility categories without pagination
//
//	@Summary		List all facility categories
//	@Description	get all facility categories with their facilities
//	@Tags			Facility Category Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	responses.GetFacilityCategoriesResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/facility-categories/all [get]
func (c Controller) GetAllFacilityCategories(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllFacilityCategoriesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllFacilityCategories")))
		responseData *responses.GetFacilityCategoriesResponse
		err          error
	)

	// Get all facility categories
	facilityCategories, err := c.facilityCategoryService().GetAllFacilityCategories(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Map to response
	responseData = &responses.GetFacilityCategoriesResponse{
		FacilityCategories: new(responses.GetFacilityCategoriesResponse).Collection(facilityCategories),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetFacilityCategoryByID get facility category by ID
//
//	@Summary		Get facility category by ID
//	@Description	Get a facility category with its facilities
//	@Tags			Facility Category Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"facility category id"
//	@Success		200	{object}	responses.GetFacilityCategoryByIDResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/facility-categories/{id} [get]
func (c Controller) GetFacilityCategoryByID(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetFacilityCategoryByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetFacilityCategoryByID"), attribute.Int("id", id)))
		err       error
	)

	// Get facility category by ID
	facilityCategory, err := c.facilityCategoryService().GetFacilityCategoryByID(ctx, id)
	if err != nil {
		// Handle specific errors
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Map to response
	responseData := new(responses.GetFacilityCategoryByIDResponse).Make(facilityCategory)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
