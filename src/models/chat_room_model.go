package models

import "time"

type ChatRoom struct {
	BaseModel

	// Fields
	LastMessageAt *time.Time `json:"lastMessageAt" gorm:"column:last_message_at;"`

	// Relationships
	// Convention: participant_1_id < participant_2_id (enforced in repository)
	Participant1ID *int          `json:"participant1Id" gorm:"column:participant_1_id;not null;uniqueIndex:idx_chat_room_pair;"`
	Participant1   *User         `json:"participant1,omitempty" gorm:"foreignKey:Participant1ID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Participant2ID *int          `json:"participant2Id" gorm:"column:participant_2_id;not null;uniqueIndex:idx_chat_room_pair;"`
	Participant2   *User         `json:"participant2,omitempty" gorm:"foreignKey:Participant2ID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LastMessageID  *int          `json:"lastMessageId" gorm:"column:last_message_id;"`
	LastMessage    *ChatMessage  `json:"lastMessage,omitempty" gorm:"-"`
	Messages       []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:RoomID;"`
}
