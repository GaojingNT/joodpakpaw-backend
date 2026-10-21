package pet_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/modules/pet_module/responses"
)

// GetAllPets lists all pets without pagination
//
//	@Summary		List all pets
//	@Description	get all pets
//	@Tags			Pet Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	responses.GetPetsResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/pets [get]
func (c Controller) GetAllPets(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllPetsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllPets")))
		responseData *responses.GetPetsResponse
		err          error
	)

	// Get all pets
	pets, err := c.petService().GetAllPets(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Map to response
	responseData = &responses.GetPetsResponse{
		Pets: new(responses.GetPetsResponse).Collection(pets),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
