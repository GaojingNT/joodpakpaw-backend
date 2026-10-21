package report_category_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/models"
)

func (s Service) GetAllReportCategories(ctx context.Context, reportType string) ([]models.ReportCategory, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllReportCategoriesService", trace.WithAttributes(attribute.String("service", "GetAllReportCategories")))

	categories, err := s.reportCategoryRepository().GetAllReportCategories(ctx, reportType)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)
	return categories, nil
}
