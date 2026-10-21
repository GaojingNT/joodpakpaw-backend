package responses

import (
	"joodpakpaw/internal/http_response"
	"joodpakpaw/src/models"
)

type (
	GetVerificationByIDResponse struct {
		http_response.OkResponse
		Verification VerificationDetailItem `json:"verification"`
	}
	VerificationDetailItem struct {
		ID                        int                        `json:"id"`
		User                      UserInVerificationResponse `json:"user"`
		Status                    string                     `json:"status"`
		IsResubmit                bool                       `json:"isResubmit"`
		SubmittedAt               string                     `json:"submittedAt"`
		RejectDescription         *string                    `json:"rejectDescription"`
		PreviousRejectDescription *string                    `json:"previousRejectDescription"`
		IdPhoto                   string                     `json:"idPhoto"`
		SelfieWithId              string                     `json:"selfieWithId"`
	}
)

func (r *GetVerificationByIDResponse) Make(v models.UserVerification) *GetVerificationByIDResponse {
	return &GetVerificationByIDResponse{
		OkResponse: http_response.OkResponse{Code: "OK", Message: "Success"},
		Verification: VerificationDetailItem{
			ID:                        v.ID,
			User:                      makeUserResponse(v.User),
			Status:                    string(v.Status),
			IsResubmit:                v.IsResubmit,
			SubmittedAt:               v.SubmittedAt.Format("2006-01-02 15:04:05"),
			RejectDescription:         v.RejectDescription,
			PreviousRejectDescription: v.PreviousRejectDescription,
			IdPhoto:                   v.IdPhoto,
			SelfieWithId:              v.SelfieWithId,
		},
	}
}
