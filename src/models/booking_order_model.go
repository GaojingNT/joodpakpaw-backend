package models

import (
	"time"

	"gorm.io/gorm"
)

type BookingOrder struct {
	BaseModel

	// Fields
	CheckInDate  time.Time `json:"checkInDate" gorm:"column:check_in_date;type:date;not null;"`
	CheckOutDate time.Time `json:"checkOutDate" gorm:"column:check_out_date;type:date;not null;"`
	RoomsBooked  int       `json:"roomsBooked" gorm:"column:rooms_booked;not null;"`
	PricePerRoom float64   `json:"pricePerRoom" gorm:"column:price_per_room;type:decimal(13,2);not null;"`
	TotalPrice   float64   `json:"totalPrice" gorm:"column:total_price;type:decimal(13,2);not null;"`

	// Relationships
	RoomCategoryID         *int                  `json:"roomCategoryId" gorm:"column:room_category_id;not null;"`
	RoomCategory           *RoomCategory         `json:"roomCategory,omitempty" gorm:"foreignKey:RoomCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoomTypeID             *int                  `json:"roomTypeId" gorm:"column:room_type_id;"`
	RoomType               *RoomCategory         `json:"roomType,omitempty" gorm:"foreignKey:RoomTypeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PropertyID             *int                  `json:"propertyId" gorm:"column:property_id;not null;"`
	Property               *Property             `json:"property,omitempty" gorm:"foreignKey:PropertyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BookingID              *int                  `json:"bookingId" gorm:"column:booking_id;not null;"`
	Booking                *Booking              `json:"booking,omitempty" gorm:"foreignKey:BookingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RoomCategoryCalendarID *int                  `json:"roomCategoryCalendarId" gorm:"column:room_category_calendar_id;"`
	RoomCategoryCalendar   *RoomCategoryCalendar `json:"roomCategoryCalendar,omitempty" gorm:"foreignKey:RoomCategoryCalendarID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
