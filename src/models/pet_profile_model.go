package models

import (
	"time"

	"gorm.io/gorm"
)

type PetSex string

const (
	PetSexMale   PetSex = "male"
	PetSexFemale PetSex = "female"
	PetSexOther  PetSex = "other"
)

type PetProfile struct {
	BaseModel

	// Fields
	Name      string    `json:"name" gorm:"column:name;size:100;not null;"`
	Image     string    `json:"image" gorm:"column:image;type:text;"`
	Sex       PetSex    `json:"sex" gorm:"column:pet_sex;type:pet_sex;"`
	BirthDate time.Time `json:"birthDate" gorm:"column:birth_date;type:date;"`
	ContactName string    `json:"contactName" gorm:"column:contact_name;size:100;not null;"`
	ContactPhone string    `json:"contactPhone" gorm:"column:contact_phone;size:20;not null;"`
	Note      string    `json:"note" gorm:"column:note;type:text;"`
	

	// Relationships
	GuestID           *int              `json:"guestId" gorm:"column:guest_id;not null;"`
	Guest             *User             `json:"guest,omitempty" gorm:"foreignKey:GuestID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BreedID           *int              `json:"breedId" gorm:"column:breed_id;not null;"`
	Breed             *PetBreed         `json:"breed,omitempty" gorm:"foreignKey:BreedID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PetTypeID         *int              `json:"petTypeId" gorm:"column:pet_type_id;not null;"`
	PetType           *Pet              `json:"petType,omitempty" gorm:"foreignKey:PetTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PetEnergyID       *int              `json:"petEnergyId" gorm:"column:pet_energy_id;not null;"`
	PetEnergy         *PetProfileEnergy `json:"petEnergy,omitempty" gorm:"foreignKey:PetEnergyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PetSizeCategoryID *int              `json:"petSizeCategoryId" gorm:"column:pet_size_category_id;"`
	PetSizeCategory   *PetSizeCategory  `json:"petSizeCategory,omitempty" gorm:"foreignKey:PetSizeCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
