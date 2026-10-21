package responses

import (
	"joodpakpaw/src/models"
)

type AddPropertyToFavoriteResponse struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	PropertyCount int    `json:"propertyCount"`
	Message       string `json:"message"`
}

func (r *AddPropertyToFavoriteResponse) Make(folder *models.FavoriteFolder, isNewFolder bool) *AddPropertyToFavoriteResponse {
	r.ID = folder.ID
	r.Name = folder.Name
	r.PropertyCount = len(folder.Properties)

	if isNewFolder {
		r.Message = "Property added to new folder successfully"
	} else {
		r.Message = "Property added to folder successfully"
	}

	return r
}
