package responses

import "joodpakpaw/src/models"

type (
	AdminHostProfileResponse struct {
		ID        int    `json:"id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Avatar    string `json:"avatar"`
	}

	AdminPropertyListItem struct {
		ID              int                      `json:"id"`
		Title           string                   `json:"title"`
		HostProfile     AdminHostProfileResponse  `json:"hostProfile"`
		IsActive        bool                     `json:"isActive"`
		IsBannedByAdmin bool                     `json:"isBannedByAdmin"`
		ReportCount     int                      `json:"reportCount"`
		CreatedAt       string                   `json:"createdAt"`
	}

	AdminGetPropertiesCollection struct{}

	AdminGetPropertyByIDResponse struct {
		ID              int                      `json:"id"`
		Title           string                   `json:"title"`
		HostProfile     AdminHostProfileResponse  `json:"hostProfile"`
		IsActive        bool                     `json:"isActive"`
		IsBannedByAdmin bool                     `json:"isBannedByAdmin"`
		ReportCount     int                      `json:"reportCount"`
		CreatedAt       string                   `json:"createdAt"`
		UpdatedAt       string                   `json:"updatedAt"`
	}
)

func (g *AdminGetPropertiesCollection) Collection(properties []models.Property) []AdminPropertyListItem {
	result := make([]AdminPropertyListItem, 0, len(properties))
	for _, p := range properties {
		result = append(result, toAdminPropertyListItem(p))
	}
	return result
}

func (r *AdminGetPropertyByIDResponse) Make(p models.Property) *AdminGetPropertyByIDResponse {
	return &AdminGetPropertyByIDResponse{
		ID:              p.ID,
		Title:           p.Title,
		HostProfile:     toAdminHostProfile(p.Host),
		IsActive:        p.IsActive,
		IsBannedByAdmin: p.IsBannedByAdmin,
		ReportCount:     p.ReportCount,
		CreatedAt:       p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toAdminPropertyListItem(p models.Property) AdminPropertyListItem {
	return AdminPropertyListItem{
		ID:              p.ID,
		Title:           p.Title,
		HostProfile:     toAdminHostProfile(p.Host),
		IsActive:        p.IsActive,
		IsBannedByAdmin: p.IsBannedByAdmin,
		ReportCount:     p.ReportCount,
		CreatedAt:       p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toAdminHostProfile(host *models.User) AdminHostProfileResponse {
	if host == nil {
		return AdminHostProfileResponse{}
	}
	return AdminHostProfileResponse{
		ID:        host.ID,
		FirstName: host.FirstName,
		LastName:  host.LastName,
		Avatar:    host.Avatar,
	}
}
