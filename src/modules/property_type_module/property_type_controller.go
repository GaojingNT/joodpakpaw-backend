package property_type_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/modules/property_type_module/responses"
)

// GetAllPropertyTypes lists all property types without pagination
//
//	@Summary		List all property types
//	@Description	get all property types
//	@Tags			Property Type Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	responses.GetPropertyTypesResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/property-types/all [get]
func (c Controller) GetAllPropertyTypes(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllPropertyTypesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllPropertyTypes")))
		responseData *responses.GetPropertyTypesResponse
		err          error
	)

	// Get all property types
	propertyTypes, err := c.propertyTypeService().GetAllPropertyTypes(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Map to response
	responseData = &responses.GetPropertyTypesResponse{
		PropertyTypes: new(responses.GetPropertyTypesResponse).Collection(propertyTypes),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetPropertyTypeByID get property type by ID
//
//	@Summary		Get property type by ID
//	@Description	Get a property type
//	@Tags			Property Type Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"property type id"
//	@Success		200	{object}	responses.GetPropertyTypeByIDResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/property-types/{id} [get]
func (c Controller) GetPropertyTypeByID(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetPropertyTypeByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPropertyTypeByID"), attribute.Int("id", id)))
		err       error
	)

	// Get property type by ID
	propertyType, err := c.propertyTypeService().GetPropertyTypeByID(ctx, id)
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
	responseData := new(responses.GetPropertyTypeByIDResponse).Make(propertyType)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

