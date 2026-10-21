package models

import "gorm.io/gorm"

type BookingReviewImage struct {
	BaseModel

	// Fields
	FileName          string `json:"fileName" gorm:"column:file_name;size:255;"`
	FilePath          string `json:"filePath" gorm:"column:file_path;type:text;"`
	FileContentType   string `json:"fileContentType" gorm:"column:file_content_type;size:100;"`
	FileExtensionName string `json:"fileExtensionName" gorm:"column:file_extension_name;size:50;"`
	FileSize          int    `json:"fileSize" gorm:"column:file_size;"`

	// Relationships
	BookingID *int     `json:"bookingId" gorm:"column:booking_id;not null;"`
	Booking   *Booking `json:"booking,omitempty" gorm:"foreignKey:BookingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
