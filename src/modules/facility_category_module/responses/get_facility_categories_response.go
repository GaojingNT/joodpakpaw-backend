package responses

import "joodpakpaw/src/models"

type (
	GetFacilityCategoriesResponse struct {
		FacilityCategories []FacilityCategoryResponse `json:"facilityCategories"`
	}
	FacilityCategoryResponse struct {
		ID         int                `json:"id"`
		Name       string             `json:"name"`
		Facilities []FacilityResponse `json:"facilities"`
	}
	FacilityResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (response *GetFacilityCategoriesResponse) Collection(facilityCategories []models.FacilityCategory) []FacilityCategoryResponse {
	facilityCategoryResponses := make([]FacilityCategoryResponse, 0)

	for _, category := range facilityCategories {
		facilityCategoryResponses = append(facilityCategoryResponses, FacilityCategoryResponse{
			ID:         category.ID,
			Name:       category.Name,
			Facilities: mapFacilities(category.Facilities),
		})
	}

	return facilityCategoryResponses
}

func mapFacilities(facilities []models.Facility) []FacilityResponse {
	facilityResponses := make([]FacilityResponse, 0)

	for _, facility := range facilities {
		facilityResponses = append(facilityResponses, FacilityResponse{
			ID:   facility.ID,
			Name: facility.Name,
		})
	}

	return facilityResponses
}
