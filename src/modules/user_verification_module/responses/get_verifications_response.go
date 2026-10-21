package responses

import "joodpakpaw/src/models"

type (
	GetVerificationsCollection struct {
		VerificationListItem
	}
	VerificationListItem struct {
		ID          int                        `json:"id"`
		User        UserInVerificationResponse `json:"user"`
		Status      string                     `json:"status"`
		IsResubmit  bool                       `json:"isResubmit"`
		SubmittedAt string                     `json:"submittedAt"`
	}
)

func (r *GetVerificationsCollection) Collection(verifications []models.UserVerification) []GetVerificationsCollection {
	items := make([]GetVerificationsCollection, 0, len(verifications))
	for _, v := range verifications {
		items = append(items, GetVerificationsCollection{
			VerificationListItem: VerificationListItem{
				ID:          v.ID,
				User:        makeUserResponse(v.User),
				Status:      string(v.Status),
				IsResubmit:  v.IsResubmit,
				SubmittedAt: v.SubmittedAt.Format("2006-01-02 15:04:05"),
			},
		})
	}
	return items
}
