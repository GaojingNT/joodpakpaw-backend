package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type (
	PropertiesForHostResponse struct {
		ID                       int                     `json:"id"`
		Title                    string                  `json:"title"`
		Province                 string                  `json:"province"`
		District                 string                  `json:"district"`
		AvgRating                float64                 `json:"avgRating"`
		IsActive                 bool                    `json:"isActive"`
		IsBannedByAdmin          bool                    `json:"isBannedByAdmin"`
		IsBannedBeforeAccountBan bool                    `json:"isBannedBeforeAccountBan"`
		UnbanRequestCount        int64                   `json:"unbanRequestCount"`
		HasPendingUnbanRequest   bool                    `json:"hasPendingUnbanRequest"`
		PublishedAt              *time.Time              `json:"publishedAt"`
		RoomCount                int                     `json:"roomCount"`
		TotalRooms               int                     `json:"totalRoomCategories"`
		Images                   []PropertyImageResponse `json:"images"`
		HostID                   *int                    `json:"hostId,omitempty"`
	}

	PropertiesForHostCollection struct{}

	GetPropertiesForHostResponse struct {
		Properties []PropertiesForHostResponse `json:"properties"`
	}
)

// Collection converts slice of Property models to PropertiesForHostResponse
func (r *PropertiesForHostCollection) Collection(properties []models.Property, unbanCounts map[int]int64, pendingUnban map[int]bool) []PropertiesForHostResponse {
	result := make([]PropertiesForHostResponse, 0, len(properties))

	for _, property := range properties {
		// Convert images
		images := make([]PropertyImageResponse, 0, len(property.Images))
		for _, img := range property.Images {
			images = append(images, PropertyImageResponse{
				ID:       img.ID,
				FilePath: img.FilePath,
			})
		}

		// Calculate average rating from bookings through booking orders
		var avgRating float64
		var ratingCount int
		var totalRating int

		// Use a map to track unique bookings
		uniqueBookings := make(map[int]*models.Booking)
		for _, order := range property.BookingOrders {
			if order.Booking != nil && order.Booking.Rating != nil {
				uniqueBookings[order.Booking.ID] = order.Booking
			}
		}

		// Calculate rating from unique bookings
		for _, booking := range uniqueBookings {
			if booking.Rating != nil {
				totalRating += *booking.Rating
				ratingCount++
			}
		}

		if ratingCount > 0 {
			avgRating = float64(totalRating) / float64(ratingCount)
		}

		// Count room categories (types of rooms)
		roomCount := len(property.RoomsCategory)
		totalRooms := len(property.RoomsCategory)

		result = append(result, PropertiesForHostResponse{
			ID:                     property.ID,
			Title:                  property.Title,
			Province:               property.Province,
			District:               property.District,
			AvgRating:              avgRating,
			IsActive:                  property.IsActive,
			IsBannedByAdmin:           property.IsBannedByAdmin,
			IsBannedBeforeAccountBan:  property.IsBannedBeforeAccountBan,
			UnbanRequestCount:         unbanCounts[property.ID],
			HasPendingUnbanRequest: pendingUnban[property.ID],
			PublishedAt:            property.PublishedAt,
			RoomCount:              roomCount,
			TotalRooms:             totalRooms,
			Images:                 images,
			HostID:                 property.HostID,
		})
	}

	return result
}
