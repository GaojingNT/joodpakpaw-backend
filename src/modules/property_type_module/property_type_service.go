package property_type_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/models"
)

func (s Service) GetAllPropertyTypes(ctx context.Context) ([]models.PropertyType, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllPropertyTypesService", trace.WithAttributes(attribute.String("service", "GetAllPropertyTypes")))

	propertyTypes, err := s.propertyTypeRepository().GetAllPropertyTypes(ctx)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	s.tracer.TraceEnd(childSpan)

	return propertyTypes, nil
}

func (s Service) GetPropertyTypeByID(ctx context.Context, id int) (models.PropertyType, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPropertyTypeByIDService", trace.WithAttributes(attribute.String("service", "GetPropertyTypeByID"), attribute.Int64("id", int64(id))))

	propertyType, err := s.propertyTypeRepository().GetPropertyTypeByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return models.PropertyType{}, err
	}

	s.tracer.TraceEnd(childSpan)

	return propertyType, nil
}

