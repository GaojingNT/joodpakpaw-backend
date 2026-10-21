package validator

import (
	"regexp"
	"unicode"

	"github.com/go-playground/validator"
)

// PasswordStrengthConfig กำหนดค่า requirement ของ password
type PasswordStrengthConfig struct {
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
	MinLength        int
}

// DefaultPasswordStrengthConfig ค่า default ของ password strength
var DefaultPasswordStrengthConfig = PasswordStrengthConfig{
	RequireUppercase: true,
	RequireLowercase: true,
	RequireNumber:    true,
	RequireSpecial:   true,
	MinLength:        8,
}

// ValidatePasswordStrength ตรวจสอบความแข็งแกร่งของ password
func ValidatePasswordStrength(password string, config PasswordStrengthConfig) bool {
	// ตรวจสอบความยาวขั้นต่ำ
	if len(password) < config.MinLength {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	// วนลูปตรวจสอบแต่ละตัวอักษร
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// ตรวจสอบตามเงื่อนไขที่กำหนด
	if config.RequireUppercase && !hasUpper {
		return false
	}
	if config.RequireLowercase && !hasLower {
		return false
	}
	if config.RequireNumber && !hasNumber {
		return false
	}
	if config.RequireSpecial && !hasSpecial {
		return false
	}

	return true
}

// Custom validator functions for go-playground/validator

// passwordStrength - validator tag: password_strength
func passwordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return ValidatePasswordStrength(password, DefaultPasswordStrengthConfig)
}

// hasUppercase - validator tag: has_uppercase
func hasUppercase(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return regexp.MustCompile(`[A-Z]`).MatchString(password)
}

// hasLowercase - validator tag: has_lowercase
func hasLowercase(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return regexp.MustCompile(`[a-z]`).MatchString(password)
}

// hasNumber - validator tag: has_number
func hasNumber(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return regexp.MustCompile(`[0-9]`).MatchString(password)
}

// hasSpecial - validator tag: has_special
func hasSpecial(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
}

// RegisterPasswordValidators ลงทะเบียน custom validators ทั้งหมด
func RegisterPasswordValidators(v *validator.Validate) {
	v.RegisterValidation("password_strength", passwordStrength)
	v.RegisterValidation("has_uppercase", hasUppercase)
	v.RegisterValidation("has_lowercase", hasLowercase)
	v.RegisterValidation("has_number", hasNumber)
	v.RegisterValidation("has_special", hasSpecial)
}
