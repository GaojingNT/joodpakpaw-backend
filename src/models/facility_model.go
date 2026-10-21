package models

import "gorm.io/gorm"

type Facility struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:100;not null;"`

	// Relationships
	CategoryID *int              `json:"categoryId" gorm:"column:category_id;"`
	Category   *FacilityCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Properties []Property        `json:"properties,omitempty" gorm:"many2many:properties_has_facilities;constraint:OnDelete:CASCADE;"`
	RoomCategories []RoomCategory `json:"roomCategories,omitempty" gorm:"many2many:room_category_has_facilities;constraint:OnDelete:CASCADE;"`
	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}

// FacilitySearchable returns searchable columns
func FacilitySearchable() []string {
	return []string{"tbl_facilities.name"}
}

// FacilityCategorySearchable returns searchable columns from joined category table
func FacilityCategorySearchable() []string {
	return []string{"tbl_facility_categories.name"}
}
