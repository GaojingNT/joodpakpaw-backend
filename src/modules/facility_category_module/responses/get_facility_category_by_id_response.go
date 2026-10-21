package responses

import "joodpakpaw/src/models"

type GetFacilityCategoryByIDResponse struct {
	ID         int                `json:"id"`
	Name       string             `json:"name"`
	Facilities []FacilityResponse `json:"facilities"`
}

func (response *GetFacilityCategoryByIDResponse) Make(category models.FacilityCategory) *GetFacilityCategoryByIDResponse {
	return &GetFacilityCategoryByIDResponse{
		ID:         category.ID,
		Name:       category.Name,
		Facilities: mapFacilities(category.Facilities),
	}
}
