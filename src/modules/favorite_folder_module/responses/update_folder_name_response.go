package responses

import (
	"joodpakpaw/src/models"
)

type UpdateFolderNameResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	UpdatedAt string `json:"updatedAt"`
	Message   string `json:"message"`
}

func (r *UpdateFolderNameResponse) Make(folder *models.FavoriteFolder) *UpdateFolderNameResponse {
	r.ID = folder.ID
	r.Name = folder.Name
	r.UpdatedAt = folder.UpdatedAt.Format("2006-01-02 15:04:05")
	r.Message = "Folder name updated successfully"
	return r
}
