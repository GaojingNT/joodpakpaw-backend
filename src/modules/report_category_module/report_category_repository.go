package report_category_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/src/models"
)

func (r Repository) GetAllReportCategories(ctx context.Context, reportType string) ([]models.ReportCategory, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllReportCategoriesRepository", trace.WithAttributes(attribute.String("repository", "GetAllReportCategories")))
		categories   []models.ReportCategory
		err          error
	)

	utils.Block{
		Try: func() {
			tx := r.db
			if reportType != "" {
				tx = tx.Where("type = ?", reportType)
			}
			if err = tx.Find(&categories).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}
	return categories, nil
}
