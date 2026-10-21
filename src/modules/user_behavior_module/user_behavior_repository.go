package user_behavior_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/src/models"
)

func (r Repository) GetOrCreateBehaviorCategoryByName(ctx context.Context, name string) (*models.UserBehaviorCategory, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetOrCreateBehaviorCategoryByNameRepository", trace.WithAttributes(attribute.String("repository", "GetOrCreateBehaviorCategoryByName")))
		category     = models.UserBehaviorCategory{Name: name}
		err          error
	)
	defer r.tracer.TraceEnd(childSpan)

	utils.Block{
		Try: func() {
			if err = r.db.Where("name = ?", name).FirstOrCreate(&category).Error; err != nil {
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

	return &category, err
}

func (r Repository) GetAllUserBehaviors(ctx context.Context, userID int) ([]models.UserBehavior, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllUserBehaviorsRepository", trace.WithAttributes(attribute.String("repository", "GetAllUserBehaviors")))
		behaviors    = make([]models.UserBehavior, 0)
		err          error
	)
	defer r.tracer.TraceEnd(childSpan)

	utils.Block{
		Try: func() {
			if err = r.db.
				Where("user_id = ?", userID).
				Preload("BehaviorCategory").
				Order("id DESC").
				Find(&behaviors).Error; err != nil {
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

	return behaviors, err
}

func (r Repository) CreateUserBehavior(ctx context.Context, behavior *models.UserBehavior) (*models.UserBehavior, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateUserBehaviorRepository", trace.WithAttributes(attribute.String("repository", "CreateUserBehavior")))
		err          error
	)
	defer r.tracer.TraceEnd(childSpan)

	utils.Block{
		Try: func() {
			if err = r.db.Create(behavior).Error; err != nil {
				utils.Throw(err)
			}
			// Preload category for response
			if err = r.db.Preload("BehaviorCategory").First(behavior, behavior.ID).Error; err != nil {
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

	return behavior, err
}
