package responses

import "joodpakpaw/src/models"

type CreateUserRecentViewResponse struct {
	ID         int    `json:"id"`
	UserID     int    `json:"userId"`
	PropertyID int    `json:"propertyId"`
	ViewedAt   string `json:"viewedAt"`
}

func (r *CreateUserRecentViewResponse) Make(record *models.UserRecentView) *CreateUserRecentViewResponse {
	r.ID = record.ID
	r.UserID = record.UserID
	r.PropertyID = record.PropertyID
	r.ViewedAt = record.ViewedAt.Format("2006-01-02 15:04:05")
	return r
}
