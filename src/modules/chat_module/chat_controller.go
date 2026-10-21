package chat_module

import (
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/validator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/chat_module/dtos"
	"joodpakpaw/src/modules/chat_module/responses"
)

// GetAllChatRooms lists all chat rooms for the authenticated user
//
//	@Summary		List all chat rooms
//	@Description	Get all chat rooms for the authenticated user, with the latest message and the other participant's info
//	@Tags			Chat Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	responses.GetChatRoomsResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/chat-rooms [get]
func (c Controller) GetAllChatRooms(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllChatRoomsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllChatRooms")))
		responseData *responses.GetChatRoomsResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	responseData, err = c.chatService().GetAllChatRooms(ctx, authUser.ID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetChatRoomById retrieves a chat room by ID with all messages
//
//	@Summary		Get chat room by ID
//	@Description	Get a chat room with all messages and sender info. Only accessible by room participants.
//	@Tags			Chat Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"chat room id"
//	@Success		200	{object}	responses.GetChatRoomByIdResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/chat-rooms/{id} [get]
func (c Controller) GetChatRoomById(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetChatRoomByIdController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetChatRoomById"), attribute.Int("id", id)))
		responseData *responses.GetChatRoomByIdResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	responseData, err = c.chatService().GetChatRoomById(ctx, id, authUser.ID)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateChatRoom creates or opens a chat room and sends an initial message
//
//	@Summary		Create chat room
//	@Description	Find or create a chat room with a host, and optionally send an initial message (text and/or images)
//	@Tags			Chat Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			participantId	formData	int		true	"User ID of the person to chat with"
//	@Param			content		formData	string	false	"Text message content"
//	@Param			images		formData	file	false	"Image attachments (can send multiple)"
//	@Success		201	{object}	responses.CreateChatRoomResponse
//	@Failure		400	{object}	exception.ErrorResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/chat-rooms [post]
func (c Controller) CreateChatRoom(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateChatRoomController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateChatRoom")))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Parse multipart form
	dto := new(dtos.CreateChatRoom)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Validate DTO
	if errs := validator.Validate(*dto); errs != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errs...)
	}

	// Cannot chat with yourself
	if dto.ParticipantID == authUser.ID {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}

	// Collect uploaded image files — supports both "images" and "images[]" field names
	form, err := f.MultipartForm()
	if err == nil && form != nil {
		allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
		maxSize := int64(10 * 1024 * 1024) // 10MB

		// Merge files from both "images" and "images[]"
		var allFiles []*multipart.FileHeader
		allFiles = append(allFiles, form.File["images"]...)
		allFiles = append(allFiles, form.File["images[]"]...)

		for _, file := range allFiles {
			if file.Size > maxSize {
				return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.AvatarFileSizeTooLargeResponseError, exception.ErrAvatarFileSizeTooLarge)
			}
			ext := ""
			for i := len(file.Filename) - 1; i >= 0; i-- {
				if file.Filename[i] == '.' {
					ext = strings.ToLower(file.Filename[i:])
					break
				}
			}
			if !allowedExtensions[ext] {
				return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidFileFormatResponseError, exception.ErrInvalidFileFormat)
			}
			dto.Images = append(dto.Images, file)
		}
	}

	responseData, err := c.chatService().CreateChatRoom(ctx, dto, authUser.ID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusCreated).JSON(responseData)
}

// SendMessage sends a message to a chat room
//
//	@Summary		Send message
//	@Description	Send a text message and/or image attachments to an existing chat room. Must be a participant of the room.
//	@Tags			Chat Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"chat room id"
//	@Param			content	formData	string	false	"Text message content"
//	@Param			images	formData	file	false	"Image attachments (supports images or images[])"
//	@Success		201	{object}	responses.SendMessageResponse
//	@Failure		400	{object}	exception.ErrorResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/chat-rooms/{id}/messages [post]
func (c Controller) SendMessage(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "SendMessageController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "SendMessage"), attribute.Int("id", id)))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Parse multipart form
	dto := new(dtos.SendMessage)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Collect uploaded image files — supports both "images" and "images[]"
	form, err := f.MultipartForm()
	if err == nil && form != nil {
		allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
		maxSize := int64(10 * 1024 * 1024) // 10MB

		var allFiles []*multipart.FileHeader
		allFiles = append(allFiles, form.File["images"]...)
		allFiles = append(allFiles, form.File["images[]"]...)

		for _, file := range allFiles {
			if file.Size > maxSize {
				return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.AvatarFileSizeTooLargeResponseError, exception.ErrAvatarFileSizeTooLarge)
			}
			ext := ""
			for i := len(file.Filename) - 1; i >= 0; i-- {
				if file.Filename[i] == '.' {
					ext = strings.ToLower(file.Filename[i:])
					break
				}
			}
			if !allowedExtensions[ext] {
				return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidFileFormatResponseError, exception.ErrInvalidFileFormat)
			}
			dto.Images = append(dto.Images, file)
		}
	}

	// Must have at least content or images
	if strings.TrimSpace(dto.Content) == "" && len(dto.Images) == 0 {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}

	responseData, err := c.chatService().SendMessage(ctx, dto, id, authUser.ID)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusCreated).JSON(responseData)
}

// MarkMessagesAsRead marks all unread messages in the room as read
//
//	@Summary		Mark messages as read
//	@Description	Mark all unread messages in the chat room as read (only messages sent by the other participant)
//	@Tags			Chat Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"chat room id"
//	@Success		204
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/chat-rooms/{id}/messages/read [patch]
func (c Controller) MarkMessagesAsRead(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "MarkMessagesAsReadController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "MarkMessagesAsRead"), attribute.Int("id", id)))
	)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	if err := c.chatService().MarkMessagesAsRead(ctx, id, authUser.ID); err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.SendStatus(fiber.StatusNoContent)
}

// DeleteChatRoom deletes a chat room and all its messages
//
//	@Summary		Delete chat room
//	@Description	Delete a chat room and all its messages. Only participants of the room can delete it.
//	@Tags			Chat Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"chat room id"
//	@Success		204
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/chat-rooms/{id} [delete]
func (c Controller) DeleteChatRoom(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeleteChatRoomController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteChatRoom"), attribute.Int("id", id)))
	)

	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}
	if err := c.chatService().DeleteChatRoom(ctx, id, authUser.ID); err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}
	c.m.tracer.TraceEnd(span)
	return f.SendStatus(fiber.StatusNoContent)
}