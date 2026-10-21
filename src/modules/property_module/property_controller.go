package property_module

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/http_response"
	"joodpakpaw/internal/utils"
	"joodpakpaw/internal/validator"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/chat_module"
	"joodpakpaw/src/modules/property_module/dtos"
	"joodpakpaw/src/modules/property_module/responses"
)

// GetProperties lists all existing properties
//
//	@Summary		List properties
//	@Description	get properties with optional filtering modes:
//	@Description	- Top Rated: ?minRating=4.0 (filter by avg rating, sorted DESC)
//	@Description	- Best Value: ?sortBy=price (filter by below-average price, sorted ASC)
//	@Description	- New Listings: ?days=30 (published within N days, sorted by newest; fallback to 20 most recent if < 5 results)
//	@Description	- Trending: ?sortBy=trending (most booked properties, sorted by bookingCount DESC)
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page		query		string	false	"page number"
//	@Param			limit		query		string	false	"page size"
//	@Param			search		query		string	false	"search query"
//	@Param			searchBy	query		string	false	"field to search by"
//	@Param			minRating	query		number	false	"Top Rated: minimum avg rating filter (e.g. 4.0)"
//	@Param			sortBy		query		string	false	"Sort mode: 'price' (Best Value) or 'trending' (Most Booked)"
//	@Param			days		query		integer	false	"New Listings: published within last N days (e.g. 30)"
//	@Success		200			{object}	paginator.Pagination
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/properties [get]
func (c Controller) GetProperties(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetPropertiesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetProperties")))
		responseData *paginator.Pagination
		err          error
	)

	// Get authenticated user from context (optional)
	var userID string
	if authUser, ok := f.Locals("authUser").(*models.User); ok && authUser != nil {
		userID = strconv.Itoa(authUser.ID)
	}

	// Get query strings
	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)
	querySearch := f.Query("search")
	querySearchBy := f.Query("searchBy")
	queryMinRating := f.QueryFloat("minRating", 0)
	querySortBy := f.Query("sortBy")
	queryDays := f.QueryInt("days", 0)
	queryPetTypes := f.Query("petTypes")
	queryMaxPrice := f.QueryFloat("maxPrice", 0)

	// Get paginate values
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
		paginator.WithAttributes("search", querySearch),
		paginator.WithAttributes("search_by", querySearchBy),
		paginator.WithAttributes("searchable", models.PropertySearchable()),
		paginator.WithAttributes("min_rating", queryMinRating),
		paginator.WithAttributes("sort_by", querySortBy),
		paginator.WithAttributes("days", queryDays),
		paginator.WithAttributes("pet_types", queryPetTypes),
		paginator.WithAttributes("max_price", queryMaxPrice),
		paginator.WithAttributes("exclude_host_id", userID),
	)

	// Make cache indexing (include userID for user-specific favorites)
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"properties"}, "GetProperties", paginate.Page, paginate.Limit, paginate.Attributes, userID)

	// Get data from cache
	responseData, err = c.m.cacher.PaginationCache(ctx, cacheKey, cacheTags, paginate, func(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
		return c.propertyService().GetProperties(ctx, paginate, userID)
	})
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetAllProperties lists all properties without pagination
//
//	@Summary		List all properties
//	@Description	get all properties without pagination
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			search		query		string	false	"search query"
//	@Param			searchBy	query		string	false	"field to search by"
//	@Success		200			{object}	responses.GetAllPropertiesResponse
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/all [get]
func (c Controller) GetAllProperties(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllPropertiesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllProperties")))
		responseData *responses.GetAllPropertiesResponse
		err          error
	)

	// Get authenticated user from context (optional)
	var userID string
	if authUser, ok := f.Locals("authUser").(*models.User); ok && authUser != nil {
		userID = strconv.Itoa(authUser.ID)
	}

	// Get query strings
	querySearch := f.Query("search")
	querySearchBy := f.Query("searchBy")

	// Make cache indexing (include userID for user-specific favorites)
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"properties"}, "GetAllProperties", querySearch, querySearchBy, userID)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, map[string]string{
		"search":   querySearch,
		"searchBy": querySearchBy,
		"userID":   userID,
	}, func(ctx context.Context, params map[string]string) (*responses.GetAllPropertiesResponse, error) {
		return c.propertyService().GetAllProperties(ctx, params, params["userID"])
	})
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetProperty get existing property by ID
//
//	@Summary		Get property
//	@Description	Get property
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			property_id	path		int	true	"property id"
//	@Success		200		{object}	responses.GetPropertyByIDResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/{property_id} [get]
func (c Controller) GetProperty(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetPropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetProperty"), attribute.Int("id", id)))
		responseData *responses.GetPropertyByIDResponse
		err          error
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"properties"}, "GetProperty", id)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, id, c.propertyService().GetProperty)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Check if authenticated guest has an existing chat with this property's host (not cached — user-specific)
	if authUser, ok := f.Locals("authUser").(*models.User); ok && authUser != nil && responseData.Property.HostID != nil {
		responseData.Property.ChatRoomID = chat_module.NewRepository().FindChatRoomID(ctx, authUser.ID, *responseData.Property.HostID)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetHostReviews lists all reviews for a host's properties with pagination
//
//	@Summary		Get host reviews
//	@Description	Get paginated reviews across all properties of a host
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			host_id	path		int		true	"host user id"
//	@Param			page	query		int		false	"page number"
//	@Param			limit	query		int		false	"page size"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/hosts/{host_id}/reviews [get]
func (c Controller) GetHostReviews(f *fiber.Ctx) error {
	var (
		hostID, _ = f.ParamsInt("host_id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetHostReviewsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetHostReviews"), attribute.Int("hostId", hostID)))
	)
	defer c.m.tracer.TraceEnd(span)

	pagination := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 10)),
		paginator.WithOrderBy("reviewed_at"),
		paginator.WithSort("desc"),
	)

	result, err := c.propertyService().GetHostReviews(ctx, hostID, pagination)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}

// GetHostProfile gets public profile of a host with their properties
//
//	@Summary		Get host public profile
//	@Description	Get host profile including avgRating, reviewCount, totalProperties and all properties
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			host_id	path		int	true	"host user id"
//	@Success		200		{object}	responses.GetHostProfileResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/hosts/{host_id}/profile [get]
func (c Controller) GetHostProfile(f *fiber.Ctx) error {
	var (
		hostID, _ = f.ParamsInt("host_id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetHostProfileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetHostProfile"), attribute.Int("hostId", hostID)))
	)
	defer c.m.tracer.TraceEnd(span)

	responseData, err := c.propertyService().GetHostProfile(ctx, hostID)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Attach chat room ID between authenticated guest and this host (not cached — user-specific)
	if authUser, ok := f.Locals("authUser").(*models.User); ok && authUser != nil && authUser.ID != hostID {
		responseData.Host.ChatRoomID = chat_module.NewRepository().FindChatRoomID(ctx, authUser.ID, hostID)
	}

	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateProperty create new property
//
//	@Summary		Create property
//	@Description	Create an new property (hostId is automatically set from authenticated user)
//	@Tags			Property Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			title				formData	string	true	"property title"
//	@Param			description			formData	string	false	"property description"
//	@Param			basePrice			formData	int		true	"property base price"
//	@Param			propertyTypeId		formData	int		false	"property type id"
//	@Param			facilities			formData	string	false	"facilities array as JSON string (e.g., [1,2,3])"
//	@Param			images				formData	file	false	"property images (multiple files allowed)"
//	@Success		200					{object}	responses.CreatePropertyResponse
//	@Failure		400					{object}	exception.ErrorResponse
//	@Failure		401					{object}	exception.ErrorResponse
//	@Failure		404					{object}	exception.ErrorResponse
//	@Failure		500					{object}	exception.ErrorResponse
//	@Router			/api/v1/properties [post]
func (c Controller) CreateProperty(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreatePropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateProperty")))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	if authUser.IsBannedAllProperties {
		return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.HostBannedAllPropertiesResponseError, exception.ErrHostBannedAllProperties)
	}

	// Create data transfer object
	dto := new(dtos.CreateProperty)

	// Parse multipart form data
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Set hostId from authenticated user
	dto.HostID = &authUser.ID

	// Get uploaded images from multipart form
	form, err := f.MultipartForm()
	if err == nil && form != nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	// Parse facilities array from form-data
	if facilities, paramErr := utils.ParseFormDataArray(dto.FacilitiesString, "facilities"); paramErr != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, *paramErr)
	} else {
		dto.Facilities = facilities
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	property, err := c.propertyService().CreateProperty(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear property cache
	c.m.cacher.Tag("properties").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(new(responses.CreatePropertyResponse).Make(*property))
}

// UpdateProperty update existing property
//
//	@Summary		Update property
//	@Description	Update an existing property
//	@Tags			Property Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			property_id		path		int		true	"property id"
//	@Param			title			formData	string	false	"property title"
//	@Param			description		formData	string	false	"property description"
//	@Param			latitude		formData	number	false	"latitude"
//	@Param			longitude		formData	number	false	"longitude"
//	@Param			address			formData	string	false	"address"
//	@Param			province		formData	string	false	"province"
//	@Param			district		formData	string	false	"district"
//	@Param			subdistrict		formData	string	false	"subdistrict"
//	@Param			country			formData	string	false	"country"
//	@Param			zipcode			formData	string	false	"zipcode"
//	@Param			hostId			formData	int		false	"host id"
//	@Param			propertyTypeId	formData	int		false	"property type id"
//	@Param			images			formData	file	false	"property images"
//	@Param			deleteImageIds	formData	string	false	"image IDs to delete as JSON string (e.g., [1,2,3])"
//	@Param			facilities		formData	string	false	"facilities array as JSON string (e.g., [1,2,3])"
//	@Success		200				{object}	http_response.OkResponse
//	@Failure		400				{object}	exception.ErrorResponse
//	@Failure		401				{object}	exception.ErrorResponse
//	@Failure		404				{object}	exception.ErrorResponse
//	@Failure		500				{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/{property_id} [put]
func (c Controller) UpdateProperty(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "UpdatePropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateProperty"), attribute.Int("id", id)))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Verify property belongs to the authenticated user before updating
	_, err = c.propertyService().GetPropertyForHostByID(ctx, authUser.ID, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Create data transfer object
	dto := new(dtos.UpdateProperty)

	// Parse form data
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Get uploaded images from multipart form
	form, err := f.MultipartForm()
	if err == nil && form != nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	// Parse deleteImageIds array from form-data
	if deleteImageIds, paramErr := utils.ParseFormDataArray(dto.DeleteImageIdsString, "deleteImageIds"); paramErr != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, *paramErr)
	} else {
		dto.DeleteImageIds = deleteImageIds
	}

	// Parse facilities array from form-data
	if facilities, paramErr := utils.ParseFormDataArray(dto.FacilitiesString, "facilities"); paramErr != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, *paramErr)
	} else {
		dto.Facilities = facilities
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.propertyService().UpdateProperty(ctx, id, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear property cache
	c.m.cacher.Tag("properties").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Property updated successfully")
}

// UnbanProperty removes admin ban from a property (admin)
//
//	@Summary		Unban property
//	@Description	Remove admin ban from a property, host will be able to reactivate it
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			property_id	path		int	true	"property id"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/properties/{property_id}/unban [patch]
func (c Controller) UnbanProperty(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "UnbanPropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UnbanProperty"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	err := c.propertyService().UnbanProperty(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.cacher.Tag("properties").Flush(ctx)

	return http_response.HttpOkResponse(f, "OK", "Property unbanned successfully")
}

// AdminDeactivateProperty deactivates a property and cancels active bookings (admin)
//
//	@Summary		Admin deactivate property
//	@Description	Deactivate a property and cancel pending/confirmed bookings (admin)
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int		true	"property id"
//	@Param			adminNote	body		object	false	"admin note"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/properties/{id}/deactivate [patch]
func (c Controller) AdminDeactivateProperty(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AdminDeactivatePropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AdminDeactivateProperty"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	var body struct {
		AdminNote string `json:"adminNote"`
	}
	_ = f.BodyParser(&body)

	err := c.propertyService().AdminDeactivateProperty(ctx, id, authUser.ID, body.AdminNote)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.cacher.Tag("properties").Flush(ctx)
	return http_response.HttpOkResponse(f, "OK", "Property deactivated successfully")
}

// AdminUnbanProperty unbans a property (admin)
//
//	@Summary		Admin unban property
//	@Description	Unban a property so the host can re-activate it (admin)
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int		true	"property id"
//	@Param			adminNote	body		object	false	"admin note"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/properties/{id}/unban [patch]
func (c Controller) AdminUnbanProperty(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AdminUnbanPropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AdminUnbanProperty"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	var body struct {
		AdminNote string `json:"adminNote"`
	}
	_ = f.BodyParser(&body)

	err := c.propertyService().AdminUnbanProperty(ctx, id, authUser.ID, body.AdminNote)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.cacher.Tag("properties").Flush(ctx)
	return http_response.HttpOkResponse(f, "OK", "Property unbanned successfully")
}

// AdminGetProperties lists all properties for admin
//
//	@Summary		Admin list properties
//	@Description	Get paginated properties with host profile and ban status (admin)
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"page number"
//	@Param			limit		query		int		false	"page size"
//	@Param			search		query		string	false	"search query"
//	@Param			searchBy	query		string	false	"field to search by"
//	@Success		200			{object}	paginator.Pagination
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/properties/paginate [get]
func (c Controller) AdminGetProperties(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AdminGetPropertiesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AdminGetProperties")))
	)
	defer c.m.tracer.TraceEnd(span)

	paginate := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 20)),
		paginator.WithAttributes("search", f.Query("search")),
		paginator.WithAttributes("search_by", f.Query("searchBy")),
		paginator.WithAttributes("searchable", models.PropertySearchable()),
	)

	result, err := c.propertyService().AdminGetProperties(ctx, paginate)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}

// AdminGetPropertyByID gets a property by ID for admin
//
//	@Summary		Admin get property by ID
//	@Description	Get property detail with host profile and ban status (admin)
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"property id"
//	@Success		200	{object}	responses.AdminGetPropertyByIDResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/properties/{id} [get]
func (c Controller) AdminGetPropertyByID(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AdminGetPropertyByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AdminGetPropertyByID"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	result, err := c.propertyService().AdminGetPropertyByID(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	return f.Status(fiber.StatusOK).JSON(result)
}

// UpdatePropertyIsActive update property IsActive status
//
//	@Summary		Update property IsActive
//	@Description	Update property IsActive status (true/false)
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			property_id	path		int								true	"property id"
//	@Param			body		body		dtos.UpdatePropertyIsActive		true	"IsActive status"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/{property_id}/is-active [patch]
func (c Controller) UpdatePropertyIsActive(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "UpdatePropertyIsActiveController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdatePropertyIsActive"), attribute.Int("id", id)))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.UpdatePropertyIsActive)

	// Parse JSON body
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.propertyService().UpdatePropertyIsActive(ctx, id, dto)
	if err != nil {
		switch err {
		case exception.ErrForbidden:
			return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.ForbiddenResponseError, err)
		case exception.ErrPropertyRequiresRoomCategory:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.PropertyRequiresRoomCategoryResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear property cache
	c.m.cacher.Tag("properties").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Property IsActive updated successfully")
}

// DeleteProperty remove existing property by ID
//
//	@Summary		Delete property
//	@Description	Delete property
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			property_id	path		int	true	"property id"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/{property_id} [delete]
func (c Controller) DeleteProperty(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeletePropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteProperty"), attribute.Int("id", id)))
		err       error
	)

	// Call service function
	err = c.propertyService().DeleteProperty(ctx, id)
	if err != nil {
		// Handle specific errors
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear property cache
	c.m.cacher.Tag("properties").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Property deleted successfully")
}

// GetPropertiesForHost get all properties for authenticated host (no pagination)
//
//	@Summary		Get all properties for host
//	@Description	Get all properties for the authenticated host user without pagination
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{object}	responses.GetPropertiesForHostResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/all [get]
func (c Controller) GetPropertiesForHost(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetPropertiesForHostController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPropertiesForHost")))
		responseData *responses.GetPropertiesForHostResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"properties", "host"}, "GetPropertiesForHost", authUser.ID)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, authUser.ID, c.propertyService().GetPropertiesForHost)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetPropertiesForHostPaginate get properties for authenticated host with pagination
//
//	@Summary		Get properties for host (paginated)
//	@Description	Get properties for the authenticated host user with pagination support
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		string	false	"page number"
//	@Param			limit		query		string	false	"page size"
//	@Param			search		query		string	false	"search query"
//	@Param			searchBy	query		string	false	"field to search by"
//	@Success		200			{object}	paginator.Pagination
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties [get]
func (c Controller) GetPropertiesForHostPaginate(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetPropertiesForHostPaginateController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPropertiesForHostPaginate")))
		responseData *paginator.Pagination
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Get query strings
	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)
	querySearch := f.Query("search")
	querySearchBy := f.Query("searchBy")

	// Get paginate values
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
		paginator.WithAttributes("search", querySearch),
		paginator.WithAttributes("search_by", querySearchBy),
		paginator.WithAttributes("searchable", models.PropertySearchable()),
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"properties", "host"}, "GetPropertiesForHostPaginate", authUser.ID, paginate.Page, paginate.Limit, paginate.Attributes)

	// Get data from cache
	responseData, err = c.m.cacher.PaginationCache(ctx, cacheKey, cacheTags, paginate, func(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
		return c.propertyService().GetPropertiesForHostPaginate(ctx, authUser.ID, paginate)
	})
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetPropertyForHostByID get existing property by ID for authenticated host
//
//	@Summary		Get property for host by ID
//	@Description	Get property by ID for the authenticated host user
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			property_id	path		int	true	"property id"
//	@Success		200			{object}	responses.GetPropertyForHostByIDResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id} [get]
func (c Controller) GetPropertyForHostByID(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetPropertyForHostByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPropertyForHostByID"), attribute.Int("id", id)))
		responseData *responses.GetPropertyForHostByIDResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"properties", "host"}, "GetPropertyForHostByID", authUser.ID, id)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, map[string]int{
		"hostId":     authUser.ID,
		"propertyId": id,
	}, func(ctx context.Context, params map[string]int) (*responses.GetPropertyForHostByIDResponse, error) {
		return c.propertyService().GetPropertyForHostByID(ctx, params["hostId"], params["propertyId"])
	})
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// AddPropertyImages adds images to an existing property
//
//	@Summary		Add property images
//	@Description	Add images to an existing property (host only)
//	@Tags			Property Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"Property ID"
//	@Param			images	formData	file	true	"Property images (can upload multiple)"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{id}/images [post]
func (c Controller) AddPropertyImages(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AddPropertyImagesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AddPropertyImages"), attribute.Int("propertyId", id)))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Create data transfer object
	dto := new(dtos.AddPropertyImages)

	// Get multipart form data
	form, err := f.MultipartForm()
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Get images from form
	if form != nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	// Validate DTO
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	propertyImages, err := c.propertyService().AddPropertyImages(ctx, id, authUser.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear property cache
	c.m.cacher.Tag("properties").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(propertyImages)
}

// GetRecommendedPropertiesByProperty returns ML-recommended properties similar to a given property ID
//
//	@Summary		Get recommended properties by property
//	@Description	Get property recommendations similar to the given property from ML service
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			property_id	path		int		true	"source property id"
//	@Param			limit		query		int		false	"number of recommendations to return"
//	@Success		200			{object}	responses.GetAllPropertiesResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/recommend/property/{property_id} [get]
func (c Controller) GetRecommendedPropertiesByProperty(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		limit         = f.QueryInt("limit", 0)
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "GetRecommendedPropertiesByPropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetRecommendedPropertiesByProperty"), attribute.Int("propertyId", propertyID)))
	)
	defer c.m.tracer.TraceEnd(span)

	var userID string
	if authUser, ok := f.Locals("authUser").(*models.User); ok && authUser != nil {
		userID = strconv.Itoa(authUser.ID)
	}

	result, err := c.propertyService().GetRecommendedPropertiesByProperty(ctx, propertyID, limit, userID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}

// GetRecommendedProperties returns ML-recommended properties for the authenticated user (paginated)
//
//	@Summary		Get recommended properties
//	@Description	Get property recommendations for the authenticated user from ML service (paginated)
//	@Tags			Property Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/recommend [get]
func (c Controller) GetRecommendedProperties(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetRecommendedPropertiesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetRecommendedProperties")))
		responseData *paginator.Pagination
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Get query strings
	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)

	// Build paginator
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
	)

	responseData, err = c.propertyService().GetRecommendedProperties(ctx, authUser.ID, paginate, strconv.Itoa(authUser.ID))
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
