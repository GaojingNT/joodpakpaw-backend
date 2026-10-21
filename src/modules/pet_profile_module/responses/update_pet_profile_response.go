package responses

import (
	"joodpakpaw/src/models"
)

type UpdatePetProfileResponse struct {
	ID                int           `json:"id"`
	Name              string        `json:"name"`
	Image             string        `json:"image,omitempty"`
	Sex               models.PetSex `json:"sex"`
	BirthDate         string        `json:"birthDate"`
	ContactName       string        `json:"contactName"`
	ContactPhone      string        `json:"contactPhone"`
	Note              string        `json:"note,omitempty"`
	PetSizeCategoryID *int          `json:"petSizeId"`
	PetEnergyID       *int          `json:"petEnergyId"`
	UpdatedAt         string        `json:"updatedAt"`
}

func (r *UpdatePetProfileResponse) Make(petProfile models.PetProfile) *UpdatePetProfileResponse {
	r.ID = petProfile.ID
	r.Name = petProfile.Name
	r.Image = petProfile.Image
	r.Sex = petProfile.Sex
	r.BirthDate = petProfile.BirthDate.Format("2006-01-02")
	r.ContactName = petProfile.ContactName
	r.ContactPhone = petProfile.ContactPhone
	r.Note = petProfile.Note
	r.PetSizeCategoryID = petProfile.PetSizeCategoryID
	r.PetEnergyID = petProfile.PetEnergyID
	r.UpdatedAt = petProfile.UpdatedAt.Format("2006-01-02 15:04:05")

	return r
}
