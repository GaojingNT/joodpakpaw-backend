package responses

import (
	"joodpakpaw/src/models"
)

type GetChatRoomsResponse struct {
	ChatRooms []ChatRoomItem `json:"chatRooms"`
}

type ChatRoomItem struct {
	ID            int              `json:"id"`
	LastMessageAt *string          `json:"lastMessageAt"`
	LastMessage   *LastMessageItem `json:"lastMessage"`
	OtherUser     UserSummary      `json:"otherUser"`
}

type LastMessageItem struct {
	ID       int     `json:"id"`
	Content  string  `json:"content"`
	ReadAt   *string `json:"readAt"`
	SenderID *int    `json:"senderId"`
}

type UserSummary struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Avatar    string `json:"avatar"`
}

func (r *GetChatRoomsResponse) Make(rooms []models.ChatRoom, authUserID int) *GetChatRoomsResponse {
	r.ChatRooms = make([]ChatRoomItem, len(rooms))

	for i, room := range rooms {
		item := ChatRoomItem{
			ID: room.ID,
		}

		if room.LastMessageAt != nil {
			t := room.LastMessageAt.Format("2006-01-02 15:04:05")
			item.LastMessageAt = &t
		}

		if room.LastMessage != nil {
			lastMsg := &LastMessageItem{
				ID:       room.LastMessage.ID,
				Content:  room.LastMessage.Content,
				SenderID: room.LastMessage.SenderID,
			}
			if room.LastMessage.ReadAt != nil {
				t := room.LastMessage.ReadAt.Format("2006-01-02 15:04:05")
				lastMsg.ReadAt = &t
			}
			item.LastMessage = lastMsg
		}

		// Show the other participant (not the authenticated user)
		if room.Participant1ID != nil && *room.Participant1ID == authUserID {
			if room.Participant2 != nil {
				item.OtherUser = UserSummary{
					ID:        room.Participant2.ID,
					FirstName: room.Participant2.FirstName,
					LastName:  room.Participant2.LastName,
					Avatar:    room.Participant2.Avatar,
				}
			}
		} else {
			if room.Participant1 != nil {
				item.OtherUser = UserSummary{
					ID:        room.Participant1.ID,
					FirstName: room.Participant1.FirstName,
					LastName:  room.Participant1.LastName,
					Avatar:    room.Participant1.Avatar,
				}
			}
		}

		r.ChatRooms[i] = item
	}

	return r
}
