package models

import "gorm.io/gorm"

type PetBreed struct {
	BaseModel

	// Fields
	Breed    string  `json:"breed" gorm:"column:breed;size:100;not null;"`
	ImageURL *string `json:"imageUrl" gorm:"column:image_url;type:text;"`

	// Relationships
	PetID      *int         `json:"petId" gorm:"column:pet_id;not null;"`
	Pet        *Pet         `json:"pet,omitempty" gorm:"foreignKey:PetID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Profiles   []PetProfile `json:"profiles,omitempty" gorm:"foreignKey:BreedID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
