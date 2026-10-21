package models

type ChatMessageAttachment struct {
	BaseModel

	// Fields
	FileName          string `json:"fileName" gorm:"column:file_name;size:255;"`
	FilePath          string `json:"filePath" gorm:"column:file_path;type:text;"`
	FileContentType   string `json:"fileContentType" gorm:"column:file_content_type;size:100;"`
	FileExtensionName string `json:"fileExtensionName" gorm:"column:file_extension_name;size:50;"`
	FileSize          int    `json:"fileSize" gorm:"column:file_size;"`

	// Relationships
	MessageID *int         `json:"messageId" gorm:"column:message_id;not null;index;"`
	Message   *ChatMessage `json:"message,omitempty" gorm:"foreignKey:MessageID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
