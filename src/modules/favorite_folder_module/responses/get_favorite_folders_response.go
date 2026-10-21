package responses

import (
	"joodpakpaw/src/models"
)

type GetFavoriteFoldersResponse struct {
	FavoriteFolders []FavoriteFolderListItem `json:"favoriteFolders"`
}

type FavoriteFolderListItem struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	PropertyCount   int     `json:"propertyCount"`
	FirstPropertyImage *string `json:"firstPropertyImage,omitempty"`
	CreatedAt       string  `json:"createdAt"`
}

func (r *GetFavoriteFoldersResponse) Make(folders []models.FavoriteFolder) *GetFavoriteFoldersResponse {
	r.FavoriteFolders = make([]FavoriteFolderListItem, len(folders))

	for i, folder := range folders {
		item := FavoriteFolderListItem{
			ID:            folder.ID,
			Name:          folder.Name,
			PropertyCount: len(folder.Properties),
			CreatedAt:     folder.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// Get first property image if available
		if len(folder.Properties) > 0 {
			firstProperty := folder.Properties[0]
			if len(firstProperty.Images) > 0 {
				item.FirstPropertyImage = &firstProperty.Images[0].FilePath
			}
		}

		r.FavoriteFolders[i] = item
	}

	return r
}
