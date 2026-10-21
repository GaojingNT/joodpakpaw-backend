package room_category_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/http_response"
	"joodpakpaw/internal/utils"
	"joodpakpaw/internal/validator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/room_category_module/dtos"
	"joodpakpaw/src/modules/room_category_module/responses"
)

// GetAllRoomCategories lists all room categories for a property
//
//	@Summary		List room categories by property ID
//	@Description	get all room categories for a specific property with pets
//	@Tags			Room Category Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			property_id	path		int	true	"property id"
//	@Success		200			{object}	responses.GetRoomCategoriesResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id}/room-categories [get]
func (c Controller) GetAllRoomCategories(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "GetAllRoomCategoriesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllRoomCategories"), attribute.Int("propertyId", propertyID)))
		responseData  *responses.GetRoomCategoriesResponse
		err           error
	)

	// Get room categories by property ID
	roomCategories, err := c.roomCategoryService().GetRoomCategoriesByPropertyID(ctx, propertyID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Map to response
	responseData = &responses.GetRoomCategoriesResponse{
		RoomCategories: new(responses.GetRoomCategoriesResponse).Collection(roomCategories),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetRoomCategoryByID get room category by ID
//
//	@Summary		Get room category by ID
//	@Description	Get a room category with its pets under a property
//	@Tags			Room Category Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			property_id	path		int	true	"property id"
//	@Param			id			path		int	true	"room category id"
//	@Success		200			{object}	responses.GetRoomCategoryByIDResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id}/room-categories/{id} [get]
func (c Controller) GetRoomCategoryByID(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		id, _         = f.ParamsInt("id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "GetRoomCategoryByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetRoomCategoryByID"), attribute.Int("propertyId", propertyID), attribute.Int("id", id)))
		err           error
	)

	// Get room category by ID
	roomCategory, err := c.roomCategoryService().GetRoomCategoryByID(ctx, id)
	if err != nil {
		// Handle specific errors
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Verify room category belongs to the property
	if roomCategory.PropertyID == nil || *roomCategory.PropertyID != propertyID {
		return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, exception.ErrRecordNotFound)
	}

	// Map to response
	responseData := new(responses.GetRoomCategoryByIDResponse).Make(roomCategory)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateRoomCategory create new room category
//
//	@Summary		Create room category
//	@Description	Create a new room category under a property
//	@Tags			Room Category Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			property_id			path		int		true	"property id"
//	@Param			name				formData	string	true	"room category name"
//	@Param			totalRoom			formData	int		true	"total number of rooms"
//	@Param			numberOfPets		formData	int		false	"number of pets allowed"
//	@Param			basePrice			formData	number	true	"base price"
//	@Param			petSizeCategoryId	formData	int		false	"pet size category id (single selection)"
//	@Param			pets				formData	string	false	"pets array as JSON string (e.g., [1,2,3])"
//	@Param			facilities			formData	string	false	"facilities array as JSON string (e.g., [1,2,3])"
//	@Param			images				formData	file	false	"room category images (multiple allowed)"
//	@Success		200					{object}	http_response.OkResponse
//	@Failure		400					{object}	exception.ErrorResponse
//	@Failure		401					{object}	exception.ErrorResponse
//	@Failure		404					{object}	exception.ErrorResponse
//	@Failure		500					{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id}/room-categories [post]
func (c Controller) CreateRoomCategory(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "CreateRoomCategoryController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateRoomCategory"), attribute.Int("propertyId", propertyID)))
		err           error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Verify property belongs to the authenticated user
	_, err = c.propertyService().GetPropertyForHostByID(ctx, authUser.ID, propertyID)
	if err != nil {
		// Return 404 for both not found and unauthorized access to hide property existence
		return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, exception.ErrRecordNotFound)
	}

	// Create data transfer object
	dto := new(dtos.CreateRoomCategory)

	// Parse form data
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Set property ID from path parameter
	dto.PropertyID = &propertyID

	// Parse pets array from form-data
	if pets, paramErr := utils.ParseFormDataArray(dto.PetsString, "pets"); paramErr != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, *paramErr)
	} else {
		dto.Pets = pets
	}

	// Parse facilities array from form-data
	if facilities, paramErr := utils.ParseFormDataArray(dto.FacilitiesString, "facilities"); paramErr != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, *paramErr)
	} else {
		dto.Facilities = facilities
	}

	// Parse images from multipart form
	if form, formErr := f.MultipartForm(); formErr == nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.roomCategoryService().CreateRoomCategory(ctx, dto)
	if err != nil {
		switch err {
		case exception.ErrMaxRoomCategoriesExceeded:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.MaxRoomCategoriesExceededResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Room category created successfully")
}

// UpdateRoomCategory update existing room category by ID
//
//	@Summary		Update room category
//	@Description	Update room category under a property
//	@Tags			Room Category Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			property_id			path		int		true	"property id"
//	@Param			id					path		int		true	"room category id"
//	@Param			name				formData	string	false	"room category name"
//	@Param			totalRoom			formData	int		false	"total number of rooms"
//	@Param			numberOfPets		formData	int		false	"number of pets allowed"
//	@Param			basePrice			formData	number	false	"base price"
//	@Param			petSizeCategoryId	formData	int		false	"pet size category id"
//	@Param			pets				formData	string	false	"pets array as JSON string (e.g., [1,2,3])"
//	@Param			facilities			formData	string	false	"facilities array as JSON string (e.g., [1,2,3])"
//	@Param			images				formData	file	false	"new room category images to upload (multiple allowed)"
//	@Param			deleteImageIds		formData	string	false	"image IDs to delete as JSON string (e.g., [1,2,3])"
//	@Success		200					{object}	http_response.OkResponse
//	@Failure		400					{object}	exception.ErrorResponse
//	@Failure		401					{object}	exception.ErrorResponse
//	@Failure		404					{object}	exception.ErrorResponse
//	@Failure		500					{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id}/room-categories/{id} [put]
func (c Controller) UpdateRoomCategory(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		id, _         = f.ParamsInt("id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "UpdateRoomCategoryController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateRoomCategory"), attribute.Int("propertyId", propertyID), attribute.Int("id", id)))
		err           error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Verify property belongs to the authenticated user
	_, err = c.propertyService().GetPropertyForHostByID(ctx, authUser.ID, propertyID)
	if err != nil {
		// Return 404 for both not found and unauthorized access to hide property existence
		return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, exception.ErrRecordNotFound)
	}

	// Verify room category belongs to the property before updating
	roomCategory, err := c.roomCategoryService().GetRoomCategoryByID(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Verify room category belongs to the property
	if roomCategory.PropertyID == nil || *roomCategory.PropertyID != propertyID {
		return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, exception.ErrRecordNotFound)
	}

	// Create data transfer object
	dto := new(dtos.UpdateRoomCategory)

	// Parse form data
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Set property ID from path parameter
	dto.PropertyID = &propertyID

	// Parse pets array from form-data
	if pets, paramErr := utils.ParseFormDataArray(dto.PetsString, "pets"); paramErr != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, *paramErr)
	} else {
		dto.Pets = pets
	}

	// Parse facilities array from form-data
	if facilities, paramErr := utils.ParseFormDataArray(dto.FacilitiesString, "facilities"); paramErr != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, *paramErr)
	} else {
		dto.Facilities = facilities
	}

	// Parse deleteImageIds array from form-data
	if deleteImageIds, paramErr := utils.ParseFormDataArray(dto.DeleteImageIdsString, "deleteImageIds"); paramErr != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, *paramErr)
	} else {
		dto.DeleteImageIds = deleteImageIds
	}

	// Parse new images from multipart form
	if form, formErr := f.MultipartForm(); formErr == nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.roomCategoryService().UpdateRoomCategory(ctx, id, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Room category updated successfully")
}

// DeleteRoomCategory remove existing room category by ID
//
//	@Summary		Delete room category
//	@Description	Delete room category under a property
//	@Tags			Room Category Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			property_id	path		int	true	"property id"
//	@Param			id			path		int	true	"room category id"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id}/room-categories/{id} [delete]
func (c Controller) DeleteRoomCategory(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		id, _         = f.ParamsInt("id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "DeleteRoomCategoryController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteRoomCategory"), attribute.Int("propertyId", propertyID), attribute.Int("id", id)))
		err           error
	)

	// Verify room category belongs to the property before deleting
	roomCategory, err := c.roomCategoryService().GetRoomCategoryByID(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Verify room category belongs to the property
	if roomCategory.PropertyID == nil || *roomCategory.PropertyID != propertyID {
		return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, exception.ErrRecordNotFound)
	}

	// Call service function to delete
	err = c.roomCategoryService().DeleteRoomCategory(ctx, id)
	if err != nil {
		// Handle specific errors
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Room category deleted successfully")
}

// SearchAvailableRoomCategories search available room categories based on criteria
//
//	@Summary		Search available room categories
//	@Description	Search for room categories with available rooms based on check-in/out dates and number of pets
//	@Tags			Room Category Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			property_id		path		int		true	"property id"
//	@Param			checkInDate		query		string	true	"check-in date (YYYY-MM-DD)"
//	@Param			checkOutDate	query		string	true	"check-out date (YYYY-MM-DD)"
//	@Param			numberOfPets	query		int		true	"number of pets"
//	@Param			petSizeId		query		int		false	"pet size category id filter: small id (all rooms), medium id (medium+large), large id (large only)"
//	@Success		200				{object}	responses.SearchAvailableRoomCategoriesResponse
//	@Failure		400				{object}	exception.ErrorResponse
//	@Failure		500				{object}	exception.ErrorResponse
//	@Router			/api/v1/properties/{property_id}/room-categories/search [get]
func (c Controller) SearchAvailableRoomCategories(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "SearchAvailableRoomCategoriesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "SearchAvailableRoomCategories"), attribute.Int("propertyId", propertyID)))
		checkInDate   = f.Query("checkInDate")
		checkOutDate  = f.Query("checkOutDate")
		numberOfPets  = f.QueryInt("numberOfPets")
		petSizeId     = f.QueryInt("petSizeId")
		petSizeIdPtr  *int
		err           error
	)

	// Validate required parameters
	if checkInDate == "" || checkOutDate == "" || numberOfPets <= 0 {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}

	// Map petSizeId → filter rooms by pet size category
	// small id  → code >= 0 → all rooms
	// medium id → code >= 1 → medium + large
	// large id  → code >= 2 → large only
	if petSizeId > 0 {
		petSizeIdPtr = &petSizeId
	}

	// Call service function
	roomCategories, calendarsMap, err := c.roomCategoryService().SearchAvailableRoomCategories(ctx, propertyID, checkInDate, checkOutDate, numberOfPets, petSizeIdPtr)
	if err != nil {
		switch err {
		case exception.ErrInvalidDateRange:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidDateRangeResponseError, err)
		case exception.ErrInvalidDateFormat:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidDateFormatResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Map to response
	responseData := &responses.SearchAvailableRoomCategoriesResponse{
		RoomCategories: new(responses.SearchAvailableRoomCategoriesResponse).Collection(roomCategories, calendarsMap, numberOfPets),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
