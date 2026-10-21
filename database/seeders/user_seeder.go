package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/internal/encryption"
	"joodpakpaw/src/models"
)

func SeedUsers(db *gorm.DB) error {
	// Find host role
	var hostRole models.Role
	if err := db.Where("name = ?", "host").First(&hostRole).Error; err != nil {
		log.Printf("Failed to find host role: %v", err)
		return err
	}

	users := []struct {
		Email     string
		FirstName string
		LastName  string
		Password  string
	}{
		{
			Email:     "host1@example.com",
			FirstName: "Host",
			LastName:  "One",
			Password:  "Password@1234",
		},
		{
			Email:     "host2@example.com",
			FirstName: "Host",
			LastName:  "Two",
			Password:  "Password@1234",
		},
		{
			Email:     "host3@example.com",
			FirstName: "Host",
			LastName:  "Three",
			Password:  "Password@1234",
		},
		{
			Email:     "host4@example.com",
			FirstName: "Host",
			LastName:  "Four",
			Password:  "Password@1234",
		},
		{
			Email:     "host5@example.com",
			FirstName: "Host",
			LastName:  "Five",
			Password:  "Password@1234",
		},
		{
			Email:     "host6@example.com",
			FirstName: "Host",
			LastName:  "Six",
			Password:  "Password@1234",
		},
	}

	for _, u := range users {
		var existing models.User
		result := db.Where("email = ?", u.Email).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			hashedPassword := encryption.EncryptPassword(u.Password, "")

			newUser := models.User{
				Email:          u.Email,
				FirstName:      u.FirstName,
				LastName:       u.LastName,
				Password:       hashedPassword,
				IsVerified:     true,
				IsBecomeToHost: true,
				IsVerifiedHost: true,
				Roles:          []models.Role{hostRole},
			}

			if err := db.Create(&newUser).Error; err != nil {
				log.Printf("Failed to create user '%s': %v", u.Email, err)
				return err
			}
			log.Printf("✓ User '%s' created successfully", u.Email)
		} else if result.Error != nil {
			log.Printf("Error checking user '%s': %v", u.Email, result.Error)
			return result.Error
		} else {
			log.Printf("- User '%s' already exists, skipping", u.Email)
		}
	}

	return nil
}
