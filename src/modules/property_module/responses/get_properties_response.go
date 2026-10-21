package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type (
	GetPropertiesCollection struct {
		Properties []PropertyListResponse `json:"properties"`
	}
	GetAllPropertiesResponse struct {
		Properties []PropertyListResponse `json:"properties"`
	}
	PropertyListResponse struct {
		ID             int                     `json:"id"`
		Title          string                  `json:"title"`
		Description    string                  `json:"description"`
		Province       string                  `json:"province"`
		District       string                  `json:"district"`
		Country        string                  `json:"country"`
		Latitude       float64                 `json:"latitude"`
		Longitude      float64                 `json:"longitude"`
		Zipcode        string                  `json:"zipcode"`
		IsActive                 bool                    `json:"isActive"`
		IsBannedByAdmin          bool                    `json:"isBannedByAdmin"`
		IsBannedBeforeAccountBan bool                    `json:"isBannedBeforeAccountBan"`
		IsFavorite               bool                    `json:"isFavorite"`
		AvgRating      float64                 `json:"avgRating"`
		BookingCount   int                     `json:"bookingCount"`
		TotalRooms     int                     `json:"totalRoomCategories"`
		HostID         *int                    `json:"hostId,omitempty"`
		PropertyTypeID *int                    `json:"propertyTypeId,omitempty"`
		Images         []PropertyImageResponse `json:"images"`
		RoomCategory   *RoomCategoryResponse   `json:"roomCategories,omitempty"`
		Pets           []PetResponse           `json:"pets"`
		CreatedAt      time.Time               `json:"createdAt"`
		PublishedAt    *time.Time              `json:"publishedAt"`
	}
	PropertyImageResponse struct {
		ID                int    `json:"id"`
		FileName          string `json:"fileName"`
		FilePath          string `json:"filePath"`
		FileContentType   string `json:"fileContentType"`
		FileExtensionName string `json:"fileExtensionName"`
		FileSize          int    `json:"fileSize"`
	}
	RoomCategoryResponse struct {
		ID        int     `json:"id"`
		Name      string  `json:"name"`
		TotalRoom int     `json:"totalRoom"`
		BasePrice float64 `json:"basePrice"`
	}
	PetResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (response *GetPropertiesCollection) Collection(properties []models.Property, favoritedPropertyIDs map[int]bool) []PropertyListResponse {
	propertyResponses := make([]PropertyListResponse, 0)

	for _, property := range properties {
		if property.IsBannedByAdmin {
			continue
		}
		// Map images
		images := make([]PropertyImageResponse, 0)
		for _, img := range property.Images {
			images = append(images, PropertyImageResponse{
				ID:                img.ID,
				FileName:          img.FileName,
				FilePath:          img.FilePath,
				FileContentType:   img.FileContentType,
				FileExtensionName: img.FileExtensionName,
				FileSize:          img.FileSize,
			})
		}

		// Calculate average rating and booking count from preloaded booking orders
		var avgRating float64
		var ratingCount int
		var totalRating int

		// Use a map to track unique bookings (deduplicate by booking ID)
		uniqueBookings := make(map[int]*models.Booking)
		for _, order := range property.BookingOrders {
			if order.Booking != nil {
				uniqueBookings[order.Booking.ID] = order.Booking
			}
		}

		// Calculate rating and booking count from unique bookings
		bookingCount := len(uniqueBookings)
		for _, booking := range uniqueBookings {
			if booking.Rating != nil {
				totalRating += *booking.Rating
				ratingCount++
			}
		}

		if ratingCount > 0 {
			avgRating = float64(totalRating) / float64(ratingCount)
		}

		// Get room category with lowest base price
		var roomCategory *RoomCategoryResponse
		if len(property.RoomsCategory) > 0 {
			cheapestRoom := property.RoomsCategory[0]
			for _, room := range property.RoomsCategory {
				if room.BasePrice < cheapestRoom.BasePrice {
					cheapestRoom = room
				}
			}
			roomCategory = &RoomCategoryResponse{
				ID:        cheapestRoom.ID,
				Name:      cheapestRoom.Name,
				TotalRoom: cheapestRoom.TotalRoom,
				BasePrice: cheapestRoom.BasePrice,
			}
		}

		// Collect all unique pets from all room categories
		petsMap := make(map[int]models.Pet)
		for _, room := range property.RoomsCategory {
			for _, pet := range room.Pets {
				petsMap[pet.ID] = pet
			}
		}
		pets := make([]PetResponse, 0)
		for _, pet := range petsMap {
			pets = append(pets, PetResponse{
				ID:   pet.ID,
				Name: pet.Name,
			})
		}

		// Count total room categories
		totalRooms := len(property.RoomsCategory)

		// Check if property is favorited
		isFavorite := false
		if favoritedPropertyIDs != nil {
			isFavorite = favoritedPropertyIDs[property.ID]
		}

		propertyResponses = append(propertyResponses, PropertyListResponse{
			ID:              property.ID,
			CreatedAt:       property.CreatedAt,
			PublishedAt:     property.PublishedAt,
			Title:           property.Title,
			Description:     property.Description,
			Province:        property.Province,
			District:        property.District,
			Country:         property.Country,
			Latitude:        property.Latitude,
			Longitude:       property.Longitude,
			Zipcode:         property.Zipcode,
			IsActive:                 property.IsActive,
			IsBannedByAdmin:          property.IsBannedByAdmin,
			IsBannedBeforeAccountBan: property.IsBannedBeforeAccountBan,
			IsFavorite:               isFavorite,
			AvgRating:      avgRating,
			BookingCount:   bookingCount,
			TotalRooms:     totalRooms,
			HostID:         property.HostID,
			PropertyTypeID: property.PropertyTypeID,
			Images:         images,
			RoomCategory:   roomCategory,
			Pets:           pets,
		})
	}

	return propertyResponses
}
