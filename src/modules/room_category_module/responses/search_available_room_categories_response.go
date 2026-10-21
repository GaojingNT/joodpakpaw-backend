package responses

import "joodpakpaw/src/models"

type (
	SearchAvailableRoomCategoriesResponse struct {
		RoomCategories []AvailableRoomCategoryResponse `json:"roomCategories"`
	}

	AvailableRoomCategoryResponse struct {
		ID                     int                      `json:"id"`
		Name                   string                   `json:"name"`
		NumberOfPets           int                      `json:"numberOfPets"`
		TotalRoom              int                      `json:"totalRoom"`
		PetSizeCategory        *PetSizeCategoryResponse `json:"petSizeCategory,omitempty"`
		IsAvailable            bool                     `json:"isAvailable"`
		IsCapacityInsufficient bool                     `json:"isCapacityInsufficient"`
		IsBooked               bool                     `json:"isBooked"`
		Availability           AvailabilityResponse     `json:"availability"`
		PropertyResponse       *PropertyResponse        `json:"property,omitempty"`
		Pets                   []PetResponse            `json:"pets,omitempty"`
		Facilities             []FacilityResponse       `json:"facilities,omitempty"`
	}

	AvailabilityResponse struct {
		RoomsNeeded    int                `json:"roomsNeeded"`
		AvailableRooms int                `json:"availableRooms"`
		Dates          []DateAvailability `json:"dates"`
		TotalPrice     float64            `json:"totalPrice"`
	}
	PropertyResponse struct {
		ID      int    `json:"id"`
		Title    string `json:"title"`
		Address string `json:"address"`
	}

	DateAvailability struct {
		Date            string  `json:"date"`
		AvailableRooms  int     `json:"availableRooms"`
		PriceAdjustment float64 `json:"priceAdjustment"`
		FinalPrice      float64 `json:"finalPrice"`
	}
)

func (response *SearchAvailableRoomCategoriesResponse) Collection(
	roomCategories []models.RoomCategory,
	calendarsMap map[int][]models.RoomCategoryCalendar,
	numberOfPets int,
) []AvailableRoomCategoryResponse {
	roomCategoryResponses := make([]AvailableRoomCategoryResponse, 0)

	for _, roomCategory := range roomCategories {
		// Get calendars for this room category
		calendars := calendarsMap[roomCategory.ID]

		// Calculate rooms needed for booking
		roomsNeeded := (numberOfPets + roomCategory.NumberOfPets - 1) / roomCategory.NumberOfPets

		// Check capacity: total capacity = TotalRoom * NumberOfPets per room
		totalCapacity := roomCategory.TotalRoom * roomCategory.NumberOfPets
		isCapacityInsufficient := numberOfPets > totalCapacity

		// Calculate availability
		minAvailableRooms := roomCategory.TotalRoom
		totalPrice := 0.0
		dateAvailabilities := make([]DateAvailability, 0)
		isBooked := false

		for _, calendar := range calendars {
			if calendar.AvailableRooms < minAvailableRooms {
				minAvailableRooms = calendar.AvailableRooms
			}

			// Check if this date has insufficient available rooms
			if calendar.AvailableRooms < roomsNeeded {
				isBooked = true
			}

			finalPrice := roomCategory.BasePrice + calendar.PriceAdjustment
			totalPrice += finalPrice * float64(roomsNeeded)

			dateAvailabilities = append(dateAvailabilities, DateAvailability{
				Date:            calendar.Date.Format("2006-01-02"),
				AvailableRooms:  calendar.AvailableRooms,
				PriceAdjustment: calendar.PriceAdjustment,
				FinalPrice:      finalPrice,
			})
		}

		// Determine if room category is available
		isAvailable := !isCapacityInsufficient && !isBooked && len(calendars) > 0

		// Build pets
		var pets []PetResponse
		if len(roomCategory.Pets) > 0 {
			pets = make([]PetResponse, 0, len(roomCategory.Pets))
			for _, pet := range roomCategory.Pets {
				pets = append(pets, PetResponse{
					ID:   pet.ID,
					Name: pet.Name,
				})
			}
		}

		// Build facilities
		var facilities []FacilityResponse
		if len(roomCategory.Facilities) > 0 {
			facilities = make([]FacilityResponse, 0, len(roomCategory.Facilities))
			for _, facility := range roomCategory.Facilities {
				facilities = append(facilities, FacilityResponse{
					ID:   facility.ID,
					Name: facility.Name,
				})
			}
		}

		// Build pet size category
		var petSizeCategory *PetSizeCategoryResponse
		if roomCategory.PetSizeCategory != nil {
			petSizeCategory = &PetSizeCategoryResponse{
				ID:   roomCategory.PetSizeCategory.ID,
				Name: roomCategory.PetSizeCategory.Name,
				Code: roomCategory.PetSizeCategory.Code,
				Min:  roomCategory.PetSizeCategory.Min,
				Max:  roomCategory.PetSizeCategory.Max,
			}
		}

	// Build property
	var property *PropertyResponse
	if roomCategory.Property != nil {
		property = &PropertyResponse{
			ID:      roomCategory.Property.ID,
			Title:   roomCategory.Property.Title,
			Address: roomCategory.Property.Address,
		}
	}

		roomCategoryResponses = append(roomCategoryResponses, AvailableRoomCategoryResponse{
			ID:                     roomCategory.ID,
			Name:                   roomCategory.Name,
			NumberOfPets:           roomCategory.NumberOfPets,
			TotalRoom:              roomCategory.TotalRoom,
			PetSizeCategory:        petSizeCategory,
			IsAvailable:            isAvailable,
			IsCapacityInsufficient: isCapacityInsufficient,
			IsBooked:               isBooked,
			Availability: AvailabilityResponse{
				RoomsNeeded:    roomsNeeded,
				AvailableRooms: minAvailableRooms,
				Dates:          dateAvailabilities,
				TotalPrice:     totalPrice,
			},
			PropertyResponse: property,
			Pets:       pets,
			Facilities: facilities,
		})
	}

	return roomCategoryResponses
}
