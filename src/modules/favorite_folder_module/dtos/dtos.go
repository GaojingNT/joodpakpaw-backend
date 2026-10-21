package dtos

type (
	// AddPropertyToFavorite DTO for adding property to favorite folder
	AddPropertyToFavorite struct {
		PropertyID *int    `json:"propertyId" validate:"omitempty"`
		FolderID   *int    `json:"folderId" validate:"omitempty"`
		FolderName *string `json:"folderName" validate:"omitempty,max=255"`
	}

	// UpdateFolderName DTO for updating folder name
	UpdateFolderName struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	}
)
