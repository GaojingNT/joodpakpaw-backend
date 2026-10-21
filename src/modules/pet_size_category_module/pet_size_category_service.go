package pet_size_category_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/models"
)

func (s Service) GetAllPetSizeCategories(ctx context.Context) ([]models.PetSizeCategory, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllPetSizeCategoriesService", trace.WithAttributes(attribute.String("service", "GetAllPetSizeCategories")))

	petSizeCategories, err := s.petSizeCategoryRepository().GetAllPetSizeCategories(ctx)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return petSizeCategories, nil
}
