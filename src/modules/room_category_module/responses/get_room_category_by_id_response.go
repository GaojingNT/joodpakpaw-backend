package responses

import "joodpakpaw/src/models"

type GetRoomCategoryByIDResponse struct {
	ID              int                         `json:"id"`
	Name            string                      `json:"name"`
	TotalRoom       int                         `json:"totalRoom"`
	NumberOfPets    int                         `json:"numberOfPets"`
	BasePrice       float64                     `json:"basePrice"`
	PropertyID      *int                        `json:"propertyId,omitempty"`
	PetSizeCategory *PetSizeCategoryResponse    `json:"petSizeCategory,omitempty"`
	Images          []RoomCategoryImageResponse `json:"images"`
	Pets            []PetResponse               `json:"pets,omitempty"`
	Facilities      []FacilityResponse          `json:"facilities,omitempty"`
}

type PetSizeCategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code int    `json:"code"`
	Min  int    `json:"min"`
	Max  int    `json:"max"`
}

type FacilityResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (response *GetRoomCategoryByIDResponse) Make(roomCategory models.RoomCategory) *GetRoomCategoryByIDResponse {
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

	return &GetRoomCategoryByIDResponse{
		ID:              roomCategory.ID,
		Name:            roomCategory.Name,
		TotalRoom:       roomCategory.TotalRoom,
		NumberOfPets:    roomCategory.NumberOfPets,
		BasePrice:       roomCategory.BasePrice,
		PropertyID:      roomCategory.PropertyID,
		PetSizeCategory: petSizeCategory,
		Images:          images,
		Pets:            pets,
		Facilities:      facilities,
	}
}
