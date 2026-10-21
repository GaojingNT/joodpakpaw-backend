package responses

import "joodpakpaw/src/models"

type UserBehaviorItem struct {
	ID               int    `json:"id"`
	UserID           int    `json:"userId"`
	ReferenceTable   string `json:"referenceTable"`
	ReferenceRecord  string `json:"referenceRecord"`
	BehaviorCategory string `json:"behaviorCategory"`
	CreatedAt        string `json:"createdAt"`
}

type GetUserBehaviorsResponse struct {
	UserBehaviors []UserBehaviorItem `json:"userBehaviors"`
}

func (r *GetUserBehaviorsResponse) Make(behaviors []models.UserBehavior) *GetUserBehaviorsResponse {
	r.UserBehaviors = make([]UserBehaviorItem, len(behaviors))
	for i, b := range behaviors {
		item := UserBehaviorItem{
			ID:              b.ID,
			UserID:          b.UserID,
			ReferenceTable:  b.ReferenceTable,
			ReferenceRecord: b.ReferenceRecord,
			CreatedAt:       b.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if b.BehaviorCategory != nil {
			item.BehaviorCategory = b.BehaviorCategory.Name
		}
		r.UserBehaviors[i] = item
	}
	return r
}
