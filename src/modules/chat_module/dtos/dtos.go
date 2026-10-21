package dtos

import "mime/multipart"

type CreateChatRoom struct {
	ParticipantID int    `json:"participantId" form:"participantId" validate:"required"`
	Content       string `json:"content" form:"content"`

	// Image files — populated in controller, not from form parser
	Images []*multipart.FileHeader `json:"-" form:"-" validate:"-"`
}

type SendMessage struct {
	Content string `json:"content" form:"content"`

	// Image files — populated in controller, not from form parser
	Images []*multipart.FileHeader `json:"-" form:"-" validate:"-"`
}
