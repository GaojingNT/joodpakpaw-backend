package report_category_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/modules/report_category_module/responses"
)

// GetAllReportCategories lists all report categories
//
//	@Summary		List all report categories
//	@Description	Get all report categories, optionally filtered by type (property, review)
//	@Tags			Report Category Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			type	query		string	false	"filter by type: property, review"
//	@Success		200		{object}	responses.GetReportCategoriesResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/report-categories [get]
func (c Controller) GetAllReportCategories(f *fiber.Ctx) error {
	var (
		ctx, span  = c.m.tracer.TraceStart(f.Context(), "GetAllReportCategoriesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllReportCategories")))
		reportType = f.Query("type")
	)

	categories, err := c.reportCategoryService().GetAllReportCategories(ctx, reportType)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	responseData := &responses.GetReportCategoriesResponse{
		ReportCategories: new(responses.GetReportCategoriesResponse).Collection(categories),
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
