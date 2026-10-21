package models

import "gorm.io/gorm"

type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusResolved  ReportStatus = "resolved"
	ReportStatusDismissed ReportStatus = "dismissed"
)

type ReportAction string

const (
	ReportActionDismissReport        ReportAction = "dismiss_report"
	ReportActionWarnUser             ReportAction = "warn_user"
	ReportActionDeactivateProperty   ReportAction = "deactivate_property"
	ReportActionDeleteReview         ReportAction = "delete_review"
	ReportActionBanUserReview        ReportAction = "ban_user_review"
	ReportActionDeactivateProperties ReportAction = "deactivate_all_properties"
	ReportActionUnbanProperty        ReportAction = "unban_property"
	ReportActionUnbanAllProperties   ReportAction = "unban_all_properties"
	ReportActionUnbanReview          ReportAction = "unban_review"
)

type Report struct {
	BaseModel

	// Fields
	Description    string        `json:"description" gorm:"column:description;type:text;"`
	Status         ReportStatus  `json:"status" gorm:"column:status;type:report_status;default:'pending';"`
	AdminNote      string        `json:"adminNote" gorm:"column:admin_note;type:text;"`
	ResolvedAction *ReportAction `json:"resolvedAction" gorm:"column:resolved_action;size:50;"`

	// Relationships
	ReportCategoryID *int            `json:"reportCategoryId" gorm:"column:report_category_id;not null;"`
	ReportCategory   *ReportCategory `json:"reportCategory,omitempty" gorm:"foreignKey:ReportCategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	ReporterID       *int            `json:"reporterId" gorm:"column:reporter_id;not null;"`
	Reporter         *User           `json:"reporter,omitempty" gorm:"foreignKey:ReporterID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BookingID        *int            `json:"bookingId,omitempty" gorm:"column:booking_id;"`
	Booking          *Booking        `json:"booking,omitempty" gorm:"foreignKey:BookingID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PropertyID       *int            `json:"propertyId,omitempty" gorm:"column:property_id;"`
	Property         *Property       `json:"property,omitempty" gorm:"foreignKey:PropertyID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Images           []ReportImage   `json:"images,omitempty" gorm:"foreignKey:ReportID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}
