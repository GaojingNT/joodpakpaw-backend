package responses

import (
	"joodpakpaw/src/models"
)

type GetBookingsForHostResponse struct {
	Bookings []BookingForHostItem `json:"bookings"`
}

type BookingForHostItem struct {
	ID            int                  `json:"id"`
	CheckInDate   string               `json:"checkInDate"`
	CheckOutDate  string               `json:"checkOutDate"`
	NumberOfPets  int                  `json:"numberOfPets"`
	TotalPrice    float64              `json:"totalPrice"`
	BookingStatus models.BookingStatus `json:"bookingStatus"`
	Rating        *int                 `json:"rating,omitempty"`
	Guest         *GuestInfo           `json:"guest,omitempty"`
	GuestNote     string               `json:"guestNote,omitempty"`
	Property      *PropertySummary     `json:"property,omitempty"`
	CreatedAt     string               `json:"createdAt"`
}

type GuestInfo struct {
	UserID    int    `json:"userId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (r *GetBookingsForHostResponse) Make(bookings []models.Booking) *GetBookingsForHostResponse {
	r.Bookings = make([]BookingForHostItem, len(bookings))

	for i, booking := range bookings {
		bookingItem := BookingForHostItem{
			ID:            booking.ID,
			CheckInDate:   booking.CheckInDate.Format("2006-01-02"),
			CheckOutDate:  booking.CheckOutDate.Format("2006-01-02"),
			NumberOfPets:  booking.NumberOfPets,
			TotalPrice:    booking.TotalPrice,
			BookingStatus: booking.BookingStatus,
			Rating:        booking.Rating,
			CreatedAt:     booking.CreatedAt.Format("2006-01-02 15:04:05"),
			GuestNote:     booking.GuestNote,
		}

		// Add guest information
		if booking.Guest != nil {
			bookingItem.Guest = &GuestInfo{
				UserID:    booking.Guest.ID,
				FirstName: booking.Guest.FirstName,
				LastName:  booking.Guest.LastName,
			}
		}

		// Get property and room categories from booking orders
		if len(booking.BookingOrders) > 0 {
			// Set property info from first order
			firstOrder := booking.BookingOrders[0]
			if firstOrder.Property != nil && firstOrder.PropertyID != nil {
				// Collect unique room categories and sum rooms booked from all orders
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

				// Create property with room categories inside
				bookingItem.Property = &PropertySummary{
					ID:             *firstOrder.PropertyID,
					Title:          firstOrder.Property.Title,
					Province:       firstOrder.Property.Province,
					RoomCategories: roomCategories,
				}
			}
		}

		r.Bookings[i] = bookingItem
	}

	return r
}
