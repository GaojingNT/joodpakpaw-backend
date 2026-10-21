package seeders

import (
	"log"
	"strings"

	"gorm.io/gorm"
	"joodpakpaw/src/models"
)

func SeedGroupPermissionsWithPermissions(db *gorm.DB) error {
	groupPermissionsWithPermissions := map[string][]struct {
		Name        string
		Description string
	}{
		"user_management": {
			{Name: "view_users", Description: ""},
			{Name: "create_users", Description: ""},
			{Name: "update_users", Description: ""},
			{Name: "delete_users", Description: ""},
		},
		"role_management": {
			{Name: "view_roles", Description: ""},
			{Name: "create_roles", Description: ""},
			{Name: "update_roles", Description: ""},
			{Name: "delete_roles", Description: ""},
		},
		"permission_management": {
			{Name: "view_permissions", Description: ""},
			{Name: "create_permissions", Description: ""},
			{Name: "update_permissions", Description: ""},
			{Name: "delete_permissions", Description: ""},
		},
		"group_permission_management": {
			{Name: "view_group_permissions", Description: ""},
			{Name: "create_group_permissions", Description: ""},
			{Name: "update_group_permissions", Description: ""},
			{Name: "delete_group_permissions", Description: ""},
		},
		"property_management": {
			{Name: "view_properties", Description: ""},
			{Name: "view_host_properties", Description: ""},
			{Name: "create_properties", Description: ""},
			{Name: "create_property_images", Description: ""},
			{Name: "update_properties", Description: ""},
			{Name: "delete_properties", Description: ""},
			{Name: "manage_properties", Description: "Permission to manage properties, including creating, updating, and deleting properties and their related entities."},
		},
		"property_type_management": {
			{Name: "view_property_types", Description: ""},
			{Name: "create_property_types", Description: ""},
			{Name: "update_property_types", Description: ""},
			{Name: "delete_property_types", Description: ""},
		},
		"property_image_management": {
			{Name: "view_property_images", Description: ""},
			{Name: "create_property_images", Description: ""},
			{Name: "update_property_images", Description: ""},
			{Name: "delete_property_images", Description: ""},
		},
		"facility_management": {
			{Name: "view_facilities", Description: ""},
			{Name: "create_facilities", Description: ""},
			{Name: "update_facilities", Description: ""},
			{Name: "delete_facilities", Description: ""},
		},
		"facility_category_management": {
			{Name: "view_facility_categories", Description: ""},
			{Name: "create_facility_categories", Description: ""},
			{Name: "update_facility_categories", Description: ""},
			{Name: "delete_facility_categories", Description: ""},
		},
		"room_management": {
			{Name: "view_rooms", Description: ""},
			{Name: "create_rooms", Description: ""},
			{Name: "update_rooms", Description: ""},
			{Name: "delete_rooms", Description: ""},
		},
		"room_category_management": {
			{Name: "view_room_categories", Description: ""},
			{Name: "create_room_categories", Description: ""},
			{Name: "update_room_categories", Description: ""},
			{Name: "delete_room_categories", Description: ""},
		},
		"room_category_calendar_management": {
			{Name: "view_room_category_calendars", Description: ""},
			{Name: "create_room_category_calendars", Description: ""},
			{Name: "update_room_category_calendars", Description: ""},
			{Name: "delete_room_category_calendars", Description: ""},
		},
		"booking_management": {
			{Name: "view_bookings", Description: ""},
			{Name: "view_booking_for_hosts", Description: ""},
			{Name: "create_booking_reviews", Description: ""},
			{Name: "create_bookings", Description: ""},
			{Name: "update_bookings", Description: ""},
			{Name: "confirm_booking_payments", Description: ""},
			{Name: "cancel_bookings", Description: ""},
			{Name: "success_bookings", Description: ""},
			{Name: "delete_bookings", Description: ""},
		},
		"booking_order_management": {
			{Name: "view_booking_orders", Description: ""},
			{Name: "create_booking_orders", Description: ""},
			{Name: "update_booking_orders", Description: ""},
			{Name: "delete_booking_orders", Description: ""},
		},
		"pet_management": {
			{Name: "view_pets", Description: ""},
			{Name: "create_pets", Description: ""},
			{Name: "update_pets", Description: ""},
			{Name: "delete_pets", Description: ""},
		},
		"pet_breed_management": {
			{Name: "view_pet_breeds", Description: ""},
			{Name: "create_pet_breeds", Description: ""},
			{Name: "update_pet_breeds", Description: ""},
			{Name: "delete_pet_breeds", Description: ""},
		},
		"pet_profile_management": {
			{Name: "view_pet_profiles", Description: ""},
			{Name: "create_pet_profiles", Description: ""},
			{Name: "update_pet_profiles", Description: ""},
			{Name: "delete_pet_profiles", Description: ""},
		},
		"pet_profile_energy_management": {
			{Name: "view_pet_profile_energies", Description: ""},
			{Name: "create_pet_profile_energies", Description: ""},
			{Name: "update_pet_profile_energies", Description: ""},
			{Name: "delete_pet_profile_energies", Description: ""},
		},
		"favorite_folder_management": {
			{Name: "view_favorite_folders", Description: ""},
			{Name: "create_favorite_folders", Description: ""},
			{Name: "update_favorite_folders", Description: ""},
			{Name: "delete_favorite_folders", Description: ""},
		},
		"audit_log_management": {
			{Name: "view_audit_logs", Description: ""},
			{Name: "create_audit_logs", Description: ""},
		},
		"notification_management": {
			{Name: "view_notifications", Description: ""},
			{Name: "create_notifications", Description: ""},
			{Name: "update_notifications", Description: ""},
			{Name: "delete_notifications", Description: ""},
		},
		"user_verification_management": {
			{Name: "view_user_verifications", Description: ""},
			{Name: "manage_user_verifications", Description: ""},
		},
		"admin_dashboard_management": {
			{Name: "view_admin_dashboard", Description: ""},
		},
		"report_management": {
			{Name: "view_reports", Description: ""},
			{Name: "action_reports", Description: ""},
		},
	}

	for groupName, permissions := range groupPermissionsWithPermissions {
		// Check if the group permission exists
		var group models.GroupPermission
		if err := db.Where("name = ?", groupName).First(&group).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create the group permission if it does not exist
				group = models.GroupPermission{Name: groupName}
				if err := db.Create(&group).Error; err != nil {
					log.Printf("Failed to create group permission '%s': %v", groupName, err)
					continue
				}
				log.Printf(" Created group permission '%s'", groupName)
			} else {
				log.Printf("Error finding group permission '%s': %v", groupName, err)
				continue
			}
		} else {
			log.Printf("- Group permission '%s' already exists", groupName)
		}

		// Link permissions to the group permission
		for _, perm := range permissions {
			var permission models.Permission
			if err := db.Where("name = ?", perm.Name).First(&permission).Error; err != nil {
				// If permission doesn't exist, create it
				permission = models.Permission{
					Name:        strings.ToLower(strings.ReplaceAll(perm.Name, " ", "_")),
					Description: perm.Description,
				}
				if err := db.Create(&permission).Error; err != nil {
					log.Printf("Failed to create permission '%s': %v", perm.Name, err)
					continue
				}
				log.Printf("   Created permission '%s'", perm.Name)
			} else {
				// Update the description if the permission exists
				permission.Description = perm.Description
				if err := db.Save(&permission).Error; err != nil {
					log.Printf("Failed to update permission '%s': %v", perm.Name, err)
					continue
				}
			}

			// Check if permission is already linked to the group
			var count int64
			db.Model(&group).Where("id = ?", permission.ID).Association("Permissions").Count()
			if count == 0 {
				// Add the permission to the group
				if err := db.Model(&group).Association("Permissions").Append(&permission); err != nil {
					log.Printf("Failed to link permission '%s' to group '%s': %v", perm.Name, groupName, err)
				} else {
					log.Printf("   Linked permission '%s' to group '%s'", perm.Name, groupName)
				}
			}
		}
	}

	log.Println(" Seeding group permissions with permissions completed")
	return nil
}
