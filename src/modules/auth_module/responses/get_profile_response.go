package responses

import (
	"time"

	"joodpakpaw/src/models"
)

type (
	GetProfileResponse struct {
		ID             int    `json:"id"`
		Email          string `json:"email"`
		FirstName      string `json:"firstName"`
		LastName       string `json:"lastName"`
		Avatar         string `json:"avatar"`
		Phone          string `json:"phone"`
		IsVerified                          bool   `json:"isVerified"`
		IsBecomeToHost                      bool   `json:"isBecomeToHost"`
		IsVerifiedHost                      bool   `json:"isVerifiedHost"`
		IsBannedAllProperties               bool   `json:"isBannedAllProperties"`
		UnbanAllPropertiesRequestCount      int64  `json:"unbanAllPropertiesRequestCount"`
		HasPendingUnbanAllPropertiesRequest bool   `json:"hasPendingUnbanAllPropertiesRequest"`

		// Relations
		Role []GetProfileRoleResponse `json:"role"`

		// Timestamp & Audit fields
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
	// Role and permissions
	GetProfileRoleResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		// Relations
		Permissions []GetProfileRolePermissionResponse `json:"permissions"`
	}
	GetProfileRoleGroupRoleResponse struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
	}

	GetProfileRolePermissionResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (response *GetProfileResponse) Make(user *models.User, unbanAllCount int64, hasPendingUnbanAll bool) *GetProfileResponse {
	response.ID = user.ID
	response.Email = user.Email
	response.FirstName = user.FirstName
	response.LastName = user.LastName
	response.Avatar = user.Avatar
	response.Phone = user.Phone
	response.IsVerified = user.IsVerified
	response.IsBecomeToHost = user.IsBecomeToHost
	response.IsVerifiedHost = user.IsVerifiedHost
	response.IsBannedAllProperties = user.IsBannedAllProperties
	response.UnbanAllPropertiesRequestCount = unbanAllCount
	response.HasPendingUnbanAllPropertiesRequest = hasPendingUnbanAll
	response.CreatedAt = user.CreatedAt
	response.UpdatedAt = user.UpdatedAt

	// Map roles
	response.Role = make([]GetProfileRoleResponse, 0)
	for _, role := range user.Roles {
		roleResponse := GetProfileRoleResponse{
			ID:   role.ID,
			Name: role.Name,
		}

		// Map permissions
		roleResponse.Permissions = make([]GetProfileRolePermissionResponse, 0)
		for _, perm := range role.Permissions {
			roleResponse.Permissions = append(roleResponse.Permissions, GetProfileRolePermissionResponse{
				ID:   perm.ID,
				Name: perm.Name,
			})
		}

		response.Role = append(response.Role, roleResponse)
	}

	return response
}
