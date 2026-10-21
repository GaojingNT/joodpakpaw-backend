package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedReportCategories(db *gorm.DB) {
	categories := []models.ReportCategory{
		{Name: "misleading_photos", Label: "Misleading Photos", Type: models.ReportTypeProperty},
		{Name: "scam_fraud", Label: "Scam / Fraud", Type: models.ReportTypeProperty},
		{Name: "wrong_location", Label: "Wrong Location", Type: models.ReportTypeProperty},
		{Name: "inaccurate_info", Label: "Inaccurate Info", Type: models.ReportTypeProperty},
		{Name: "inappropriate_content", Label: "Inappropriate Content", Type: models.ReportTypeReview},
		{Name: "fake_review", Label: "Fake Review", Type: models.ReportTypeReview},
		{Name: "hate_speech", Label: "Hate Speech", Type: models.ReportTypeReview},
		{Name: "other", Label: "Other", Type: models.ReportTypeReview},
		{Name: "other", Label: "Other", Type: models.ReportTypeProperty},
		{Name: "unban_property", Label: "Unban Property", Type: models.ReportTypeUnbanProperty},
		{Name: "unban_all_properties", Label: "Unban All Properties", Type: models.ReportTypeUnbanProperty},
		{Name: "unban_review", Label: "Unban Review", Type: models.ReportTypeUnbanReview},
	}

	for _, category := range categories {
		var existing models.ReportCategory
		if err := db.Where("name = ? AND type = ?", category.Name, category.Type).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&category).Error; err != nil {
					log.Printf("Failed to create report category '%s': %v", category.Name, err)
					continue
				}
				log.Printf("✓ Created report category '%s'", category.Name)
			} else {
				log.Printf("Error finding report category '%s': %v", category.Name, err)
			}
		} else {
			log.Printf("- Report category '%s' already exists", category.Name)
		}
	}

	log.Println("✓ Seeding report categories completed")
}
