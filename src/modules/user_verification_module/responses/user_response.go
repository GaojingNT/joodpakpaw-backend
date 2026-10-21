package responses

import "joodpakpaw/src/models"

type UserInVerificationResponse struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
}

func makeUserResponse(u *models.User) UserInVerificationResponse {
	if u == nil {
		return UserInVerificationResponse{}
	}
	return UserInVerificationResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
		Email:     u.Email,
	}
}
