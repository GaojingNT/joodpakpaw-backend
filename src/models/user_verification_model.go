package models

import "time"

type VerificationStatus string

const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusApproved VerificationStatus = "approved"
	VerificationStatusRejected VerificationStatus = "rejected"
	VerificationStatusRevoked  VerificationStatus = "revoked"
)

type UserVerification struct {
	BaseModel
	UserID                    int                `json:"userId"            gorm:"column:user_id;not null;index;"`
	User                      *User              `json:"user,omitempty"    gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status                    VerificationStatus `json:"status"            gorm:"column:status;type:varchar(20);default:'pending';not null;"`
	IsResubmit                bool               `json:"isResubmit"        gorm:"column:is_resubmit;default:false;not null;"`
	SubmittedAt               time.Time          `json:"submittedAt"       gorm:"column:submitted_at;not null;"`
	RejectDescription         *string            `json:"rejectDescription"         gorm:"column:reject_description;type:text;"`
	PreviousRejectDescription *string            `json:"previousRejectDescription" gorm:"column:previous_reject_description;type:text;"`
	IdPhoto                   string             `json:"idPhoto"           gorm:"column:id_photo;type:text;not null;"`
	SelfieWithId              string             `json:"selfieWithId"      gorm:"column:selfie_with_id;type:text;not null;"`
}
