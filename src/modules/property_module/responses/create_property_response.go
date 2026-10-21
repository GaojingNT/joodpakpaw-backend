package responses

import (
	"joodpakpaw/internal/http_response"
	"joodpakpaw/src/models"
)

type (
	CreatePropertyResponse struct {
		http_response.OkResponse
		Data PropertyCreatedData `json:"data"`
	}
	PropertyCreatedData struct {
		ID int `json:"id"`
	}
)

func (response *CreatePropertyResponse) Make(property models.Property) *CreatePropertyResponse {
	response = &CreatePropertyResponse{
		OkResponse: http_response.OkResponse{Code: "OK", Message: "Property created successfully"},
		Data: PropertyCreatedData{
			ID: property.ID,
		},
	}

	return response
}
