package responses

import "joodpakpaw/src/models"

type (
	GetFacilitiesResponse struct {
		Facilities []FacilityResponse `json:"facilities"`
	}
	FacilityResponse struct {
		ID       int               `json:"id"`
		Name     string            `json:"name"`
		Category *CategoryResponse `json:"category,omitempty"`
	}
	CategoryResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (response *GetFacilitiesResponse) Collection(facilities []models.Facility) []FacilityResponse {
	facilityResponses := make([]FacilityResponse, 0)

	for _, facility := range facilities {
		var category *CategoryResponse
		if facility.Category != nil {
			category = &CategoryResponse{
				ID:   facility.Category.ID,
				Name: facility.Category.Name,
			}
		}

		facilityResponses = append(facilityResponses, FacilityResponse{
			ID:       facility.ID,
			Name:     facility.Name,
			Category: category,
		})
	}

	return facilityResponses
}
