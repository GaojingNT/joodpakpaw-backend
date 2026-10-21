package responses

import "joodpakpaw/src/models"

type GetPropertyTypeByIDResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
}

func (response *GetPropertyTypeByIDResponse) Make(propertyType models.PropertyType) *GetPropertyTypeByIDResponse {
	return &GetPropertyTypeByIDResponse{
		ID:    propertyType.ID,
		Name:  propertyType.Name,
		Label: propertyType.Label,
	}
}
