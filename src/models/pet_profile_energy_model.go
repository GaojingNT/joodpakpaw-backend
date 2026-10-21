package models

type PetProfileEnergy struct {
	BaseModel

	// Fields
	Rating      int    `json:"rating" gorm:"column:rating;not null;"`
	Description string `json:"description" gorm:"column:description;type:text;"`

	// Relationships
	Profiles []PetProfile `json:"profiles,omitempty" gorm:"foreignKey:PetEnergyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
