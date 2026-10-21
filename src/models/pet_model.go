package models

import "gorm.io/gorm"

type Pet struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:50;not null;"` //cat or dog

	// Relationships
	Breeds     []PetBreed   `json:"breeds,omitempty" gorm:"foreignKey:PetID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Profiles   []PetProfile `json:"profiles,omitempty" gorm:"foreignKey:PetTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoomCategorys []RoomCategory `json:"roomCategories,omitempty" gorm:"many2many:room_has_pets;constraint:OnDelete:CASCADE;"`
	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
