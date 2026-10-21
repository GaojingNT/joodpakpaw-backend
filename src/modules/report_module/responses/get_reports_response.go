package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type (
	ReportListItem struct {
		ID                  int                    `json:"id"`
		Type                models.ReportType      `json:"type"`
		Status              models.ReportStatus    `json:"status"`
		ResolvedAction      *models.ReportAction   `json:"resolvedAction"`
		ReportCategoryLabel string                 `json:"reportCategoryLabel"`
		Reporter            *UserProfileResponse   `json:"reporter"`
		CreatedAt           time.Time              `json:"createdAt"`

		// type=property only
		Property    *PropertyBriefResponse `json:"property,omitempty"`
		ReportCount *int                   `json:"reportCount,omitempty"`

		// type=review only
		ReviewAuthor *UserProfileResponse `json:"reviewAuthor,omitempty"`
	}

	GetReportsCollection struct{}
)

func (g *GetReportsCollection) Collection(reports []models.Report) []ReportListItem {
	result := make([]ReportListItem, 0, len(reports))

	propertyCounts := make(map[int]int)
	for _, r := range reports {
		if r.Property != nil {
			propertyCounts[r.Property.ID]++
		}
	}
	propertySeq := make(map[int]int)

	for _, r := range reports {
		item := ReportListItem{
			ID:             r.ID,
			Status:         r.Status,
			ResolvedAction: r.ResolvedAction,
			CreatedAt:      r.CreatedAt,
		}

		if r.ReportCategory != nil {
			item.Type = r.ReportCategory.Type
			item.ReportCategoryLabel = r.ReportCategory.Label
		}

		if r.Reporter != nil {
			item.Reporter = &UserProfileResponse{
				ID:        r.Reporter.ID,
				FirstName: r.Reporter.FirstName,
				LastName:  r.Reporter.LastName,
				Avatar:    r.Reporter.Avatar,
			}
		}

		switch item.Type {
		case models.ReportTypeProperty:
			if r.Property != nil {
				item.Property = &PropertyBriefResponse{ID: r.Property.ID, Title: r.Property.Title}
				propertySeq[r.Property.ID]++
				count := propertyCounts[r.Property.ID] - propertySeq[r.Property.ID] + 1
				item.ReportCount = &count
			}

		case models.ReportTypeUnbanProperty:
			if r.Property != nil {
				item.Property = &PropertyBriefResponse{ID: r.Property.ID, Title: r.Property.Title}
			}

		case models.ReportTypeReview:
			if r.Property != nil {
				item.Property = &PropertyBriefResponse{ID: r.Property.ID, Title: r.Property.Title}
			}
			if r.Booking != nil && r.Booking.Guest != nil {
				item.ReviewAuthor = &UserProfileResponse{
					ID:        r.Booking.Guest.ID,
					FirstName: r.Booking.Guest.FirstName,
					LastName:  r.Booking.Guest.LastName,
					Avatar:    r.Booking.Guest.Avatar,
				}
			}
		}

		result = append(result, item)
	}

	return result
}
