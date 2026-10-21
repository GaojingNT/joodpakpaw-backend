package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type (
	GetPropertyForHostByIDResponse struct {
		Property PropertyForHostDetailResponse `json:"property"`
	}

	PropertyForHostDetailResponse struct {
		ID                          int                        `json:"id"`
		Title                       string                     `json:"title"`
		Description                 string                     `json:"description"`
		Province                    string                     `json:"province"`
		District                    string                     `json:"district"`
		Country                     string                     `json:"country"`
		Latitude                    float64                    `json:"latitude"`
		Longitude                   float64                    `json:"longitude"`
		IsActive                    bool                       `json:"isActive"`
		IsBannedByAdmin             bool                       `json:"isBannedByAdmin"`
		IsBannedBeforeAccountBan    bool                       `json:"isBannedBeforeAccountBan"`
		UnbanRequestCount           int64                      `json:"unbanRequestCount"`
		HasPendingUnbanRequest      bool                       `json:"hasPendingUnbanRequest"`
		AvgRating                   float64                    `json:"avgRating"`
		ReviewCount                 int                        `json:"reviewCount"`
		TotalRooms                  int                        `json:"totalRoomCategories"`
		Images                      []PropertyImageResponse    `json:"images"`
		Facilities                  []FacilityResponse         `json:"facilities"`
		RoomCategories              []RoomCategoryInfoResponse `json:"roomCategories"`
		Reviews                     []ReviewResponse           `json:"reviews"`
		PropertyType                *PropertyTypeResponse      `json:"propertyType,omitempty"`
		HostID                      *int                       `json:"hostId,omitempty"`
		Host                        *HostProfileResponse       `json:"host,omitempty"`
		HostAvgRating               float64                    `json:"hostAvgRating"`
		PropertyTypeID              *int                       `json:"propertyTypeId,omitempty"`
		PublishedAt                 *time.Time                 `json:"publishedAt"`
		CreatedAt                   string                     `json:"createdAt"`
		UpdatedAt                   string                     `json:"updatedAt"`
	}

	PropertyTypeResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	RoomCategoryInfoResponse struct {
		ID           int                           `json:"id"`
		Name         string                        `json:"name"`
		TotalRoom    int                           `json:"totalRoom"`
		BasePrice    float64                       `json:"basePrice"`
		NumberOfPets int                           `json:"numberOfPets"`
		Images       []RoomCategoryImageResponse   `json:"images"`
		Pets         []RoomPetResponse             `json:"pets"`
		Facilities   []FacilityResponse            `json:"facilities"`
	}
	RoomCategoryImageResponse struct {
		ID                int    `json:"id"`
		FileName          string `json:"fileName"`
		FilePath          string `json:"filePath"`
		FileContentType   string `json:"fileContentType"`
		FileExtensionName string `json:"fileExtensionName"`
		FileSize          int    `json:"fileSize"`
	}

	RoomPetResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	FacilityResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (r *GetPropertyForHostByIDResponse) Make(property models.Property, hostAvgRating float64, unbanCount int64, hasPendingUnban bool) *GetPropertyForHostByIDResponse {
	// Map images
	images := make([]PropertyImageResponse, 0, len(property.Images))
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

	// Calculate rating from unique bookings and build reviews
	reviews := make([]ReviewResponse, 0)
	for _, booking := range uniqueBookings {
		if booking.Rating != nil {
			totalRating += *booking.Rating
			ratingCount++

			reviewImages := make([]ReviewImageResponse, 0, len(booking.ReviewImages))
			for _, img := range booking.ReviewImages {
				reviewImages = append(reviewImages, ReviewImageResponse{
					ID:                img.ID,
					FileName:          img.FileName,
					FilePath:          img.FilePath,
					FileContentType:   img.FileContentType,
					FileExtensionName: img.FileExtensionName,
					FileSize:          img.FileSize,
				})
			}

			review := ReviewResponse{
				BookingID:    booking.ID,
				Comment:      booking.Comment,
				Rating:       *booking.Rating,
				ReviewImages: reviewImages,
			}
			if booking.Guest != nil {
				review.FirstName = booking.Guest.FirstName
				review.LastName = booking.Guest.LastName
				review.Avatar = booking.Guest.Avatar
			}
			reviews = append(reviews, review)
		}
	}
	if ratingCount > 0 {
		avgRating = float64(totalRating) / float64(ratingCount)
	}

	// Get property type if exists
	var propertyType *PropertyTypeResponse
	if property.PropertyType != nil {
		propertyType = &PropertyTypeResponse{
			ID:   property.PropertyType.ID,
			Name: property.PropertyType.Name,
		}
	}

	// Map facilities
	facilities := make([]FacilityResponse, 0, len(property.Facilities))
	for _, facility := range property.Facilities {
		facilities = append(facilities, FacilityResponse{
			ID:   facility.ID,
			Name: facility.Name,
		})
	}

	// Map room categories
	roomCategories := make([]RoomCategoryInfoResponse, 0, len(property.RoomsCategory))
	for _, room := range property.RoomsCategory {
		// Map pets for this room
		pets := make([]RoomPetResponse, 0, len(room.Pets))
		for _, pet := range room.Pets {
			pets = append(pets, RoomPetResponse{
				ID:   pet.ID,
				Name: pet.Name,
			})
		}

		// Map facilities for this room
		roomFacilities := make([]FacilityResponse, 0, len(room.Facilities))
		for _, facility := range room.Facilities {
			roomFacilities = append(roomFacilities, FacilityResponse{
				ID:   facility.ID,
				Name: facility.Name,
			})
		}

		roomImages := make([]RoomCategoryImageResponse, 0, len(room.Images))
		for _, img := range room.Images {
			roomImages = append(roomImages, RoomCategoryImageResponse{
				ID:                img.ID,
				FileName:          img.FileName,
				FilePath:          img.FilePath,
				FileContentType:   img.FileContentType,
				FileExtensionName: img.FileExtensionName,
				FileSize:          img.FileSize,
			})
		}

		roomCategories = append(roomCategories, RoomCategoryInfoResponse{
			ID:           room.ID,
			Name:         room.Name,
			BasePrice:    room.BasePrice,
			TotalRoom:    room.TotalRoom,
			NumberOfPets: room.NumberOfPets,
			Images:       roomImages,
			Pets:         pets,
			Facilities:   roomFacilities,
		})
	}

	// Map host profile
	var hostProfile *HostProfileResponse
	if property.Host != nil {
		hostProfile = &HostProfileResponse{
			ID:        property.Host.ID,
			FirstName: property.Host.FirstName,
			LastName:  property.Host.LastName,
			Avatar:    property.Host.Avatar,
		}
	}

	return &GetPropertyForHostByIDResponse{
		Property: PropertyForHostDetailResponse{
			ID:                     property.ID,
			Title:                  property.Title,
			Description:            property.Description,
			Province:               property.Province,
			District:               property.District,
			Country:                property.Country,
			Latitude:               property.Latitude,
			Longitude:              property.Longitude,
			IsActive:                  property.IsActive,
			IsBannedByAdmin:           property.IsBannedByAdmin,
			IsBannedBeforeAccountBan:  property.IsBannedBeforeAccountBan,
			UnbanRequestCount:         unbanCount,
			HasPendingUnbanRequest: hasPendingUnban,
			AvgRating:              avgRating,
			ReviewCount:            ratingCount,
			TotalRooms:             len(property.RoomsCategory),
			Images:                 images,
			Facilities:             facilities,
			RoomCategories:         roomCategories,
			Reviews:                reviews,
			PropertyType:           propertyType,
			HostID:                 property.HostID,
			Host:                   hostProfile,
			HostAvgRating:          hostAvgRating,
			PropertyTypeID:         property.PropertyTypeID,
			PublishedAt:            property.PublishedAt,
		},
	}
}
