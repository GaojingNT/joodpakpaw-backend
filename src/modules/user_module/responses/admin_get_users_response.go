package responses

import "joodpakpaw/src/models"

type (
	AdminUserListItem struct {
		ID                    int      `json:"id"`
		Email                 string   `json:"email"`
		FirstName             string   `json:"firstName"`
		LastName              string   `json:"lastName"`
		Avatar                string   `json:"avatar"`
		IsVerifiedHost        bool     `json:"isVerifiedHost"`
		IsBannedAllProperties bool     `json:"isBannedAllProperties"`
		IsBannedReview        bool     `json:"isBannedReview"`
		Roles                 []string `json:"roles"`
		CreatedAt             string   `json:"createdAt"`
	}

	AdminGetUsersCollection struct{}

	AdminGetUserByIDResponse struct {
		ID                    int      `json:"id"`
		Email                 string   `json:"email"`
		FirstName             string   `json:"firstName"`
		LastName              string   `json:"lastName"`
		Avatar                string   `json:"avatar"`
		IsVerifiedHost        bool     `json:"isVerifiedHost"`
		IsBannedAllProperties bool     `json:"isBannedAllProperties"`
		IsBannedReview        bool     `json:"isBannedReview"`
		Roles                 []string `json:"roles"`
		Violations            int64    `json:"violations"`
		CreatedAt             string   `json:"createdAt"`
		UpdatedAt             string   `json:"updatedAt"`
	}
)

func (g *AdminGetUsersCollection) Collection(users []models.User) []AdminUserListItem {
	result := make([]AdminUserListItem, 0, len(users))
	for _, u := range users {
		result = append(result, toAdminUserListItem(u))
	}
	return result
}

func (r *AdminGetUserByIDResponse) Make(u models.User, violations int64) *AdminGetUserByIDResponse {
	roles := toRoleNames(u.Roles)
	return &AdminGetUserByIDResponse{
		ID:                    u.ID,
		Email:                 u.Email,
		FirstName:             u.FirstName,
		LastName:              u.LastName,
		Avatar:                u.Avatar,
		IsVerifiedHost:        u.IsVerifiedHost,
		IsBannedAllProperties: u.IsBannedAllProperties,
		IsBannedReview:        u.IsBannedReview,
		Roles:                 roles,
		Violations:            violations,
		CreatedAt:             u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             u.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toAdminUserListItem(u models.User) AdminUserListItem {
	return AdminUserListItem{
		ID:                    u.ID,
		Email:                 u.Email,
		FirstName:             u.FirstName,
		LastName:              u.LastName,
		Avatar:                u.Avatar,
		IsVerifiedHost:        u.IsVerifiedHost,
		IsBannedAllProperties: u.IsBannedAllProperties,
		IsBannedReview:        u.IsBannedReview,
		Roles:                 toRoleNames(u.Roles),
		CreatedAt:             u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toRoleNames(roles []models.Role) []string {
	names := make([]string, 0, len(roles))
	for _, r := range roles {
		names = append(names, r.Name)
	}
	return names
}
