package responses

import "joodpakpaw/src/models"

type (
	UserProfileResponse struct {
		ID        int    `json:"id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Avatar    string `json:"avatar"`
	}

	UserDetailResponse struct {
		ID        int    `json:"id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
		Avatar    string `json:"avatar"`
	}

	PropertyBriefResponse struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}
)

func toUserDetail(u *models.User) *UserDetailResponse {
	if u == nil {
		return nil
	}
	return &UserDetailResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Avatar:    u.Avatar,
	}
}
