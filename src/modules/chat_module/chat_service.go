package chat_module

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/chat_module/dtos"
	"joodpakpaw/src/modules/chat_module/responses"
)

// GetAllChatRooms retrieves all chat rooms for the authenticated user
func (s Service) GetAllChatRooms(ctx context.Context, userID int) (*responses.GetChatRoomsResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetAllChatRoomsService", trace.WithAttributes(attribute.String("service", "GetAllChatRooms")))

	rooms, err := s.chatRepository().GetAllChatRooms(ctx, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	response := new(responses.GetChatRoomsResponse).Make(rooms, userID)

	s.tracer.TraceEnd(childSpan)

	return response, nil
}

// GetChatRoomById retrieves a chat room by ID with all messages and sender info,
// and automatically marks unread messages (sent to the caller) as read.
func (s Service) GetChatRoomById(ctx context.Context, roomID int, userID int) (*responses.GetChatRoomByIdResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetChatRoomByIdService", trace.WithAttributes(attribute.String("service", "GetChatRoomById")))
	defer s.tracer.TraceEnd(childSpan)

	// Mark as read first so the fetched messages already have read_at populated
	if markErr := s.chatRepository().MarkMessagesAsRead(ctx, roomID, userID); markErr == nil && s.hub != nil {
		p1ID, p2ID, pErr := s.chatRepository().GetChatRoomParticipantIDs(ctx, roomID)
		if pErr == nil {
			otherID := p1ID
			if p1ID == userID {
				otherID = p2ID
			}
			s.hub.EmitToUser(otherID, "messages_read", map[string]int{"roomId": roomID})
		}
	}

	room, err := s.chatRepository().GetChatRoomById(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}

	return new(responses.GetChatRoomByIdResponse).Make(room, userID), nil
}

// CreateChatRoom finds or creates a chat room, then sends an initial message (text + optional images)
func (s Service) CreateChatRoom(ctx context.Context, dto *dtos.CreateChatRoom, authUserID int) (*responses.CreateChatRoomResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateChatRoomService", trace.WithAttributes(attribute.String("service", "CreateChatRoom")))
	defer s.tracer.TraceEnd(childSpan)

	// Find or create room
	room, isNew, err := s.chatRepository().FindOrCreateChatRoom(ctx, authUserID, dto.ParticipantID)
	if err != nil {
		return nil, err
	}

	// If no message content and no images, return room without a message
	hasContent := strings.TrimSpace(dto.Content) != ""
	hasImages := len(dto.Images) > 0
	if !hasContent && !hasImages {
		return new(responses.CreateChatRoomResponse).Make(room, nil, authUserID), nil
	}

	// Upload images and build attachments
	attachments := make([]models.ChatMessageAttachment, 0, len(dto.Images))
	for _, imageFile := range dto.Images {
		// Get file extension
		fileExt := ""
		for i := len(imageFile.Filename) - 1; i >= 0; i-- {
			if imageFile.Filename[i] == '.' {
				fileExt = strings.ToLower(imageFile.Filename[i:])
				break
			}
		}

		fileName := fmt.Sprintf("chat_%d_%d_%d%s", room.ID, authUserID, time.Now().UnixNano(), fileExt)
		storagePath := "chat-attachments"

		if err := s.fileSystem.Put(ctx, storagePath, fileName, imageFile); err != nil {
			s.logger.Error("Failed to upload chat attachment", zap.Error(err))
			return nil, err
		}

		filePath := s.fileSystem.GetPublicURL(storagePath, fileName)

		attachments = append(attachments, models.ChatMessageAttachment{
			FileName:          imageFile.Filename,
			FilePath:          filePath,
			FileContentType:   imageFile.Header.Get("Content-Type"),
			FileExtensionName: fileExt,
			FileSize:          int(imageFile.Size),
		})
	}

	// Create message (content may be empty if only images)
	content := dto.Content
	message, err := s.chatRepository().CreateChatMessage(ctx, room.ID, authUserID, content, attachments)
	if err != nil {
		return nil, err
	}

	response := new(responses.CreateChatRoomResponse).Make(room, message, authUserID)

	// Emit to the recipient
	if s.hub != nil && message != nil {
		recipientID := dto.ParticipantID
		if isNew {
			// New room: send full room info so frontend can prepend without re-fetching
			lastMsgAt := message.CreatedAt.Format("2006-01-02 15:04:05")
			chatRoomItem := responses.ChatRoomItem{
				ID:            room.ID,
				LastMessageAt: &lastMsgAt,
				LastMessage: &responses.LastMessageItem{
					ID:       message.ID,
					Content:  message.Content,
					ReadAt:   nil,
					SenderID: message.SenderID,
				},
			}
			// otherUser from recipient's perspective = sender (authUserID)
			if room.Participant1ID != nil && *room.Participant1ID == authUserID && room.Participant1 != nil {
				chatRoomItem.OtherUser = responses.UserSummary{
					ID:        room.Participant1.ID,
					FirstName: room.Participant1.FirstName,
					LastName:  room.Participant1.LastName,
					Avatar:    room.Participant1.Avatar,
				}
			} else if room.Participant2 != nil {
				chatRoomItem.OtherUser = responses.UserSummary{
					ID:        room.Participant2.ID,
					FirstName: room.Participant2.FirstName,
					LastName:  room.Participant2.LastName,
					Avatar:    room.Participant2.Avatar,
				}
			}
			s.hub.EmitToUser(recipientID, "new_chat_room", chatRoomItem)
		} else {
			s.hub.EmitToUser(recipientID, "new_message", map[string]interface{}{
				"roomId":  room.ID,
				"message": new(responses.SendMessageResponse).Make(message),
			})
		}
	}

	return response, nil
}

// SendMessage sends a message (text and/or images) to an existing chat room
func (s Service) SendMessage(ctx context.Context, dto *dtos.SendMessage, roomID int, authUserID int) (*responses.SendMessageResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "SendMessageService", trace.WithAttributes(attribute.String("service", "SendMessage")))
	defer s.tracer.TraceEnd(childSpan)

	// Verify user is a participant of this room
	if err := s.chatRepository().VerifyChatRoomParticipant(ctx, roomID, authUserID); err != nil {
		return nil, err
	}

	// Upload images and build attachments
	attachments := make([]models.ChatMessageAttachment, 0, len(dto.Images))
	for _, imageFile := range dto.Images {
		fileExt := ""
		for i := len(imageFile.Filename) - 1; i >= 0; i-- {
			if imageFile.Filename[i] == '.' {
				fileExt = strings.ToLower(imageFile.Filename[i:])
				break
			}
		}

		fileName := fmt.Sprintf("chat_%d_%d_%d%s", roomID, authUserID, time.Now().UnixNano(), fileExt)
		storagePath := "chat-attachments"

		if err := s.fileSystem.Put(ctx, storagePath, fileName, imageFile); err != nil {
			s.logger.Error("Failed to upload chat attachment", zap.Error(err))
			return nil, err
		}

		attachments = append(attachments, models.ChatMessageAttachment{
			FileName:          imageFile.Filename,
			FilePath:          s.fileSystem.GetPublicURL(storagePath, fileName),
			FileContentType:   imageFile.Header.Get("Content-Type"),
			FileExtensionName: fileExt,
			FileSize:          int(imageFile.Size),
		})
	}

	message, err := s.chatRepository().CreateChatMessage(ctx, roomID, authUserID, dto.Content, attachments)
	if err != nil {
		return nil, err
	}

	response := new(responses.SendMessageResponse).Make(message)

	// Emit new_message to the recipient via Socket.IO
	if s.hub != nil {
		p1ID, p2ID, pErr := s.chatRepository().GetChatRoomParticipantIDs(ctx, roomID)
		if pErr == nil {
			recipientID := p1ID
			if p1ID == authUserID {
				recipientID = p2ID
			}
			s.hub.EmitToUser(recipientID, "new_message", map[string]interface{}{
				"roomId":  roomID,
				"message": response,
			})
		}
	}

	return response, nil
}

// MarkMessagesAsRead marks all unread messages in the room (not sent by reader) as read
func (s Service) MarkMessagesAsRead(ctx context.Context, roomID int, readerID int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "MarkMessagesAsReadService", trace.WithAttributes(attribute.String("service", "MarkMessagesAsRead")))
	defer s.tracer.TraceEnd(childSpan)

	// Verify user is a participant of this room
	if err := s.chatRepository().VerifyChatRoomParticipant(ctx, roomID, readerID); err != nil {
		return err
	}

	if err := s.chatRepository().MarkMessagesAsRead(ctx, roomID, readerID); err != nil {
		return err
	}

	// Emit messages_read to the other participant (the message sender)
	if s.hub != nil {
		p1ID, p2ID, pErr := s.chatRepository().GetChatRoomParticipantIDs(ctx, roomID)
		if pErr == nil {
			otherID := p1ID
			if p1ID == readerID {
				otherID = p2ID
			}
			s.hub.EmitToUser(otherID, "messages_read", map[string]int{"roomId": roomID})
		}
	}

	return nil
}

func (s Service) DeleteChatRoom(ctx context.Context, roomID int, authUserID int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteChatRoomService", trace.WithAttributes(attribute.String("service", "DeleteChatRoom")))
	defer s.tracer.TraceEnd(childSpan)

	err := s.chatRepository().DeleteChatRoom(ctx, roomID, authUserID)
	if err != nil {
		return err
	}
	return nil
}