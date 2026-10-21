package models

import "gorm.io/gorm"

type Role struct {
	BaseModel
	// BaseAuditorModel // For role management, we need to know who created and updated the role

	// Fields
	Name        string `json:"name" gorm:"column:name;size:100;uniqueIndex;not null"`
	Label       string `json:"label" gorm:"column:label;size:100;null;"`
	Description string `json:"description" gorm:"column:description;type:text;null;"`

	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:roles_permissions;constraint:OnDelete:CASCADE;"`
	Users       []User       `json:"users,omitempty" gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;"`

	// Soft deletes
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

type RolePermission struct {
	RoleID       int `json:"roleID"`
	PermissionID int `json:"permissionID"`
}

func (m *RolePermission) TableName() string {
	return "roles_permissions"
}
