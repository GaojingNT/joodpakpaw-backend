package responses

import (
	"joodpakpaw/internal/http_response"
	"joodpakpaw/src/models"
)

type (
	GetUserByIDResponse struct {
		http_response.OkResponse
		User UserResponse `json:"user"`
	}
	UserResponse struct {
		ID             int    `json:"id"`
		Email          string `json:"email"`
		FirstName      string `json:"firstName"`
		LastName       string `json:"lastName"`
		IsVerifiedHost bool   `json:"isVerifiedHost"`
		IsBannedReview bool   `json:"isBannedReview"`
		CreatedAt      string `json:"createdAt"`
		UpdatedAt      string `json:"updatedAt"`
		// Role
		RoleID      *int     `json:"roleId,omitempty"`
		RoleName    string   `json:"roleName,omitempty"`
		Permissions []string `json:"permissions,omitempty"`
	}
)

func (response *GetUserByIDResponse) Make(user models.User) *GetUserByIDResponse {
	response = &GetUserByIDResponse{
		OkResponse: http_response.OkResponse{Code: "OK", Message: "Success"},
		User: UserResponse{
			ID:             user.ID,
			Email:          user.Email,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			IsVerifiedHost: user.IsVerifiedHost,
			IsBannedReview: user.IsBannedReview,
			CreatedAt:      user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      user.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	return response
}
