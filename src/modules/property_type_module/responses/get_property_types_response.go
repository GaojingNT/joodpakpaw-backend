package responses

import "joodpakpaw/src/models"

type (
	GetPropertyTypesResponse struct {
		PropertyTypes []PropertyTypeResponse `json:"propertyTypes"`
	}
	PropertyTypeResponse struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Label string `json:"label"`
	}
)

func (response *GetPropertyTypesResponse) Collection(propertyTypes []models.PropertyType) []PropertyTypeResponse {
	propertyTypeResponses := make([]PropertyTypeResponse, 0)

	for _, propertyType := range propertyTypes {
		propertyTypeResponses = append(propertyTypeResponses, PropertyTypeResponse{
			ID:    propertyType.ID,
			Name:  propertyType.Name,
			Label: propertyType.Label,
		})
	}

	return propertyTypeResponses
}
