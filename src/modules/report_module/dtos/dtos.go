package dtos

import (
	"mime/multipart"

	"joodpakpaw/src/models"
)

type (
	CreateReportCategory struct {
		Name string            `json:"name" validate:"required,max=255"`
		Type models.ReportType `json:"type" validate:"required,oneof=review property"`
	}

	UpdateReportCategory struct {
		Name string            `json:"name" validate:"omitempty,max=255"`
		Type models.ReportType `json:"type" validate:"omitempty,oneof=review property"`
	}

	CreateReportReview struct {
		ReportCategoryID *int                    `json:"reportCategoryId" form:"reportCategoryId" validate:"required"`
		BookingID        *int                    `json:"bookingId" form:"bookingId" validate:"required"`
		Description      string                  `json:"description" form:"description" validate:"omitempty"`
		Images           []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty"`
	}

	CreateReportProperty struct {
		ReportCategoryID *int                    `json:"reportCategoryId" form:"reportCategoryId" validate:"required"`
		PropertyID       *int                    `json:"propertyId" form:"propertyId" validate:"required"`
		Description      string                  `json:"description" form:"description" validate:"omitempty"`
		Images           []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty"`
	}

	// CreateReportUnbanProperty - propertyId optional: ระบุ = unban property เดียว, ไม่ระบุ = unban ทั้งหมด
	CreateReportUnbanProperty struct {
		ReportCategoryID *int                    `json:"reportCategoryId" form:"reportCategoryId" validate:"required"`
		PropertyID       *int                    `json:"propertyId" form:"propertyId" validate:"omitempty"`
		Description      string                  `json:"description" form:"description" validate:"omitempty"`
		Images           []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty"`
	}

	CreateReportUnbanReview struct {
		ReportCategoryID *int                    `json:"reportCategoryId" form:"reportCategoryId" validate:"required"`
		Description      string                  `json:"description" form:"description" validate:"omitempty"`
		Images           []*multipart.FileHeader `json:"images" form:"images" validate:"omitempty"`
	}

	UpdateReportStatus struct {
		Status models.ReportStatus `json:"status" validate:"required,oneof=pending resolved dismissed"`
	}

	ActionReport struct {
		Action    string `json:"action" validate:"required,oneof=dismiss_report warn_user deactivate_property delete_review ban_user_review deactivate_all_properties unban_property unban_all_properties unban_review"`
		AdminNote string `json:"adminNote" validate:"required"`
	}
)
