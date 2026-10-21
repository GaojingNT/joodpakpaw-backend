package responses

import "joodpakpaw/src/models"

type GuestProfile struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Avatar    string `json:"avatar"`
}

type HostReviewItem struct {
	BookingID    int                  `json:"id"`
	Rating       int                  `json:"rating"`
	Comment      string               `json:"comment"`
	PropertyName string               `json:"propertyName"`
	ReviewedAt   *string              `json:"reviewedAt"`
	GuestProfile *GuestProfile        `json:"guestProfile"`
	ReviewImages []HostReviewImage    `json:"reviewImages"`
}

type HostReviewImage struct {
	ID                int    `json:"id"`
	FileName          string `json:"fileName"`
	FilePath          string `json:"filePath"`
	FileContentType   string `json:"fileContentType"`
	FileExtensionName string `json:"fileExtensionName"`
	FileSize          int    `json:"fileSize"`
}

type HostReviewsCollection struct{}

func (c *HostReviewsCollection) Collection(bookings []models.Booking) []HostReviewItem {
	items := make([]HostReviewItem, 0, len(bookings))
	for _, b := range bookings {
		propertyName := ""
		if len(b.BookingOrders) > 0 && b.BookingOrders[0].Property != nil {
			propertyName = b.BookingOrders[0].Property.Title
		}

		var reviewedAt *string
		if b.ReviewedAt != nil && b.ReviewedAt.Valid {
			t := b.ReviewedAt.Time.Format("2006-01-02 15:04:05")
			reviewedAt = &t
		}

		rating := 0
		if b.Rating != nil {
			rating = *b.Rating
		}

		var guestProfile *GuestProfile
		if b.Guest != nil {
			guestProfile = &GuestProfile{
				ID:        b.Guest.ID,
				FirstName: b.Guest.FirstName,
				LastName:  b.Guest.LastName,
				Avatar:    b.Guest.Avatar,
			}
		}

		reviewImages := make([]HostReviewImage, 0, len(b.ReviewImages))
		for _, img := range b.ReviewImages {
			reviewImages = append(reviewImages, HostReviewImage{
				ID:                img.ID,
				FileName:          img.FileName,
				FilePath:          img.FilePath,
				FileContentType:   img.FileContentType,
				FileExtensionName: img.FileExtensionName,
				FileSize:          img.FileSize,
			})
		}

		items = append(items, HostReviewItem{
			BookingID:    b.ID,
			Rating:       rating,
			Comment:      b.Comment,
			PropertyName: propertyName,
			ReviewedAt:   reviewedAt,
			GuestProfile: guestProfile,
			ReviewImages: reviewImages,
		})
	}
	return items
}
