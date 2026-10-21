package responses

import "joodpakpaw/src/models"

type GetPetBreedByIDResponse struct {
	ID       int     `json:"id"`
	Breed    string  `json:"breed"`
	ImageURL *string `json:"imageUrl"`
	PetID    int     `json:"petId"`
	PetName  string  `json:"petName,omitempty"`
}

func (response *GetPetBreedByIDResponse) Make(petBreed models.PetBreed) *GetPetBreedByIDResponse {
	response.ID = petBreed.ID
	response.Breed = petBreed.Breed
	response.ImageURL = petBreed.ImageURL

	if petBreed.PetID != nil {
		response.PetID = *petBreed.PetID
	}

	if petBreed.Pet != nil {
		response.PetName = petBreed.Pet.Name
	}

	return response
}
