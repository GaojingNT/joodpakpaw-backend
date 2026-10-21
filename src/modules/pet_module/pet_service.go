package pet_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/models"
)

func (s Service) GetAllPets(ctx context.Context) ([]models.Pet, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllPetsService", trace.WithAttributes(attribute.String("service", "GetAllPets")))

	pets, err := s.petRepository().GetAllPets(ctx)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return pets, nil
}
