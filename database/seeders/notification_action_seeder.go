package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedNotificationActions(db *gorm.DB) error {
	actions := []models.NotificationAction{
		{Name: "booking"},
		{Name: "review"},
		{Name: "payment"},
		{Name: "reminder"},
		{Name: "cancel"},
		{Name: "report"},
		{Name: "verification"},
	}

	for _, action := range actions {
		var existing models.NotificationAction
		result := db.Where("name = ?", action.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&action).Error; err != nil {
				log.Printf("Failed to create notification action '%s': %v", action.Name, err)
				return err
			}
			log.Printf("✓ Notification action '%s' created successfully", action.Name)
		} else if result.Error != nil {
			log.Printf("Error checking notification action '%s': %v", action.Name, result.Error)
			return result.Error
		} else {
			log.Printf("- Notification action '%s' already exists, skipping", action.Name)
		}
	}

	return nil
}
