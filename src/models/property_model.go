package models

import (
	"time"

	"gorm.io/gorm"
)

type Property struct {
	BaseModel
	// Fields
	Title       string  `json:"title" gorm:"column:title;size:255;not null;"`
	Description string  `json:"description" gorm:"column:description;type:text;"`
	Address     string  `json:"address" gorm:"column:address;type:text;"`
	Province    string  `json:"province" gorm:"column:province;size:100;"`
	District    string  `json:"district" gorm:"column:district;size:100;"`
	Subdistrict string  `json:"subdistrict" gorm:"column:subdistrict;size:100;"`
	Country     string  `json:"country" gorm:"column:country;size:100;"`
	Zipcode     string  `json:"zipcode" gorm:"column:zipcode;size:10;"`
	Latitude    float64 `json:"latitude" gorm:"column:latitude;"`
	Longitude   float64 `json:"longitude" gorm:"column:longitude;"`
	IsActive                 bool       `json:"isActive" gorm:"column:is_active;default:false;"`
	IsBannedByAdmin          bool       `json:"isBannedByAdmin" gorm:"column:is_banned_by_admin;default:false;"`
	IsBannedBeforeAccountBan bool       `json:"isBannedBeforeAccountBan" gorm:"column:is_banned_before_account_ban;default:false;"`
	ReportCount              int        `json:"reportCount" gorm:"column:report_count;default:0;"`
	PublishedAt              *time.Time `json:"publishedAt" gorm:"column:published_at;"`

	// Relationships
	HostID          *int             `json:"hostId" gorm:"column:host_id;not null;"`
	Host            *User            `json:"host,omitempty" gorm:"foreignKey:HostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PropertyType    *PropertyType    `json:"propertyType,omitempty" gorm:"foreignKey:PropertyTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PropertyTypeID  *int             `json:"propertyTypeId" gorm:"column:property_type_id;not null;"`
	Images          []PropertyImage  `json:"images,omitempty" gorm:"foreignKey:PropertyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Facilities      []Facility       `json:"facilities,omitempty" gorm:"many2many:properties_has_facilities;constraint:OnDelete:CASCADE;"`
	FavoriteFolders []FavoriteFolder `json:"favoriteFolders,omitempty" gorm:"many2many:favorite_properties;joinForeignKey:property_id;joinReferences:folder_id;constraint:OnDelete:CASCADE;"`
	RoomsCategory   []RoomCategory   `json:"rooms,omitempty" gorm:"foreignKey:PropertyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BookingOrders   []BookingOrder   `json:"bookingOrders,omitempty" gorm:"foreignKey:PropertyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
// Property searchable attributes
func PropertySearchable() []string {
	return []string{"title", "address", "province", "district", "subdistrict", "country"}
}

// PropertyPublicSearchable returns searchable fields for the public listing endpoint.
func PropertyPublicSearchable() []string {
	return []string{"title", "province", "district", "description"}
}
