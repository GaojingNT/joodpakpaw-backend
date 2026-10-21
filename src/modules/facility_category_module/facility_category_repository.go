package facility_category_module

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

func (r Repository) GetAllFacilityCategories(ctx context.Context) ([]models.FacilityCategory, error) {
	var (
		_, childSpan         = r.tracer.TraceStart(ctx, "GetAllFacilityCategoriesRepository", trace.WithAttributes(attribute.String("repository", "GetAllFacilityCategories")))
		facilityCategories []models.FacilityCategory
		err                  error
	)

	utils.Block{
		Try: func() {
			// Get all facility categories with their facilities (only id and name)
			if err = r.db.
				Preload("Facilities", func(db *gorm.DB) *gorm.DB {
					return db.Select("id", "name", "category_id")
				}).
				Find(&facilityCategories).Error; err != nil {
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

	return facilityCategories, nil
}

func (r Repository) GetFacilityCategoryByID(ctx context.Context, id int) (models.FacilityCategory, error) {
	var (
		_, childSpan       = r.tracer.TraceStart(ctx, "GetFacilityCategoryByIDRepository", trace.WithAttributes(attribute.String("repository", "GetFacilityCategoryByID"), attribute.Int64("id", int64(id))))
		facilityCategory models.FacilityCategory
		err                error
	)

	utils.Block{
		Try: func() {
			// Get facility category by ID with its facilities (only id and name)
			if err = r.db.
				Preload("Facilities", func(db *gorm.DB) *gorm.DB {
					return db.Select("id", "name", "category_id")
				}).
				First(&facilityCategory, id).Error; err != nil {
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
		return facilityCategory, err
	}

	return facilityCategory, nil
}
