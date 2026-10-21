package responses

import "joodpakpaw/src/models"

type (
	GetRoomCategoriesResponse struct {
		RoomCategories []RoomCategoryResponse `json:"roomCategories"`
	}
	RoomCategoryImageResponse struct {
		ID                int    `json:"id"`
		FileName          string `json:"fileName"`
		FilePath          string `json:"filePath"`
		FileContentType   string `json:"fileContentType"`
		FileExtensionName string `json:"fileExtensionName"`
		FileSize          int    `json:"fileSize"`
	}
	RoomCategoryResponse struct {
		ID                   int                         `json:"id"`
		Name                 string                      `json:"name"`
		TotalRoom            int                         `json:"totalRoom"`
		NumberOfPets         int                         `json:"numberOfPets"`
		BasePrice            float64                     `json:"basePrice"`
		PropertyID           *int                        `json:"propertyId,omitempty"`
		PetSizeCategoryName  *string                     `json:"petSizeCategoryName"`
		Images               []RoomCategoryImageResponse `json:"images"`
		Pets                 []PetResponse               `json:"pets,omitempty"`
	}
	PetResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (response *GetRoomCategoriesResponse) Collection(roomCategories []models.RoomCategory) []RoomCategoryResponse {
	roomCategoryResponses := make([]RoomCategoryResponse, 0)

	for _, roomCategory := range roomCategories {
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

		var petSizeCategoryName *string
		if roomCategory.PetSizeCategory != nil {
			petSizeCategoryName = &roomCategory.PetSizeCategory.Name
		}

		images := make([]RoomCategoryImageResponse, 0, len(roomCategory.Images))
		for _, img := range roomCategory.Images {
			images = append(images, RoomCategoryImageResponse{
				ID:                img.ID,
				FileName:          img.FileName,
				FilePath:          img.FilePath,
				FileContentType:   img.FileContentType,
				FileExtensionName: img.FileExtensionName,
				FileSize:          img.FileSize,
			})
		}

		roomCategoryResponses = append(roomCategoryResponses, RoomCategoryResponse{
			ID:                  roomCategory.ID,
			Name:                roomCategory.Name,
			TotalRoom:           roomCategory.TotalRoom,
			NumberOfPets:        roomCategory.NumberOfPets,
			BasePrice:           roomCategory.BasePrice,
			PropertyID:          roomCategory.PropertyID,
			PetSizeCategoryName: petSizeCategoryName,
			Images:              images,
			Pets:                pets,
		})
	}

	return roomCategoryResponses
}
