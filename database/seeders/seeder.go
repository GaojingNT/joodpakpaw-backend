package seeders

import (
	"log"

	"gorm.io/gorm"
)

// SeedDatabase รวบรวมทุก Seeder ไว้ในที่เดียว
func SeedDatabase(db *gorm.DB) {
	log.Println("Seeding database...")

	SeedGroupPermissionsWithPermissions(db)
	SeedRoles(db)
	SeedNotificationActions(db)
	SeedAdminPermissions(db)
	SeedHostPermissions(db)
	SeedGuestPermissions(db)
	SeedPropertyTypes(db)
	SeedPets(db)
	SeedPetBreeds(db)
	SeedPetSizeCategories(db)
	SeedFacilityCategoriesWithFacilities(db)
	SeedUserBehaviorCategories(db)
	SeedReportCategories(db)
	SeedUsers(db)
	SeedProperties(db)
	SeedPropertyImages(db)

	log.Println("Database seeding completed successfully!")
}
