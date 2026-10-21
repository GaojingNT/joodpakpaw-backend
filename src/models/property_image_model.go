package models

import "gorm.io/gorm"

type PropertyImage struct {
	BaseModel

	// Fields
	FileName          string `json:"fileName" gorm:"column:file_name;size:255;"`
	FilePath          string `json:"filePath" gorm:"column:file_path;type:text;"`
	FileContentType   string `json:"fileContentType" gorm:"column:file_content_type;size:100;"`
	FileExtensionName string `json:"fileExtensionName" gorm:"column:file_extension_name;size:50;"`
	FileSize          int    `json:"fileSize" gorm:"column:file_size;"`

	// Relationships
	PropertyID *int      `json:"propertyId" gorm:"column:property_id;not null;"`
	Property   *Property `json:"property,omitempty" gorm:"foreignKey:PropertyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
