package pet_breed_module

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/pet_breed_module/dtos"
	"joodpakpaw/src/modules/pet_breed_module/responses"
)

func (s Service) GetPetBreedsPaginate(ctx context.Context, pagination *paginator.Pagination, petId int) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPetBreedsPaginateService", trace.WithAttributes(attribute.String("service", "GetPetBreedsPaginate"), attribute.Int("petId", petId)))

	result, err := s.petBreedRepository().GetPetBreedsPaginate(ctx, pagination, petId)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	// Transform data to response
	if result != nil && result.Data != nil {
		result.Data = new(responses.GetPetBreedsResponse).Collection(result.Data.([]models.PetBreed))
	}

	s.tracer.TraceEnd(childSpan)

	return result, nil
}

func (s Service) GetAllPetBreeds(ctx context.Context, petId int) ([]models.PetBreed, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllPetBreedsService", trace.WithAttributes(attribute.String("service", "GetAllPetBreeds"), attribute.Int("petId", petId)))

	petBreeds, err := s.petBreedRepository().GetAllPetBreeds(ctx, petId)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return petBreeds, nil
}

func (s Service) CreatePetBreed(ctx context.Context, dto *dtos.CreatePetBreed) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreatePetBreedService", trace.WithAttributes(attribute.String("service", "CreatePetBreed"), attribute.Int("petId", *dto.PetID)))

	petBreed := &models.PetBreed{
		Breed: dto.Breed,
		PetID: dto.PetID,
	}

	if dto.ImageFile != nil {
		ext := filepath.Ext(dto.ImageFile.Filename)
		fileName := fmt.Sprintf("pet_breed_%d_%d%s", *dto.PetID, time.Now().UnixNano(), ext)
		storagePath := "pet-breeds"

		if err := s.fileSystem.Put(ctx, storagePath, fileName, dto.ImageFile); err != nil {
			s.logger.Error("Failed to upload pet breed image", zap.Error(err))
			s.tracer.TraceEnd(childSpan)
			return err
		}

		publicURL := s.fileSystem.GetPublicURL(storagePath, fileName)
		petBreed.ImageURL = &publicURL
	}

	if err := s.petBreedRepository().CreatePetBreed(ctx, petBreed); err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) UpdatePetBreed(ctx context.Context, petId int, breedId int, dto *dtos.UpdatePetBreed) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdatePetBreedService", trace.WithAttributes(attribute.String("service", "UpdatePetBreed"), attribute.Int("petId", petId), attribute.Int("breedId", breedId)))

	if _, err := s.petBreedRepository().GetPetBreedByID(ctx, petId, breedId); err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	petBreed := &models.PetBreed{
		Breed: dto.Breed,
	}

	if dto.ImageFile != nil {
		ext := filepath.Ext(dto.ImageFile.Filename)
		fileName := fmt.Sprintf("pet_breed_%d_%d%s", breedId, time.Now().UnixNano(), ext)
		storagePath := "pet-breeds"

		if err := s.fileSystem.Put(ctx, storagePath, fileName, dto.ImageFile); err != nil {
			s.logger.Error("Failed to upload pet breed image", zap.Error(err))
			s.tracer.TraceEnd(childSpan)
			return err
		}

		publicURL := s.fileSystem.GetPublicURL(storagePath, fileName)
		petBreed.ImageURL = &publicURL
	}

	if err := s.petBreedRepository().UpdatePetBreed(ctx, breedId, petBreed); err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) GetPetBreedByID(ctx context.Context, petId int, breedId int) (models.PetBreed, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPetBreedByIDService", trace.WithAttributes(attribute.String("service", "GetPetBreedByID"), attribute.Int("petId", petId), attribute.Int("breedId", breedId)))

	petBreed, err := s.petBreedRepository().GetPetBreedByID(ctx, petId, breedId)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return petBreed, err
	}

	s.tracer.TraceEnd(childSpan)

	return petBreed, nil
}
