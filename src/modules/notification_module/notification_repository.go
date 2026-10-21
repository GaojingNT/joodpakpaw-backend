package notification_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/src/models"
)

// CreateNotification saves a new notification to database
func (r Repository) CreateNotification(ctx context.Context, userID int, roleName string, actionName string, title, message, referenceTable string, referenceID *int) (*models.Notification, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateNotificationRepository", trace.WithAttributes(attribute.String("repository", "CreateNotification")))
		notification = &models.Notification{
			Title:          title,
			Message:        message,
			Category:       models.NotificationCategoryMessage,
			Type:           models.NotificationTypeInfo,
			IsRead:         false,
			UserID:         &userID,
			ReferenceTable: referenceTable,
			ReferenceID:    referenceID,
		}
		err error
	)

	// Lookup role by name
	if roleName != "" {
		var role models.Role
		if dbErr := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; dbErr == nil {
			notification.RoleID = &role.ID
		}
	}

	// Lookup notification action by name
	if actionName != "" {
		var action models.NotificationAction
		if dbErr := r.db.WithContext(ctx).Where("name = ?", actionName).First(&action).Error; dbErr == nil {
			notification.NotificationActionID = &action.ID
		}
	}

	utils.Block{
		Try: func() {
			if err = r.db.WithContext(ctx).Create(notification).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	// Preload Role and NotificationAction for realtime push
	r.db.WithContext(ctx).Preload("Role").Preload("NotificationAction").First(notification, notification.ID)

	return notification, nil
}

// GetNotificationsByUserID retrieves all notifications for a user
func (r Repository) GetNotificationsByUserID(ctx context.Context, userID int) ([]models.Notification, error) {
	var (
		_, childSpan  = r.tracer.TraceStart(ctx, "GetNotificationsByUserIDRepository", trace.WithAttributes(attribute.String("repository", "GetNotificationsByUserID"), attribute.Int("userId", userID)))
		notifications = make([]models.Notification, 0)
		err           error
	)

	utils.Block{
		Try: func() {
			if err = r.db.WithContext(ctx).
				Preload("Role").
				Preload("NotificationAction").
				Where("user_id = ?", userID).
				Order("created_at DESC").
				Find(&notifications).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// DeleteAllNotifications soft-deletes all notifications for a user
func (r Repository) DeleteAllNotifications(ctx context.Context, userID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteAllNotificationsRepository", trace.WithAttributes(attribute.String("repository", "DeleteAllNotifications"), attribute.Int("userId", userID)))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.WithContext(ctx).
				Where("user_id = ?", userID).
				Delete(&models.Notification{}).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)
	return err
}

// DeleteNotificationByID soft-deletes a notification by ID for a specific user
func (r Repository) DeleteNotificationByID(ctx context.Context, userID int, notificationID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteNotificationByIDRepository", trace.WithAttributes(attribute.String("repository", "DeleteNotificationByID"), attribute.Int("notificationId", notificationID)))
		err          error
	)

	utils.Block{
		Try: func() {
			result := r.db.WithContext(ctx).
				Where("id = ? AND user_id = ?", notificationID, userID).
				Delete(&models.Notification{})
			if result.Error != nil {
				utils.Throw(result.Error)
			}
			if result.RowsAffected == 0 {
				utils.Throw(exception.ErrRecordNotFound)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			if err != exception.ErrRecordNotFound {
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)
	return err
}
