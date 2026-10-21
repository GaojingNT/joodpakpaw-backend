package dtos

import "mime/multipart"

type (
	CreateUser struct {
		Email           string `json:"email" form:"email" query:"email" validate:"required,email,max=100"`
		Password        string `json:"password" form:"password" query:"password" validate:"required,min=8,has_uppercase,has_lowercase,has_number,has_special"`
		ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" query:"confirmPassword" validate:"required,eqfield=Password"`
	}
	UpdateUser struct {
		FirstName  string                `json:"firstName" form:"firstName" query:"firstName" validate:"omitempty,max=50"`
		LastName   string                `json:"lastName" form:"lastName" query:"lastName" validate:"omitempty,max=50"`
		Avatar     string                `json:"avatar" form:"avatar" query:"avatar" validate:"omitempty,max=255"`
		AvatarFile *multipart.FileHeader `json:"-" form:"-" query:"-" validate:"-"`
	}
	VerifyOTP struct {
		Email string `json:"email" form:"email" query:"email" validate:"required,email,max=100"`
		OTP   string `json:"otp" form:"otp" query:"otp" validate:"required,len=6"`
	}
	AddRoleToUser struct {
		RoleName string `json:"roleName" validate:"required"`
	}
	PatchChangePassword struct {
		OldPassword     string `json:"oldPassword" form:"oldPassword" query:"oldPassword" validate:"required,min=8"`
		NewPassword     string `json:"newPassword" form:"newPassword" query:"newPassword" validate:"required,min=8,has_uppercase,has_lowercase,has_number,has_special"`
		ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" query:"confirmPassword" validate:"required,eqfield=NewPassword"`
	}
	AdminActionUser struct {
		AdminNote string `json:"adminNote"`
	}
)
