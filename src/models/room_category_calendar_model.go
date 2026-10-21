package models

import (
	"time"

	"gorm.io/gorm"
)

type RoomCategoryCalendar struct {
	BaseModel

	// Fields
	Date              time.Time `json:"date" gorm:"column:date;type:date;not null;"`
	AvailableRooms    int       `json:"availableRooms" gorm:"column:available_rooms;not null;"`
	BookedRooms       int       `json:"bookedRooms" gorm:"column:booked_rooms;default:0;"`
	UnavailableRooms  int       `json:"unavailableRooms" gorm:"column:unavailable_rooms;default:0;"`
	PriceAdjustment   float64   `json:"priceAdjustment" gorm:"column:price_adjustment;type:decimal(13,2);default:0;"`

	// Relationships
	RoomCategoryID *int          `json:"roomCategoryId" gorm:"column:room_category_id;not null;"`
	RoomCategory   *RoomCategory `json:"roomCategory,omitempty" gorm:"foreignKey:RoomCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BookingOrders  []BookingOrder `json:"bookingOrders,omitempty" gorm:"foreignKey:RoomCategoryCalendarID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
