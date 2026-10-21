package responses

import "joodpakpaw/src/models"

type GetFacilityByIDResponse struct {
	ID       int               `json:"id"`
	Name     string            `json:"name"`
	Category *CategoryResponse `json:"category,omitempty"`
}

func (response *GetFacilityByIDResponse) Make(facility models.Facility) *GetFacilityByIDResponse {
	var category *CategoryResponse
	if facility.Category != nil {
		category = &CategoryResponse{
			ID:   facility.Category.ID,
			Name: facility.Category.Name,
		}
	}

	return &GetFacilityByIDResponse{
		ID:       facility.ID,
		Name:     facility.Name,
		Category: category,
	}
}
