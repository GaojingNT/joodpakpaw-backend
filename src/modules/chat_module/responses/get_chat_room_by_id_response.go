package responses

import (
	"joodpakpaw/src/models"
)

type GetChatRoomByIdResponse struct {
	ID            int           `json:"id"`
	LastMessageAt *string       `json:"lastMessageAt"`
	OtherUser     UserSummary   `json:"otherUser"`
	Messages      []MessageItem `json:"messages"`
}

type MessageItem struct {
	ID          int              `json:"id"`
	Content     string           `json:"content"`
	ReadAt      *string          `json:"readAt"`
	CreatedAt   string           `json:"createdAt"`
	Sender      UserSummary      `json:"sender"`
	Attachments []AttachmentItem `json:"attachments"`
}



func (r *GetChatRoomByIdResponse) Make(room *models.ChatRoom, authUserID int) *GetChatRoomByIdResponse {
	r.ID = room.ID

	if room.LastMessageAt != nil {
		t := room.LastMessageAt.Format("2006-01-02 15:04:05")
		r.LastMessageAt = &t
	}

	// Show the other participant (not the authenticated user)
	if room.Participant1ID != nil && *room.Participant1ID == authUserID {
		if room.Participant2 != nil {
			r.OtherUser = UserSummary{
				ID:        room.Participant2.ID,
				FirstName: room.Participant2.FirstName,
				LastName:  room.Participant2.LastName,
				Avatar:    room.Participant2.Avatar,
			}
		}
	} else {
		if room.Participant1 != nil {
			r.OtherUser = UserSummary{
				ID:        room.Participant1.ID,
				FirstName: room.Participant1.FirstName,
				LastName:  room.Participant1.LastName,
				Avatar:    room.Participant1.Avatar,
			}
		}
	}

	// Map messages
	r.Messages = make([]MessageItem, len(room.Messages))
	for i, msg := range room.Messages {
		item := MessageItem{
			ID:          msg.ID,
			Content:     msg.Content,
			CreatedAt:   msg.CreatedAt.Format("2006-01-02 15:04:05"),
			Attachments: make([]AttachmentItem, len(msg.Attachments)),
		}

		if msg.ReadAt != nil {
			t := msg.ReadAt.Format("2006-01-02 15:04:05")
			item.ReadAt = &t
		}

		if msg.Sender != nil {
			item.Sender = UserSummary{
				ID:        msg.Sender.ID,
				FirstName: msg.Sender.FirstName,
				LastName:  msg.Sender.LastName,
				Avatar:    msg.Sender.Avatar,
			}
		}

		for j, a := range msg.Attachments {
			item.Attachments[j] = AttachmentItem{
				ID:          a.ID,
				FileName:    a.FileName,
				FilePath:    a.FilePath,
				ContentType: a.FileContentType,
				FileSize:    a.FileSize,
			}
		}

		r.Messages[i] = item
	}

	return r
}
