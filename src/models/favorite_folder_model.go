package models

import "gorm.io/gorm"

type FavoriteFolder struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:255;not null;"`

	// Relationships
	UserID     *string    `json:"userId" gorm:"column:user_id;size:255;not null;"`
	User       *User      `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Properties []Property `json:"properties,omitempty" gorm:"many2many:favorite_properties;joinForeignKey:folder_id;joinReferences:property_id;constraint:OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
