package responses

import "joodpakpaw/src/models"

type SendMessageResponse struct {
	ID          int                `json:"id"`
	Content     string             `json:"content"`
	ReadAt      *string            `json:"readAt"`
	CreatedAt   string             `json:"createdAt"`
	Sender      UserSummary        `json:"sender"`
	Attachments []AttachmentItem   `json:"attachments"`
}

type AttachmentItem struct {
	ID          int    `json:"id"`
	FileName    string `json:"fileName"`
	FilePath    string `json:"filePath"`
	ContentType string `json:"contentType"`
	FileSize    int    `json:"fileSize"`
}

func (r *SendMessageResponse) Make(message *models.ChatMessage) *SendMessageResponse {
	r.ID = message.ID
	r.Content = message.Content
	r.CreatedAt = message.CreatedAt.Format("2006-01-02 15:04:05")

	if message.ReadAt != nil {
		t := message.ReadAt.Format("2006-01-02 15:04:05")
		r.ReadAt = &t
	}

	if message.Sender != nil {
		r.Sender = UserSummary{
			ID:        message.Sender.ID,
			FirstName: message.Sender.FirstName,
			LastName:  message.Sender.LastName,
			Avatar:    message.Sender.Avatar,
		}
	}

	r.Attachments = make([]AttachmentItem, len(message.Attachments))
	for i, a := range message.Attachments {
		r.Attachments[i] = AttachmentItem{
			ID:          a.ID,
			FileName:    a.FileName,
			FilePath:    a.FilePath,
			ContentType: a.FileContentType,
			FileSize:    a.FileSize,
		}
	}

	return r
}
