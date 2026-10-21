package dtos

import (
	"mime/multipart"

	"joodpakpaw/src/models"
)

type CreatePetProfile struct {
	Name              string                `json:"name" form:"name" validate:"required,min=1,max=100"`
	Image             string                `json:"image" form:"image"`
	ImageFile         *multipart.FileHeader `json:"-" form:"-" validate:"-"`
	Sex               models.PetSex         `json:"sex" form:"sex" validate:"required,oneof=male female other"`
	BirthDate         string                `json:"birthDate" form:"birthDate" validate:"required"`
	ContactName       string                `json:"contactName" form:"contactName" validate:"required,min=1,max=100"`
	ContactPhone      string                `json:"contactPhone" form:"contactPhone" validate:"required,min=1,max=20"`
	Note              string                `json:"note" form:"note"`
	PetTypeID         *int                  `json:"petId" form:"petId" validate:"required"`
	BreedID           *int                  `json:"petBreedId" form:"petBreedId" validate:"required"`
	PetSizeCategoryID *int                  `json:"petSizeId" form:"petSizeId" validate:"required"`
	PetEnergyID       *int                  `json:"petEnergyId" form:"petEnergyId" validate:"required"`
	GuestID           *int                  `json:"guestId" form:"guestId"` // Set from authenticated user
}

type UpdatePetProfile struct {
	Name              *string               `json:"name" form:"name" `
	Image             *string               `json:"image" form:"image"`
	ImageFile         *multipart.FileHeader `json:"-" form:"-" validate:"-"`
	PetSizeCategoryID *int                  `json:"petSizeId" form:"petSizeId"`
	PetEnergyID       *int                  `json:"petEnergyId" form:"petEnergyId"`
	ContactName       *string               `json:"contactName" form:"contactName"`
	ContactPhone      *string               `json:"contactPhone" form:"contactPhone"`
	Note              *string               `json:"note" form:"note"`
}
