package models

import (
	"time"

	"gorm.io/gorm"
)

// AuditLogAction enum type for audit log actions
type AuditLogAction string

const (
	AuditLogActionView   AuditLogAction = "view"
	AuditLogActionCreate AuditLogAction = "create"
	AuditLogActionUpdate AuditLogAction = "update"
	AuditLogActionDelete AuditLogAction = "delete"
)

type AuditLog struct {
	ID string `json:"id" gorm:"column:id;primarykey;type:varchar(255);"`

	// Fields
	ModelName  string         `json:"modelName" gorm:"column:model_name;type:varchar(255);index;"`
	Messages   string         `json:"messages" gorm:"column:messages;type:text;"`
	RecordID   string         `json:"recordId" gorm:"column:record_id;type:varchar(255);index;"`
	Action     AuditLogAction `json:"action" gorm:"column:action;type:varchar(50);index;"`
	OldValue   string         `json:"oldValue" gorm:"column:old_value;type:text;"`
	NewValue   string         `json:"newValue" gorm:"column:new_value;type:text;"`
	EditValue  string         `json:"editValue" gorm:"column:edit_value;type:text;"`
	UserAgent  string         `json:"userAgent" gorm:"column:user_agent;type:varchar(500);"`
	CreatedAt  time.Time      `json:"createdAt" gorm:"column:created_at;default:CURRENT_TIMESTAMP;not null;"`
	
	// Relations
	UserID *int `json:"userId" gorm:"column:user_id;index;"`
	User   *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}

// TableName specifies the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}
