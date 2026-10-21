package pet_profile_module

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/http_response"
	"joodpakpaw/internal/validator"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/pet_profile_module/dtos"
	"joodpakpaw/src/modules/pet_profile_module/responses"
)

// GetAllPetProfiles lists all pet profiles
//
//	@Summary		List all pet profiles
//	@Description	Get all pet profiles with pet type name and calculated age
//	@Tags			Pet Profile Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	responses.GetPetProfilesResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/profiles/pet-profiles/all [get]
func (c Controller) GetAllPetProfiles(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllPetProfilesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllPetProfiles")))
		responseData *responses.GetPetProfilesResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		c.m.tracer.TraceEnd(span)
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Get pet profiles for authenticated user only
	responseData, err = c.petProfileService().GetAllPetProfiles(ctx, authUser.ID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetPetProfile get existing pet profile by ID
//
//	@Summary		Get pet profile
//	@Description	Get pet profile by ID with full details including contact, breed, and energy rating
//	@Tags			Pet Profile Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			pet_profile_id	path		int	true	"pet profile id"
//	@Success		200				{object}	responses.GetPetProfileByIDResponse
//	@Failure		401				{object}	exception.ErrorResponse
//	@Failure		404				{object}	exception.ErrorResponse
//	@Failure		500				{object}	exception.ErrorResponse
//	@Router			/api/v1/profiles/pet-profiles/{pet_profile_id} [get]
func (c Controller) GetPetProfile(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetPetProfileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPetProfile"), attribute.Int("id", id)))
		responseData *responses.GetPetProfileByIDResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		c.m.tracer.TraceEnd(span)
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing (include userID to separate cache per user)
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"pet_profiles"}, "GetPetProfile", id, authUser.ID)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, []interface{}{id, authUser.ID}, func(ctx context.Context, params []interface{}) (*responses.GetPetProfileByIDResponse, error) {
		return c.petProfileService().GetPetProfileByID(ctx, params[0].(int), params[1].(int))
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

// CreatePetProfile creates a new pet profile
//
//	@Summary		Create pet profile
//	@Description	Create a new pet profile (guestId is automatically set from authenticated user)
//	@Tags			Pet Profile Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name			formData	string	true	"Pet name"
//	@Param			sex				formData	string	true	"Pet sex (male, female, other)"
//	@Param			birthDate		formData	string	true	"Birth date (YYYY-MM-DD)"
//	@Param			contactName		formData	string	true	"Contact name"
//	@Param			contactPhone	formData	string	true	"Contact phone"
//	@Param			note			formData	string	false	"Note"
//	@Param			petId			formData	int		true	"Pet type ID"
//	@Param			petBreedId		formData	int		true	"Pet breed ID"
//	@Param			petSizeId		formData	int		true	"Pet size category ID"
//	@Param			petEnergyId		formData	int		true	"Pet energy ID"
//	@Param			image			formData	file	false	"Pet image (max 10MB, formats: jpg, jpeg, png, webp, gif)"
//	@Success		201				{object}	responses.CreatePetProfileResponse
//	@Failure		400				{object}	exception.ErrorResponse
//	@Failure		401				{object}	exception.ErrorResponse
//	@Failure		404				{object}	exception.ErrorResponse
//	@Failure		500				{object}	exception.ErrorResponse
//	@Router			/api/v1/profiles/pet-profiles [post]
func (c Controller) CreatePetProfile(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreatePetProfileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreatePetProfile")))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Create data transfer object
	dto := new(dtos.CreatePetProfile)

	// Parse multipart form data
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Set guest ID from authenticated user
	dto.GuestID = &authUser.ID

	// Get image file if provided
	imageFile, err := f.FormFile("image")
	if err == nil && imageFile != nil {
		// Image file is provided - validate file

		// 1. Check file size (max 10MB)
		maxSize := int64(10 * 1024 * 1024) // 10MB in bytes
		if imageFile.Size > maxSize {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.AvatarFileSizeTooLargeResponseError, exception.ErrAvatarFileSizeTooLarge)
		}

		// 2. Check file format (allow only jpg, jpeg, png, webp, gif)
		allowedExtensions := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".webp": true,
			".gif":  true,
		}

		fileExt := ""
		filename := imageFile.Filename
		for i := len(filename) - 1; i >= 0; i-- {
			if filename[i] == '.' {
				fileExt = filename[i:]
				break
			}
		}

		// Convert to lowercase for comparison
		fileExtLower := ""
		for _, char := range fileExt {
			if char >= 'A' && char <= 'Z' {
				fileExtLower += string(char + 32)
			} else {
				fileExtLower += string(char)
			}
		}

		if !allowedExtensions[fileExtLower] {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidFileFormatResponseError, exception.ErrInvalidFileFormat)
		}

		dto.ImageFile = imageFile
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	petProfile, err := c.petProfileService().CreatePetProfile(ctx, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrInvalidDateFormat:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidDateFormatResponseError, err)
		case exception.ErrFutureBirthDate:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.FutureBirthDateResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear pet profile cache
	c.m.cacher.Tag("pet_profiles").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusCreated).JSON(petProfile)
}

// UpdatePetProfile updates an existing pet profile
//
//	@Summary		Update pet profile
//	@Description	Update an existing pet profile
//	@Tags			Pet Profile Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			pet_profile_id	path		int		true	"pet profile id"
//	@Param			name			formData	string	false	"Pet name"
//	@Param			contactName		formData	string	false	"Contact name"
//	@Param			contactPhone	formData	string	false	"Contact phone"
//	@Param			note			formData	string	false	"Note"
//	@Param			petSizeId		formData	int		false	"Pet size category ID"
//	@Param			petEnergyId		formData	int		false	"Pet energy ID"
//	@Param			image			formData	file	false	"Pet image (max 10MB, formats: jpg, jpeg, png, webp, gif)"
//	@Success		200				{object}	http_response.OkResponse
//	@Failure		400				{object}	exception.ErrorResponse
//	@Failure		401				{object}	exception.ErrorResponse
//	@Failure		404				{object}	exception.ErrorResponse
//	@Failure		500				{object}	exception.ErrorResponse
//	@Router			/api/v1/profiles/pet-profiles/{pet_profile_id} [put]
func (c Controller) UpdatePetProfile(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "UpdatePetProfileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdatePetProfile"), attribute.Int("id", id)))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Create data transfer object
	dto := new(dtos.UpdatePetProfile)

	// Parse multipart form data
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Get image file if provided
	imageFile, err := f.FormFile("image")
	if err == nil && imageFile != nil {
		// Image file is provided - validate file

		// 1. Check file size (max 10MB)
		maxSize := int64(10 * 1024 * 1024) // 10MB in bytes
		if imageFile.Size > maxSize {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.AvatarFileSizeTooLargeResponseError, exception.ErrAvatarFileSizeTooLarge)
		}

		// 2. Check file format (allow only jpg, jpeg, png, webp, gif)
		allowedExtensions := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".webp": true,
			".gif":  true,
		}

		fileExt := ""
		filename := imageFile.Filename
		for i := len(filename) - 1; i >= 0; i-- {
			if filename[i] == '.' {
				fileExt = filename[i:]
				break
			}
		}

		// Convert to lowercase for comparison
		fileExtLower := ""
		for _, char := range fileExt {
			if char >= 'A' && char <= 'Z' {
				fileExtLower += string(char + 32)
			} else {
				fileExtLower += string(char)
			}
		}

		if !allowedExtensions[fileExtLower] {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidFileFormatResponseError, exception.ErrInvalidFileFormat)
		}

		dto.ImageFile = imageFile
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.petProfileService().UpdatePetProfile(ctx, id, authUser.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear pet profile cache
	c.m.cacher.Tag("pet_profiles").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Pet profile updated successfully")
}

// DeletePetProfile deletes an existing pet profile
//
//	@Summary		Delete pet profile
//	@Description	Delete an existing pet profile (soft delete)
//	@Tags			Pet Profile Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			pet_profile_id	path	int	true	"pet profile id"
//	@Success		204
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/profiles/pet-profiles/{pet_profile_id} [delete]
func (c Controller) DeletePetProfile(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeletePetProfileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeletePetProfile"), attribute.Int("id", id)))
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Call service function
	err := c.petProfileService().DeletePetProfile(ctx, id, authUser.ID)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear pet profile cache
	c.m.cacher.Tag("pet_profiles").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.SendStatus(fiber.StatusNoContent)
}
