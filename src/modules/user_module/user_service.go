package user_module

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"joodpakpaw/internal/encryption"
	"joodpakpaw/internal/exception"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/action_log_module"
	"joodpakpaw/src/modules/notification_module"
	"joodpakpaw/src/modules/user_module/dtos"
	"joodpakpaw/src/modules/user_module/responses"
)

func (s Service) AdminGetUsers(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AdminGetUsersService", trace.WithAttributes(attribute.String("service", "AdminGetUsers")))
	defer s.tracer.TraceEnd(childSpan)

	result, err := s.userRepository().AdminGetUsersPaginate(ctx, paginate)
	if err != nil {
		return nil, err
	}

	if result.Data != nil {
		result.Data = new(responses.AdminGetUsersCollection).Collection(result.Data.([]models.User))
	}
	return result, nil
}

func (s Service) AdminGetUserByID(ctx context.Context, id int) (*responses.AdminGetUserByIDResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AdminGetUserByIDService", trace.WithAttributes(attribute.String("service", "AdminGetUserByID")))
	defer s.tracer.TraceEnd(childSpan)

	user, err := s.userRepository().AdminGetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	violations, err := s.userRepository().CountUserViolations(ctx, id)
	if err != nil {
		violations = 0
	}

	return new(responses.AdminGetUserByIDResponse).Make(user, violations), nil
}

func (s Service) GetUsers(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetUsersService", trace.WithAttributes(attribute.String("service", "GetUsers")))

	result, err := s.userRepository().GetUserPaginate(ctx, paginate)

	// Convert the result to a collection of responses
	if result.Data != nil {
		result.Data = new(responses.GetUsersCollection).Collection(result.Data.([]models.User))
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) GetUser(ctx context.Context, id int) (*responses.GetUserByIDResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetUserService", trace.WithAttributes(attribute.String("service", "GetUser")))

	user, err := s.userRepository().GetUserByID(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return new(responses.GetUserByIDResponse).Make(user), err
}

func (s Service) CreateUser(ctx context.Context, userDto *dtos.CreateUser) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreateUserService", trace.WithAttributes(attribute.String("service", "CreateUser")))

	// Encrypt password
	encryptedPassword := encryption.EncryptPassword(userDto.Password, "")

	// Generate OTP
	otp := s.mailService.GenerateOTP()

	// Calculate OTP expiration (10 minutes from now)
	otpExpiredAt := sql.NullTime{
		Time:  time.Now().Add(10 * time.Minute),
		Valid: true,
	}

	user := new(models.User)
	user.Email = userDto.Email
	user.Password = encryptedPassword
	user.OTP = otp
	user.OTPExpiredAt = &otpExpiredAt
	user.IsVerified = false

	// Create user in database
	err := s.userRepository().CreateUser(ctx, user)
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

func (s Service) UpdateUser(ctx context.Context, id int, userDto *dtos.UpdateUser) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdateUserService", trace.WithAttributes(attribute.String("service", "UpdateUser")))

	user := new(models.User)
	user.FirstName = userDto.FirstName
	user.LastName = userDto.LastName

	// Handle avatar file upload if provided
	if userDto.AvatarFile != nil {
		// Generate unique filename
		fileExt := ""
		if len(userDto.AvatarFile.Filename) > 0 {
			// Get file extension
			for i := len(userDto.AvatarFile.Filename) - 1; i >= 0; i-- {
				if userDto.AvatarFile.Filename[i] == '.' {
					fileExt = userDto.AvatarFile.Filename[i:]
					break
				}
			}
		}

		// Generate filename: user_avatar_{userId}_{timestamp}{extension}
		fileName := fmt.Sprintf("user_avatar_%d_%d%s", id, time.Now().Unix(), fileExt)

		// Define the path in storage
		storagePath := "users/avatars"

		// Upload to MinIO/storage
		err := s.fileSystem.Put(ctx, storagePath, fileName, userDto.AvatarFile)
		if err != nil {
			s.logger.Error("Failed to upload avatar to storage", zap.Error(err), zap.Int("userId", id))
			s.tracer.TraceEnd(childSpan)
			return err
		}

		// Get the public URL for the uploaded file
		avatarURL := s.fileSystem.GetPublicURL(storagePath, fileName)

		// Set avatar URL in user model
		user.Avatar = avatarURL
	}

	err := s.userRepository().UpdateUser(ctx, id, user)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) DeleteUser(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "DeleteUserService", trace.WithAttributes(attribute.String("service", "DeleteUser")))

	err := s.userRepository().DeleteUser(ctx, id)

	s.tracer.TraceEnd(childSpan)

	return err
}

func (s Service) VerifyOTP(ctx context.Context, verifyDto *dtos.VerifyOTP) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "VerifyOTPService", trace.WithAttributes(attribute.String("service", "VerifyOTP")))

	// Get user by email
	user, err := s.userRepository().GetUserByEmail(ctx, verifyDto.Email)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	// Check if user is already verified
	if user.IsVerified {
		s.tracer.TraceEnd(childSpan)
		return exception.ErrInvalidRequestParameter
	}

	// Check if OTP matches
	if user.OTP != verifyDto.OTP {
		s.tracer.TraceEnd(childSpan)
		return exception.ErrInvalidRequestParameter
	}

	// Check if OTP is expired
	if user.OTPExpiredAt == nil || !user.OTPExpiredAt.Valid || time.Now().After(user.OTPExpiredAt.Time) {
		s.tracer.TraceEnd(childSpan)
		return exception.ErrInvalidRequestParameter
	}

	// Verify user
	err = s.userRepository().VerifyUserOTP(ctx, verifyDto.Email)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) ChangePassword(ctx context.Context, id int, changePasswordDto *dtos.PatchChangePassword) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "ChangePasswordService", trace.WithAttributes(attribute.String("service", "ChangePassword")))

	// Get user by ID
	user, err := s.userRepository().GetUserByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	// Verify old password
	if !encryption.VerifyPassword(changePasswordDto.OldPassword, user.Password) {
		s.tracer.TraceEnd(childSpan)
		return exception.ErrPasswordNotMatch
	}

	// Check if new password is same as old password
	if encryption.VerifyPassword(changePasswordDto.NewPassword, user.Password) {
		s.tracer.TraceEnd(childSpan)
		return exception.ErrNewPasswordSameAsOld
	}

	// Encrypt new password
	encryptedPassword := encryption.EncryptPassword(changePasswordDto.NewPassword, "")

	// Update password
	err = s.userRepository().ChangeUserPassword(ctx, id, encryptedPassword)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) AddRoleToUser(ctx context.Context, id int, dto *dtos.AddRoleToUser) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AddRoleToUserService", trace.WithAttributes(attribute.String("service", "AddRoleToUser")))

	err := s.userRepository().AddRoleToUser(ctx, id, dto.RoleName)

	s.tracer.TraceEnd(childSpan)
	return err
}

func (s Service) UpgradeToHost(ctx context.Context, id int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpgradeToHostService", trace.WithAttributes(attribute.String("service", "UpgradeToHost")))

	err := s.userRepository().UpgradeUserToHost(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}

	s.tracer.TraceEnd(childSpan)

	return nil
}

func (s Service) AdminDeactivateAllProperties(ctx context.Context, userID int, adminID int, adminNote string) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AdminDeactivateAllPropertiesService", trace.WithAttributes(attribute.String("service", "AdminDeactivateAllProperties")))
	defer s.tracer.TraceEnd(childSpan)

	// Collect affected guest IDs BEFORE cancelling bookings (for notifications)
	propertyIDs, _ := s.userRepository().GetPropertyIDsByUserID(ctx, userID)
	var affectedGuestIDs []int
	if len(propertyIDs) > 0 {
		affectedGuestIDs, _ = s.userRepository().GetActiveBookingGuestIDsByPropertyIDs(ctx, propertyIDs)
	}

	if err := s.userRepository().AdminDeactivateAllProperties(ctx, userID); err != nil {
		return err
	}

	// Notify guests whose bookings were cancelled
	for _, guestID := range affectedGuestIDs {
		gID := guestID
		go notification_module.NewService().SendToUser(ctx, gID, "guest", "cancel", "Your booking has been cancelled because the property was deactivated by admin", "", "bookings", nil)
	}

	// Notify host
	go notification_module.NewService().SendToUser(ctx, userID, "host", "report", "All your properties have been deactivated by admin", adminNote, "reports", nil)
	go notification_module.NewService().SendToUser(ctx, userID, "host", "report", "Your host verification has been revoked", "", "reports", nil)

	go func() {
		logEntry := &models.ActionLog{
			ModelName: "User",
			RecordID:  fmt.Sprintf("%d", userID),
			Action:    "deactivate_all_properties",
			Messages:  adminNote,
			ActorID:   &adminID,
			TargetID:  &userID,
		}
		_ = action_log_module.NewRepository().CreateActionLog(ctx, logEntry)
	}()

	return nil
}

func (s Service) AdminUnbanAllProperties(ctx context.Context, userID int, adminID int, adminNote string) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "AdminUnbanAllPropertiesService", trace.WithAttributes(attribute.String("service", "AdminUnbanAllProperties")))
	defer s.tracer.TraceEnd(childSpan)

	if err := s.userRepository().AdminUnbanAllProperties(ctx, userID); err != nil {
		return err
	}

	// Notify host
	go notification_module.NewService().SendToUser(ctx, userID, "host", "report", "Your account has been reinstated. Re-publish your properties from the host dashboard.", adminNote, "reports", nil)

	go func() {
		logEntry := &models.ActionLog{
			ModelName: "User",
			RecordID:  fmt.Sprintf("%d", userID),
			Action:    "unban_all_properties",
			Messages:  adminNote,
			ActorID:   &adminID,
			TargetID:  &userID,
		}
		_ = action_log_module.NewRepository().CreateActionLog(ctx, logEntry)
	}()

	return nil
}
