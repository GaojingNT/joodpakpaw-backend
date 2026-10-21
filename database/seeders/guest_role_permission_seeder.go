package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedGuestPermissions(db *gorm.DB) {
	// Define the roles we want to assign permissions to
	roleNames := []string{"guest"}

	// Define the permission names that all roles should have
	permissionNames := []string{
		// user management
		"view_users",
		"update_users",
		// pet profile management
		"view_pet_profiles",
		"create_pet_profiles",
		"update_pet_profiles",
		"delete_pet_profiles",
		// favorite management
		"view_favorite_folders",
		"create_favorite_folders",
		"update_favorite_folders",
		"delete_favorite_folders",
		// booking management
		"view_bookings",
		"create_bookings",
		"confirm_booking_payments",
		"cancel_bookings",
		"create_booking_reviews",
		//
		"view_properties",
		"create_properties",	
		

	}

	// Get the permissions from the database
	var permissions []models.Permission
	if err := db.Where("name IN ?", permissionNames).Find(&permissions).Error; err != nil {
		log.Printf("Failed to fetch permissions: %v", err)
		return
	}

	// For each role, find it and assign permissions
	for _, roleName := range roleNames {
		var role models.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			log.Printf("%s role not found: %v", roleName, err)
			continue
		}

		// Assign permissions to the role
		if err := db.Model(&role).Association("Permissions").Replace(&permissions); err != nil {
			log.Printf("Failed to assign permissions to %s role: %v", roleName, err)
			continue
		}

		log.Printf("Successfully assigned %d permissions to %s role", len(permissions), roleName)
	}
}
