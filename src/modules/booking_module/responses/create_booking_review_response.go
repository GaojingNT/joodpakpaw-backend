package responses

import "joodpakpaw/src/models"

type (
	CreateBookingReviewResponse struct {
		ID            int                          `json:"id"`
		Rating        *int                         `json:"rating"`
		Comment       string                       `json:"comment"`
		BookingStatus string                       `json:"bookingStatus"`
		ReviewImages  []BookingReviewImageResponse `json:"reviewImages"`
	}
	BookingReviewImageResponse struct {
		ID                int    `json:"id"`
		FileName          string `json:"fileName"`
		FilePath          string `json:"filePath"`
		FileContentType   string `json:"fileContentType"`
		FileExtensionName string `json:"fileExtensionName"`
		FileSize          int    `json:"fileSize"`
	}
)

func (r *CreateBookingReviewResponse) Make(booking *models.Booking) *CreateBookingReviewResponse {
	reviewImages := make([]BookingReviewImageResponse, 0, len(booking.ReviewImages))
	for _, img := range booking.ReviewImages {
		reviewImages = append(reviewImages, BookingReviewImageResponse{
			ID:                img.ID,
			FileName:          img.FileName,
			FilePath:          img.FilePath,
			FileContentType:   img.FileContentType,
			FileExtensionName: img.FileExtensionName,
			FileSize:          img.FileSize,
		})
	}

	return &CreateBookingReviewResponse{
		ID:            booking.ID,
		Rating:        booking.Rating,
		Comment:       booking.Comment,
		BookingStatus: string(booking.BookingStatus),
		ReviewImages:  reviewImages,
	}
}
