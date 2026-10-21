package models

import "time"

type UserRecentView struct {
	ID int `json:"id" gorm:"column:id;primarykey;autoIncrement;"`

	// Fields
	ViewedAt time.Time `json:"viewedAt" gorm:"column:viewed_at;default:CURRENT_TIMESTAMP;not null;index;"`

	// Relations
	UserID     int       `json:"userId" gorm:"column:user_id;not null;index;"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PropertyID int       `json:"propertyId" gorm:"column:property_id;not null;index;"`
	Property   *Property `json:"property,omitempty" gorm:"foreignKey:PropertyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}


