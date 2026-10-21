
package responses

import "joodpakpaw/src/models"

type (
	GetPetBreedsResponse struct {
		PetBreeds []PetBreedResponse `json:"petBreeds"`
	}
	PetBreedResponse struct {
		ID       int     `json:"id"`
		Breed    string  `json:"breed"`
		ImageURL *string `json:"imageUrl"`
		PetID    int     `json:"petId"`
		PetName  string  `json:"petName,omitempty"`
	}
)

func (response *GetPetBreedsResponse) Collection(petBreeds []models.PetBreed) []PetBreedResponse {
	petBreedResponses := make([]PetBreedResponse, 0)

	for _, petBreed := range petBreeds {
		item := PetBreedResponse{
			ID:       petBreed.ID,
			Breed:    petBreed.Breed,
			ImageURL: petBreed.ImageURL,
		}

		if petBreed.PetID != nil {
			item.PetID = *petBreed.PetID
		}

		if petBreed.Pet != nil {
			item.PetName = petBreed.Pet.Name
		}

		petBreedResponses = append(petBreedResponses, item)
	}

	return petBreedResponses
}
