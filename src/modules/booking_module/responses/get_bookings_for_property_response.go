package responses

import (
	"joodpakpaw/src/models"
)

type GetBookingsForPropertyResponse struct {
	Bookings []BookingForPropertyItem `json:"bookings"`
}

type BookingForPropertyItem struct {
	ID            int                  `json:"id"`
	CheckInDate   string               `json:"checkInDate"`
	CheckOutDate  string               `json:"checkOutDate"`
	NumberOfPets  int                  `json:"numberOfPets"`
	TotalPrice    float64              `json:"totalPrice"`
	BookingStatus models.BookingStatus `json:"bookingStatus"`
	Guest         *GuestInfo           `json:"guest,omitempty"`
	RoomCategories []RoomCategorySummary `json:"roomCategories,omitempty"`
	GuestNote     string               `json:"guestNote,omitempty"`
	CreatedAt     string               `json:"createdAt"`
}

func (r *GetBookingsForPropertyResponse) Make(bookings []models.Booking) *GetBookingsForPropertyResponse {
	r.Bookings = make([]BookingForPropertyItem, len(bookings))

	for i, booking := range bookings {
		bookingItem := BookingForPropertyItem{
			ID:            booking.ID,
			CheckInDate:   booking.CheckInDate.Format("2006-01-02"),
			CheckOutDate:  booking.CheckOutDate.Format("2006-01-02"),
			NumberOfPets:  booking.NumberOfPets,
			TotalPrice:    booking.TotalPrice,
			BookingStatus: booking.BookingStatus,
			GuestNote:     booking.GuestNote,
			CreatedAt:     booking.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// Add guest information
		if booking.Guest != nil {
			bookingItem.Guest = &GuestInfo{
				UserID:    booking.Guest.ID,
				FirstName: booking.Guest.FirstName,
				LastName:  booking.Guest.LastName,
			}
		}

		// Get room categories from booking orders
		if len(booking.BookingOrders) > 0 {
			// Collect unique room categories and sum rooms booked
			roomCategoryMap := make(map[int]RoomCategorySummary)
			for _, order := range booking.BookingOrders {
				if order.RoomCategory != nil && order.RoomCategoryID != nil {
					// If category exists, add to existing rooms booked
					if existing, ok := roomCategoryMap[*order.RoomCategoryID]; ok {
						existing.RoomsBooked += order.RoomsBooked
						roomCategoryMap[*order.RoomCategoryID] = existing
					} else {
						// Create new entry
						roomCategoryMap[*order.RoomCategoryID] = RoomCategorySummary{
							ID:          *order.RoomCategoryID,
							Name:        order.RoomCategory.Name,
							RoomsBooked: order.RoomsBooked,
						}
					}
				}
			}

			// Convert map to slice
			roomCategories := make([]RoomCategorySummary, 0, len(roomCategoryMap))
			for _, roomCategory := range roomCategoryMap {
				roomCategories = append(roomCategories, roomCategory)
			}

			bookingItem.RoomCategories = roomCategories
		}

		r.Bookings[i] = bookingItem
	}

	return r
}
