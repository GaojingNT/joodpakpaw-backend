package responses

import (
	"joodpakpaw/src/models"
)

type GetPetProfileByIDResponse struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Image        string        `json:"image,omitempty"`
	Sex          models.PetSex `json:"sex"`
	BirthDate    string        `json:"birthDate"`
	PetBreedName string        `json:"petBreedName,omitempty"`
	PetSizeName  string        `json:"petSizeName,omitempty"`
	Rating       int           `json:"rating"`
	ContactName  string        `json:"contactName"`
	ContactPhone string        `json:"contactPhone"`
	Note         string        `json:"note,omitempty"`
}



func (r *GetPetProfileByIDResponse) Make(petProfile models.PetProfile) *GetPetProfileByIDResponse {
	r.ID = petProfile.ID
	r.Name = petProfile.Name
	r.Image = petProfile.Image
	r.Sex = petProfile.Sex
	r.BirthDate = petProfile.BirthDate.Format("2006-01-02")
	r.Note = petProfile.Note
	r.ContactName = petProfile.ContactName
	r.ContactPhone = petProfile.ContactPhone

	// Get pet breed name from Breed
	if petProfile.Breed != nil {
		r.PetBreedName = petProfile.Breed.Breed
	}

	// Get pet size name from PetSizeCategory
	if petProfile.PetSizeCategory != nil {
		r.PetSizeName = petProfile.PetSizeCategory.Name
	}

	// Get rating from PetEnergy
	if petProfile.PetEnergy != nil {
		r.Rating = petProfile.PetEnergy.Rating
	}

	return r
}
