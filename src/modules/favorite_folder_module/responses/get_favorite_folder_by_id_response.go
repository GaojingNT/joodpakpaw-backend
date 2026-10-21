package responses

import (
	"joodpakpaw/src/models"
)

type GetFavoriteFolderByIdResponse struct {
	ID         int                        `json:"id"`
	Name       string                     `json:"name"`
	Properties []FavoritePropertyListItem `json:"properties"`
	CreatedAt  string                     `json:"createdAt"`
}

type FavoritePropertyListItem struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	Province   string  `json:"province"`
	District   string  `json:"district,omitempty"`
	FirstImage *string `json:"firstImage,omitempty"`
}

func (r *GetFavoriteFolderByIdResponse) Make(folder *models.FavoriteFolder) *GetFavoriteFolderByIdResponse {
	r.ID = folder.ID
	r.Name = folder.Name
	r.CreatedAt = folder.CreatedAt.Format("2006-01-02 15:04:05")

	// Map properties
	r.Properties = make([]FavoritePropertyListItem, len(folder.Properties))
	for i, property := range folder.Properties {
		item := FavoritePropertyListItem{
			ID:       property.ID,
			Title:    property.Title,
			Province: property.Province,
			District: property.District,
		}

		// Get first image if available
		if len(property.Images) > 0 {
			item.FirstImage = &property.Images[0].FilePath
		}

		r.Properties[i] = item
	}

	return r
}
