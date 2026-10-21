package pet_breed_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
)

func (r Repository) GetPetBreedsPaginate(ctx context.Context, pagination *paginator.Pagination, petId int) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPetBreedsPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetPetBreedsPaginate"), attribute.Int("petId", petId)))
		petBreeds    = make([]models.PetBreed, 0)
		err          error
	)

	// Create pagination query
	tx := r.db.Where("pet_id = ?", petId)

	utils.Block{
		Try: func() {
			// Execute query with pagination
			if err = tx.
				Preload("Pet").
				Scopes(paginator.Paginate(petBreeds, pagination, r.db.Where("pet_id = ?", petId))).
				Find(&petBreeds).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	// Set data
	pagination.Data = petBreeds

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) GetAllPetBreeds(ctx context.Context, petId int) ([]models.PetBreed, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllPetBreedsRepository", trace.WithAttributes(attribute.String("repository", "GetAllPetBreeds"), attribute.Int("petId", petId)))
		petBreeds    []models.PetBreed
		err          error
	)

	utils.Block{
		Try: func() {
			// Get all pet breeds by pet ID
			if err = r.db.Where("pet_id = ?", petId).Preload("Pet").Find(&petBreeds).Error; err != nil {
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

	return petBreeds, nil
}

func (r Repository) CreatePetBreed(ctx context.Context, petBreed *models.PetBreed) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreatePetBreedRepository", trace.WithAttributes(attribute.String("repository", "CreatePetBreed")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Create(petBreed).Error; err != nil {
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

	if err != nil {
		return err
	}

	return nil
}

func (r Repository) UpdatePetBreed(ctx context.Context, breedId int, petBreed *models.PetBreed) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdatePetBreedRepository", trace.WithAttributes(attribute.String("repository", "UpdatePetBreed"), attribute.Int64("breedId", int64(breedId))))
		err          error
	)

	utils.Block{
		Try: func() {
			updateMap := make(map[string]interface{})

			if petBreed.Breed != "" {
				updateMap["breed"] = petBreed.Breed
			}
			if petBreed.ImageURL != nil {
				updateMap["image_url"] = petBreed.ImageURL
			}

			if len(updateMap) > 0 {
				if err = r.db.Model(&models.PetBreed{}).Where("id = ?", breedId).Updates(updateMap).Error; err != nil {
					utils.Throw(err)
				}
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
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return err
	}

	return nil
}

func (r Repository) GetPetBreedByID(ctx context.Context, petId int, breedId int) (models.PetBreed, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPetBreedByIDRepository", trace.WithAttributes(attribute.String("repository", "GetPetBreedByID"), attribute.Int("petId", petId), attribute.Int("breedId", breedId)))
		petBreed     models.PetBreed
		err          error
	)

	utils.Block{
		Try: func() {
			// Get pet breed by ID with pet preloaded
			if err = r.db.Where("pet_id = ?", petId).Preload("Pet").First(&petBreed, breedId).Error; err != nil {
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
		return petBreed, err
	}

	return petBreed, nil
}
