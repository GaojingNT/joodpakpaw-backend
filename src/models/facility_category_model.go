package models

import "gorm.io/gorm"

type FacilityCategory struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:255;not null;"`

	// Relationships
	Facilities []Facility `json:"facilities,omitempty" gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
