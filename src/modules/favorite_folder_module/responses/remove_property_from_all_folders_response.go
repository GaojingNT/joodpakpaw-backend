package responses

type RemovePropertyFromAllFoldersResponse struct {
	Message         string `json:"message"`
	FoldersAffected int    `json:"foldersAffected"`
	FoldersDeleted  int    `json:"foldersDeleted"`
}

func (r *RemovePropertyFromAllFoldersResponse) Make(foldersAffected int, foldersDeleted int) *RemovePropertyFromAllFoldersResponse {
	r.Message = "Property removed from all folders successfully"
	r.FoldersAffected = foldersAffected
	r.FoldersDeleted = foldersDeleted
	return r
}
