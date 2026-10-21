package pet_breed_module

import (
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/http_response"
	"joodpakpaw/internal/validator"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/modules/pet_breed_module/dtos"
	"joodpakpaw/src/modules/pet_breed_module/responses"
)

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func validateImageFile(f *fiber.Ctx, file *multipart.FileHeader) error {
	if file.Size > 10*1024*1024 {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.AvatarFileSizeTooLargeResponseError, exception.ErrAvatarFileSizeTooLarge)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidFileFormatResponseError, exception.ErrInvalidFileFormat)
	}
	return nil
}

// CreatePetBreed creates a new pet breed
//
//	@Summary		Create pet breed
//	@Description	Create a new breed under a pet
//	@Tags			Pet Breed Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			petId	path		int		true	"pet id"
//	@Param			breed	formData	string	true	"breed name"
//	@Param			image	formData	file	false	"breed image (max 10MB, formats: jpg, jpeg, png, webp)"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/pets/{petId}/pet-breeds [post]
func (c Controller) CreatePetBreed(f *fiber.Ctx) error {
	var (
		petId, _  = f.ParamsInt("petId")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreatePetBreedController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreatePetBreed"), attribute.Int("petId", petId)))
		err       error
	)

	dto := new(dtos.CreatePetBreed)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	dto.PetID = &petId

	if imageFile, fileErr := f.FormFile("image"); fileErr == nil && imageFile != nil {
		if validateErr := validateImageFile(f, imageFile); validateErr != nil {
			return validateErr
		}
		dto.ImageFile = imageFile
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	err = c.petBreedService().CreatePetBreed(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Pet breed created successfully")
}

// UpdatePetBreed updates an existing pet breed
//
//	@Summary		Update pet breed
//	@Description	Update an existing pet breed by ID
//	@Tags			Pet Breed Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			petId	path		int		true	"pet id"
//	@Param			breedId	path		int		true	"breed id"
//	@Param			breed	formData	string	false	"breed name"
//	@Param			image	formData	file	false	"breed image (max 10MB, formats: jpg, jpeg, png, webp)"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/pets/{petId}/pet-breeds/{breedId} [patch]
func (c Controller) UpdatePetBreed(f *fiber.Ctx) error {
	var (
		petId, _   = f.ParamsInt("petId")
		breedId, _ = f.ParamsInt("breedId")
		ctx, span  = c.m.tracer.TraceStart(f.Context(), "UpdatePetBreedController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdatePetBreed"), attribute.Int("petId", petId), attribute.Int("breedId", breedId)))
		err        error
	)

	dto := new(dtos.UpdatePetBreed)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	if imageFile, fileErr := f.FormFile("image"); fileErr == nil && imageFile != nil {
		if validateErr := validateImageFile(f, imageFile); validateErr != nil {
			return validateErr
		}
		dto.ImageFile = imageFile
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	err = c.petBreedService().UpdatePetBreed(ctx, petId, breedId, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Pet breed updated successfully")
}

// GetPetBreeds lists pet breeds with pagination
//
//	@Summary		List pet breeds
//	@Description	get pet breeds with pagination by pet ID
//	@Tags			Pet Breed Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			petId	path		int		true	"pet id"
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/pets/{petId}/pet-breeds [get]
func (c Controller) GetPetBreeds(f *fiber.Ctx) error {
	var (
		petId, _     = f.ParamsInt("petId")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetPetBreedsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPetBreeds"), attribute.Int("petId", petId)))
		responseData *paginator.Pagination
		err          error
	)

	// Get query strings
	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)

	// Get paginate values
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
	)

	// Get data
	responseData, err = c.petBreedService().GetPetBreedsPaginate(ctx, paginate, petId)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetAllPetBreeds lists all pet breeds without pagination
//
//	@Summary		List all pet breeds
//	@Description	get all pet breeds by pet ID
//	@Tags			Pet Breed Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			petId	path		int	true	"pet id"
//	@Success		200		{object}	responses.GetPetBreedsResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/pets/{petId}/pet-breeds/all [get]
func (c Controller) GetAllPetBreeds(f *fiber.Ctx) error {
	var (
		petId, _     = f.ParamsInt("petId")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllPetBreedsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllPetBreeds"), attribute.Int("petId", petId)))
		responseData *responses.GetPetBreedsResponse
		err          error
	)

	// Get all pet breeds
	petBreeds, err := c.petBreedService().GetAllPetBreeds(ctx, petId)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Map to response
	responseData = &responses.GetPetBreedsResponse{
		PetBreeds: new(responses.GetPetBreedsResponse).Collection(petBreeds),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetPetBreedByID get pet breed by ID
//
//	@Summary		Get pet breed by ID
//	@Description	Get a pet breed by ID
//	@Tags			Pet Breed Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			petId	path		int	true	"pet id"
//	@Param			breedId	path		int	true	"breed id"
//	@Success		200		{object}	responses.GetPetBreedByIDResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/pets/{petId}/pet-breeds/{breedId} [get]
func (c Controller) GetPetBreedByID(f *fiber.Ctx) error {
	var (
		petId, _   = f.ParamsInt("petId")
		breedId, _ = f.ParamsInt("breedId")
		ctx, span  = c.m.tracer.TraceStart(f.Context(), "GetPetBreedByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPetBreedByID"), attribute.Int("petId", petId), attribute.Int("breedId", breedId)))
		err        error
	)

	// Get pet breed by ID
	petBreed, err := c.petBreedService().GetPetBreedByID(ctx, petId, breedId)
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
	responseData := new(responses.GetPetBreedByIDResponse).Make(petBreed)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
