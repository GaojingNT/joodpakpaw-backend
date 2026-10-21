package responses
import "joodpakpaw/src/models"

type (
GetPetSizeCategoriesResponse struct {
	PetSizeCategories []PetSizeCategoryResponse `json:"petSizeCategories"`
	}
PetSizeCategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code int    `json:"code"`
	Min  int    `json:"min"`
	Max  int    `json:"max"`
}
)
func (response *GetPetSizeCategoriesResponse) Collection(petSizeCategories []models.PetSizeCategory) []PetSizeCategoryResponse {
	petSizeCategoryResponses := make([]PetSizeCategoryResponse, 0)
	for _, petSizeCategory := range petSizeCategories {
		petSizeCategoryResponses = append(petSizeCategoryResponses, PetSizeCategoryResponse{
			ID:   petSizeCategory.ID,
			Name: petSizeCategory.Name,
			Code: petSizeCategory.Code,
			Min:  petSizeCategory.Min,
			Max:  petSizeCategory.Max,
		})
	}
	return petSizeCategoryResponses
}