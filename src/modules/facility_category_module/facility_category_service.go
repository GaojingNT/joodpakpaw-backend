package facility_category_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/models"
)

func (s Service) GetAllFacilityCategories(ctx context.Context) ([]models.FacilityCategory, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllFacilityCategoriesService", trace.WithAttributes(attribute.String("service", "GetAllFacilityCategories")))

	facilityCategories, err := s.facilityCategoryRepository().GetAllFacilityCategories(ctx)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return facilityCategories, nil
}

func (s Service) GetFacilityCategoryByID(ctx context.Context, id int) (models.FacilityCategory, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetFacilityCategoryByIDService", trace.WithAttributes(attribute.String("service", "GetFacilityCategoryByID"), attribute.Int64("id", int64(id))))

	facilityCategory, err := s.facilityCategoryRepository().GetFacilityCategoryByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return models.FacilityCategory{}, err
	}

	s.tracer.TraceEnd(childSpan)

	return facilityCategory, nil
}
