package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type (
	UserProfileResponse struct {
		ID        int      `json:"id"`
		FirstName string   `json:"firstName"`
		LastName  string   `json:"lastName"`
		Avatar    string   `json:"avatar"`
		Roles     []string `json:"roles"`
	}

	ActionLogListItem struct {
		ID        int                  `json:"id"`
		ModelName string               `json:"modelName"`
		RecordID  string               `json:"recordId"`
		Action    string               `json:"action"`
		Messages  string               `json:"messages"`
		Actor     *UserProfileResponse `json:"actor"`
		Target    *UserProfileResponse `json:"target"`
		CreatedAt time.Time            `json:"createdAt"`
	}

	GetActionLogsCollection struct{}
)

func (g *GetActionLogsCollection) Collection(logs []models.ActionLog) []ActionLogListItem {
	result := make([]ActionLogListItem, 0, len(logs))

	for _, l := range logs {
		item := ActionLogListItem{
			ID:        l.ID,
			ModelName: l.ModelName,
			RecordID:  l.RecordID,
			Action:    l.Action,
			Messages:  l.Messages,
			CreatedAt: l.CreatedAt,
		}

		if l.Actor != nil {
			item.Actor = toUserProfile(l.Actor)
		}

		if l.Target != nil {
			item.Target = toUserProfile(l.Target)
		}

		result = append(result, item)
	}

	return result
}

func toUserProfile(u *models.User) *UserProfileResponse {
	if u == nil {
		return nil
	}
	roles := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}
	return &UserProfileResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
		Roles:     roles,
	}
}
