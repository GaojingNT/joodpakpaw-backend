package responses

import (
	"joodpakpaw/src/models"
)

type CreateBookingResponse struct {
	ID            int                      `json:"id"`
	CheckInDate   string                   `json:"checkInDate"`
	CheckOutDate  string                   `json:"checkOutDate"`
	NumberOfPets  int                      `json:"numberOfPets"`
	TotalPrice    float64                  `json:"totalPrice"`
	BookingStatus models.BookingStatus     `json:"bookingStatus"`
	GuestID       *int                     `json:"guestId"`
	GuestNote     string                   `json:"guestNote,omitempty"`
	BookingOrders []BookingOrderResponse   `json:"bookingOrders,omitempty"`
	PetProfiles   []PetProfileInfo         `json:"petProfiles,omitempty"`
	CreatedAt     string                   `json:"createdAt"`
	UpdatedAt     string                   `json:"updatedAt"`
}

type BookingOrderResponse struct {
	ID                     int     `json:"id"`
	CheckInDate            string  `json:"checkInDate"`
	CheckOutDate           string  `json:"checkOutDate"`
	RoomsBooked            int     `json:"roomsBooked"`
	PricePerRoom           float64 `json:"pricePerRoom"`
	TotalPrice             float64 `json:"totalPrice"`
	RoomCategoryID         *int    `json:"roomCategoryId"`
	RoomTypeID             *int    `json:"roomTypeId"`
	PropertyID             *int    `json:"propertyId"`
	RoomCategoryCalendarID *int    `json:"roomCategoryCalendarId"`
}

func (r *CreateBookingResponse) Make(booking models.Booking) *CreateBookingResponse {
	r.ID = booking.ID
	r.CheckInDate = booking.CheckInDate.Format("2006-01-02")
	r.CheckOutDate = booking.CheckOutDate.Format("2006-01-02")
	r.NumberOfPets = booking.NumberOfPets
	r.TotalPrice = booking.TotalPrice
	r.BookingStatus = booking.BookingStatus
	r.GuestNote = booking.GuestNote
	r.GuestID = booking.GuestID
	r.CreatedAt = booking.CreatedAt.Format("2006-01-02 15:04:05")
	r.UpdatedAt = booking.UpdatedAt.Format("2006-01-02 15:04:05")

	// Map pet profiles if available
	r.PetProfiles = mapPetProfiles(booking.PetProfiles)

	// Map booking orders if available
	if len(booking.BookingOrders) > 0 {
		r.BookingOrders = make([]BookingOrderResponse, len(booking.BookingOrders))
		for i, order := range booking.BookingOrders {
			r.BookingOrders[i] = BookingOrderResponse{
				ID:                     order.ID,
				CheckInDate:            order.CheckInDate.Format("2006-01-02"),
				CheckOutDate:           order.CheckOutDate.Format("2006-01-02"),
				RoomsBooked:            order.RoomsBooked,
				PricePerRoom:           order.PricePerRoom,
				TotalPrice:             order.TotalPrice,
				RoomCategoryID:         order.RoomCategoryID,
				PropertyID:             order.PropertyID,
				RoomCategoryCalendarID: order.RoomCategoryCalendarID,
			}
		}
	}

	return r
}
