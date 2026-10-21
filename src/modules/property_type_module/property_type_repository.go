package property_type_module

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

func (r Repository) GetAllPropertyTypes(ctx context.Context) ([]models.PropertyType, error) {
	var (
		_, childSpan  = r.tracer.TraceStart(ctx, "GetAllPropertyTypesRepository", trace.WithAttributes(attribute.String("repository", "GetAllPropertyTypes")))
		propertyTypes []models.PropertyType
		err           error
	)

	utils.Block{
		Try: func() {
			// Get all property types
			if err = r.db.Find(&propertyTypes).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return propertyTypes, nil
}

func (r Repository) GetPropertyTypeByID(ctx context.Context, id int) (models.PropertyType, error) {
	var (
		_, childSpan   = r.tracer.TraceStart(ctx, "GetPropertyTypeByIDRepository", trace.WithAttributes(attribute.String("repository", "GetPropertyTypeByID"), attribute.Int64("id", int64(id))))
		propertyType models.PropertyType
		err            error
	)

	utils.Block{
		Try: func() {
			// Get property type by ID
			if err = r.db.First(&propertyType, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return propertyType, err
	}

	return propertyType, nil
}

