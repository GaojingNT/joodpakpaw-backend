package pet_module

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

func (r Repository) GetAllPets(ctx context.Context) ([]models.Pet, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllPetsRepository", trace.WithAttributes(attribute.String("repository", "GetAllPets")))
		pets         []models.Pet
		err          error
	)

	utils.Block{
		Try: func() {
			// Get all pets
			if err = r.db.Find(&pets).Error; err != nil {
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

	return pets, nil
}
