package responses

import (
	"joodpakpaw/src/models"
)

type CreatePetProfileResponse struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	Image            string        `json:"image,omitempty"`
	Sex              models.PetSex `json:"sex"`
	BirthDate        string        `json:"birthDate"`
	ContactName      string        `json:"contactName"`
	ContactPhone     string        `json:"contactPhone"`
	Note             string        `json:"note,omitempty"`
	GuestID          *int          `json:"guestId"`
	PetTypeID        *int          `json:"petId"`
	BreedID          *int          `json:"petBreedId"`
	PetSizeCategoryID *int         `json:"petSizeId"`
	PetEnergyID      *int          `json:"petEnergyId"`
	CreatedAt        string        `json:"createdAt"`
}

func (r *CreatePetProfileResponse) Make(petProfile models.PetProfile) *CreatePetProfileResponse {
	r.ID = petProfile.ID
	r.Name = petProfile.Name
	r.Image = petProfile.Image
	r.Sex = petProfile.Sex
	r.BirthDate = petProfile.BirthDate.Format("2006-01-02")
	r.ContactName = petProfile.ContactName
	r.ContactPhone = petProfile.ContactPhone
	r.Note = petProfile.Note
	r.GuestID = petProfile.GuestID
	r.PetTypeID = petProfile.PetTypeID
	r.BreedID = petProfile.BreedID
	r.PetSizeCategoryID = petProfile.PetSizeCategoryID
	r.PetEnergyID = petProfile.PetEnergyID
	r.CreatedAt = petProfile.CreatedAt.Format("2006-01-02 15:04:05")

	return r
}
