package auth_module

import (
	"context"
	"database/sql"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"joodpakpaw/internal/encryption"
	"joodpakpaw/internal/exception"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/auth_module/dtos"
	"joodpakpaw/src/modules/auth_module/responses"
)

func (s Service) Authenticate(ctx context.Context, dto *dtos.SignInDto) (map[string]string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AuthenticateService", trace.WithAttributes(attribute.String("service", "Authenticate")))

	tokens, err := s.authRepository().Authenticate(ctx, dto.Email, dto.Password)

	s.tracer.TraceEnd(childSpan)

	return tokens, err
}

func (s Service) AuthenticateWithGoogle(ctx context.Context, dto *dtos.GoogleSignInDto) (map[string]string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AuthenticateWithGoogleService", trace.WithAttributes(attribute.String("service", "AuthenticateWithGoogle")))

	tokens, err := s.authRepository().AuthenticateWithGoogle(ctx, dto.IDToken)

	s.tracer.TraceEnd(childSpan)

	return tokens, err
}

func (s Service) ForgotPassword(ctx context.Context, dto *dtos.ForgotPasswordDto) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "ForgotPasswordService", trace.WithAttributes(attribute.String("service", "ForgotPassword")))

	err := s.authRepository().ForgotPassword(ctx, dto.Email)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) ResetPassword(ctx context.Context, dto *dtos.ResetPasswordDto) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "ResetPasswordService", trace.WithAttributes(attribute.String("service", "ResetPassword")))

	err := s.authRepository().ResetPassword(ctx, dto.Email, dto.Password, dto.ResetPasswordToken)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) SignUp(ctx context.Context, dto *dtos.SignUpDto) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "SignUpService", trace.WithAttributes(attribute.String("service", "SignUp")))

	// Generate OTP
	otp := s.mailService.GenerateOTP()

	// Calculate OTP expiration (10 minutes from now)
	otpExpiredAt := sql.NullTime{
		Time:  time.Now().UTC().Add(10 * time.Minute),
		Valid: true,
	}

	// Check if user already exists by email
	existingUser, err := s.authRepository().GetUserByEmail(ctx, dto.Email)

	if err == nil && existingUser != nil {
		// User exists - check if already verified
		if existingUser.IsVerified {
			// User is already verified - cannot sign up again
			s.tracer.TraceEnd(childSpan)
			return exception.ErrUserAlreadyExists
		}

		// User exists but not verified - update OTP and send email
		err = s.authRepository().ResendOTP(ctx, dto.Email, otp, &otpExpiredAt)
		if err != nil {
			s.tracer.TraceEnd(childSpan)
			return err
		}

		// Send OTP via email
		err = s.mailService.SendOTP(existingUser.Email, otp)
		if err != nil {
			s.logger.Error("Failed to send OTP email", zap.Error(err), zap.String("email", existingUser.Email))
			// Don't return error here - OTP is already updated
		}

		s.tracer.TraceEnd(childSpan)
		return nil
	}

	// Check if phone number already exists (only if phone is provided)
	if dto.Phone != "" {
		existingPhoneUser, err := s.authRepository().GetUserByPhone(ctx, dto.Phone)
		if err == nil && existingPhoneUser != nil && existingPhoneUser.IsVerified {
			// Phone number already exists and is verified
			s.tracer.TraceEnd(childSpan)
			return exception.ErrPhoneAlreadyExists
		}
	}

	// User doesn't exist - create new user
	encryptedPassword := encryption.EncryptPassword(dto.Password, "")

	user := new(models.User)
	user.Email = dto.Email
	user.FirstName = dto.FirstName
	user.LastName = dto.LastName
	user.Phone = dto.Phone
	user.Password = encryptedPassword
	user.OTP = otp
	user.OTPExpiredAt = &otpExpiredAt
	user.IsVerified = false

	// Create user in database with guest role
	err = s.authRepository().CreateUserWithRole(ctx, user, "guest")
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	// Send OTP via email
	err = s.mailService.SendOTP(user.Email, otp)
	if err != nil {
		s.logger.Error("Failed to send OTP email", zap.Error(err), zap.String("email", user.Email))
		// Don't return error here - user is already created
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) VerifyOTP(ctx context.Context, dto *dtos.VerifyOTPDto) (map[string]string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "VerifyOTPService", trace.WithAttributes(attribute.String("service", "VerifyOTP")))

	tokens, err := s.authRepository().VerifyOTP(ctx, dto.Email, dto.OTP)

	s.tracer.TraceEnd(childSpan)

	return tokens, err
}

func (s Service) RefreshToken(ctx context.Context, dto *dtos.RefreshTokenDto) (map[string]string, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "RefreshTokenService", trace.WithAttributes(attribute.String("service", "RefreshToken")))

	tokens, err := s.authRepository().RefreshToken(ctx, dto.RefreshToken)

	s.tracer.TraceEnd(childSpan)

	return tokens, err
}

func (s Service) ResendOTP(ctx context.Context, dto *dtos.ResendOTPDto) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "ResendOTPService", trace.WithAttributes(attribute.String("service", "ResendOTP")))

	// Generate new OTP
	otp := s.mailService.GenerateOTP()

	// Calculate OTP expiration (10 minutes from now)
	otpExpiredAt := sql.NullTime{
		Time:  time.Now().UTC().Add(10 * time.Minute),
		Valid: true,
	}

	// Update OTP in database
	err := s.authRepository().ResendOTP(ctx, dto.Email, otp, &otpExpiredAt)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	// Send OTP via email
	err = s.mailService.SendOTP(dto.Email, otp)
	if err != nil {
		s.logger.Error("Failed to send OTP email", zap.Error(err), zap.String("email", dto.Email))
		// Don't return error here - OTP is already updated in database
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}
func (s Service) GetProfile(ctx context.Context, userID int) (*responses.GetProfileResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetProfileService", trace.WithAttributes(attribute.String("service", "GetProfile")))

	profile, err := s.authRepository().GetProfile(ctx, userID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	unbanAllCount, _ := s.authRepository().CountUnbanAllPropertyRequests(ctx, userID)
	hasPendingUnbanAll, _ := s.authRepository().HasPendingUnbanAllPropertyRequest(ctx, userID)

	s.tracer.TraceEnd(childSpan)

	return new(responses.GetProfileResponse).Make(profile, unbanAllCount, hasPendingUnbanAll), err
}
