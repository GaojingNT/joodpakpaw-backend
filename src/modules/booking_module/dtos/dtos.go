package dtos

import "mime/multipart"

type (
	// CreateBooking DTO for creating a new booking
	CreateBooking struct {
		CheckInDate   string               `json:"checkInDate" form:"checkInDate" validate:"omitempty"`
		CheckOutDate  string               `json:"checkOutDate" form:"checkOutDate" validate:"omitempty"`
		GuestID       *int                 `json:"guestId" validate:"required"`
		GuestNote     string               `json:"guestNote" validate:"omitempty,max=255"`
		BookingOrders []CreateBookingOrder `json:"bookingOrders" validate:"required,min=1,dive"`
		PetProfileIDs []int                `json:"petProfileIds" validate:"omitempty,dive,min=1"`
		NumberOfPets  *int                 `json:"numberOfPets" validate:"omitempty,min=0"`
	}

	// CreateBookingOrder DTO for booking order within a booking
	CreateBookingOrder struct {
		CheckInDate    string `json:"checkInDate" validate:"required"`
		CheckOutDate   string `json:"checkOutDate" validate:"required"`
		RoomsBooked    int    `json:"roomsBooked" validate:"required,min=1"`
		RoomCategoryID *int   `json:"roomCategoryId" validate:"required"`
		RoomTypeID     *int   `json:"roomTypeId" validate:"omitempty"`
	}

	// UpdateBookingStatus DTO for updating booking status
	UpdateBookingStatus struct {
		BookingStatus string `json:"bookingStatus" validate:"required,oneof=pending confirmed cancelled successful"`
	}

	// CreateBookingReview DTO for creating a booking review
	CreateBookingReview struct {
		Rating  int                     `json:"rating" form:"rating" validate:"required,min=1,max=5"`
		Comment string                  `json:"comment" form:"comment" validate:"omitempty,max=500"`
		Images  []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty,dive"`
	}
)
