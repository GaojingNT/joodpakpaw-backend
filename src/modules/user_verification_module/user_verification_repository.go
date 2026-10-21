package user_verification_module

import (
	"context"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
)

func (r Repository) GetVerificationsPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan  = r.tracer.TraceStart(ctx, "GetVerificationsPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetVerificationsPaginate")))
		verifications = make([]models.UserVerification, 0)
		err           error
	)

	statusFilter, _ := pagination.GetStringAttribute("status")

	tx := r.db
	if statusFilter != "" {
		tx = tx.Where("status = ?", statusFilter)
	}

	utils.Block{
		Try: func() {
			if err = tx.
				Preload("User").
				Scopes(paginator.Paginate(verifications, pagination, tx)).
				Find(&verifications).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
	}.Do()

	pagination.Data = verifications
	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}
	return pagination, nil
}

func (r Repository) GetVerificationByID(ctx context.Context, id int) (models.UserVerification, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetVerificationByIDRepository", trace.WithAttributes(attribute.String("repository", "GetVerificationByID"), attribute.Int64("id", int64(id))))
		verification models.UserVerification
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Preload("User").First(&verification, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return verification, err
	}
	return verification, nil
}

func (r Repository) GetLatestVerificationByUserID(ctx context.Context, userID int) (models.UserVerification, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetLatestVerificationByUserIDRepository", trace.WithAttributes(attribute.String("repository", "GetLatestVerificationByUserID"), attribute.Int64("userID", int64(userID))))
		verification models.UserVerification
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Preload("User").
				Where("user_id = ?", userID).
				Order("submitted_at DESC").
				First(&verification).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return verification, err
	}
	return verification, nil
}

func (r Repository) HasPendingVerification(ctx context.Context, userID int) (bool, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "HasPendingVerificationRepository", trace.WithAttributes(attribute.String("repository", "HasPendingVerification"), attribute.Int64("userID", int64(userID))))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.UserVerification{}).
				Where("user_id = ? AND status = ?", userID, models.VerificationStatusPending).
				Count(&count).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r Repository) HasRejectedVerification(ctx context.Context, userID int) (bool, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "HasRejectedVerificationRepository", trace.WithAttributes(attribute.String("repository", "HasRejectedVerification"), attribute.Int64("userID", int64(userID))))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.UserVerification{}).
				Where("user_id = ? AND status = ?", userID, models.VerificationStatusRejected).
				Count(&count).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r Repository) CreateVerification(ctx context.Context, verification *models.UserVerification) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateVerificationRepository", trace.WithAttributes(attribute.String("repository", "CreateVerification")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				return tx.Create(verification).Error
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return err
	}
	return nil
}

func (r Repository) ResubmitVerification(ctx context.Context, id int, idPhoto string, selfieWithId string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ResubmitVerificationRepository", trace.WithAttributes(attribute.String("repository", "ResubmitVerification"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Fetch current record to save existing rejectDescription before clearing
				var current models.UserVerification
				if err = tx.First(&current, id).Error; err != nil {
					return err
				}
				return tx.Model(&models.UserVerification{}).
					Where("id = ?", id).
					Updates(map[string]any{
						"status":                      models.VerificationStatusPending,
						"previous_reject_description": current.RejectDescription,
						"reject_description":          nil,
						"id_photo":                    idPhoto,
						"selfie_with_id":              selfieWithId,
						"is_resubmit":                 true,
						"submitted_at":                gorm.Expr("NOW()"),
					}).Error
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return err
	}
	return nil
}
func (r Repository) RevokeVerificationByUserID(ctx context.Context, userID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "RevokeVerificationByUserIDRepository", trace.WithAttributes(attribute.String("repository", "RevokeVerificationByUserID"), attribute.Int64("userID", int64(userID))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Model(&models.UserVerification{}).
				Where("user_id = ? AND status = ?", userID, models.VerificationStatusApproved).
				Update("status", models.VerificationStatusRevoked).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			r.logger.Error(err.Error())
			sentry.CaptureException(err)
			exception.SqlErrorMessage = err.Error()
			err = exception.ErrDbQueryStatement
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)
	return err
}

func (r Repository) ApproveVerification(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ApproveVerificationRepository", trace.WithAttributes(attribute.String("repository", "ApproveVerification"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				var v models.UserVerification
				if err = tx.First(&v, id).Error; err != nil {
					return err
				}
				// Update verification status to approved
				if err = tx.Model(&models.UserVerification{}).
					Where("id = ?", id).
					Updates(map[string]interface{}{
						"status":             models.VerificationStatusApproved,
						"reject_description": nil,
					}).Error; err != nil {
					return err
				}

				// Update user IsVerifiedHost to true
				return tx.Model(&models.User{}).
					Where("id = ?", v.UserID).
					Update("is_verified_host", true).Error
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return err
	}
	return nil
}

func (r Repository) RejectVerification(ctx context.Context, id int, rejectDescription string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "RejectVerificationRepository", trace.WithAttributes(attribute.String("repository", "RejectVerification"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				var v models.UserVerification
				if err = tx.First(&v, id).Error; err != nil {
					return err
				}
				return tx.Model(&models.UserVerification{}).
					Where("id = ?", id).
					Updates(map[string]interface{}{
						"status":             models.VerificationStatusRejected,
						"reject_description": rejectDescription,
					}).Error
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return err
	}
	return nil
}
