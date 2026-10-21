package responses

import "joodpakpaw/src/models"

type (
	GetUsersCollection struct {
		UserResponse
	}
)

func (response *GetUsersCollection) Collection(users []models.User) []GetUsersCollection {
	var userCollection []GetUsersCollection

	for _, user := range users {
		userResponse := GetUsersCollection{
			UserResponse: UserResponse{
				ID:        user.ID,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,

				CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
			},
		}
		userCollection = append(userCollection, userResponse)
	}

	return userCollection
}
