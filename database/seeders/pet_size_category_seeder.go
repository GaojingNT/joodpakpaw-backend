package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedPetSizeCategories(db *gorm.DB) error {
	petSizeCategories := []models.PetSizeCategory{
		{
			Name: "small",
			Code: 0,
			Min:  0,
			Max:  14, // Under 14 kg
		},
		{
			Name: "medium",
			Code: 1,
			Min:  14,
			Max:  25, // 14-25 kg
		},
		{
			Name: "large",
			Code: 2,
			Min:  25,
			Max:  999, // Over 25 kg (using 999 as upper limit)
		},
	}

	for _, petSizeCategory := range petSizeCategories {
		// Check if pet size category already exists by code (unique constraint)
		var existingPetSizeCategory models.PetSizeCategory
		result := db.Where("code = ?", petSizeCategory.Code).First(&existingPetSizeCategory)

		if result.Error == gorm.ErrRecordNotFound {
			// Pet size category doesn't exist, create it
			if err := db.Create(&petSizeCategory).Error; err != nil {
				log.Printf("Failed to create pet size category '%s': %v", petSizeCategory.Name, err)
				return err
			}
			log.Printf("✓ Pet size category '%s' created successfully", petSizeCategory.Name)
		} else if result.Error != nil {
			// Some other error occurred
			log.Printf("Error checking pet size category '%s': %v", petSizeCategory.Name, result.Error)
			return result.Error
		} else {
			// Pet size category already exists
			log.Printf("- Pet size category '%s' (code: %d) already exists, skipping", petSizeCategory.Name, petSizeCategory.Code)
		}
	}

	return nil
}
