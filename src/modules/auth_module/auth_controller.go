package auth_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/encryption"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/validator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/auth_module/dtos"
)

// SignIn sign in
//
//	@Summary		Sign in
//	@Description	Sign in with email and password
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.SignInDto	true	"Sign in credentials"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		401		{object}	http_response.OkResponse
//	@Failure		403		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/signin [post]
func (c Controller) SignIn(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "SignInController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "SignIn")))
		tokens    map[string]string
		err       error
	)

	// Create data transfer object
	dto := new(dtos.SignInDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	tokens, err = c.authService().Authenticate(ctx, dto)
	if err != nil {
		// Handle specific authentication errors
		switch err {
		case exception.ErrAccountNotVerified:
			return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.AccountNotVerifiedResponseError, err)
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidLoginCredentialResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidLoginCredentialResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":         0,
		"message":      "OK",
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
		"tokenType":    "Bearer",
		"expiresIn":    30 * 60, // 30 minutes in seconds
	})
}

// GoogleSignIn sign in with Google
//
//	@Summary		Sign in with Google
//	@Description	Sign in (or sign up as guest) with a Google ID token from Google Identity Services
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.GoogleSignInDto	true	"Google ID token"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		401		{object}	http_response.OkResponse
//	@Failure		503		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/google [post]
func (c Controller) GoogleSignIn(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GoogleSignInController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GoogleSignIn")))
		tokens    map[string]string
		err       error
	)

	// Create data transfer object
	dto := new(dtos.GoogleSignInDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	tokens, err = c.authService().AuthenticateWithGoogle(ctx, dto)
	if err != nil {
		switch err {
		case exception.ErrGoogleSignInNotConfigured:
			return exception.HttpErrorResponseMapping(f, fiber.StatusServiceUnavailable, exception.GoogleSignInNotConfiguredResponseError, err)
		case exception.ErrInvalidGoogleToken:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidGoogleTokenResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":         0,
		"message":      "OK",
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
		"tokenType":    "Bearer",
		"expiresIn":    30 * 60, // 30 minutes in seconds
	})
}

// ForgotPassword request forgot password
//
//	@Summary		Forgot password
//	@Description	Request password reset by sending OTP to email
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.ForgotPasswordDto	true	"Email address"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		403		{object}	http_response.OkResponse
//	@Failure		404		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/forgot-password [post]
func (c Controller) ForgotPassword(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "ForgotPasswordController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ForgotPassword")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.ForgotPasswordDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.authService().ForgotPassword(ctx, dto)
	if err != nil {
		// Handle specific errors with proper HTTP status codes
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrAccountNotVerified:
			return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.AccountNotVerifiedResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}

// ResetPassword request reset password
//
//	@Summary		Reset password
//	@Description	Reset password using reset token from email
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.ResetPasswordDto	true	"Reset password data"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		401		{object}	http_response.OkResponse
//	@Failure		404		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/reset-password [post]
func (c Controller) ResetPassword(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "ResetPasswordController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ResetPassword")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.ResetPasswordDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.authService().ResetPassword(ctx, dto)
	if err != nil {
		// Handle specific errors with proper HTTP status codes
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrInvalidResetPasswordToken:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidResetPasswordTokenResponseError, err)
		case exception.ErrResetPasswordTokenExpired:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.ResetPasswordTokenExpiredResponseError, err)
		case exception.ErrResetPasswordTokenUsed:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.ResetPasswordTokenUsedResponseError, err)
		case exception.ErrPasswordNotMatch:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.PasswordNotMatchResponseError, err)
		case exception.ErrNewPasswordSameAsOld:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.NewPasswordSameAsOldResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}

// EncryptPassword encrypt password
//
//	@Summary		Encrypt password
//	@Description	Encrypt password for testing purposes
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.EncryptPasswordDto	true	"Password to encrypt"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/encrypt-password [post]
func (c Controller) EncryptPassword(f *fiber.Ctx) error {
	var (
		_, span = c.m.tracer.TraceStart(f.Context(), "EncryptPasswordController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "EncryptPassword")))
	)

	// Create data transfer object
	dto := new(dtos.EncryptPasswordDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	encryptedPassword := encryption.EncryptPassword(dto.Password, "")

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":              0,
		"message":           "OK",
		"encryptedPassword": encryptedPassword,
	})
}

// SignUp creates a new user account
//
//	@Summary		Sign up
//	@Description	Create a new user account with email verification
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.SignUpDto	true	"Sign up data"
//	@Success		201		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		409		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/sign-up [post]
func (c Controller) SignUp(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "SignUpController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "SignUp")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.SignUpDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.authService().SignUp(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusCreated).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}

// VerifyOTP verifies the OTP sent to user's email
//
//	@Summary		Verify OTP
//	@Description	Verify OTP code sent to user's email for account activation
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.VerifyOTPDto	true	"Email and OTP code"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		401		{object}	http_response.OkResponse
//	@Failure		409		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/verify-otp [post]
func (c Controller) VerifyOTP(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "VerifyOTPController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "VerifyOTP")))
		tokens    map[string]string
		err       error
	)

	// Create data transfer object
	dto := new(dtos.VerifyOTPDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	tokens, err = c.authService().VerifyOTP(ctx, dto)
	if err != nil {
		// Handle specific OTP errors with proper HTTP status codes
		switch err {
		case exception.ErrInvalidOTP:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidOTPResponseError, err)
		case exception.ErrOTPExpired:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.OTPExpiredResponseError, err)
		case exception.ErrEmailAlreadyVerified:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.EmailAlreadyVerifiedResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":         0,
		"message":      "OK",
		"accessToken":  tokens["accessToken"],
		"refreshToken": tokens["refreshToken"],
		"tokenType":    "Bearer",
		"expiresIn":    30 * 60, // 30 minutes in seconds
	})
}

// RefreshToken generates a new access token using refresh token
//
//	@Summary		Refresh access token
//	@Description	Generate a new access token using refresh token
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.RefreshTokenDto	true	"Refresh token"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/refresh-token [post]
func (c Controller) RefreshToken(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "RefreshTokenController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "RefreshToken")))
		tokens    map[string]string
		err       error
	)

	// Create data transfer object
	dto := new(dtos.RefreshTokenDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	tokens, err = c.authService().RefreshToken(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":        0,
		"message":     "OK",
		"accessToken": tokens["accessToken"],
		"tokenType":   "Bearer",
		"expiresIn":   30 * 60, // 30 minutes in seconds
	})
}

// ResendOTP resends OTP to user's email
//
//	@Summary		Resend OTP
//	@Description	Resend OTP code to user's email for account verification
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.ResendOTPDto	true	"Email address"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		404		{object}	http_response.OkResponse
//	@Failure		409		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/resend-otp [post]
func (c Controller) ResendOTP(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "ResendOTPController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ResendOTP")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.ResendOTPDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.authService().ResendOTP(ctx, dto)
	if err != nil {
		// Handle specific errors with proper HTTP status codes
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrEmailAlreadyVerified:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.EmailAlreadyVerifiedResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}

// GetProfile gets the current user's profile
//
//	@Summary		Get Profile
//	@Description	Get current user's profile information
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	http_response.OkResponse
//	@Failure		401	{object}	http_response.OkResponse
//	@Failure		500	{object}	http_response.OkResponse
//	@Router			/api/v1/auth/profile [get]
func (c Controller) GetProfile(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetProfileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetProfile")))
		err       error
	)

	// Get user from JWT token
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
	}

	// Call service function
	profile, err := c.authService().GetProfile(ctx, authUser.ID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
		"data":    profile,
	})
}
