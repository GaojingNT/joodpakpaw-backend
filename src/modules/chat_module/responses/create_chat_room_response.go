package responses

import "joodpakpaw/src/models"

type CreateChatRoomResponse struct {
	ID        int          `json:"id"`
	OtherUser UserSummary  `json:"otherUser"`
	Message   *MessageItem `json:"message"`
}

func (r *CreateChatRoomResponse) Make(room *models.ChatRoom, message *models.ChatMessage, authUserID int) *CreateChatRoomResponse {
	r.ID = room.ID

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

	if message != nil {
		msgItem := &MessageItem{
			ID:        message.ID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if message.ReadAt != nil {
			t := message.ReadAt.Format("2006-01-02 15:04:05")
			msgItem.ReadAt = &t
		}
		if message.Sender != nil {
			msgItem.Sender = UserSummary{
				ID:        message.Sender.ID,
				FirstName: message.Sender.FirstName,
				LastName:  message.Sender.LastName,
				Avatar:    message.Sender.Avatar,
			}
		}
		r.Message = msgItem
	}

	return r
}
