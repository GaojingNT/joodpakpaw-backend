package user_recent_view_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/modules/user_recent_view_module/responses"
)

func (s Service) CreateUserRecentView(ctx context.Context, userID int, propertyID int) (*responses.CreateUserRecentViewResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateUserRecentViewService", trace.WithAttributes(attribute.String("service", "CreateUserRecentView")))
	defer s.tracer.TraceEnd(childSpan)

	record, err := s.userRecentViewRepository().CreateUserRecentView(ctx, userID, propertyID)
	if err != nil {
		return nil, err
	}

	return new(responses.CreateUserRecentViewResponse).Make(record), nil
}
