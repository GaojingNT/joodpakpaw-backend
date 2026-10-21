package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedHostPermissions(db *gorm.DB) {
	// Define the roles we want to assign permissions to
	roleNames := []string{"host"}

	// Define the permission names that all roles should have
	permissionNames := []string{
		// room category management
		"view_room_categories",
		"create_room_categories",
		"update_room_categories",
		"delete_room_categories",
		// property management
		"view_properties",
		"view_host_properties",
		"create_properties",
		"update_properties",
		"delete_properties",
		"create_property_images",
		// booking management
		"view_booking_for_hosts",
		"cancel_bookings",
		"success_bookings",
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
