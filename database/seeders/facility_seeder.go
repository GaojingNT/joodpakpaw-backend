package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedFacilityCategoriesWithFacilities(db *gorm.DB) error {
	categoriesWithFacilities := map[string][]string{
		"indoor": {
			"Air Conditioning",
			"Heating",
			"Individual Pet Rooms",
			"Pet Suites",
			"Indoor Play Area",
			"Comfortable Bedding",
		},
		"outdoor": {
			"Outdoor Play Area",
			"Dog Park",
			"Cat Garden",
			"Pet Swimming Pool",
			"Walking Trail",

		},
		"entertainment": {
			"Toys and Games",
			"Interactive Play Equipment",
			"Puzzle Feeders",
			"Treat Dispensers",
			"Ball Pit",
		},
		"safety": {
			"24/7 Surveillance Camera",
			"Fire Alarm System",
			"Emergency Exit",
			"First Aid Kit",
		},
		"service": {
			"Professional Pet Care Staff",
			"Grooming Service",
			"Bathing Service",
			"Feeding Service",
		},
	}

	for categoryName, facilities := range categoriesWithFacilities {
		// Check if the category exists
		var category models.FacilityCategory
		if err := db.Where("name = ?", categoryName).First(&category).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create the category if it does not exist
				category = models.FacilityCategory{Name: categoryName}
				if err := db.Create(&category).Error; err != nil {
					log.Printf("Failed to create facility category '%s': %v", categoryName, err)
					continue
				}
				log.Printf("✓ Created facility category '%s'", categoryName)
			} else {
				log.Printf("Error finding facility category '%s': %v", categoryName, err)
				continue
			}
		} else {
			log.Printf("- Facility category '%s' already exists", categoryName)
		}

		// Create facilities for this category
		for _, facilityName := range facilities {
			var facility models.Facility
			if err := db.Where("name = ? AND category_id = ?", facilityName, category.ID).First(&facility).Error; err != nil {
				// If facility doesn't exist, create it
				categoryID := int(category.ID)
				facility = models.Facility{
					Name:       facilityName,
					CategoryID: &categoryID,
				}
				if err := db.Create(&facility).Error; err != nil {
					log.Printf("Failed to create facility '%s': %v", facilityName, err)
					continue
				}
				log.Printf("   ✓ Created facility '%s'", facilityName)
			} else {
				log.Printf("   - Facility '%s' already exists", facilityName)
			}
		}
	}

	log.Println("✓ Seeding facility categories with facilities completed")
	return nil
}
