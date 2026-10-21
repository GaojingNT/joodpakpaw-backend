package responses

import "joodpakpaw/src/models"

type CreateUserBehaviorResponse struct {
	ID              int    `json:"id"`
	UserID          int    `json:"userId"`
	ReferenceTable  string `json:"referenceTable"`
	ReferenceRecord string `json:"referenceRecord"`
	BehaviorCategory string `json:"behaviorCategory"`
	ViewedAt        string `json:"viewedAt"`
}

func (r *CreateUserBehaviorResponse) Make(behavior *models.UserBehavior) *CreateUserBehaviorResponse {
	r.ID = behavior.ID
	r.UserID = behavior.UserID
	r.ReferenceTable = behavior.ReferenceTable
	r.ReferenceRecord = behavior.ReferenceRecord
	if behavior.BehaviorCategory != nil {
		r.BehaviorCategory = behavior.BehaviorCategory.Name
	}
	r.ViewedAt = behavior.CreatedAt.Format("2006-01-02 15:04:05")
	return r
}
