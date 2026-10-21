package models

import "gorm.io/gorm"

type PropertyType struct {
	BaseModel

	// Fields
	Name  string `json:"name" gorm:"column:name;size:100;not null;"`
	Label string `json:"label" gorm:"column:label;size:100;default:'';"`

	// Relationships
	Properties []Property `json:"properties,omitempty" gorm:"foreignKey:PropertyTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
