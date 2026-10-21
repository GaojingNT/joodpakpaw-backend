package models

import "time"

type ChatMessageType string



type ChatMessage struct {
	BaseModel

	// Fields
	Content     string          `json:"content" gorm:"column:content;type:text;not null;"`
	IsDeleted   bool            `json:"isDeleted" gorm:"column:is_deleted;default:false;"`
	ReadAt      *time.Time      `json:"readAt" gorm:"column:read_at;type:timestamp;"`

	// Relationships
	RoomID      *int                     `json:"roomId" gorm:"column:room_id;not null;index;"`
	Room        *ChatRoom                `json:"room,omitempty" gorm:"foreignKey:RoomID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SenderID    *int                     `json:"senderId" gorm:"column:sender_id;not null;index;"`
	Sender      *User                    `json:"sender,omitempty" gorm:"foreignKey:SenderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Attachments []ChatMessageAttachment  `json:"attachments,omitempty" gorm:"foreignKey:MessageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
