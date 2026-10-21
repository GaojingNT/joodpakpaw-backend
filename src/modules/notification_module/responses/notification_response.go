package responses

import "joodpakpaw/src/models"

type NotificationResponse struct {
	ID             int                         `json:"id"`
	Title          string                      `json:"title"`
	Message        string                      `json:"message"`
	Category       models.NotificationCategory `json:"category"`
	Type           models.NotificationType     `json:"type"`
	IsRead         bool                        `json:"isRead"`
	ReferenceTable string                      `json:"referenceTable"`
	ReferenceID    *int                        `json:"referenceId"`
	RoleName       string                      `json:"roleName"`
	ActionName     string                      `json:"actionName"`
	CreatedAt      string                      `json:"createdAt"`
}

type GetNotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
}

func (r *NotificationResponse) Make(n models.Notification) *NotificationResponse {
	r.ID = n.ID
	r.Title = n.Title
	r.Message = n.Message
	r.Category = n.Category
	r.Type = n.Type
	r.IsRead = n.IsRead
	r.ReferenceTable = n.ReferenceTable
	r.ReferenceID = n.ReferenceID
	r.RoleName = ""
	if n.Role != nil {
		r.RoleName = n.Role.Name
	}
	r.ActionName = ""
	if n.NotificationAction != nil {
		r.ActionName = n.NotificationAction.Name
	}
	r.CreatedAt = n.CreatedAt.Format("2006-01-02 15:04:05")
	return r
}

func (r *GetNotificationsResponse) Make(notifications []models.Notification) *GetNotificationsResponse {
	r.Notifications = make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		r.Notifications[i] = *new(NotificationResponse).Make(n)
	}
	return r
}
