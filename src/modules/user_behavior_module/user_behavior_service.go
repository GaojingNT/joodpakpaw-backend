package user_behavior_module

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/user_behavior_module/responses"
)

func (s Service) GetAllUserBehaviors(ctx context.Context, userID int) (*responses.GetUserBehaviorsResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllUserBehaviorsService", trace.WithAttributes(attribute.String("service", "GetAllUserBehaviors")))
	defer s.tracer.TraceEnd(childSpan)

	behaviors, err := s.userBehaviorRepository().GetAllUserBehaviors(ctx, userID)
	if err != nil {
		return nil, err
	}

	return new(responses.GetUserBehaviorsResponse).Make(behaviors), nil
}

func (s Service) CreateViewPropertyBehavior(ctx context.Context, propertyID int, userID int) (*responses.CreateUserBehaviorResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateViewPropertyBehaviorService", trace.WithAttributes(attribute.String("service", "CreateViewPropertyBehavior")))
	defer s.tracer.TraceEnd(childSpan)

	repo := s.userBehaviorRepository()

	// Find behavior category "view_property"
	category, err := repo.GetOrCreateBehaviorCategoryByName(ctx, "view_property")
	if err != nil {
		return nil, err
	}

	// Build record
	behavior := &models.UserBehavior{
		UserID:             userID,
		ReferenceTable:     "tbl_properties",
		ReferenceRecord:    strconv.Itoa(propertyID),
		BehaviorCategoryID: category.ID,
	}

	behavior, err = repo.CreateUserBehavior(ctx, behavior)
	if err != nil {
		return nil, err
	}

	response := new(responses.CreateUserBehaviorResponse).Make(behavior)
	return response, nil
}
