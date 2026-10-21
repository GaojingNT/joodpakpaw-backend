package dtos

type (
	// CreateNotification DTO for creating a new notification
	CreateNotification struct {
		UserID  int    `json:"userId" validate:"required"`
		Title   string `json:"title" validate:"required,max=100"`
		Message string `json:"message" validate:"required"`
	}
)
