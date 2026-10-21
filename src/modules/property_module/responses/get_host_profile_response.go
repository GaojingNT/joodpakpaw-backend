package responses

import (
	"joodpakpaw/internal/http_response"
	"joodpakpaw/src/models"
)

type (
	GetHostProfileResponse struct {
		http_response.OkResponse
		Host HostProfileDetailResponse `json:"host"`
	}
	HostProfileDetailResponse struct {
		ID                               int                    `json:"id"`
		FirstName                        string                 `json:"firstName"`
		LastName                         string                 `json:"lastName"`
		Avatar                           string                 `json:"avatar"`
		ChatRoomID                       *int                   `json:"chatRoomId"`
		IsVerifiedHost                   bool                   `json:"isVerifiedHost"`
		IsBannedAllProperties            bool                   `json:"isBannedAllProperties"`
		UnbanAllPropertiesRequestCount   int64                  `json:"unbanAllPropertiesRequestCount"`
		HasPendingUnbanAllPropertiesRequest bool                `json:"hasPendingUnbanAllPropertiesRequest"`
		Roles                            []string               `json:"roles"`
		CreatedAt                        string                 `json:"createdAt"`
		AvgRating                        float64                `json:"avgRating"`
		ReviewCount                      int                    `json:"reviewCount"`
		TotalProperties                  int                    `json:"totalProperties"`
		Properties                       []PropertyListResponse `json:"properties"`
	}
)

func (r *GetHostProfileResponse) Make(host models.User, properties []models.Property, avgRating float64, reviewCount int, unbanAllCount int64, hasPendingUnbanAll bool) *GetHostProfileResponse {
	activeProperties := make([]models.Property, 0)
	for _, p := range properties {
		if p.IsActive && !p.IsBannedByAdmin {
			activeProperties = append(activeProperties, p)
		}
	}

	roles := make([]string, 0, len(host.Roles))
	for _, role := range host.Roles {
		roles = append(roles, role.Name)
	}

	r.OkResponse = http_response.OkResponse{Code: "OK", Message: "Success"}
	r.Host = HostProfileDetailResponse{
		ID:                               host.ID,
		FirstName:                        host.FirstName,
		LastName:                         host.LastName,
		Avatar:                           host.Avatar,
		IsVerifiedHost:                   host.IsVerifiedHost,
		IsBannedAllProperties:            host.IsBannedAllProperties,
		UnbanAllPropertiesRequestCount:   unbanAllCount,
		HasPendingUnbanAllPropertiesRequest: hasPendingUnbanAll,
		CreatedAt:                        host.CreatedAt.Format("2006-01-02 15:04:05"),
		AvgRating:                        avgRating,
		ReviewCount:                      reviewCount,
		TotalProperties:                  len(activeProperties),
		Properties:                       new(GetPropertiesCollection).Collection(activeProperties, nil),
		Roles:                            roles,
	}
	return r
}
