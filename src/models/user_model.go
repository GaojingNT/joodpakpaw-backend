package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type User struct {
	BaseModel

	// Fields
	Email                  string        `json:"email" gorm:"column:email;size:100;uniqueIndex;not null;"`
	FirstName              string        `json:"firstName" gorm:"column:first_name;size:100;"`
	LastName               string        `json:"lastName" gorm:"column:last_name;size:100;"`
	Avatar                 string        `json:"avatar" gorm:"column:avatar;type:text;"`
	Phone                  string        `json:"phone" gorm:"column:phone;size:20;"`
	Password               string        `json:"-" gorm:"column:password;type:text;"`
	GoogleID               *string       `json:"-" gorm:"column:google_id;size:255;uniqueIndex;"`
	ResetPasswordToken     string        `json:"-" gorm:"column:reset_password_token;type:text;"`
	ResetPasswordExpiredAt *sql.NullTime `json:"-" gorm:"column:reset_password_expired_at;type:timestamp;"`
	OTP                    string        `json:"-" gorm:"column:otp;size:6;"`
	OTPExpiredAt           *sql.NullTime `json:"-" gorm:"column:otp_expired_at;type:timestamp;"`
	IsVerified             bool          `json:"isVerified" gorm:"column:is_verified;default:false;"`
	IsBecomeToHost         bool          `json:"isBecomeToHost" gorm:"column:is_become_to_host;default:false;"`
	IsVerifiedHost         bool          `json:"isVerifiedHost" gorm:"column:is_verified_host;default:false;"`
	IsBannedReview         bool          `json:"isBannedReview" gorm:"column:is_banned_review;default:false;"`
	IsBannedAllProperties  bool       `json:"isBannedAllProperties" gorm:"column:is_banned_all_properties;default:false;"`
	BannedAt               *time.Time `json:"bannedAt"              gorm:"column:banned_at;"`
	// User has Many roles
	Roles           []Role           `json:"roles,omitempty" gorm:"many2many:user_roles;constraint:OnDelete:CASCADE;"`
	PetProfiles     []PetProfile     `json:"petProfiles,omitempty" gorm:"foreignKey:GuestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Properties      []Property       `json:"properties,omitempty" gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Bookings        []Booking        `json:"bookings,omitempty" gorm:"foreignKey:GuestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AuditLogs       []AuditLog       `json:"auditLogs,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Notifications   []Notification   `json:"notifications,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FavoriteFolders []FavoriteFolder `json:"favoriteFolders,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserBehaviors   []UserBehavior   `json:"userBehaviors,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	RecentViews     []UserRecentView     `json:"recentViews,omitempty"     gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Verifications   []UserVerification   `json:"verifications,omitempty"   gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}

// Check if user has role
func (m *User) HasRole(roleName string) bool {
	// Check preload
	if len(m.Roles) == 0 {
		return false
	}

	for _, role := range m.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// Check if user has permission
func (m *User) HasPermission(permission string) bool {
	// Check preload
	if len(m.Roles) == 0 {
		return false
	}

	for _, role := range m.Roles {
		for _, rolePermission := range role.Permissions {
			if rolePermission.Name == permission {
				return true
			}
		}
	}
	return false
}

// User searchable attributes
func UserSearchable() []string {
	return []string{"email", "first_name", "last_name"}
}
