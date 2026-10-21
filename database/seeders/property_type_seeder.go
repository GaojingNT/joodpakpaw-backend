package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedPropertyTypes(db *gorm.DB) error {
	propertyTypes := []models.PropertyType{
		{
			Name:  "house",
			Label: "House",
		},
		{
			Name:  "apartment_condo",
			Label: "Apartment/Condo",
		},
		{
			Name:  "pet_hotel",
			Label: "Pet Hotel",
		},
	}

	for _, propertyType := range propertyTypes {
		// Check if property type already exists
		var existingPropertyType models.PropertyType
		result := db.Where("name = ?", propertyType.Name).First(&existingPropertyType)

		if result.Error == gorm.ErrRecordNotFound {
			// Property type doesn't exist, create it
			if err := db.Create(&propertyType).Error; err != nil {
				log.Printf("Failed to create property type '%s': %v", propertyType.Name, err)
				return err
			}
			log.Printf("✓ Property type '%s' created successfully", propertyType.Name)
		} else if result.Error != nil {
			// Some other error occurred
			log.Printf("Error checking property type '%s': %v", propertyType.Name, result.Error)
			return result.Error
		} else {
			// Property type already exists
			log.Printf("- Property type '%s' already exists, skipping", propertyType.Name)
		}
	}

	return nil
}
