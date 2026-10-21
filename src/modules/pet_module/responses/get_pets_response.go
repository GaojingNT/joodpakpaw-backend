package responses

import "joodpakpaw/src/models"

type (
	GetPetsResponse struct {
		Pets []PetResponse `json:"pets"`
	}
	PetResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (response *GetPetsResponse) Collection(pets []models.Pet) []PetResponse {
	petResponses := make([]PetResponse, 0)

	for _, pet := range pets {
		petResponses = append(petResponses, PetResponse{
			ID:   pet.ID,
			Name: pet.Name,
		})
	}

	return petResponses
}
