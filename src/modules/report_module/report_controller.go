package report_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/http_response"
	"joodpakpaw/internal/validator"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/report_module/dtos"
)

// GetReports lists all reports with pagination, filterable by type
//
//	@Summary		List reports
//	@Description	Get paginated reports, filter by type (review|property)
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int		false	"page number"
//	@Param			limit	query		int		false	"page size"
//	@Param			type	query		string	false	"filter by type: review or property"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/reports [get]
func (c Controller) GetReports(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetReportsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetReports")))
	)
	defer c.m.tracer.TraceEnd(span)

	queryType := f.Query("type")

	pagination := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 20)),
		paginator.WithOrderBy("tbl_reports.created_at"),
		paginator.WithSort("desc"),
		paginator.WithAttributes("type", queryType),
	)

	result, err := c.reportService().GetReports(ctx, pagination)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}

// ActionReport performs an action on a report (admin)
//
//	@Summary		Action report
//	@Description	Perform an action on a report. Property: dismiss_report|warn_user|deactivate_property. Review: dismiss_report|warn_user|delete_review
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"report id"
//	@Param			body	body		dtos.ActionReport	true	"action"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		409		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/reports/{id}/action [patch]
func (c Controller) ActionReport(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "ActionReportController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ActionReport"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	dto := new(dtos.ActionReport)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}
	if errs := validator.Validate(*dto); errs != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errs...)
	}

	err := c.reportService().ActionReport(ctx, id, dto.Action, dto.AdminNote, authUser.ID)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrReportActionNotAllowed:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.ReportActionNotAllowedResponseError, err)
		case exception.ErrReportAlreadyActioned:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.ReportAlreadyActionedResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	return http_response.HttpOkResponse(f, "OK", "Report actioned successfully")
}

// GetMyReports lists paginated reports of the authenticated user
//
//	@Summary		Get my reports
//	@Description	Get paginated reports created by the current user, filter by type (review|property)
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int		false	"page number"
//	@Param			limit	query		int		false	"page size"
//	@Param			type	query		string	false	"filter by type: review or property"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/reports/me [get]
func (c Controller) GetMyReports(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetMyReportsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetMyReports")))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	pagination := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 20)),
		paginator.WithOrderBy("tbl_reports.created_at"),
		paginator.WithSort("desc"),
		paginator.WithAttributes("type", f.Query("type")),
	)

	result, err := c.reportService().GetMyReports(ctx, authUser.ID, pagination)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}

// GetMyReportByID gets a single report by ID belonging to the authenticated user
//
//	@Summary		Get my report by ID
//	@Description	Get detail of a report created by the current user
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"report id"
//	@Success		200	{object}	responses.GetReportByIDResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/reports/me/{id} [get]
func (c Controller) GetMyReportByID(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetMyReportByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetMyReportByID"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	result, err := c.reportService().GetMyReportByID(ctx, id, authUser.ID)
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

// CreateReportReview creates a report on a review (booking)
//
//	@Summary		Create report review
//	@Description	Report a review (booking), reportCategoryId must be type=review
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dtos.CreateReportReview		true	"Report review data"
//	@Success		201		{object}	responses.CreateReportResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/reports/review [post]
func (c Controller) CreateReportReview(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateReportReviewController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateReportReview")))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	dto := new(dtos.CreateReportReview)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}
	if form, err := f.MultipartForm(); err == nil && form != nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	if errs := validator.Validate(*dto); errs != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errs...)
	}

	result, err := c.reportService().CreateReportReview(ctx, authUser.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrInvalidRequestParameter:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		case exception.ErrCannotReportOwnReview:
			return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.CannotReportOwnReviewResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	return f.Status(fiber.StatusCreated).JSON(result)
}

// CreateReportProperty creates a report on a property
//
//	@Summary		Create report property
//	@Description	Report a property, reportCategoryId must be type=property
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dtos.CreateReportProperty	true	"Report property data"
//	@Success		201		{object}	responses.CreateReportResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/reports/property [post]
func (c Controller) CreateReportProperty(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateReportPropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateReportProperty")))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	dto := new(dtos.CreateReportProperty)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}
	if form, err := f.MultipartForm(); err == nil && form != nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	if errs := validator.Validate(*dto); errs != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errs...)
	}

	result, err := c.reportService().CreateReportProperty(ctx, authUser.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrInvalidRequestParameter:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		case exception.ErrCannotReportOwnProperty:
			return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.CannotReportOwnPropertyResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	return f.Status(fiber.StatusCreated).JSON(result)
}

// CreateReportUnbanProperty creates a request to unban a property (or all properties of a host)
//
//	@Summary		Create report unban property
//	@Description	Request to unban a property, reportCategoryId must be type=unban_property. propertyId is optional: set = unban single property, omit = unban all properties of host
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dtos.CreateReportUnbanProperty	true	"Unban property request data"
//	@Success		201		{object}	responses.CreateReportResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/reports/unban-property [post]
func (c Controller) CreateReportUnbanProperty(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateReportUnbanPropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateReportUnbanProperty")))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	dto := new(dtos.CreateReportUnbanProperty)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}
	if form, err := f.MultipartForm(); err == nil && form != nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	if errs := validator.Validate(*dto); errs != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errs...)
	}

	result, err := c.reportService().CreateReportUnbanProperty(ctx, authUser.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrInvalidRequestParameter:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		case exception.ErrUnbanRequestAlreadyPending:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.UnbanRequestAlreadyPendingResponseError, err)
		case exception.ErrUnbanRequestLimitExceeded:
			return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.UnbanRequestLimitExceededResponseError, err)
		case exception.ErrMustUnbanAllProperties:
			return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.MustUnbanAllPropertiesResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	return f.Status(fiber.StatusCreated).JSON(result)
}

// CreateReportUnbanReview creates a request to unban a user from writing reviews
//
//	@Summary		Create report unban review
//	@Description	Request to unban a user from writing reviews, reportCategoryId must be type=unban_review
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dtos.CreateReportUnbanReview	true	"Unban review request data"
//	@Success		201		{object}	responses.CreateReportResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/reports/unban-review [post]
func (c Controller) CreateReportUnbanReview(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateReportUnbanReviewController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateReportUnbanReview")))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	dto := new(dtos.CreateReportUnbanReview)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}
	if form, err := f.MultipartForm(); err == nil && form != nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	if errs := validator.Validate(*dto); errs != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errs...)
	}

	result, err := c.reportService().CreateReportUnbanReview(ctx, authUser.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrInvalidRequestParameter:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		case exception.ErrUnbanRequestAlreadyPending:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.UnbanRequestAlreadyPendingResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	return f.Status(fiber.StatusCreated).JSON(result)
}

// GetReportByID gets a single report by ID with full detail
//
//	@Summary		Get report by ID
//	@Description	Get report detail by ID, response fields vary by type (property|review)
//	@Tags			Report Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"report id"
//	@Success		200	{object}	responses.GetReportByIDResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/admin/reports/{id} [get]
func (c Controller) GetReportByID(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetReportByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetReportByID"), attribute.Int("id", id)))
	)
	defer c.m.tracer.TraceEnd(span)

	result, err := c.reportService().GetReportByID(ctx, id)
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
