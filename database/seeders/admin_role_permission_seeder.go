package seeders

import (
	"log"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedAdminPermissions(db *gorm.DB) {
	permissionNames := []string{
		// user management
		"view_users",
		"create_users",
		"update_users",
		"delete_users",
		// role management
		"view_roles",
		"create_roles",
		"update_roles",
		"delete_roles",
		// permission management
		"view_permissions",
		"create_permissions",
		"update_permissions",
		"delete_permissions",
		// group permission management
		"view_group_permissions",
		"create_group_permissions",
		"update_group_permissions",
		"delete_group_permissions",
		// property management
		"view_properties",
		"create_properties",
		"create_property_images",
		"update_properties",
		"delete_properties",
		// property type management
		"view_property_types",
		"create_property_types",
		"update_property_types",
		"delete_property_types",
		// property image management
		"view_property_images",
		"create_property_images",
		"update_property_images",
		"delete_property_images",
		// facility management
		"view_facilities",
		"create_facilities",
		"update_facilities",
		"delete_facilities",
		// facility category management
		"view_facility_categories",
		"create_facility_categories",
		"update_facility_categories",
		"delete_facility_categories",
		// room category management
		"view_room_categories",
		"create_room_categories",
		"update_room_categories",
		"delete_room_categories",
		// room category calendar management
		"view_room_category_calendars",
		"create_room_category_calendars",
		"update_room_category_calendars",
		"delete_room_category_calendars",
		// booking management
		"view_bookings",
		"view_booking_for_hosts",
		"create_bookings",
		"create_booking_reviews",
		"update_bookings",
		"confirm_booking_payments",
		"cancel_bookings",
		"success_bookings",
		"delete_bookings",
		// booking order management
		"view_booking_orders",
		"create_booking_orders",
		"update_booking_orders",
		"delete_booking_orders",
		// pet management
		"view_pets",
		"create_pets",
		"update_pets",
		"delete_pets",
		// pet breed management
		"view_pet_breeds",
		"create_pet_breeds",
		"update_pet_breeds",
		"delete_pet_breeds",
		// pet profile management
		"view_pet_profiles",
		"create_pet_profiles",
		"update_pet_profiles",
		"delete_pet_profiles",
		// pet profile energy management
		"view_pet_profile_energies",
		"create_pet_profile_energies",
		"update_pet_profile_energies",
		"delete_pet_profile_energies",
		// favorite folder management
		"view_favorite_folders",
		"create_favorite_folders",
		"update_favorite_folders",
		"delete_favorite_folders",
		// audit log management
		"view_audit_logs",
		"create_audit_logs",
		// notification management
		"view_notifications",
		"create_notifications",
		"update_notifications",
		"delete_notifications",
		// user verification management
		"view_user_verifications",
		"manage_user_verifications",

		//
		"manage_properties",
		// admin dashboard management
		"view_admin_dashboard",
		// report management
		"view_reports",
		"action_reports",
	}

	var permissions []models.Permission
	if err := db.Where("name IN ?", permissionNames).Find(&permissions).Error; err != nil {
		log.Printf("Failed to fetch permissions: %v", err)
		return
	}

	var role models.Role
	if err := db.Where("name = ?", "admin").First(&role).Error; err != nil {
		log.Printf("admin role not found: %v", err)
		return
	}

	if err := db.Model(&role).Association("Permissions").Replace(&permissions); err != nil {
		log.Printf("Failed to assign permissions to admin role: %v", err)
		return
	}

	log.Printf("Successfully assigned %d permissions to admin role", len(permissions))
}
