package user_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/http_response"
	"joodpakpaw/internal/validator"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/user_module/dtos"
	"joodpakpaw/src/modules/user_module/responses"
)

// GetUsers lists all existing users
//
//	@Summary		List users
//	@Description	get users
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page		query		string	false	"page number"
//	@Param			limit		query		string	false	"page size"
//	@Param			search		query		string	false	"search query"
//	@Param			searchBy	query		string	false	"field to search by"
//	@Success		200			{object}	paginator.Pagination
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/users [get]
func (c Controller) GetUsers(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetUsersController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetUsers")))
		responseData *paginator.Pagination
		err          error
	)

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
		paginator.WithAttributes("searchable", models.UserSearchable()),
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"users"}, "GetUsers", paginate.Page, paginate.Limit, paginate.Attributes)

	// Get data from cache
	responseData, err = c.m.cacher.PaginationCache(ctx, cacheKey, cacheTags, paginate, c.userService().GetUsers)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// AdminGetUsers lists all users for admin
//
//	@Summary		Admin list users
//	@Description	Get paginated users with roles and ban status (admin)
//	@Tags			User Module (Version 1)
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
//	@Router			/api/v1/admin/users/paginate [get]
func (c Controller) AdminGetUsers(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AdminGetUsersController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AdminGetUsers")))
	)
	defer c.m.tracer.TraceEnd(span)

	paginate := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 20)),
		paginator.WithAttributes("search", f.Query("search")),
		paginator.WithAttributes("search_by", f.Query("searchBy")),
		paginator.WithAttributes("searchable", models.UserSearchable()),
	)

	result, err := c.userService().AdminGetUsers(ctx, paginate)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}

// AdminGetUserByID gets a user by ID for admin
//
//	@Summary		Admin get user by ID
//	@Description	Get user detail with roles and ban status (admin)
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"user id"
//	@Success		200	{object}	responses.AdminGetUserByIDResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/users/{id} [get]
func (c Controller) AdminGetUserByID(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AdminGetUserByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AdminGetUserByID"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	result, err := c.userService().AdminGetUserByID(ctx, id)
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

// AdminDeactivateAllProperties deactivates all properties of a user (admin)
//
//	@Summary		Admin deactivate all properties
//	@Description	Deactivate all properties of a user and cancel active bookings (admin)
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int						true	"user id"
//	@Param			adminNote	body		dtos.AdminActionUser	false	"admin note"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/users/{id}/deactivate-all-properties [patch]
func (c Controller) AdminDeactivateAllProperties(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AdminDeactivateAllPropertiesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AdminDeactivateAllProperties"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	dto := new(dtos.AdminActionUser)
	_ = f.BodyParser(dto)

	err := c.userService().AdminDeactivateAllProperties(ctx, id, authUser.ID, dto.AdminNote)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.cacher.Tag("users").Flush(ctx)
	return http_response.HttpOkResponse(f, "OK", "All properties deactivated successfully")
}

// AdminUnbanAllProperties unbans all properties of a user (admin)
//
//	@Summary		Admin unban all properties
//	@Description	Unban all properties of a user (admin)
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int						true	"user id"
//	@Param			adminNote	body		dtos.AdminActionUser	false	"admin note"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/users/{id}/unban-all-properties [patch]
func (c Controller) AdminUnbanAllProperties(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AdminUnbanAllPropertiesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AdminUnbanAllProperties"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	dto := new(dtos.AdminActionUser)
	_ = f.BodyParser(dto)

	err := c.userService().AdminUnbanAllProperties(ctx, id, authUser.ID, dto.AdminNote)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.cacher.Tag("users").Flush(ctx)
	return http_response.HttpOkResponse(f, "OK", "All properties unbanned successfully")
}

// GetUser get existing user by ID
//
//	@Summary		Get user
//	@Description	Get user
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		int	true	"user id"
//	@Success		200		{object}	responses.GetUserByIDResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/users/{user_id} [get]
func (c Controller) GetUser(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetUserController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetUser"), attribute.Int("id", id)))
		responseData *responses.GetUserByIDResponse
		err          error
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"users"}, "GetUser", id)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, id, c.userService().GetUser)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}



// UpdateUser update existing user
//
//	@Summary		Update user
//	@Description	Update an existing user
//	@Tags			User Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			user_id		path		int		true	"user id"
//	@Param			firstName	formData	string	false	"user first name"
//	@Param			lastName	formData	string	false	"user last name"
//	@Param			avatar		formData	file	false	"user avatar image (max 10MB, formats: jpg, jpeg, png, webp, gif)"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		400			{object}	exception.ErrorResponse	"Invalid file format, file size too large, or invalid characters in name"
//	@Failure		404			{object}	exception.ErrorResponse	"User not found"
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/users/{user_id} [put]
func (c Controller) UpdateUser(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "UpdateUserController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateUser"), attribute.Int("id", id)))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.UpdateUser)

	// Parse form data
	dto.FirstName = f.FormValue("firstName")
	dto.LastName = f.FormValue("lastName")

	// Validate name for invalid characters (allow only letters, spaces, hyphens, apostrophes)
	if dto.FirstName != "" {
		for _, char := range dto.FirstName {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= 'ก' && char <= 'ฮ') || char == ' ' || char == '-' || char == '\'') {
				return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidCharactersInNameResponseError, exception.ErrInvalidCharactersInName)
			}
		}
	}
	if dto.LastName != "" {
		for _, char := range dto.LastName {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= 'ก' && char <= 'ฮ') || char == ' ' || char == '-' || char == '\'') {
				return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidCharactersInNameResponseError, exception.ErrInvalidCharactersInName)
			}
		}
	}

	// Get avatar file if provided
	avatarFile, err := f.FormFile("avatar")
	if err == nil {
		// Avatar file is provided - validate file

		// 1. Check file size (max 10MB)
		maxSize := int64(10 * 1024 * 1024) // 10MB in bytes
		if avatarFile.Size > maxSize {
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
		filename := avatarFile.Filename
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

		dto.AvatarFile = avatarFile
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.userService().UpdateUser(ctx, id, dto)
	if err != nil {
		// Handle specific errors
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear user cache
	c.m.cacher.Tag("users").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "User updated successfully")
}

// DeleteUser remove existing user by ID
//
//	@Summary		Delete user
//	@Description	Delete user
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		int	true	"user id"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/users/{user_id} [delete]
func (c Controller) DeleteUser(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeleteUserController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteUser"), attribute.Int("id", id)))
		err       error
	)

	// Call service function
	err = c.userService().DeleteUser(ctx, id)
	if err != nil {
		// Handle specific errors
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear user cache
	c.m.cacher.Tag("users").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "User deleted successfully")
}



// AddRoleToUser assigns a role to a user
//
//	@Summary		Add role to user
//	@Description	Assign a role to a user by role name (admin)
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			user_id		path		int		true	"user id"
//	@Param			roleName	body		string	true	"role name"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		409			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/users/{user_id}/roles [patch]
func (c Controller) AddRoleToUser(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "AddRoleToUserController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "AddRoleToUser"), attribute.Int("id", id)))
		err       error
	)

	dto := new(dtos.AddRoleToUser)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	err = c.userService().AddRoleToUser(ctx, id, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrRoleAlreadyAssigned:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.RoleAlreadyAssignedResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.cacher.Tag("users").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Role assigned successfully")
}

// ChangePassword change user password
//
//	@Summary		Change Password
//	@Description	Change user password
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			user_id			path		int		true	"user id"
//	@Param			oldPassword		body		string	true	"old password"
//	@Param			newPassword		body		string	true	"new password"
//	@Param			confirmPassword	body		string	true	"confirm new password"
//	@Success		200				{object}	http_response.OkResponse
//	@Failure		400				{object}	exception.ErrorResponse	"Password not match or new password same as old"
//	@Failure		404				{object}	exception.ErrorResponse	"User not found"
//	@Failure		500				{object}	exception.ErrorResponse
//	@Router			/api/v1/users/{user_id}/change-password [patch]
func (c Controller) ChangePassword(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "ChangePasswordController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ChangePassword"), attribute.Int("id", id)))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.PatchChangePassword)

	// Parse HTTP request body to struct variable
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.userService().ChangePassword(ctx, id, dto)
	if err != nil {
		// Handle specific errors
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrPasswordNotMatch:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.PasswordNotMatchResponseError, err)
		case exception.ErrNewPasswordSameAsOld:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.NewPasswordSameAsOldResponseError, err)
		case exception.ErrInvalidRequestParameter:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear user cache
	c.m.cacher.Tag("users").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Password changed successfully")
}

