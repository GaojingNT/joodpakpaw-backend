package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type GetPetProfilesResponse struct {
	PetProfiles []PetProfileListItem `json:"petProfiles"`
}

type PetProfileListItem struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Image             string `json:"image,omitempty"`
	BirthDate         string `json:"birthDate"`
	Age               int    `json:"age"`
	PetTypeName       string `json:"petTypeName,omitempty"`
	PetSizeCategoryId *int   `json:"petSizeCategoryId,omitempty"`
}

func (r *GetPetProfilesResponse) Make(petProfiles []models.PetProfile) *GetPetProfilesResponse {
	r.PetProfiles = make([]PetProfileListItem, len(petProfiles))

	for i, profile := range petProfiles {
		petProfileItem := PetProfileListItem{
			ID:                profile.ID,
			Name:              profile.Name,
			Image:             profile.Image,
			BirthDate:         profile.BirthDate.Format("2006-01-02"),
			Age:               calculateAge(profile.BirthDate),
			PetSizeCategoryId: profile.PetSizeCategoryID,
		}

		// Get pet type name from PetType
		if profile.PetType != nil {
			petProfileItem.PetTypeName = profile.PetType.Name
		}

		r.PetProfiles[i] = petProfileItem
	}

	return r
}

// calculateAge calculates age from birthdate
func calculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// Adjust if birthday hasn't occurred yet this year
	if now.Month() < birthDate.Month() ||
		(now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}
