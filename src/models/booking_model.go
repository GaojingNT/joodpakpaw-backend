package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusSuccessful BookingStatus = "successful"
)

type Booking struct {
	BaseModel

	// Fields
	CheckInDate   time.Time     `json:"checkInDate" gorm:"column:check_in_date;type:date;not null;"`
	CheckOutDate  time.Time     `json:"checkOutDate" gorm:"column:check_out_date;type:date;not null;"`
	NumberOfPets  int           `json:"numberOfPets" gorm:"column:number_of_pets;default:0;"`
	TotalPrice    float64       `json:"totalPrice" gorm:"column:total_price;type:decimal(13,2);"`
	BookingStatus BookingStatus `json:"bookingStatus" gorm:"column:booking_status;type:booking_status;default:'pending';"`
	GuestNote     string        `json:"guestNote" gorm:"column:guest_note;size:255;"`
	Rating        *int          `json:"rating" gorm:"column:rating;"`
	Comment       string        `json:"comment" gorm:"column:comment;size:500;"`
	SuccessedAt   *sql.NullTime `json:"successedAt" gorm:"column:successed_at;type:timestamp;"`
	ConfirmedAt   *sql.NullTime `json:"confirmedAt" gorm:"column:confirmed_at;type:timestamp;"`
	ReviewedAt    *sql.NullTime `json:"reviewedAt" gorm:"column:reviewed_at;type:timestamp;"`
	// Relationships
	GuestID       *int                 `json:"guestId" gorm:"column:guest_id;not null;"`
	Guest         *User                `json:"guest,omitempty" gorm:"foreignKey:GuestID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BookingOrders []BookingOrder       `json:"bookingOrders,omitempty" gorm:"foreignKey:BookingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PetProfiles   []PetProfile         `json:"petProfiles,omitempty" gorm:"many2many:booking_pet_profiles;constraint:OnDelete:CASCADE;"`
	ReviewImages  []BookingReviewImage `json:"reviewImages,omitempty" gorm:"foreignKey:BookingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
