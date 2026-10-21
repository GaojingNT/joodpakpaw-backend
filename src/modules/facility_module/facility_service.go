package facility_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/facility_module/responses"
)

func (s Service) GetFacilities(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetFacilitiesService", trace.WithAttributes(attribute.String("service", "GetFacilities")))

	result, err := s.facilityRepository().GetFacilityPaginate(ctx, paginate)

	// Convert the result to a collection of responses
	if result.Data != nil {
		result.Data = new(responses.GetFacilitiesResponse).Collection(result.Data.([]models.Facility))
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) GetAllFacilities(ctx context.Context) ([]models.Facility, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllFacilitiesService", trace.WithAttributes(attribute.String("service", "GetAllFacilities")))

	facilities, err := s.facilityRepository().GetAllFacilities(ctx)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return facilities, nil
}

func (s Service) GetFacilityByID(ctx context.Context, id int) (models.Facility, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetFacilityByIDService", trace.WithAttributes(attribute.String("service", "GetFacilityByID"), attribute.Int64("id", int64(id))))

	facility, err := s.facilityRepository().GetFacilityByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return models.Facility{}, err
	}

	s.tracer.TraceEnd(childSpan)

	return facility, nil
}

