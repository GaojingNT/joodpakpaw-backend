package dtos

import "mime/multipart"

type (
	CreateRoomCategory struct {
		Name              string                  `json:"name" form:"name" validate:"required,max=255"`
		TotalRoom         int                     `json:"totalRoom" form:"totalRoom" validate:"required,min=1"`
		NumberOfPets      int                     `json:"numberOfPets" form:"numberOfPets" validate:"required,omitempty,min=0"`
		BasePrice         float64                 `json:"basePrice" form:"basePrice" validate:"required,min=0"`
		PropertyID        *int                    `json:"-"`
		PetSizeCategoryID *int                    `json:"petSizeCategoryId" form:"petSizeCategoryId" validate:"omitempty"`
		PetsString        string                  `json:"-" form:"pets" validate:"omitempty"`
		Pets              []int                   `json:"pets" validate:"omitempty"`
		FacilitiesString  string                  `json:"-" form:"facilities" validate:"omitempty"`
		Facilities        []int                   `json:"facilities" validate:"omitempty"`
		Images            []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty,dive"`
	}

	UpdateRoomCategory struct {
		Name                 string                  `json:"name" form:"name" validate:"omitempty,max=255"`
		TotalRoom            int                     `json:"totalRoom" form:"totalRoom" validate:"omitempty,min=1"`
		NumberOfPets         int                     `json:"numberOfPets" form:"numberOfPets" validate:"omitempty,min=0"`
		BasePrice            float64                 `json:"basePrice" form:"basePrice" validate:"omitempty,min=0"`
		PropertyID           *int                    `json:"-"`
		PetSizeCategoryID    *int                    `json:"petSizeCategoryId" form:"petSizeCategoryId" validate:"omitempty"`
		PetsString           string                  `json:"-" form:"pets" validate:"omitempty"`
		Pets                 []int                   `json:"pets" validate:"omitempty"`
		FacilitiesString     string                  `json:"-" form:"facilities" validate:"omitempty"`
		Facilities           []int                   `json:"facilities" validate:"omitempty"`
		Images               []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty,dive"`
		DeleteImageIdsString string                  `json:"-" form:"deleteImageIds" validate:"omitempty"`
		DeleteImageIds       []int                   `json:"deleteImageIds" validate:"omitempty"`
	}
)
