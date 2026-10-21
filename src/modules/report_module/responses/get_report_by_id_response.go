package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type ReportImageResponse struct {
	ID                int    `json:"id"`
	FileName          string `json:"fileName"`
	FilePath          string `json:"filePath"`
	FileContentType   string `json:"fileContentType"`
	FileExtensionName string `json:"fileExtensionName"`
	FileSize          int    `json:"fileSize"`
}

type GetReportByIDResponse struct {
	ID                  int                    `json:"id"`
	Type                models.ReportType      `json:"type"`
	Status              models.ReportStatus    `json:"status"`
	ResolvedAction      *models.ReportAction   `json:"resolvedAction"`
	AdminNote           string                 `json:"adminNote"`
	ReportCategoryLabel string                 `json:"reportCategoryLabel"`
	Description         string                 `json:"description"`
	Images              []ReportImageResponse  `json:"images"`
	Property            *PropertyBriefResponse `json:"property,omitempty"`
	Reporter            *UserDetailResponse    `json:"reporter,omitempty"`
	PropOwner           *UserDetailResponse    `json:"propOwner,omitempty"`

	CreatedAt time.Time `json:"createdAt"`

	// type=review only
	ReviewAuthor *UserDetailResponse `json:"reviewAuthor,omitempty"`
}

func (g *GetReportByIDResponse) Make(r models.Report) *GetReportByIDResponse {
	g.ID = r.ID
	g.Status = r.Status
	g.ResolvedAction = r.ResolvedAction
	g.AdminNote = r.AdminNote
	g.Description = r.Description
	g.CreatedAt = r.CreatedAt
	g.Reporter = toUserDetail(r.Reporter)

	images := make([]ReportImageResponse, 0)
	for _, img := range r.Images {
		images = append(images, ReportImageResponse{
			ID:                img.ID,
			FileName:          img.FileName,
			FilePath:          img.FilePath,
			FileContentType:   img.FileContentType,
			FileExtensionName: img.FileExtensionName,
			FileSize:          img.FileSize,
		})
	}
	g.Images = images

	if r.ReportCategory != nil {
		g.Type = r.ReportCategory.Type
		g.ReportCategoryLabel = r.ReportCategory.Label
	}

	switch g.Type {
	case models.ReportTypeProperty:
		if r.Property != nil {
			g.Property = &PropertyBriefResponse{ID: r.Property.ID, Title: r.Property.Title}
			g.PropOwner = toUserDetail(r.Property.Host)
		}

	case models.ReportTypeUnbanProperty:
		if r.Property != nil {
			g.Property = &PropertyBriefResponse{ID: r.Property.ID, Title: r.Property.Title}
			g.PropOwner = toUserDetail(r.Property.Host)
		}

	case models.ReportTypeReview:
		if r.Property != nil {
			g.Property = &PropertyBriefResponse{ID: r.Property.ID, Title: r.Property.Title}
			g.PropOwner = toUserDetail(r.Property.Host)
		} else if r.Booking != nil && len(r.Booking.BookingOrders) > 0 && r.Booking.BookingOrders[0].Property != nil {
			prop := r.Booking.BookingOrders[0].Property
			g.Property = &PropertyBriefResponse{ID: prop.ID, Title: prop.Title}
			g.PropOwner = toUserDetail(prop.Host)
		}
		if r.Booking != nil {
			g.ReviewAuthor = toUserDetail(r.Booking.Guest)
		}
	}

	return g
}
