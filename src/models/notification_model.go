package models

import "gorm.io/gorm"

// NotificationCategory enum type for notification categories
type NotificationCategory string

const (
	NotificationCategoryMessage  NotificationCategory = "message"
	NotificationCategoryReminder NotificationCategory = "reminder"
	NotificationCategoryNews     NotificationCategory = "news"
	NotificationCategoryAlert    NotificationCategory = "alert"
)

// NotificationType enum type for notification types
type NotificationType string

const (
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeDefault NotificationType = "default"
)

type Notification struct {
	BaseModel

	// Fields
	Title           string               `json:"title" gorm:"column:title;type:varchar(100);not null;"`
	Message         string               `json:"message" gorm:"column:message;type:text;"`
	Category        NotificationCategory `json:"category" gorm:"column:category;type:varchar(50);index;"`
	Type            NotificationType     `json:"type" gorm:"column:type;type:varchar(50);index;"`
	IsRead          bool                 `json:"isRead" gorm:"column:is_read;default:false;"`
	ReferenceTable  string               `json:"referenceTable" gorm:"column:reference_table;type:varchar(100);"`
	ReferenceID     *int                 `json:"referenceId" gorm:"column:reference_id;"`

	// Relations
	UserID             *int                `json:"userId" gorm:"column:user_id;index;"`
	User               *User               `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	RoleID             *int                `json:"roleId" gorm:"column:role_id;index;"`
	Role               *Role               `json:"role,omitempty" gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	NotificationActionID *int              `json:"notificationActionId" gorm:"column:notification_action_id;index;"`
	NotificationAction   *NotificationAction `json:"notificationAction,omitempty" gorm:"foreignKey:NotificationActionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
