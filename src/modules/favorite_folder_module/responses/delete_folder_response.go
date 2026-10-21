package responses

type DeleteFolderResponse struct {
	Message string `json:"message"`
}

func (r *DeleteFolderResponse) Make() *DeleteFolderResponse {
	r.Message = "Folder deleted successfully"
	return r
}
