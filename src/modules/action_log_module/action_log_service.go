package action_log_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/action_log_module/responses"
)

func (s Service) GetActionLogs(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetActionLogsService", trace.WithAttributes(attribute.String("service", "GetActionLogs")))
	defer s.tracer.TraceEnd(span)

	result, err := s.actionLogRepository().GetActionLogsPaginate(ctx, pagination)
	if err != nil {
		return nil, err
	}

	if result.Data != nil {
		result.Data = new(responses.GetActionLogsCollection).Collection(result.Data.([]models.ActionLog))
	}

	return result, nil
}
