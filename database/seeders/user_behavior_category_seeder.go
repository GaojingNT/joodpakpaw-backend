package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedUserBehaviorCategories(db *gorm.DB) error {
	categories := []models.UserBehaviorCategory{
		{Name: "view_property"},
		{Name: "create_booking"},
	}

	for _, category := range categories {
		var existing models.UserBehaviorCategory
		result := db.Where("name = ?", category.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&category).Error; err != nil {
				log.Printf("Failed to create user behavior category '%s': %v", category.Name, err)
				return err
			}
			log.Printf("✓ User behavior category '%s' created successfully", category.Name)
		} else if result.Error != nil {
			log.Printf("Error checking user behavior category '%s': %v", category.Name, result.Error)
			return result.Error
		} else {
			log.Printf("- User behavior category '%s' already exists, skipping", category.Name)
		}
	}

	return nil
}
