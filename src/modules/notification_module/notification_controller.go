package notification_module

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/jwt"
	"joodpakpaw/src/models"
)

// GetNotifications retrieves all notifications for the authenticated user
//
//	@Summary		Get notifications
//	@Description	Get all notifications for the authenticated user
//	@Tags			Notification Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	responses.GetNotificationsResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/notifications [get]
func (c Controller) GetNotifications(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetNotificationsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetNotifications")))
		err       error
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	result, err := c.notificationService().GetNotificationsByUserID(ctx, authUser.ID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.Status(fiber.StatusOK).JSON(result)
}

// DeleteAllNotifications deletes all notifications for the authenticated user
//
//	@Summary		Delete all notifications
//	@Description	Delete all notifications for the authenticated user
//	@Tags			Notification Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		204
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/notifications [delete]
func (c Controller) DeleteAllNotifications(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeleteAllNotificationsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteAllNotifications")))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	if err := c.notificationService().DeleteAllNotifications(ctx, authUser.ID); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.SendStatus(fiber.StatusNoContent)
}

// DeleteNotificationByID deletes a notification by ID for the authenticated user
//
//	@Summary		Delete notification by ID
//	@Description	Delete a specific notification by ID for the authenticated user
//	@Tags			Notification Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Notification ID"
//	@Success		204
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/notifications/{id} [delete]
func (c Controller) DeleteNotificationByID(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeleteNotificationByIDController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteNotificationByID")))
	)
	defer c.m.tracer.TraceEnd(span)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	notificationID, err := f.ParamsInt("id")
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.NotFoundResponseError, exception.ErrInvalidRequestParameter)
	}

	if err := c.notificationService().DeleteNotificationByID(ctx, authUser.ID, notificationID); err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	return f.SendStatus(fiber.StatusNoContent)
}

// WebSocketUpgrade validates JWT token before upgrading to WebSocket
func (c Controller) WebSocketUpgrade(f *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(f) {
		return fiber.ErrUpgradeRequired
	}

	token := f.Query("token")
	if token == "" {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.TokenMissingResponseError, exception.ErrTokenMissing)
	}

	payload, err := jwt.JwtToken.Payload(token)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidAuthorizationTokenResponseError, exception.ErrInvalidAuthorizationToken)
	}

	userMap, ok := payload["user"].(map[string]interface{})
	if !ok {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidTokenClaimResponseError, exception.ErrInvalidTokenClaim)
	}

	userID := int(userMap["id"].(float64))
	f.Locals("wsUserID", userID)

	return f.Next()
}

// WebSocketHandler handles the persistent WebSocket connection
func (c Controller) WebSocketHandler() func(*websocket.Conn) {
	return func(conn *websocket.Conn) {
		userID := conn.Locals("wsUserID").(int)

		c.m.hub.Register(userID, conn)
		defer c.m.hub.Unregister(userID)

		// Keep connection alive, handle ping/close frames from client
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}
}

// SSEHandler handles Server-Sent Events for notifications (works through HTTP proxies)
func (c Controller) SSEHandler(f *fiber.Ctx) error {
	token := f.Query("token")
	if token == "" {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.TokenMissingResponseError, exception.ErrTokenMissing)
	}

	payload, err := jwt.JwtToken.Payload(token)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidAuthorizationTokenResponseError, exception.ErrInvalidAuthorizationToken)
	}

	userMap, ok := payload["user"].(map[string]interface{})
	if !ok {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidTokenClaimResponseError, exception.ErrInvalidTokenClaim)
	}

	userID := int(userMap["id"].(float64))

	f.Set("Content-Type", "text/event-stream")
	f.Set("Cache-Control", "no-cache")
	f.Set("Connection", "keep-alive")
	f.Set("X-Accel-Buffering", "no")

	ch := c.m.hub.RegisterSSE(userID)

	f.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer c.m.hub.UnregisterSSE(userID)

		// Send initial ping to confirm connection is open
		if _, err := fmt.Fprintf(w, ": connected\n\n"); err != nil {
			return
		}
		w.Flush()

		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if _, err := fmt.Fprintf(w, "data: %s\n\n", msg); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				// Heartbeat — keeps connection alive and detects client disconnect
				if _, err := fmt.Fprintf(w, ": ping\n\n"); err != nil {
					return
				}
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}
