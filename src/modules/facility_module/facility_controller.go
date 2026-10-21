package facility_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/modules/facility_module/responses"
)

// GetFacilities lists facilities with pagination
//
//	@Summary		List facilities
//	@Description	get facilities with pagination, search and category filter
//	@Tags			Facility Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page		query		string	false	"page number"
//	@Param			limit		query		string	false	"page size"
//	@Param			search		query		string	false	"search query"
//	@Param			searchBy	query		string	false	"search by specific field (name or category name)"
//	@Param			category	query		string	false	"filter by category name"
//	@Success		200			{object}	paginator.Pagination
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/facilities [get]
func (c Controller) GetFacilities(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetFacilitiesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetFacilities")))
		responseData *paginator.Pagination
		err          error
	)

	// Get query strings
	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)
	querySearch := f.Query("search")
	querySearchBy := f.Query("searchBy")
	queryCategory := f.Query("category")

	// Get paginate values
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
		paginator.WithAttributes("search", querySearch),
		paginator.WithAttributes("search_by", querySearchBy),
		paginator.WithAttributes("category", queryCategory),
	)

	// Get data
	responseData, err = c.facilityService().GetFacilities(ctx, paginate)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetAllFacilities lists all facilities without pagination
//
//	@Summary		List all facilities
//	@Description	get all facilities with categories
//	@Tags			Facility Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	responses.GetFacilitiesResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/facilities/all [get]
func (c Controller) GetAllFacilities(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllFacilitiesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllFacilities")))
		responseData *responses.GetFacilitiesResponse
		err          error
	)

	// Get all facilities
	facilities, err := c.facilityService().GetAllFacilities(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Map to response
	responseData = &responses.GetFacilitiesResponse{
		Facilities: new(responses.GetFacilitiesResponse).Collection(facilities),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetFacilityByID get facility by ID
//
//	@Summary		Get facility by ID
//	@Description	Get a facility with its category
//	@Tags			Facility Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"facility id"
//	@Success		200	{object}	responses.GetFacilityByIDResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/facilities/{id} [get]
func (c Controller) GetFacilityByID(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetFacilityByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetFacilityByID"), attribute.Int("id", id)))
		err       error
	)

	// Get facility by ID
	facility, err := c.facilityService().GetFacilityByID(ctx, id)
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
	responseData := new(responses.GetFacilityByIDResponse).Make(facility)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

