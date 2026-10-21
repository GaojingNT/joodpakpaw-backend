package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type (
	GetReportCategoriesResponse struct {
		ReportCategories []ReportCategoryResponse `json:"reportCategories"`
	}
	ReportCategoryResponse struct {
		ID        int                `json:"id"`
		Name      string             `json:"name"`
		Label     string             `json:"label"`
		Type      models.ReportType  `json:"type"`
		CreatedAt time.Time          `json:"createdAt"`
		UpdatedAt time.Time          `json:"updatedAt"`
	}
)

func (r *GetReportCategoriesResponse) Collection(categories []models.ReportCategory) []ReportCategoryResponse {
	result := make([]ReportCategoryResponse, 0)
	for _, c := range categories {
		result = append(result, ReportCategoryResponse{
			ID:        c.ID,
			Name:      c.Name,
			Label:     c.Label,
			Type:      c.Type,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}
	return result
}
