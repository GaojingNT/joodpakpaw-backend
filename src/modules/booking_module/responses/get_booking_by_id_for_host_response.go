package responses

import (
	"joodpakpaw/src/models"
)

type GetBookingByIDForHostResponse struct {
	ID            int                  `json:"id"`
	CheckInDate   string               `json:"checkInDate"`
	CheckOutDate  string               `json:"checkOutDate"`
	NumberOfPets  int                  `json:"numberOfPets"`
	TotalPrice    float64              `json:"totalPrice"`
	BookingStatus models.BookingStatus `json:"bookingStatus"`
	GuestNote     string               `json:"guestNote,omitempty"`
	Rating        *int                 `json:"rating,omitempty"`
	Comment       string               `json:"comment,omitempty"`
	SuccessedAt   string               `json:"successedAt,omitempty"`
	ConfirmedAt   string               `json:"confirmedAt,omitempty"`
	Guest         *GuestDetailInfo     `json:"guest,omitempty"`
	BookingOrders []BookingOrderDetail `json:"bookingOrders,omitempty"`
	PetProfiles   []PetProfileInfo     `json:"petProfiles,omitempty"`
	CreatedAt     string               `json:"createdAt"`
	UpdatedAt     string               `json:"updatedAt"`
}

type GuestDetailInfo struct {
	UserID    int    `json:"userId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
}

func (r *GetBookingByIDForHostResponse) Make(booking models.Booking) *GetBookingByIDForHostResponse {
	r.ID = booking.ID
	r.CheckInDate = booking.CheckInDate.Format("2006-01-02")
	r.CheckOutDate = booking.CheckOutDate.Format("2006-01-02")
	r.NumberOfPets = booking.NumberOfPets
	r.TotalPrice = booking.TotalPrice
	r.BookingStatus = booking.BookingStatus
	r.GuestNote = booking.GuestNote
	r.Rating = booking.Rating
	r.Comment = booking.Comment
	r.CreatedAt = booking.CreatedAt.Format("2006-01-02 15:04:05")
	r.UpdatedAt = booking.UpdatedAt.Format("2006-01-02 15:04:05")

	// Format nullable timestamps
	if booking.SuccessedAt != nil && booking.SuccessedAt.Valid {
		r.SuccessedAt = booking.SuccessedAt.Time.Format("2006-01-02 15:04:05")
	}
	if booking.ConfirmedAt != nil && booking.ConfirmedAt.Valid {
		r.ConfirmedAt = booking.ConfirmedAt.Time.Format("2006-01-02 15:04:05")
	}

	// Map guest info if available
	if booking.Guest != nil {
		r.Guest = &GuestDetailInfo{
			UserID:    booking.Guest.ID,
			FirstName: booking.Guest.FirstName,
			LastName:  booking.Guest.LastName,
			Email:     booking.Guest.Email,
			Phone:     booking.Guest.Phone,
		}
	}

	// Map pet profiles if available
	r.PetProfiles = mapPetProfiles(booking.PetProfiles)

	// Map booking orders if available
	if len(booking.BookingOrders) > 0 {
		r.BookingOrders = make([]BookingOrderDetail, len(booking.BookingOrders))
		for i, order := range booking.BookingOrders {
			r.BookingOrders[i] = BookingOrderDetail{
				ID:             order.ID,
				CheckInDate:    order.CheckInDate.Format("2006-01-02"),
				CheckOutDate:   order.CheckOutDate.Format("2006-01-02"),
				RoomsBooked:    order.RoomsBooked,
				PricePerRoom:   order.PricePerRoom,
				TotalPrice:     order.TotalPrice,
				PropertyID:     order.PropertyID,
				RoomCategoryID: order.RoomCategoryID,
			}

			// Map property if available
			if order.Property != nil {
				r.BookingOrders[i].Property = &PropertyInfo{
					ID:       order.Property.ID,
					Title:    order.Property.Title,
					Address:  order.Property.Address,
					District: order.Property.District,
					Province: order.Property.Province,
				}
			}

			// Map room category if available
			if order.RoomCategory != nil {
				r.BookingOrders[i].RoomCategory = &RoomCategoryInfo{
					ID:   order.RoomCategory.ID,
					Name: order.RoomCategory.Name,
				}
			}
		}
	}

	return r
}
