package favorite_folder_module

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/validator"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/favorite_folder_module/dtos"
	"joodpakpaw/src/modules/favorite_folder_module/responses"
)

// GetAllFavoriteFolders lists all favorite folders without pagination
//
//	@Summary		List all favorite folders
//	@Description	Get all favorite folders for authenticated user
//	@Tags			Favorite Folder Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	responses.GetFavoriteFoldersResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/favorite-folders/all [get]
func (c Controller) GetAllFavoriteFolders(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllFavoriteFoldersController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllFavoriteFolders")))
		responseData *responses.GetFavoriteFoldersResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Convert user ID to string
	userID := strconv.Itoa(authUser.ID)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"favorite_folders"}, "GetAllFavoriteFolders", userID)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, userID, c.favoriteFolderService().GetAllFavoriteFolders)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetFavoriteFoldersPaginate lists favorite folders with pagination
//
//	@Summary		List favorite folders with pagination
//	@Description	Get favorite folders for authenticated user with pagination
//	@Tags			Favorite Folder Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	false	"page number"	default(1)
//	@Param			limit	query		int	false	"page size"		default(20)
//	@Success		200		{object}	paginator.Pagination
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/favorite-folders [get]
func (c Controller) GetFavoriteFoldersPaginate(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetFavoriteFoldersPaginateController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetFavoriteFoldersPaginate")))
		responseData *paginator.Pagination
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Convert user ID to string
	userID := strconv.Itoa(authUser.ID)

	// Get query strings
	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)

	// Get paginate values
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"favorite_folders"}, "GetFavoriteFoldersPaginate", userID, paginate.Page, paginate.Limit)

	// Get data from cache
	responseData, err = c.m.cacher.PaginationCache(ctx, cacheKey, cacheTags, paginate, func(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
		return c.favoriteFolderService().GetFavoriteFoldersPaginate(ctx, pagination, userID)
	})
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetFavoriteFolderById get favorite folder by ID
//
//	@Summary		Get favorite folder by ID
//	@Description	Get favorite folder details with all properties
//	@Tags			Favorite Folder Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"favorite folder id"
//	@Success		200	{object}	responses.GetFavoriteFolderByIdResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/favorite-folders/{id} [get]
func (c Controller) GetFavoriteFolderById(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetFavoriteFolderByIdController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetFavoriteFolderById"), attribute.Int("id", id)))
		responseData *responses.GetFavoriteFolderByIdResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Convert user ID to string
	userID := strconv.Itoa(authUser.ID)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"favorite_folders"}, "GetFavoriteFolderById", id, userID)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, []interface{}{id, userID}, func(ctx context.Context, params []interface{}) (*responses.GetFavoriteFolderByIdResponse, error) {
		return c.favoriteFolderService().GetFavoriteFolderById(ctx, params[0].(int), params[1].(string))
	})
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// AddPropertyToFavorite adds a property to a favorite folder
//
//	@Summary		Add property to favorite folder
//	@Description	Add a property to favorite folder. Supports: create empty folder (folderName only), create folder with property, add property to existing folder (folderId + propertyId), or auto-name folder with property (propertyId only)
//	@Tags			Favorite Folder Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dtos.AddPropertyToFavorite	true	"Add property to favorite request"
//	@Success		200		{object}	responses.AddPropertyToFavoriteResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/favorite-folders [post]
func (c Controller) AddPropertyToFavorite(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "AddPropertyToFavoriteController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AddPropertyToFavorite")))
		responseData *responses.AddPropertyToFavoriteResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Convert user ID to string
	userID := strconv.Itoa(authUser.ID)

	// Parse request body
	dto := new(dtos.AddPropertyToFavorite)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Validate request
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service
	responseData, err = c.favoriteFolderService().AddPropertyToFavorite(ctx, dto, userID)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		if err == exception.ErrPropertyAlreadyInFolder {
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.PropertyAlreadyInFolderResponseError, err)
		}
		if err == exception.ErrPropertyAlreadyInSameFolder {
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.PropertyAlreadyInSameFolderResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear cache
	c.m.cacher.Tag("favorite_folders").Flush(ctx)
	c.m.cacher.Tag("properties").Flush(ctx)


	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// UpdateFolderName updates the name of a favorite folder
//
//	@Summary		Update folder name
//	@Description	Update the name of a favorite folder
//	@Tags			Favorite Folder Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"favorite folder id"
//	@Param			body	body		dtos.UpdateFolderName	true	"Update folder name request"
//	@Success		200		{object}	responses.UpdateFolderNameResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/favorite-folders/{id} [put]
func (c Controller) UpdateFolderName(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "UpdateFolderNameController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateFolderName"), attribute.Int("id", id)))
		responseData *responses.UpdateFolderNameResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Convert user ID to string
	userID := strconv.Itoa(authUser.ID)

	// Parse request body
	dto := new(dtos.UpdateFolderName)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Validate request
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service
	responseData, err = c.favoriteFolderService().UpdateFolderName(ctx, id, dto.Name, userID)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear cache
	c.m.cacher.Tag("favorite_folders").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// DeleteFolder soft deletes a favorite folder
//
//	@Summary		Delete folder
//	@Description	Soft delete a favorite folder
//	@Tags			Favorite Folder Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"favorite folder id"
//	@Success		200	{object}	responses.DeleteFolderResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/favorite-folders/{id} [delete]
func (c Controller) DeleteFolder(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "DeleteFolderController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteFolder"), attribute.Int("id", id)))
		responseData *responses.DeleteFolderResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Convert user ID to string
	userID := strconv.Itoa(authUser.ID)

	// Call service
	responseData, err = c.favoriteFolderService().DeleteFolder(ctx, id, userID)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear cache
	c.m.cacher.Tag("favorite_folders").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// RemovePropertyFromAllFolders removes a property from all favorite folders of the user
//
//	@Summary		Remove property from all folders
//	@Description	Remove a property from all favorite folders of the authenticated user
//	@Tags			Favorite Folder Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			propertyId	path		int	true	"property id"
//	@Success		200			{object}	responses.RemovePropertyFromAllFoldersResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/favorite-properties/{propertyId} [delete]
func (c Controller) RemovePropertyFromAllFolders(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("propertyId")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "RemovePropertyFromAllFoldersController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "RemovePropertyFromAllFolders"), attribute.Int("propertyId", propertyID)))
		responseData  *responses.RemovePropertyFromAllFoldersResponse
		err           error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Convert user ID to string
	userID := strconv.Itoa(authUser.ID)

	// Call service
	responseData, err = c.favoriteFolderService().RemovePropertyFromAllFolders(ctx, propertyID, userID)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear cache
	c.m.cacher.Tag("favorite_folders").Flush(ctx)
	c.m.cacher.Tag("properties").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}


