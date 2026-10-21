package dtos

import "mime/multipart"

type (
	CreatePetBreed struct {
		Breed     string                `json:"breed" form:"breed" validate:"required,max=100"`
		ImageFile *multipart.FileHeader `json:"-" form:"-" validate:"-"`
		PetID     *int                  `json:"-"`
	}

	UpdatePetBreed struct {
		Breed     string                `json:"breed" form:"breed" validate:"omitempty,max=100"`
		ImageFile *multipart.FileHeader `json:"-" form:"-" validate:"-"`
	}
)
