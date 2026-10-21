package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

// SeedPetProfileEnergies seeds energy levels 1-5 in order; the frontend sends the level itself as petEnergyId
func SeedPetProfileEnergies(db *gorm.DB) error {
	petProfileEnergies := []models.PetProfileEnergy{
		{Rating: 1, Description: "Lazy"},
		{Rating: 2, Description: "Calm"},
		{Rating: 3, Description: "Moderate"},
		{Rating: 4, Description: "Active"},
		{Rating: 5, Description: "Hyper"},
	}

	for _, petProfileEnergy := range petProfileEnergies {
		// Check if pet profile energy already exists by rating
		var existingPetProfileEnergy models.PetProfileEnergy
		result := db.Where("rating = ?", petProfileEnergy.Rating).First(&existingPetProfileEnergy)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&petProfileEnergy).Error; err != nil {
				log.Printf("Failed to create pet profile energy '%s': %v", petProfileEnergy.Description, err)
				return err
			}
			log.Printf("✓ Pet profile energy '%s' created successfully", petProfileEnergy.Description)
		} else if result.Error != nil {
			log.Printf("Error checking pet profile energy '%s': %v", petProfileEnergy.Description, result.Error)
			return result.Error
		} else {
			log.Printf("- Pet profile energy '%s' (rating: %d) already exists, skipping", petProfileEnergy.Description, petProfileEnergy.Rating)
		}
	}

	return nil
}
