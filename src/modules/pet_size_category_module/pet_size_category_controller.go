package pet_size_category_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/modules/pet_size_category_module/responses"
)

// GetAllPetSizeCategories lists all pet size categories without pagination
//
//	@Summary		List all pet size categories
//	@Description	get all pet size categories
//	@Tags			Pet Size Category Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	responses.GetPetSizeCategoriesResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/pet-size-categories [get]
func (c Controller) GetAllPetSizeCategories(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllPetSizeCategoriesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllPetSizeCategories")))
		responseData *responses.GetPetSizeCategoriesResponse
		err          error
	)

	// Get all pet size categories
	petSizeCategories, err := c.petSizeCategoryService().GetAllPetSizeCategories(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Map to response
	responseData = &responses.GetPetSizeCategoriesResponse{
		PetSizeCategories: new(responses.GetPetSizeCategoriesResponse).Collection(petSizeCategories),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
