package admin_dashboard_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/modules/admin_dashboard_module/responses"
)

func (s Service) GetDashboard(ctx context.Context) (*responses.DashboardResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetDashboardService", trace.WithAttributes(attribute.String("service", "GetDashboard")))
	defer s.tracer.TraceEnd(childSpan)

	repo := s.dashboardRepository()

	pendingReports, err := repo.CountPendingReports(ctx)
	if err != nil {
		return nil, err
	}
	pendingReportsSinceYesterday, err := repo.CountPendingReportsThisMonth(ctx)
	if err != nil {
		return nil, err
	}

	totalUsers, err := repo.CountTotalUsers(ctx)
	if err != nil {
		return nil, err
	}
	newUsersThisMonth, err := repo.CountNewUsersThisMonth(ctx)
	if err != nil {
		return nil, err
	}

	activeProperties, err := repo.CountActiveProperties(ctx)
	if err != nil {
		return nil, err
	}
	activePropertiesThisMonth, err := repo.CountActivePropertiesPublishedThisMonth(ctx)
	if err != nil {
		return nil, err
	}

	pendingVerifications, err := repo.CountPendingVerifications(ctx)
	if err != nil {
		return nil, err
	}
	pendingVerificationsSinceMonth, err := repo.CountPendingVerificationsSinceMonth(ctx)
	if err != nil {
		return nil, err
	}

	bannedHosts, err := repo.CountBannedAllProperties(ctx)
	if err != nil {
		return nil, err
	}
	bannedHostsThisMonth, err := repo.CountBannedAllPropertiesThisMonth(ctx)
	if err != nil {
		return nil, err
	}

	return &responses.DashboardResponse{
		PendingReports: responses.StatWithChange{
			Total:     pendingReports,
			ThisMonth: pendingReportsSinceYesterday,
		},
		TotalUsers: responses.StatWithChange{
			Total:     totalUsers,
			ThisMonth: newUsersThisMonth,
		},
		ActiveProperties: responses.StatWithChange{
			Total:     activeProperties,
			ThisMonth: activePropertiesThisMonth,
		},
		PendingVerifications: responses.StatWithChange{
			Total:     pendingVerifications,
			ThisMonth: pendingVerificationsSinceMonth,
		},
		BannedHosts: responses.StatWithChange{
			Total:     bannedHosts,
			ThisMonth: bannedHostsThisMonth,
		},
	}, nil
}
