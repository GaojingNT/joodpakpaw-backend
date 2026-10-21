package pet_size_category_module

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

func (r Repository) GetAllPetSizeCategories(ctx context.Context) ([]models.PetSizeCategory, error) {
	var (
		_, childSpan        = r.tracer.TraceStart(ctx, "GetAllPetSizeCategoriesRepository", trace.WithAttributes(attribute.String("repository", "GetAllPetSizeCategories")))
		petSizeCategories   []models.PetSizeCategory
		err                 error
	)

	utils.Block{
		Try: func() {
			// Get all pet size categories
			if err = r.db.Find(&petSizeCategories).Error; err != nil {
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

	return petSizeCategories, nil
}
