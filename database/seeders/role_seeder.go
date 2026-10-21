package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedRoles(db *gorm.DB) error {
	roles := []models.Role{
		{
			Name:        "guest",
			Label:       "Guest",
			Description: "Guest user role with basic access",
		},
		{
			Name:        "host",
			Label:       "Host",
			Description: "Host user role with property management access",
		},
		{
			Name:        "admin",
			Label:       "Admin",
			Description: "Administrator role with full system access",
		},
	}

	for _, role := range roles {
		// Check if role already exists
		var existingRole models.Role
		result := db.Where("name = ?", role.Name).First(&existingRole)

		if result.Error == gorm.ErrRecordNotFound {
			// Role doesn't exist, create it
			if err := db.Create(&role).Error; err != nil {
				log.Printf("Failed to create role '%s': %v", role.Name, err)
				return err
			}
			log.Printf("✓ Role '%s' created successfully", role.Name)
		} else if result.Error != nil {
			// Some other error occurred
			log.Printf("Error checking role '%s': %v", role.Name, result.Error)
			return result.Error
		} else {
			// Role already exists
			log.Printf("- Role '%s' already exists, skipping", role.Name)
		}
	}

	return nil
}
