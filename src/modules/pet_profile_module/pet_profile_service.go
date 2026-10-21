package pet_profile_module

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/pet_profile_module/dtos"
	"joodpakpaw/src/modules/pet_profile_module/responses"
)

// GetAllPetProfiles retrieves all pet profiles for a specific user
func (s Service) GetAllPetProfiles(ctx context.Context, guestID int) (*responses.GetPetProfilesResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetAllPetProfilesService", trace.WithAttributes(attribute.String("service", "GetAllPetProfiles")))
	defer s.tracer.TraceEnd(span)

	petProfiles, err := s.petProfileRepository().GetAllPetProfiles(ctx, guestID)
	if err != nil {
		return nil, err
	}

	response := new(responses.GetPetProfilesResponse).Make(petProfiles)
	return response, nil
}

// GetPetProfileByID retrieves a pet profile by ID (filtered by owner)
func (s Service) GetPetProfileByID(ctx context.Context, id int, guestID int) (*responses.GetPetProfileByIDResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetPetProfileByIDService", trace.WithAttributes(attribute.String("service", "GetPetProfileByID")))
	defer s.tracer.TraceEnd(span)

	petProfile, err := s.petProfileRepository().GetPetProfileByID(ctx, id, guestID)
	if err != nil {
		return nil, err
	}

	response := new(responses.GetPetProfileByIDResponse).Make(*petProfile)
	return response, nil
}

// CreatePetProfile creates a new pet profile
func (s Service) CreatePetProfile(ctx context.Context, dto *dtos.CreatePetProfile) (*responses.CreatePetProfileResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "CreatePetProfileService", trace.WithAttributes(attribute.String("service", "CreatePetProfile")))
	defer s.tracer.TraceEnd(span)

	// Validate that all foreign key entities exist
	// Validate PetType exists
	_, err := s.petProfileRepository().GetPetTypeByID(ctx, *dto.PetTypeID)
	if err != nil {
		return nil, exception.ErrRecordNotFound
	}

	// Validate Breed exists
	_, err = s.petProfileRepository().GetBreedByID(ctx, *dto.BreedID)
	if err != nil {
		return nil, exception.ErrRecordNotFound
	}

	// Validate PetSizeCategory exists
	_, err = s.petProfileRepository().GetPetSizeCategoryByID(ctx, *dto.PetSizeCategoryID)
	if err != nil {
		return nil, exception.ErrRecordNotFound
	}

	// Validate PetEnergy exists
	_, err = s.petProfileRepository().GetPetEnergyByID(ctx, *dto.PetEnergyID)
	if err != nil {
		return nil, exception.ErrRecordNotFound
	}

	// Parse birthdate
	birthDate, err := time.Parse("2006-01-02", dto.BirthDate)
	if err != nil {
		return nil, exception.ErrInvalidDateFormat
	}

	// Validate birthdate is not in the future
	if birthDate.After(time.Now()) {
		return nil, exception.ErrFutureBirthDate
	}

	// Handle image file upload if provided
	imageURL := dto.Image
	if dto.ImageFile != nil {
		// Generate unique filename
		fileExt := ""
		if len(dto.ImageFile.Filename) > 0 {
			for i := len(dto.ImageFile.Filename) - 1; i >= 0; i-- {
				if dto.ImageFile.Filename[i] == '.' {
					fileExt = dto.ImageFile.Filename[i:]
					break
				}
			}
		}

		// Generate filename: pet_profile_{guestId}_{timestamp}{extension}
		fileName := fmt.Sprintf("pet_profile_%d_%d%s", *dto.GuestID, time.Now().Unix(), fileExt)

		// Define the path in storage
		storagePath := "pet-profiles"

		// Upload to MinIO/storage
		err := s.fileSystem.Put(ctx, storagePath, fileName, dto.ImageFile)
		if err != nil {
			s.logger.Error("Failed to upload pet profile image to storage", zap.Error(err))
			return nil, err
		}

		// Get the public URL for the uploaded file
		imageURL = s.fileSystem.GetPublicURL(storagePath, fileName)
	}

	// Create pet profile model from DTO
	petProfile := &models.PetProfile{
		Name:              dto.Name,
		Image:             imageURL,
		Sex:               dto.Sex,
		BirthDate:         birthDate,
		ContactName:       dto.ContactName,
		ContactPhone:      dto.ContactPhone,
		Note:              dto.Note,
		GuestID:           dto.GuestID,
		PetTypeID:         dto.PetTypeID,
		BreedID:           dto.BreedID,
		PetSizeCategoryID: dto.PetSizeCategoryID,
		PetEnergyID:       dto.PetEnergyID,
	}

	// Create pet profile in database
	err = s.petProfileRepository().CreatePetProfile(ctx, petProfile)
	if err != nil {
		return nil, err
	}

	// Build response
	response := new(responses.CreatePetProfileResponse).Make(*petProfile)

	return response, nil
}

// UpdatePetProfile updates an existing pet profile (filtered by owner)
func (s Service) UpdatePetProfile(ctx context.Context, id int, guestID int, dto *dtos.UpdatePetProfile) error {
	ctx, span := s.tracer.TraceStart(ctx, "UpdatePetProfileService", trace.WithAttributes(attribute.String("service", "UpdatePetProfile")))
	defer s.tracer.TraceEnd(span)

	// Validate foreign keys if provided
	if dto.PetSizeCategoryID != nil {
		_, err := s.petProfileRepository().GetPetSizeCategoryByID(ctx, *dto.PetSizeCategoryID)
		if err != nil {
			return exception.ErrRecordNotFound
		}
	}

	if dto.PetEnergyID != nil {
		_, err := s.petProfileRepository().GetPetEnergyByID(ctx, *dto.PetEnergyID)
		if err != nil {
			return exception.ErrRecordNotFound
		}
	}

	// Build update model from DTO
	petProfile := new(models.PetProfile)

	if dto.Name != nil {
		petProfile.Name = *dto.Name
	}
	if dto.ContactName != nil {
		petProfile.ContactName = *dto.ContactName
	}
	if dto.ContactPhone != nil {
		petProfile.ContactPhone = *dto.ContactPhone
	}
	if dto.Note != nil {
		petProfile.Note = *dto.Note
	}
	if dto.PetSizeCategoryID != nil {
		petProfile.PetSizeCategoryID = dto.PetSizeCategoryID
	}
	if dto.PetEnergyID != nil {
		petProfile.PetEnergyID = dto.PetEnergyID
	}

	// Handle image file upload if provided
	if dto.ImageFile != nil {
		// Generate unique filename
		fileExt := ""
		if len(dto.ImageFile.Filename) > 0 {
			for i := len(dto.ImageFile.Filename) - 1; i >= 0; i-- {
				if dto.ImageFile.Filename[i] == '.' {
					fileExt = dto.ImageFile.Filename[i:]
					break
				}
			}
		}

		// Generate filename: pet_profile_{id}_{timestamp}{extension}
		fileName := fmt.Sprintf("pet_profile_%d_%d%s", id, time.Now().Unix(), fileExt)

		// Define the path in storage
		storagePath := "pet-profiles"

		// Upload to MinIO/storage
		err := s.fileSystem.Put(ctx, storagePath, fileName, dto.ImageFile)
		if err != nil {
			s.logger.Error("Failed to upload pet profile image to storage", zap.Error(err), zap.Int("petProfileId", id))
			return err
		}

		// Get the public URL for the uploaded file
		imageURL := s.fileSystem.GetPublicURL(storagePath, fileName)
		petProfile.Image = imageURL
	}

	// Update pet profile in database
	err := s.petProfileRepository().UpdatePetProfile(ctx, id, guestID, petProfile)
	if err != nil {
		return err
	}

	return nil
}

// DeletePetProfile deletes a pet profile (filtered by owner)
func (s Service) DeletePetProfile(ctx context.Context, id int, guestID int) error {
	ctx, span := s.tracer.TraceStart(ctx, "DeletePetProfileService", trace.WithAttributes(attribute.String("service", "DeletePetProfile")))
	defer s.tracer.TraceEnd(span)

	// Delete pet profile through repository
	err := s.petProfileRepository().DeletePetProfile(ctx, id, guestID)
	if err != nil {
		return err
	}

	return nil
}
