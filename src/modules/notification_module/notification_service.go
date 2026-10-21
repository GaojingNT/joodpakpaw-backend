package notification_module

import (
	"context"

	"github.com/goccy/go-json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/notification_module/responses"
)

// SendToUser saves notification to DB and pushes to connected WebSocket client
func (s Service) SendToUser(ctx context.Context, userID int, roleName string, actionName string, title, message, referenceTable string, referenceID *int) error {
	ctx, span := s.tracer.TraceStart(ctx, "SendToUserService", trace.WithAttributes(attribute.String("service", "SendToUser")))
	defer s.tracer.TraceEnd(span)

	notification, err := s.notificationRepository().CreateNotification(ctx, userID, roleName, actionName, title, message, referenceTable, referenceID)
	if err != nil {
		return err
	}

	s.pushToClient(userID, notification)
	return nil
}

// DeleteAllNotifications deletes all notifications for a user
func (s Service) DeleteAllNotifications(ctx context.Context, userID int) error {
	ctx, span := s.tracer.TraceStart(ctx, "DeleteAllNotificationsService", trace.WithAttributes(attribute.String("service", "DeleteAllNotifications")))
	defer s.tracer.TraceEnd(span)

	return s.notificationRepository().DeleteAllNotifications(ctx, userID)
}

// DeleteNotificationByID deletes a notification by ID for a user
func (s Service) DeleteNotificationByID(ctx context.Context, userID int, notificationID int) error {
	ctx, span := s.tracer.TraceStart(ctx, "DeleteNotificationByIDService", trace.WithAttributes(attribute.String("service", "DeleteNotificationByID")))
	defer s.tracer.TraceEnd(span)

	return s.notificationRepository().DeleteNotificationByID(ctx, userID, notificationID)
}

// GetNotificationsByUserID retrieves all notifications for a user
func (s Service) GetNotificationsByUserID(ctx context.Context, userID int) (*responses.GetNotificationsResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetNotificationsByUserIDService", trace.WithAttributes(attribute.String("service", "GetNotificationsByUserID")))
	defer s.tracer.TraceEnd(span)

	notifications, err := s.notificationRepository().GetNotificationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return new(responses.GetNotificationsResponse).Make(notifications), nil
}

// pushToClient pushes notification payload to connected WebSocket client
func (s Service) pushToClient(userID int, notification *models.Notification) {
	if s.hub == nil {
		return
	}

	data, err := json.Marshal(new(responses.NotificationResponse).Make(*notification))
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	s.hub.SendToUser(userID, data)
}
