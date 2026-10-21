package models

import "gorm.io/gorm"

type RoomCategory struct {
	BaseModel

	// Fields
	Name      string  `json:"name" gorm:"column:name;size:255;"`
	TotalRoom int     `json:"totalRoom" gorm:"column:total_room;"`
	NumberOfPets int     `json:"numberOfPets" gorm:"column:number_of_pets;"`
	BasePrice float64 `json:"basePrice" gorm:"column:base_price;type:decimal(13,2);"`

	// Relationships
	PropertyID        *int                   `json:"propertyId" gorm:"column:property_id;"`
	Property          *Property              `json:"property,omitempty" gorm:"foreignKey:PropertyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PetSizeCategoryID *int                   `json:"petSizeCategoryId" gorm:"column:pet_size_category_id;"`
	PetSizeCategory   *PetSizeCategory       `json:"petSizeCategory,omitempty" gorm:"foreignKey:PetSizeCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Calendars         []RoomCategoryCalendar `json:"calendars,omitempty" gorm:"foreignKey:RoomCategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BookingOrders     []BookingOrder         `json:"bookingOrders,omitempty" gorm:"foreignKey:RoomTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Images            []RoomCategoryImage    `json:"images,omitempty" gorm:"foreignKey:RoomCategoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Pets              []Pet                  `json:"pets,omitempty" gorm:"many2many:room_has_pets;constraint:OnDelete:CASCADE;"`
	Facilities        []Facility             `json:"facilities,omitempty" gorm:"many2many:room_category_has_facilities;constraint:OnDelete:CASCADE;"`
	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
