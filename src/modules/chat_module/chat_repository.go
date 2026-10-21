package chat_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/src/models"
)

// ExistsChatRoom returns true if a chat room already exists between userA and userB
func (r Repository) ExistsChatRoom(ctx context.Context, userA int, userB int) bool {
	p1, p2 := userA, userB
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	var count int64
	r.db.WithContext(ctx).Model(&models.ChatRoom{}).
		Where("participant_1_id = ? AND participant_2_id = ?", p1, p2).
		Count(&count)
	return count > 0
}

// FindChatRoomID returns the chat room ID if one exists between userA and userB, or nil if not found
func (r Repository) FindChatRoomID(ctx context.Context, userA int, userB int) *int {
	p1, p2 := userA, userB
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	var room models.ChatRoom
	err := r.db.WithContext(ctx).Model(&models.ChatRoom{}).
		Select("id").
		Where("participant_1_id = ? AND participant_2_id = ?", p1, p2).
		First(&room).Error
	if err != nil {
		return nil
	}
	return &room.ID
}

// GetChatRoomParticipantIDs returns participant1ID and participant2ID for a room
func (r Repository) GetChatRoomParticipantIDs(ctx context.Context, roomID int) (p1ID int, p2ID int, err error) {
	_, childSpan := r.tracer.TraceStart(ctx, "GetChatRoomParticipantIDsRepository", trace.WithAttributes(attribute.String("repository", "GetChatRoomParticipantIDs")))

	utils.Block{
		Try: func() {
			var room models.ChatRoom
			if err = r.db.Select("participant_1_id, participant_2_id").First(&room, roomID).Error; err != nil {
				utils.Throw(err)
			}
			if room.Participant1ID != nil {
				p1ID = *room.Participant1ID
			}
			if room.Participant2ID != nil {
				p2ID = *room.Participant2ID
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

	return p1ID, p2ID, err
}

// GetAllChatRooms retrieves all chat rooms for a user (as either participant)
func (r Repository) GetAllChatRooms(ctx context.Context, userID int) ([]models.ChatRoom, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllChatRoomsRepository", trace.WithAttributes(attribute.String("repository", "GetAllChatRooms")))
		rooms        = make([]models.ChatRoom, 0)
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("userID", userID))

	utils.Block{
		Try: func() {
			if err = r.db.
				Where("participant_1_id = ? OR participant_2_id = ?", userID, userID).
				Preload("Participant1").
				Preload("Participant2").
				Order("last_message_at DESC").
				Find(&rooms).Error; err != nil {
				utils.Throw(err)
			}

			// Manually populate LastMessage (gorm:"-" prevents auto preload)
			for i, room := range rooms {
				if room.LastMessageID != nil {
					var msg models.ChatMessage
					if err = r.db.First(&msg, *room.LastMessageID).Error; err == nil {
						rooms[i].LastMessage = &msg
					}
					err = nil
				}
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

	return rooms, nil
}

// GetChatRoomById retrieves a single chat room by ID with all messages and sender info
func (r Repository) GetChatRoomById(ctx context.Context, roomID int, userID int) (*models.ChatRoom, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetChatRoomByIdRepository", trace.WithAttributes(attribute.String("repository", "GetChatRoomById")))
		room         models.ChatRoom
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("roomID", roomID))
	r.tracer.SetAttributes(childSpan, attribute.Int("userID", userID))

	utils.Block{
		Try: func() {
			if err = r.db.
				Where("id = ? AND (participant_1_id = ? OR participant_2_id = ?)", roomID, userID, userID).
				Preload("Participant1").
				Preload("Participant2").
				Preload("Messages", func(db *gorm.DB) *gorm.DB {
					return db.Order("created_at ASC")
				}).
				Preload("Messages.Sender").
				Preload("Messages.Attachments").
				First(&room).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return &room, nil
}

// FindOrCreateChatRoom finds existing room between two users, or creates a new one.
// Convention: participant_1_id is always the smaller ID to guarantee uniqueness.
func (r Repository) FindOrCreateChatRoom(ctx context.Context, userA int, userB int) (*models.ChatRoom, bool, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "FindOrCreateChatRoomRepository", trace.WithAttributes(attribute.String("repository", "FindOrCreateChatRoom")))
		room         models.ChatRoom
		created      bool
		err          error
	)

	// Enforce convention: smaller ID always goes to participant_1
	p1, p2 := userA, userB
	if p1 > p2 {
		p1, p2 = p2, p1
	}

	r.tracer.SetAttributes(childSpan, attribute.Int("participant1ID", p1))
	r.tracer.SetAttributes(childSpan, attribute.Int("participant2ID", p2))

	utils.Block{
		Try: func() {
			findErr := r.db.
				Where("participant_1_id = ? AND participant_2_id = ?", p1, p2).
				Preload("Participant1").
				Preload("Participant2").
				First(&room).Error

			if findErr == nil {
				return
			}

			if findErr.Error() != "record not found" {
				utils.Throw(findErr)
			}

			// Create new room
			room = models.ChatRoom{
				Participant1ID: &p1,
				Participant2ID: &p2,
			}
			if err = r.db.Create(&room).Error; err != nil {
				utils.Throw(err)
			}

			// Reload with associations
			if err = r.db.
				Preload("Participant1").
				Preload("Participant2").
				First(&room, room.ID).Error; err != nil {
				utils.Throw(err)
			}

			created = true
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
		return nil, false, err
	}

	return &room, created, nil
}

// MarkMessagesAsRead sets read_at = now for all unread messages NOT sent by readerID
func (r Repository) MarkMessagesAsRead(ctx context.Context, roomID int, readerID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "MarkMessagesAsReadRepository", trace.WithAttributes(attribute.String("repository", "MarkMessagesAsRead")))
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("roomID", roomID))
	r.tracer.SetAttributes(childSpan, attribute.Int("readerID", readerID))

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.ChatMessage{}).
				Where("room_id = ? AND sender_id != ? AND read_at IS NULL", roomID, readerID).
				Update("read_at", time.Now()).Error; err != nil {
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

// VerifyChatRoomParticipant checks that the given user belongs to the room
func (r Repository) VerifyChatRoomParticipant(ctx context.Context, roomID int, userID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "VerifyChatRoomParticipantRepository", trace.WithAttributes(attribute.String("repository", "VerifyChatRoomParticipant")))
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("roomID", roomID))
	r.tracer.SetAttributes(childSpan, attribute.Int("userID", userID))

	utils.Block{
		Try: func() {
			var room models.ChatRoom
			if err = r.db.
				Where("id = ? AND (participant_1_id = ? OR participant_2_id = ?)", roomID, userID, userID).
				First(&room).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	return err
}

// DeleteChatRoom hard-deletes a chat room and all its messages. Only participants can delete.
func (r Repository) DeleteChatRoom(ctx context.Context, roomID int, userID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteChatRoomRepository", trace.WithAttributes(attribute.String("repository", "DeleteChatRoom")))
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("roomID", roomID))
	r.tracer.SetAttributes(childSpan, attribute.Int("userID", userID))

	utils.Block{
		Try: func() {
			var room models.ChatRoom
			if err = r.db.
				Where("id = ? AND (participant_1_id = ? OR participant_2_id = ?)", roomID, userID, userID).
				First(&room).Error; err != nil {
				utils.Throw(err)
			}

			if err = r.db.Where("room_id = ?", roomID).Delete(&models.ChatMessage{}).Error; err != nil {
				utils.Throw(err)
			}
			if err = r.db.Delete(&room).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)

			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else {
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	return err
}

// CreateChatMessage creates a new message and updates room's LastMessageID and LastMessageAt
func (r Repository) CreateChatMessage(ctx context.Context, roomID int, senderID int, content string, attachments []models.ChatMessageAttachment) (*models.ChatMessage, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateChatMessageRepository", trace.WithAttributes(attribute.String("repository", "CreateChatMessage")))
		message      models.ChatMessage
		err          error
	)

	r.tracer.SetAttributes(childSpan, attribute.Int("roomID", roomID))
	r.tracer.SetAttributes(childSpan, attribute.Int("senderID", senderID))

	utils.Block{
		Try: func() {
			message = models.ChatMessage{
				Content:     content,
				RoomID:      &roomID,
				SenderID:    &senderID,
				Attachments: attachments,
			}

			if err = r.db.Create(&message).Error; err != nil {
				utils.Throw(err)
			}

			// Update room LastMessageID and LastMessageAt
			now := message.CreatedAt
			if err = r.db.Model(&models.ChatRoom{}).
				Where("id = ?", roomID).
				Updates(map[string]interface{}{
					"last_message_id": message.ID,
					"last_message_at": now,
				}).Error; err != nil {
				utils.Throw(err)
			}

			// Reload message with Sender and Attachments
			if err = r.db.Preload("Sender").Preload("Attachments").First(&message, message.ID).Error; err != nil {
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

	return &message, nil
}
