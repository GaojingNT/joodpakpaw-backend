package pet_profile_module

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

// GetAllPetProfiles retrieves all pet profiles for a specific user
func (r Repository) GetAllPetProfiles(ctx context.Context, guestID int) ([]models.PetProfile, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllPetProfilesRepository", trace.WithAttributes(attribute.String("repository", "GetAllPetProfiles")))
		petProfiles  = make([]models.PetProfile, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query - filter by guest_id
			if err = r.db.WithContext(ctx).
				Where("guest_id = ?", guestID).
				Preload("PetType").
				Order("created_at DESC").
				Find(&petProfiles).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return petProfiles, nil
}

// GetPetProfileByID retrieves a single pet profile by ID with all associations (filtered by owner)
func (r Repository) GetPetProfileByID(ctx context.Context, id int, guestID int) (*models.PetProfile, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetPetProfileByIDRepository", trace.WithAttributes(attribute.String("repository", "GetPetProfileByID"), attribute.Int("petProfileId", id)))
		petProfile   models.PetProfile
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query - filter by both id and guest_id to ensure ownership
			if err = r.db.WithContext(ctx).
				Where("id = ? AND guest_id = ?", id, guestID).
				Preload("Guest").
				Preload("Breed").
				Preload("PetEnergy").
				Preload("PetSizeCategory").
				First(&petProfile).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found
			if err.Error() == "record not found" {
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

	return &petProfile, nil
}

// CreatePetProfile creates a new pet profile
func (r Repository) CreatePetProfile(ctx context.Context, petProfile *models.PetProfile) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreatePetProfileRepository", trace.WithAttributes(attribute.String("repository", "CreatePetProfile")))
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.WithContext(ctx).Create(&petProfile).Error; err != nil {
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

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

// GetPetTypeByID retrieves a pet type by ID (helper for validation)
func (r Repository) GetPetTypeByID(ctx context.Context, id int) (*models.Pet, error) {
	var petType models.Pet
	if err := r.db.WithContext(ctx).First(&petType, id).Error; err != nil {
		return nil, err
	}
	return &petType, nil
}

// GetBreedByID retrieves a breed by ID (helper for validation)
func (r Repository) GetBreedByID(ctx context.Context, id int) (*models.PetBreed, error) {
	var breed models.PetBreed
	if err := r.db.WithContext(ctx).First(&breed, id).Error; err != nil {
		return nil, err
	}
	return &breed, nil
}

// GetPetSizeCategoryByID retrieves a pet size category by ID (helper for validation)
func (r Repository) GetPetSizeCategoryByID(ctx context.Context, id int) (*models.PetSizeCategory, error) {
	var petSizeCategory models.PetSizeCategory
	if err := r.db.WithContext(ctx).First(&petSizeCategory, id).Error; err != nil {
		return nil, err
	}
	return &petSizeCategory, nil
}

// GetPetEnergyByID retrieves a pet energy by ID (helper for validation)
func (r Repository) GetPetEnergyByID(ctx context.Context, id int) (*models.PetProfileEnergy, error) {
	var petEnergy models.PetProfileEnergy
	if err := r.db.WithContext(ctx).First(&petEnergy, id).Error; err != nil {
		return nil, err
	}
	return &petEnergy, nil
}

// UpdatePetProfile updates an existing pet profile (filtered by owner)
func (r Repository) UpdatePetProfile(ctx context.Context, id int, guestID int, petProfile *models.PetProfile) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdatePetProfileRepository", trace.WithAttributes(attribute.String("repository", "UpdatePetProfile"), attribute.Int("petProfileId", id)))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Verify pet profile belongs to user
				var existing models.PetProfile
				if err = tx.Where("id = ? AND guest_id = ?", id, guestID).First(&existing).Error; err != nil {
					return err
				}

				// Build update map dynamically to only update non-empty fields
				updateMap := make(map[string]interface{})

				if petProfile.Name != "" {
					updateMap["name"] = petProfile.Name
				}
				if petProfile.Image != "" {
					updateMap["image"] = petProfile.Image
				}
				if petProfile.ContactName != "" {
					updateMap["contact_name"] = petProfile.ContactName
				}
				if petProfile.ContactPhone != "" {
					updateMap["contact_phone"] = petProfile.ContactPhone
				}
				if petProfile.Note != "" {
					updateMap["note"] = petProfile.Note
				}
				if petProfile.PetSizeCategoryID != nil {
					updateMap["pet_size_category_id"] = petProfile.PetSizeCategoryID
				}
				if petProfile.PetEnergyID != nil {
					updateMap["pet_energy_id"] = petProfile.PetEnergyID
				}

				// Execute query only if there are fields to update
				if len(updateMap) > 0 {
					if err = tx.
						Model(&existing).
						Updates(updateMap).Error; err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err.Error() == "record not found" {
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
		return err
	}

	return nil
}

// DeletePetProfile deletes a pet profile (soft delete, filtered by owner)
func (r Repository) DeletePetProfile(ctx context.Context, id int, guestID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeletePetProfileRepository", trace.WithAttributes(attribute.String("repository", "DeletePetProfile"), attribute.Int("petProfileId", id)))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Check if pet profile exists and belongs to user
				var petProfile models.PetProfile
				if err = tx.Where("id = ? AND guest_id = ?", id, guestID).First(&petProfile).Error; err != nil {
					return err
				}

				// Execute soft delete
				if err = tx.Delete(&petProfile).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err.Error() == "record not found" {
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
		return err
	}

	return nil
}
