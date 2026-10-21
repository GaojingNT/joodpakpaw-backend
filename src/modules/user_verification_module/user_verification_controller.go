package user_verification_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/http_response"
	"joodpakpaw/internal/validator"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/user_verification_module/dtos"
)

// GetVerifications lists all verifications with pagination
//
//	@Summary		List verifications
//	@Description	Get all user verifications (admin)
//	@Tags			User Verification Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int		false	"page number"
//	@Param			limit	query		int		false	"page size"
//	@Param			status	query		string	false	"filter by status: pending, approved, rejected"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/user-verifications [get]
func (c Controller) GetVerifications(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetVerificationsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetVerifications")))
		responseData *paginator.Pagination
		err          error
	)

	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)
	queryStatus := f.Query("status")

	pagination := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
		paginator.WithAttributes("status", queryStatus),
	)

	responseData, err = c.verificationService().GetVerifications(ctx, pagination)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetVerificationByID gets a verification by ID
//
//	@Summary		Get verification by ID
//	@Description	Get verification detail (admin)
//	@Tags			User Verification Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"verification id"
//	@Success		200	{object}	responses.GetVerificationByIDResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/user-verifications/{id} [get]
func (c Controller) GetVerificationByID(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetVerificationByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetVerificationByID"), attribute.Int("id", id)))
		err       error
	)

	responseData, err := c.verificationService().GetVerificationByID(ctx, id)
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

// GetMyVerification gets the current user's latest verification
//
//	@Summary		Get my verification
//	@Description	Get current user's latest verification status
//	@Tags			User Verification Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.GetVerificationByIDResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/user-verifications/me [get]
func (c Controller) GetMyVerification(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetMyVerificationController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetMyVerification")))
		err       error
	)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	responseData, err := c.verificationService().GetMyVerification(ctx, authUser.ID)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			c.m.tracer.TraceEnd(span)
			return f.Status(fiber.StatusOK).JSON(fiber.Map{
				"code":         "OK",
				"message":      "Success",
				"verification": nil,
			})
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// SubmitVerification submits a new verification request
//
//	@Summary		Submit verification
//	@Description	Submit identity verification with ID photo and selfie
//	@Tags			User Verification Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			idPhoto		formData	file	true	"ID card photo (jpg, jpeg, png, max 10MB)"
//	@Param			selfieWithId	formData	file	true	"Selfie with ID card (jpg, jpeg, png, max 10MB)"
//	@Success		201	{object}	http_response.OkResponse
//	@Failure		400	{object}	exception.ErrorResponse
//	@Failure		409	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/user-verifications [post]
func (c Controller) SubmitVerification(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "SubmitVerificationController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "SubmitVerification")))
		err       error
	)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	dto := new(dtos.SubmitVerification)

	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	maxSize := int64(10 * 1024 * 1024) // 10MB

	// Validate idPhoto
	idPhotoFile, err := f.FormFile("idPhoto")
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}
	if idPhotoFile.Size > maxSize {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.VerificationFileSizeTooLargeResponseError, exception.ErrVerificationFileSizeTooLarge)
	}
	if !allowedExts[toLower(getFileExt(idPhotoFile.Filename))] {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidFileFormatResponseError, exception.ErrInvalidFileFormat)
	}
	dto.IdPhotoFile = idPhotoFile

	// Validate selfieWithId
	selfieFile, err := f.FormFile("selfieWithId")
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}
	if selfieFile.Size > maxSize {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.VerificationFileSizeTooLargeResponseError, exception.ErrVerificationFileSizeTooLarge)
	}
	if !allowedExts[toLower(getFileExt(selfieFile.Filename))] {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidFileFormatResponseError, exception.ErrInvalidFileFormat)
	}
	dto.SelfieWithIdFile = selfieFile

	err = c.verificationService().SubmitVerification(ctx, authUser.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrVerificationAlreadyPending:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.VerificationAlreadyPendingResponseError, err)
		case exception.ErrVerificationAlreadyApproved:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.VerificationAlreadyApprovedResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "CREATED", "Verification submitted successfully")
}

// ApproveVerification approves a pending verification
//
//	@Summary		Approve verification
//	@Description	Approve a pending user verification (admin)
//	@Tags			User Verification Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"verification id"
//	@Success		200	{object}	http_response.OkResponse
//	@Failure		400	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/user-verifications/{id}/approve [patch]
func (c Controller) ApproveVerification(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "ApproveVerificationController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ApproveVerification"), attribute.Int("id", id)))
		err       error
	)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	err = c.verificationService().ApproveVerification(ctx, id, authUser.ID)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrVerificationNotPending:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.VerificationNotPendingResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Flush verification cache
	c.m.cacher.Tag("user_verifications").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Verification approved successfully")
}

// RejectVerification rejects a pending verification
//
//	@Summary		Reject verification
//	@Description	Reject a pending user verification with a reason (admin)
//	@Tags			User Verification Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id					path		int		true	"verification id"
//	@Param			rejectDescription	body		string	true	"reason for rejection"
//	@Success		200	{object}	http_response.OkResponse
//	@Failure		400	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/user-verifications/{id}/reject [patch]
func (c Controller) RejectVerification(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "RejectVerificationController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "RejectVerification"), attribute.Int("id", id)))
		err       error
	)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	dto := new(dtos.RejectVerification)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	err = c.verificationService().RejectVerification(ctx, id, dto, authUser.ID)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrVerificationNotPending:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.VerificationNotPendingResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Flush verification cache
	c.m.cacher.Tag("user_verifications").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Verification rejected successfully")
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := range s {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}
