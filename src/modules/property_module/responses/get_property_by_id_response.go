package responses

import (
	"joodpakpaw/internal/http_response"
	"joodpakpaw/src/models"
)

type (
	GetPropertyByIDResponse struct {
		http_response.OkResponse
		Property PropertyResponse `json:"property"`
	}
	PropertyResponse struct {
		ID         int                            `json:"id"`
		ChatRoomID *int                           `json:"chatRoomId"`
		Title      string                         `json:"title"`
		Description    string                         `json:"description"`
		Address        string                         `json:"address"`
		Province       string                         `json:"province"`
		District       string                         `json:"district"`
		Country        string                         `json:"country"`
		Zipcode        string                         `json:"zipcode"`
		Latitude       float64                        `json:"latitude"`
		Longitude      float64                        `json:"longitude"`
		IsActive        bool                           `json:"isActive"`
		IsBannedByAdmin bool                           `json:"isBannedByAdmin"`
		AvgRating      float64                        `json:"avgRating"`
		ReviewCount    int                            `json:"reviewCount"`
		TotalRooms     int                            `json:"totalRoomCategories"`
		HostID         *int                           `json:"hostId,omitempty"`
		Host           *HostProfileResponse           `json:"host,omitempty"`
		HostAvgRating  float64                        `json:"hostAvgRating"`
		PropertyTypeID *int                           `json:"propertyTypeId,omitempty"`
		RoomCategories []RoomCategoryDetailResponse   `json:"roomCategories"`
		Facilities     []FacilityDetailResponse       `json:"facilities"`
		Images         []PropertyImageDetailResponse  `json:"images"`
		Reviews        []ReviewResponse               `json:"reviews"`
	}
	RoomCategoryDetailResponse struct {
		ID           int                              `json:"id"`
		Name         string                           `json:"name"`
		BasePrice    float64                          `json:"basePrice"`
		NumberOfPets int                              `json:"numberOfPets"`
		TotalRoom    int                              `json:"totalRoom"`
		Images       []RoomCategoryImageDetailResponse `json:"images"`
		Pets         []PetDetailResponse              `json:"pets"`
		Facilities   []FacilityDetailResponse         `json:"facilities"`
	}
	RoomCategoryImageDetailResponse struct {
		ID                int    `json:"id"`
		FileName          string `json:"fileName"`
		FilePath          string `json:"filePath"`
		FileContentType   string `json:"fileContentType"`
		FileExtensionName string `json:"fileExtensionName"`
		FileSize          int    `json:"fileSize"`
	}
	PetDetailResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	FacilityDetailResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	PropertyImageDetailResponse struct {
		ID                int    `json:"id"`
		FileName          string `json:"fileName"`
		FilePath          string `json:"filePath"`
		FileContentType   string `json:"fileContentType"`
		FileExtensionName string `json:"fileExtensionName"`
		FileSize          int    `json:"fileSize"`
	}
	ReviewResponse struct {
		BookingID    int                        `json:"bookingId"`
		FirstName    string                     `json:"firstName"`
		LastName     string                     `json:"lastName"`
		Avatar       string                     `json:"avatar"`
		Comment      string                     `json:"comment"`
		Rating       int                        `json:"rating"`
		ReviewImages []ReviewImageResponse      `json:"reviewImages"`
	}
	ReviewImageResponse struct {
		ID                int    `json:"id"`
		FileName          string `json:"fileName"`
		FilePath          string `json:"filePath"`
		FileContentType   string `json:"fileContentType"`
		FileExtensionName string `json:"fileExtensionName"`
		FileSize          int    `json:"fileSize"`
	}
	HostProfileResponse struct {
		ID             int    `json:"id"`
		FirstName      string `json:"firstName"`
		LastName       string `json:"lastName"`
		Avatar         string `json:"avatar"`
		IsVerifiedHost bool   `json:"isVerifiedHost"`
	}
)

func (response *GetPropertyByIDResponse) Make(property models.Property, hostAvgRating float64) *GetPropertyByIDResponse {
	// Map room categories
	roomCategories := make([]RoomCategoryDetailResponse, 0)
	for _, room := range property.RoomsCategory {
		// Map pets for this room
		pets := make([]PetDetailResponse, 0)
		for _, pet := range room.Pets {
			pets = append(pets, PetDetailResponse{
				ID:   pet.ID,
				Name: pet.Name,
			})
		}

		// Map facilities for this room
		roomFacilities := make([]FacilityDetailResponse, 0)
		for _, facility := range room.Facilities {
			roomFacilities = append(roomFacilities, FacilityDetailResponse{
				ID:   facility.ID,
				Name: facility.Name,
			})
		}

		roomImages := make([]RoomCategoryImageDetailResponse, 0, len(room.Images))
		for _, img := range room.Images {
			roomImages = append(roomImages, RoomCategoryImageDetailResponse{
				ID:                img.ID,
				FileName:          img.FileName,
				FilePath:          img.FilePath,
				FileContentType:   img.FileContentType,
				FileExtensionName: img.FileExtensionName,
				FileSize:          img.FileSize,
			})
		}

		roomCategories = append(roomCategories, RoomCategoryDetailResponse{
			ID:           room.ID,
			Name:         room.Name,
			BasePrice:    room.BasePrice,
			NumberOfPets: room.NumberOfPets,
			TotalRoom:    room.TotalRoom,
			Images:       roomImages,
			Pets:         pets,
			Facilities:   roomFacilities,
		})
	}

	// Map facilities
	facilities := make([]FacilityDetailResponse, 0)
	for _, facility := range property.Facilities {
		facilities = append(facilities, FacilityDetailResponse{
			ID:   facility.ID,
			Name: facility.Name,
		})
	}

	// Map images
	images := make([]PropertyImageDetailResponse, 0)
	for _, image := range property.Images {
		images = append(images, PropertyImageDetailResponse{
			ID:                image.ID,
			FileName:          image.FileName,
			FilePath:          image.FilePath,
			FileContentType:   image.FileContentType,
			FileExtensionName: image.FileExtensionName,
			FileSize:          image.FileSize,
		})
	}

	// Calculate average rating and review count from bookings through booking orders
	var avgRating float64
	var reviewCount int
	var totalRating int

	// Use a map to track unique bookings (avoid counting same booking multiple times)
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
			reviewCount++

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

	if reviewCount > 0 {
		avgRating = float64(totalRating) / float64(reviewCount)
	}

	// Map host profile
	var hostProfile *HostProfileResponse
	if property.Host != nil {
		hostProfile = &HostProfileResponse{
			ID:             property.Host.ID,
			FirstName:      property.Host.FirstName,
			LastName:       property.Host.LastName,
			Avatar:         property.Host.Avatar,
			IsVerifiedHost: property.Host.IsVerifiedHost,
		}
	}

	// Count total room categories
	totalRooms := len(property.RoomsCategory)

	response = &GetPropertyByIDResponse{
		OkResponse: http_response.OkResponse{Code: "OK", Message: "Success"},
		Property: PropertyResponse{
			ID:             property.ID,
			Title:          property.Title,
			Description:    property.Description,
			Address:        property.Address,
			Province:       property.Province,
			District:       property.District,
			Country:        property.Country,
			Zipcode:        property.Zipcode,
			Latitude:       property.Latitude,
			Longitude:      property.Longitude,
			IsActive:        property.IsActive,
			IsBannedByAdmin: property.IsBannedByAdmin,
			AvgRating:      avgRating,
			ReviewCount:    reviewCount,
			TotalRooms:     totalRooms,
			HostID:         property.HostID,
			Host:           hostProfile,
			HostAvgRating:  hostAvgRating,
			PropertyTypeID: property.PropertyTypeID,
			RoomCategories: roomCategories,
			Facilities:     facilities,
			Images:         images,
			Reviews:        reviews,
		},
	}

	return response
}
