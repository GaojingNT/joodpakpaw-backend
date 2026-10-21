package dtos

type (
	SignInDto struct {
		Email    string `json:"email" form:"email" query:"email" validate:"required,email"`
		Password string `json:"password" form:"password" query:"password" validate:"required,min=8"`
	}
	GoogleSignInDto struct {
		IDToken string `json:"idToken" form:"idToken" query:"idToken" validate:"required"`
	}
	SignUpDto struct {
		Email           string `json:"email" form:"email" query:"email" validate:"required,email"`
		FirstName       string `json:"firstName" form:"firstName" query:"firstName" validate:"max=100"`
		LastName        string `json:"lastName" form:"lastName" query:"lastName" validate:"max=100"`
		Phone           string `json:"phone" form:"phone" query:"phone" validate:"max=20"`
		Password        string `json:"password" form:"password" query:"password" validate:"required,min=8,has_uppercase,has_lowercase,has_number,has_special"`
		ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" query:"confirmPassword" validate:"required,eqfield=Password"`
	}
	VerifyOTPDto struct {
		Email string `json:"email" form:"email" query:"email" validate:"required,email"`
		OTP   string `json:"otp" form:"otp" query:"otp" validate:"required,len=6,numeric"`
	}
	ForgotPasswordDto struct {
		Email string `json:"email" form:"email" query:"email" validate:"required,max=50,email"`
	}
	ResetPasswordDto struct {
		Email              string `json:"email" form:"email" query:"email" validate:"required,max=50,email"`
		Password           string `json:"password" form:"password" query:"password" validate:"required,min=8,has_uppercase,has_lowercase,has_number,has_special"`
		ConfirmPassword    string `json:"confirmPassword" form:"confirmPassword" query:"confirmPassword" validate:"required,eqfield=Password"`
		ResetPasswordToken string `json:"resetPasswordToken" form:"resetPasswordToken" query:"resetPasswordToken" validate:"required"`
	}
	EncryptPasswordDto struct {
		Password string `json:"password" form:"password" query:"password" validate:"required,min=8"`
	}
	RefreshTokenDto struct {
		RefreshToken string `json:"refreshToken" form:"refreshToken" query:"refreshToken" validate:"required"`
	}
	ResendOTPDto struct {
		Email string `json:"email" form:"email" query:"email" validate:"required,email"`
	}
)
