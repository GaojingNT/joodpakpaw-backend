package admin_dashboard_module

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

func (r Repository) CountPendingReports(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountPendingReportsRepository", trace.WithAttributes(attribute.String("repository", "CountPendingReports")))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.Report{}).Where("status = ?", models.ReportStatusPending).Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountPendingReportsThisMonth(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountPendingReportsThisMonthRepository", trace.WithAttributes(attribute.String("repository", "CountPendingReportsThisMonth")))
		count        int64
		err          error
	)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.Report{}).
				Where("status = ?", models.ReportStatusPending).
				Where("created_at >= ?", startOfMonth).
				Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountTotalUsers(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountTotalUsersRepository", trace.WithAttributes(attribute.String("repository", "CountTotalUsers")))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.User{}).Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountNewUsersThisMonth(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountNewUsersThisMonthRepository", trace.WithAttributes(attribute.String("repository", "CountNewUsersThisMonth")))
		count        int64
		err          error
	)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.User{}).
				Where("created_at >= ?", startOfMonth).
				Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountActiveProperties(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountActivePropertiesRepository", trace.WithAttributes(attribute.String("repository", "CountActiveProperties")))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.Property{}).Where("is_active = ?", true).Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountActivePropertiesPublishedThisMonth(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountActivePropertiesPublishedThisMonthRepository", trace.WithAttributes(attribute.String("repository", "CountActivePropertiesPublishedThisMonth")))
		count        int64
		err          error
	)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.Property{}).
				Where("is_active = ?", true).
				Where("published_at >= ?", startOfMonth).
				Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountBannedAllProperties(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountBannedAllPropertiesRepository", trace.WithAttributes(attribute.String("repository", "CountBannedAllProperties")))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.User{}).Where("is_banned_all_properties = ?", true).Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountBannedAllPropertiesThisMonth(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountBannedAllPropertiesThisMonthRepository", trace.WithAttributes(attribute.String("repository", "CountBannedAllPropertiesThisMonth")))
		count        int64
		err          error
	)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.User{}).
				Where("is_banned_all_properties = ?", true).
				Where("banned_at >= ?", startOfMonth).
				Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountPendingVerifications(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountPendingVerificationsRepository", trace.WithAttributes(attribute.String("repository", "CountPendingVerifications")))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.UserVerification{}).Where("status = ?", models.VerificationStatusPending).Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}

func (r Repository) CountPendingVerificationsSinceMonth(ctx context.Context) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountPendingVerificationsSinceMonthRepository", trace.WithAttributes(attribute.String("repository", "CountPendingVerificationsSinceMonth")))
		count        int64
		err          error
	)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.UserVerification{}).
				Where("status = ?", models.VerificationStatusPending).
				Where("created_at >= ?", startOfMonth).
				Count(&count).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)
	return count, err
}
