package auth_module

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/goccy/go-json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/idtoken"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"joodpakpaw/config"
	"joodpakpaw/internal/encryption"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/helpers"
	"joodpakpaw/internal/jwt"
	"joodpakpaw/internal/utils"
	"joodpakpaw/src/models"
)

func (r Repository) Authenticate(ctx context.Context, email string, password string) (map[string]string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AuthenticateRepository", trace.WithAttributes(attribute.String("repository", "Authenticate")))
		authUser     *models.User
		isAuth       bool
		tokens       map[string]string
		err          error
	)

	// CASE: Authenticate user
	// 1. Get user by email
	// 2. Attempt password
	// 3. Generate access token and refresh token

	utils.Block{
		Try: func() {
			// Get user by email
			authUser, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			log.Println("authUser passed")

			// Check if user is verified
			if !authUser.IsVerified {
				err = exception.ErrAccountNotVerified
				utils.Throw(err)
			}

			// Accounts created with Google sign-in have no password
			if authUser.Password == "" {
				err = exception.ErrInvalidLoginCredential
				utils.Throw(err)
			}

			// Attempt password
			if isAuth, err = r.Attempt(ctx, authUser, password); err != nil {
				utils.Throw(err)
			} else if !isAuth {
				err = exception.ErrInvalidLoginCredential
				utils.Throw(err)
			}

			log.Println("Attempt passed")

			tokens, err = r.IssueTokens(ctx, authUser)
			if err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			} else if err == exception.ErrAccountNotVerified || err == exception.ErrInvalidLoginCredential {
				// Expected errors (no logging needed)
			} else {
				// Logging unexpected errors
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// IssueTokens generates the access token (2 hours) and refresh token (7 days) for a signed-in user
func (r Repository) IssueTokens(ctx context.Context, authUser *models.User) (map[string]string, error) {
	var (
		accessToken  string
		refreshToken string
		err          error
	)

	utils.Block{
		Try: func() {
			userPayload := map[string]interface{}{
				"user": map[string]interface{}{
					"id":        authUser.ID,
					"email":     authUser.Email,
					"firstName": authUser.FirstName,
					"lastName":  authUser.LastName,
					"roles": func() []map[string]interface{} {
						var roles []map[string]interface{}

						for _, role := range authUser.Roles {
							roles = append(roles, map[string]interface{}{
								"id":   role.ID,
								"name": role.Name,
							})
						}

						return roles
					}(),
				},
			}

			// Generate access token (2 hours)
			accessToken, err = r.GenerateJWT(
				ctx,
				userPayload,
				2*time.Hour,
			)
			if err != nil {
				utils.Throw(err)
			}

			log.Println("GenerateJWT passed")

			// Generate refresh token (7 days)
			refreshToken, err = r.GenerateJWT(
				ctx,
				map[string]interface{}{
					"user": map[string]interface{}{
						"id":    authUser.ID,
						"email": authUser.Email,
					},
					"type": "refresh",
				},
				7*24*time.Hour,
			)
			if err != nil {
				utils.Throw(err)
			}

			log.Println("GenerateRefreshToken passed")
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
		},
		Finally: nil,
	}.Do()

	// Check error
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}, nil
}

func (r Repository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserByIDRepository", trace.WithAttributes(attribute.String("repository", "GetUserByID")))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Get user by email
			if err = r.db.
				Preload("Roles.Permissions").
				First(&user, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			} else {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserByEmailRepository", trace.WithAttributes(attribute.String("repository", "GetUserByEmail")))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Get user by email
			if err = r.db.
				Preload("Roles.Permissions").
				Where("email = ?", email).
				First(&user).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			} else {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r Repository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserByPhoneRepository", trace.WithAttributes(attribute.String("repository", "GetUserByPhone")))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Get user by phone
			if err = r.db.
				Preload("Roles.Permissions").
				Where("phone = ?", phone).
				First(&user).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			} else {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r Repository) Attempt(ctx context.Context, user *models.User, password string) (bool, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AttemptRepository", trace.WithAttributes(attribute.String("repository", "Attempt")))
		isAuth       bool
		err          error
	)

	utils.Block{
		Try: func() {
			// Compare password
			isAuth = encryption.VerifyPassword(password, user.Password)
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return false, err
	}

	return isAuth, nil
}

func (r Repository) ForgotPassword(ctx context.Context, email string) error {
	var (
		_, childSpan       = r.tracer.TraceStart(ctx, "ForgotPasswordRepository", trace.WithAttributes(attribute.String("repository", "ForgotPassword")))
		user               *models.User
		resetPasswordToken string
		err                error
	)

	// CASE: Forgot password
	// 1. Get user by email
	// 2. Generate reset password token
	// 3. Update user reset password token and reset password token expired date time
	// 4. Send reset password email

	utils.Block{
		Try: func() {
			// Get user by email
			user, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			// Check if user is verified
			if !user.IsVerified {
				err = exception.ErrAccountNotVerified
				utils.Throw(err)
			}

			// Generate reset password token
			resetPasswordToken, err = helpers.RandomHexString(16)
			if err != nil {
				utils.Throw(err)
			}

			// Update user reset password token
			if err = r.db.
				Model(&user).
				Updates(map[string]interface{}{
					"reset_password_token":      resetPasswordToken,
					"reset_password_expired_at": time.Now().UTC().Add(30 * time.Minute),
				}).Error; err != nil {
				utils.Throw(err)
			}

			// Send reset password email
			if err = r.SendResetPasswordEmail(ctx, user, resetPasswordToken); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			} else if err == exception.ErrRecordNotFound || err == exception.ErrAccountNotVerified {
				// Expected errors (no logging needed)
			} else {
				// Logging unexpected errors
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) SendResetPasswordEmail(ctx context.Context, user *models.User, resetPasswordToken string) error {
	var (
		_, childSpan         = r.tracer.TraceStart(ctx, "ForgotPasswordRepository", trace.WithAttributes(attribute.String("repository", "ForgotPassword")))
		mailConn             *gomail.Dialer
		mailMessage          = gomail.NewMessage()
		emailSubject         = "Reset Password"
		emailContent         string
		resetPasswordPayload = map[string]interface{}{"email": user.Email, "resetPasswordToken": resetPasswordToken}
		payloadBytes         []byte
		base64Payload        string
		resetPasswordUri     string
		err                  error
	)

	// CASE: Send reset password email
	// 1. Make email content
	// 2. Set email subject
	// 3. Set email receiver
	// 4. Send email

	utils.Block{
		Try: func() {
			// Make email content
			// 1. Encode payload
			// 2. Body (Content with reset password button)
			// 3. Footer (Signature)

			// Encode payload JSON
			payloadBytes, err = json.Marshal(resetPasswordPayload)
			if err != nil {
				utils.Throw(err)
			}

			// Base64 encode payload
			base64Payload = base64.StdEncoding.EncodeToString(payloadBytes)

			// Make reset password URI
			resetPasswordUri = fmt.Sprintf(
				"%s/authen/reset-password?payload=%s",
				config.Global.App.AppFrontendUrl,
				base64Payload,
			)

			// Body with modern UI template
			emailContent = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h1 style="color: #333333; text-align: center;">Reset Your Password</h1>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">Hello!</p>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">You are receiving this email because we received a password reset request for your account.</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="display: inline-block; background-color: #007bff; color: #ffffff; padding: 15px 40px; text-decoration: none; border-radius: 5px; font-size: 16px; font-weight: bold;">Reset Password</a>
        </div>

        <p style="color: #666666; font-size: 14px; line-height: 1.5;">This reset password link will expire in <strong>30 minutes</strong>.</p>
        <p style="color: #666666; font-size: 14px; line-height: 1.5;">If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>

        <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">

        <p style="color: #666666; font-size: 13px; line-height: 1.5;">If you're having trouble clicking the "Reset Password" button, copy and paste the URL below into your web browser:</p>
        <p style="color: #007bff; font-size: 12px; word-break: break-all;"><a href="%s" style="color: #007bff;">%s</a></p>

        <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">

        <p style="color: #999999; font-size: 12px; text-align: center;">© 2025 JoodPakPaw. All rights reserved.</p>
    </div>
</body>
</html>`,
				resetPasswordUri,
				resetPasswordUri,
				resetPasswordUri,
			)

			// Validation email
			_, err = mail.ParseAddress(config.Global.Mail.SmtpUsername)
			if err != nil {
				utils.Throw(err)
			}
			_, err = mail.ParseAddress(user.Email)
			if err != nil {
				utils.Throw(err)
			}

			// Set sender and receiver
			mailMessage.SetHeader("From", config.Global.Mail.SmtpUsername)
			mailMessage.SetHeader("To", user.Email)

			// Set subject
			mailMessage.SetHeader("Subject", emailSubject)

			// Set body
			mailMessage.SetBody("text/html", emailContent)

			// Setup dial
			mailConn = gomail.NewDialer(
				config.Global.Mail.SmtpHost,
				config.Global.Mail.SmtpPort,
				config.Global.Mail.SmtpUsername,
				config.Global.Mail.SmtpPassword,
			)

			// Retry parameters
			maxRetries := 3
			retryDelay := time.Second * 5

			// Loop send mail
			for attempt := 1; attempt <= maxRetries; attempt++ {
				if err = mailConn.DialAndSend(mailMessage); err != nil {
					time.Sleep(retryDelay)
					continue
				}
				break
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) ResetPassword(ctx context.Context, email string, password string, resetPasswordToken string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ResetPasswordRepository", trace.WithAttributes(attribute.String("repository", "ResetPassword")))
		user         *models.User
		err          error
	)

	// CASE: Reset password
	// 1. Get user by email
	// 2. Check reset password token
	// 3. Update user password and reset password token to empty string

	utils.Block{
		Try: func() {
			// Get user by email
			user, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			// Check if reset password token is empty (already used)
			if user.ResetPasswordToken == "" {
				err = exception.ErrResetPasswordTokenUsed
				utils.Throw(err)
			}

			// Check if reset password token matches
			if user.ResetPasswordToken != resetPasswordToken {
				err = exception.ErrInvalidResetPasswordToken
				utils.Throw(err)
			}

			// Check if reset password token is expired (using UTC for comparison)
			if user.ResetPasswordExpiredAt == nil || !user.ResetPasswordExpiredAt.Valid || time.Now().UTC().After(user.ResetPasswordExpiredAt.Time.UTC()) {
				err = exception.ErrResetPasswordTokenExpired
				utils.Throw(err)
			}

			// Check if new password is same as old password
			if encryption.VerifyPassword(password, user.Password) {
				err = exception.ErrNewPasswordSameAsOld
				utils.Throw(err)
			}

			// Update user password and reset password token
			if err = r.db.
				Model(&user).
				Updates(map[string]interface{}{
					"password":                  encryption.EncryptPassword(password, ""),
					"reset_password_token":      "",
					"reset_password_expired_at": sql.NullTime{Time: time.Time{}, Valid: false},
				}).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Handle expected errors (no logging needed)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else if err == exception.ErrRecordNotFound ||
				err == exception.ErrInvalidResetPasswordToken ||
				err == exception.ErrResetPasswordTokenExpired ||
				err == exception.ErrResetPasswordTokenUsed ||
				err == exception.ErrNewPasswordSameAsOld {
				// Keep the error as is (expected errors)
			} else {
				// Log unexpected errors
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) GenerateJWT(ctx context.Context, content interface{}, duration time.Duration) (string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GenerateJWTRepository", trace.WithAttributes(attribute.String("repository", "GenerateJWT")))
		token        string
		err          error
	)

	utils.Block{
		Try: func() {
			// Generate JWT
			token, err = jwt.JwtToken.Create(
				duration,
				content,
			)
			if err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			exception.SqlErrorMessage = err.Error()
			sentry.CaptureException(e.(error))
		},
		Finally: func() {},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return "", err
	}

	return token, nil
}

func (r Repository) GetProfile (ctx context.Context, userID int) (*models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetProfileRepository", trace.WithAttributes(attribute.String("repository", "GetProfile")))
		user         models.User
		err          error
	)
	
	utils.Block{
		Try: func() {
			// Get user by ID with roles and permissions
			if err = r.db.
				Preload("Roles.Permissions").
				First(&user, userID).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound { // CASE: User not found
				err = exception.ErrRecordNotFound
			} else {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)
	
	// Check error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r Repository) CreateUserWithRole(ctx context.Context, user *models.User, roleName string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateUserWithRoleRepository", trace.WithAttributes(attribute.String("repository", "CreateUserWithRole")))
		err          error
	)

	utils.Block{
		Try: func() {
			// Use transaction to ensure both user and role assignment succeed
			err = r.db.Transaction(func(tx *gorm.DB) error {
				// Create user
				if err = tx.Create(user).Error; err != nil {
					return err
				}

				// Find role by name
				var role models.Role
				if err = tx.Where("name = ?", roleName).First(&role).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
					}
					return err
				}

				// Assign role to user
				if err = tx.Model(&user).Association("Roles").Append(&role); err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Logging
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

func (r Repository) VerifyOTP(ctx context.Context, email string, otp string) (map[string]string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "VerifyOTPRepository", trace.WithAttributes(attribute.String("repository", "VerifyOTP")))
		user         *models.User
		accessToken  string
		refreshToken string
		err          error
	)

	utils.Block{
		Try: func() {
			// Get user by email
			user, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			// Check if user is already verified
			if user.IsVerified {
				err = exception.ErrEmailAlreadyVerified
				utils.Throw(err)
			}

			// Check if OTP is expired first (using UTC for comparison)
			if user.OTPExpiredAt == nil || !user.OTPExpiredAt.Valid || time.Now().UTC().After(user.OTPExpiredAt.Time.UTC()) {
				err = exception.ErrOTPExpired
				utils.Throw(err)
			}

			// Check if OTP matches
			if user.OTP != otp {
				err = exception.ErrInvalidOTP
				utils.Throw(err)
			}

			// Verify user
			if err = r.db.Model(&user).Updates(map[string]interface{}{
				"is_verified":    true,
				"otp":            "",
				"otp_expired_at": sql.NullTime{Time: time.Time{}, Valid: false},
			}).Error; err != nil {
				utils.Throw(err)
			}

			// Reload user with roles after verification
			user, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			// Generate JWT tokens
			userPayload := map[string]interface{}{
				"user": map[string]interface{}{
					"id":        user.ID,
					"email":     user.Email,
					"firstName": user.FirstName,
					"lastName":  user.LastName,
					"roles": func() []map[string]interface{} {
						var roles []map[string]interface{}

						for _, role := range user.Roles {
							roles = append(roles, map[string]interface{}{
								"id":   role.ID,
								"name": role.Name,
							})
						}

						return roles
					}(),
				},
			}

			// Generate access token (2 hours)
			accessToken, err = r.GenerateJWT(
				ctx,
				userPayload,
				2*time.Hour,
			)
			if err != nil {
				utils.Throw(err)
			}

			// Generate refresh token (7 days)
			refreshToken, err = r.GenerateJWT(
				ctx,
				map[string]interface{}{
					"user": map[string]interface{}{
						"id":    user.ID,
						"email": user.Email,
					},
					"type": "refresh",
				},
				7*24*time.Hour,
			)
			if err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else if err != exception.ErrInvalidOTP && err != exception.ErrOTPExpired && err != exception.ErrEmailAlreadyVerified {
				// Logging (only for unexpected errors)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}, nil
}

func (r Repository) RefreshToken(ctx context.Context, refreshToken string) (map[string]string, error) {
	var (
		_, childSpan   = r.tracer.TraceStart(ctx, "RefreshTokenRepository", trace.WithAttributes(attribute.String("repository", "RefreshToken")))
		payload        map[string]interface{}
		userID         int
		user           *models.User
		newAccessToken string
		err            error
	)

	utils.Block{
		Try: func() {
			// Verify refresh token
			payload, err = jwt.JwtToken.Payload(refreshToken)
			if err != nil {
				utils.Throw(err)
			}

			// Check if token type is refresh
			if tokenType, ok := payload["type"].(string); !ok || tokenType != "refresh" {
				err = exception.ErrInvalidToken
				utils.Throw(err)
			}

			// Get user data from payload
			userData, ok := payload["user"].(map[string]interface{})
			if !ok {
				err = exception.ErrInvalidTokenClaim
				utils.Throw(err)
			}

			userID = int(userData["id"].(float64))

			// Get user from database
			user, err = r.GetUserByID(ctx, userID)
			if err != nil {
				utils.Throw(err)
			}

			// Generate new access token (30 minutes)
			newAccessToken, err = r.GenerateJWT(
				ctx,
				map[string]interface{}{
					"user": map[string]interface{}{
						"id":        user.ID,
						"email":     user.Email,
						"firstName": user.FirstName,
						"lastName":  user.LastName,
						"roles": func() []map[string]interface{} {
							var roles []map[string]interface{}

							for _, role := range user.Roles {
								roles = append(roles, map[string]interface{}{
									"id":   role.ID,
									"name": role.Name,
								})
							}

							return roles
						}(),
					},
				},
				2*time.Hour,
			)
			if err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else if err != exception.ErrInvalidToken && err != exception.ErrInvalidTokenClaim && err != exception.ErrTokenExpired {
				// Logging
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"accessToken": newAccessToken,
	}, nil
}

func (r Repository) CountUnbanAllPropertyRequests(ctx context.Context, userID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_reports.reporter_id = ?", userID).
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.property_id IS NULL").
		Where("tbl_reports.deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

func (r Repository) HasPendingUnbanAllPropertyRequest(ctx context.Context, userID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_reports.reporter_id = ?", userID).
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.property_id IS NULL").
		Where("tbl_reports.status = ?", "pending").
		Where("tbl_reports.deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func (r Repository) ResendOTP(ctx context.Context, email string, newOTP string, otpExpiredAt *sql.NullTime) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ResendOTPRepository", trace.WithAttributes(attribute.String("repository", "ResendOTP")))
		user         *models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Get user by email
			user, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			// Check if user is already verified
			if user.IsVerified {
				err = exception.ErrEmailAlreadyVerified
				utils.Throw(err)
			}

			// Update OTP and expiration time
			if err = r.db.Model(&user).Updates(map[string]interface{}{
				"otp":            newOTP,
				"otp_expired_at": otpExpiredAt,
			}).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Handle expected errors (no logging needed)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else if err == exception.ErrRecordNotFound || err == exception.ErrEmailAlreadyVerified {
				// Keep the error as is (expected errors)
			} else {
				// Log unexpected errors
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return err
	}

	return nil
}

// AuthenticateWithGoogle verifies a Google ID token and signs the user in.
// A new guest account is created on first sign-in; an existing account with the same email is linked.
func (r Repository) AuthenticateWithGoogle(ctx context.Context, idToken string) (map[string]string, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AuthenticateWithGoogleRepository", trace.WithAttributes(attribute.String("repository", "AuthenticateWithGoogle")))
		authUser     *models.User
		tokens       map[string]string
		err          error
	)

	// CASE: Authenticate user with Google
	// 1. Validate ID token (signature, audience, expiry) and require a verified email
	// 2. Find user by Google ID, otherwise link by email, otherwise create a guest user
	// 3. Generate access token and refresh token

	utils.Block{
		Try: func() {
			clientID := config.Global.Auth.GoogleClientID
			if clientID == "" {
				err = exception.ErrGoogleSignInNotConfigured
				utils.Throw(err)
			}

			payload, validateErr := idtoken.Validate(ctx, idToken, clientID)
			if validateErr != nil {
				err = exception.ErrInvalidGoogleToken
				utils.Throw(err)
			}

			email, _ := payload.Claims["email"].(string)
			emailVerified, _ := payload.Claims["email_verified"].(bool)
			if email == "" || !emailVerified || payload.Subject == "" {
				err = exception.ErrInvalidGoogleToken
				utils.Throw(err)
			}
			googleID := payload.Subject
			firstName, _ := payload.Claims["given_name"].(string)
			lastName, _ := payload.Claims["family_name"].(string)
			picture, _ := payload.Claims["picture"].(string)

			var existing models.User
			findErr := r.db.Where("google_id = ?", googleID).First(&existing).Error
			if findErr == gorm.ErrRecordNotFound {
				findErr = r.db.Where("email = ?", email).First(&existing).Error
			}

			switch {
			case findErr == nil:
				// Link Google to the existing account (Google has verified the email)
				updates := map[string]interface{}{
					"google_id":   googleID,
					"is_verified": true,
				}
				if existing.Avatar == "" && picture != "" {
					updates["avatar"] = picture
				}
				if err = r.db.Model(&existing).Updates(updates).Error; err != nil {
					utils.Throw(err)
				}
				email = existing.Email
			case findErr == gorm.ErrRecordNotFound:
				newUser := &models.User{
					Email:      email,
					FirstName:  firstName,
					LastName:   lastName,
					Avatar:     picture,
					GoogleID:   &googleID,
					IsVerified: true,
				}
				if err = r.CreateUserWithRole(ctx, newUser, "guest"); err != nil {
					utils.Throw(err)
				}
			default:
				err = findErr
				utils.Throw(err)
			}

			// Reload with roles for the token payload
			authUser, err = r.GetUserByEmail(ctx, email)
			if err != nil {
				utils.Throw(err)
			}

			tokens, err = r.IssueTokens(ctx, authUser)
			if err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == exception.ErrInvalidGoogleToken || err == exception.ErrGoogleSignInNotConfigured {
				// Expected errors (no logging needed)
			} else {
				// Logging unexpected errors
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
