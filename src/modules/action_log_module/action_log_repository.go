package action_log_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
)

func (r Repository) CreateActionLog(ctx context.Context, log *models.ActionLog) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateActionLogRepository", trace.WithAttributes(attribute.String("repository", "CreateActionLog")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Create(log).Error; err != nil {
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
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)
	return err
}

func (r Repository) GetActionLogsPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetActionLogsPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetActionLogsPaginate")))
		logs         = make([]models.ActionLog, 0)
		err          error
	)

	tx := r.db.Model(&models.ActionLog{})

	utils.Block{
		Try: func() {
			if err = tx.
				Preload("Actor").
				Preload("Actor.Roles").
				Preload("Target").
				Preload("Target.Roles").
				Order("tbl_action_logs.created_at DESC").
				Scopes(paginator.Paginate(logs, pagination, tx)).
				Find(&logs).Error; err != nil {
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
		Finally: nil,
	}.Do()

	pagination.Data = logs
	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return pagination, nil
}
