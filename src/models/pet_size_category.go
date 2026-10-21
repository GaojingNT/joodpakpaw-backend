package models

import (
	"gorm.io/gorm"
)

type PetSizeCategory struct {
	BaseModel
	Name string `json:"name" gorm:"column:name;size:100;not null;"`
	Code int    `json:"code" gorm:"column:code;not null;uniqueIndex;"`
	Min  int    `json:"min" gorm:"column:min;not null;"`
	Max  int    `json:"max" gorm:"column:max;not null;"`
	//relationships
	RoomCategories []RoomCategory `json:"roomCategories,omitempty" gorm:"foreignKey:PetSizeCategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Profiles       []PetProfile   `json:"profiles,omitempty" gorm:"foreignKey:PetSizeCategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
