package user_recent_view_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/src/models"
)

func (r Repository) CreateUserRecentView(ctx context.Context, userID int, propertyID int) (*models.UserRecentView, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateUserRecentViewRepository", trace.WithAttributes(attribute.String("repository", "CreateUserRecentView")))
		record       = &models.UserRecentView{
			UserID:     userID,
			PropertyID: propertyID,
			ViewedAt:   time.Now(),
		}
		err error
	)
	defer r.tracer.TraceEnd(childSpan)

	utils.Block{
		Try: func() {
			if err = r.db.Create(record).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
	}.Do()

	return record, err
}
