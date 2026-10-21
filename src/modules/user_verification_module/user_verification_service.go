package user_verification_module

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"joodpakpaw/internal/exception"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/action_log_module"
	"joodpakpaw/src/modules/notification_module"
	"joodpakpaw/src/modules/user_verification_module/dtos"
	"joodpakpaw/src/modules/user_verification_module/responses"
)

func (s Service) GetVerifications(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetVerificationsService", trace.WithAttributes(attribute.String("service", "GetVerifications")))

	result, err := s.verificationRepository().GetVerificationsPaginate(ctx, pagination)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	if result.Data != nil {
		result.Data = new(responses.GetVerificationsCollection).Collection(result.Data.([]models.UserVerification))
	}

	s.tracer.TraceEnd(childSpan)
	return result, nil
}

func (s Service) GetVerificationByID(ctx context.Context, id int) (*responses.GetVerificationByIDResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetVerificationByIDService", trace.WithAttributes(attribute.String("service", "GetVerificationByID")))

	verification, err := s.verificationRepository().GetVerificationByID(ctx, id)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}
	return new(responses.GetVerificationByIDResponse).Make(verification), nil
}

func (s Service) GetMyVerification(ctx context.Context, userID int) (*responses.GetVerificationByIDResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetMyVerificationService", trace.WithAttributes(attribute.String("service", "GetMyVerification")))

	verification, err := s.verificationRepository().GetLatestVerificationByUserID(ctx, userID)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}
	return new(responses.GetVerificationByIDResponse).Make(verification), nil
}

func (s Service) SubmitVerification(ctx context.Context, userID int, dto *dtos.SubmitVerification) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "SubmitVerificationService", trace.WithAttributes(attribute.String("service", "SubmitVerification")))

	repo := s.verificationRepository()
	var err error

	// Check the latest verification status to decide if submit is allowed
	isResubmit := false
	latest, latestErr := repo.GetLatestVerificationByUserID(ctx, userID)
	if latestErr != nil && latestErr != exception.ErrRecordNotFound {
		s.tracer.TraceEnd(childSpan)
		return latestErr
	}
	if latestErr == nil {
		switch latest.Status {
		case models.VerificationStatusPending:
			// Already waiting for review — cannot submit again
			s.tracer.TraceEnd(childSpan)
			return exception.ErrVerificationAlreadyPending
		case models.VerificationStatusApproved:
			// Already verified — no need to resubmit
			s.tracer.TraceEnd(childSpan)
			return exception.ErrVerificationAlreadyApproved
		case models.VerificationStatusRejected:
			// Was rejected — allow resubmission
			isResubmit = true
		case models.VerificationStatusRevoked:
			// Verification was revoked due to account ban — allow resubmission
			isResubmit = true
		}
	}

	// Upload idPhoto
	idPhotoExt := ""
	if len(dto.IdPhotoFile.Filename) > 0 {
		for i := len(dto.IdPhotoFile.Filename) - 1; i >= 0; i-- {
			if dto.IdPhotoFile.Filename[i] == '.' {
				idPhotoExt = dto.IdPhotoFile.Filename[i:]
				break
			}
		}
	}
	idPhotoName := fmt.Sprintf("verification_id_%d_%d%s", userID, time.Now().Unix(), idPhotoExt)
	idPhotoPath := "users/verifications/id-photos"
	if err = s.fileSystem.Put(ctx, idPhotoPath, idPhotoName, dto.IdPhotoFile); err != nil {
		s.logger.Error("Failed to upload id photo to storage", zap.Error(err), zap.Int("userId", userID))
		s.tracer.TraceEnd(childSpan)
		return err
	}
	idPhotoURL := s.fileSystem.GetPublicURL(idPhotoPath, idPhotoName)

	// Upload selfieWithId
	selfieExt := ""
	if len(dto.SelfieWithIdFile.Filename) > 0 {
		for i := len(dto.SelfieWithIdFile.Filename) - 1; i >= 0; i-- {
			if dto.SelfieWithIdFile.Filename[i] == '.' {
				selfieExt = dto.SelfieWithIdFile.Filename[i:]
				break
			}
		}
	}
	selfieName := fmt.Sprintf("verification_selfie_%d_%d%s", userID, time.Now().Unix(), selfieExt)
	selfiePath := "users/verifications/selfies"
	if err = s.fileSystem.Put(ctx, selfiePath, selfieName, dto.SelfieWithIdFile); err != nil {
		s.logger.Error("Failed to upload selfie to storage", zap.Error(err), zap.Int("userId", userID))
		s.tracer.TraceEnd(childSpan)
		return err
	}
	selfieURL := s.fileSystem.GetPublicURL(selfiePath, selfieName)

	// Resubmit: update existing record — keep same ID, reset status & clear rejection
	var verificationID int
	if isResubmit {
		err = repo.ResubmitVerification(ctx, latest.ID, idPhotoURL, selfieURL)
		verificationID = latest.ID
	} else {
		verification := &models.UserVerification{
			UserID:       userID,
			Status:       models.VerificationStatusPending,
			IsResubmit:   false,
			SubmittedAt:  time.Now(),
			IdPhoto:      idPhotoURL,
			SelfieWithId: selfieURL,
		}
		err = repo.CreateVerification(ctx, verification)
		verificationID = verification.ID
	}

	if err == nil {
		title := "Your host verification request has been submitted and is under review"
		go notification_module.NewService().SendToUser(ctx, userID, "host", "verification", title, "", "user_verification", &verificationID)
	}

	s.tracer.TraceEnd(childSpan)
	return err
}

func (s Service) ApproveVerification(ctx context.Context, id int, adminID int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "ApproveVerificationService", trace.WithAttributes(attribute.String("service", "ApproveVerification")))

	// Validate status is pending
	verification, err := s.verificationRepository().GetVerificationByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}
	if verification.Status != models.VerificationStatusPending {
		s.tracer.TraceEnd(childSpan)
		return exception.ErrVerificationNotPending
	}

	err = s.verificationRepository().ApproveVerification(ctx, id)

	if err == nil {
		go notification_module.NewService().SendToUser(ctx, verification.UserID, "host", "verification", "Your host verification has been approved", "", "user_verification", &id)
		go func() {
			targetID := verification.UserID
			_ = action_log_module.NewRepository().CreateActionLog(ctx, &models.ActionLog{
				ModelName: "UserVerification",
				RecordID:  fmt.Sprintf("%d", id),
				Action:    "approve_verification",
				ActorID:   &adminID,
				TargetID:  &targetID,
			})
		}()
	}

	s.tracer.TraceEnd(childSpan)
	return err
}

func (s Service) RejectVerification(ctx context.Context, id int, dto *dtos.RejectVerification, adminID int) error {
	ctx, childSpan := s.tracer.TraceStart(ctx, "RejectVerificationService", trace.WithAttributes(attribute.String("service", "RejectVerification")))

	// Validate status is pending
	verification, err := s.verificationRepository().GetVerificationByID(ctx, id)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return err
	}
	if verification.Status != models.VerificationStatusPending {
		s.tracer.TraceEnd(childSpan)
		return exception.ErrVerificationNotPending
	}

	err = s.verificationRepository().RejectVerification(ctx, id, dto.RejectDescription)

	if err == nil {
		go notification_module.NewService().SendToUser(ctx, verification.UserID, "host", "verification", "Your host verification has been rejected", "", "user_verification", &id)
		go func() {
			targetID := verification.UserID
			_ = action_log_module.NewRepository().CreateActionLog(ctx, &models.ActionLog{
				ModelName: "UserVerification",
				RecordID:  fmt.Sprintf("%d", id),
				Action:    "reject_verification",
				Messages:  dto.RejectDescription,
				ActorID:   &adminID,
				TargetID:  &targetID,
			})
		}()
	}

	s.tracer.TraceEnd(childSpan)
	return err
}

func getFileExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}
