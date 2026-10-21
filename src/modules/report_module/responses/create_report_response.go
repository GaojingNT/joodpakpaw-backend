package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type CreateReportResponse struct {
	ID               int                 `json:"id"`
	Type             models.ReportType   `json:"type"`
	Status           models.ReportStatus `json:"status"`
	Description      string              `json:"description"`
	ReportCategoryID *int                `json:"reportCategoryId"`
	ReporterID       *int                `json:"reporterId"`
	BookingID        *int                `json:"bookingId,omitempty"`
	PropertyID       *int                `json:"propertyId,omitempty"`
	CreatedAt        time.Time           `json:"createdAt"`
}

func (r *CreateReportResponse) Make(report models.Report, reportType models.ReportType) *CreateReportResponse {
	return &CreateReportResponse{
		ID:               report.ID,
		Type:             reportType,
		Status:           report.Status,
		Description:      report.Description,
		ReportCategoryID: report.ReportCategoryID,
		ReporterID:       report.ReporterID,
		BookingID:        report.BookingID,
		PropertyID:       report.PropertyID,
		CreatedAt:        report.CreatedAt,
	}
}
