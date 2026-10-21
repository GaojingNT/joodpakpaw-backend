package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedPets(db *gorm.DB) error {
	pets := []models.Pet{
		{
			Name: "dog",
		},
		{
			Name: "cat",
		},
	}

	for _, pet := range pets {
		// Check if pet already exists
		var existingPet models.Pet
		result := db.Where("name = ?", pet.Name).First(&existingPet)

		if result.Error == gorm.ErrRecordNotFound {
			// Pet doesn't exist, create it
			if err := db.Create(&pet).Error; err != nil {
				log.Printf("Failed to create pet '%s': %v", pet.Name, err)
				return err
			}
			log.Printf("✓ Pet '%s' created successfully", pet.Name)
		} else if result.Error != nil {
			// Some other error occurred
			log.Printf("Error checking pet '%s': %v", pet.Name, result.Error)
			return result.Error
		} else {
			// Pet already exists
			log.Printf("- Pet '%s' already exists, skipping", pet.Name)
		}
	}

	return nil
}
