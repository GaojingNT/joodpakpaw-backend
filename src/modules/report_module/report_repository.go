package report_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
)

func (r Repository) GetReportsPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetReportsPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetReportsPaginate")))
		reports      = make([]models.Report, 0)
		err          error
	)

	reportType, _ := pagination.GetStringAttribute("type")

	tx := r.db.Model(&models.Report{})
	if reportType != "" {
		tx = tx.Joins("JOIN tbl_report_categories rc ON rc.id = tbl_reports.report_category_id").
			Where("rc.type = ?", reportType)
	}

	utils.Block{
		Try: func() {
			if err = tx.
				Preload("ReportCategory").
				Preload("Reporter").
				Preload("Property").
				Preload("Booking.Guest").
				Scopes(paginator.Paginate(reports, pagination, tx)).
				Find(&reports).Error; err != nil {
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
		Finally: nil,
	}.Do()

	pagination.Data = reports

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) CreateReport(ctx context.Context, report *models.Report) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateReportRepository", trace.WithAttributes(attribute.String("repository", "CreateReport")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				if err = tx.Create(report).Error; err != nil {
					return err
				}
				// Increment property report_count when reporting a property
				if report.PropertyID != nil {
					if err = tx.Model(&models.Property{}).
						Where("id = ?", *report.PropertyID).
						UpdateColumn("report_count", gorm.Expr("report_count + ?", 1)).Error; err != nil {
						return err
					}
				}
				return nil
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
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return err
	}

	return nil
}

func (r Repository) HasPendingUnbanRequest(ctx context.Context, reporterID int, reportType models.ReportType, propertyID *int) (bool, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "HasPendingUnbanRequestRepository", trace.WithAttributes(attribute.String("repository", "HasPendingUnbanRequest")))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			tx := r.db.Model(&models.Report{}).
				Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
				Where("tbl_reports.reporter_id = ?", reporterID).
				Where("tbl_report_categories.type = ?", reportType).
				Where("tbl_reports.status = ?", models.ReportStatusPending).
				Where("tbl_reports.deleted_at IS NULL")

			if propertyID != nil {
				tx = tx.Where("tbl_reports.property_id = ?", *propertyID)
			} else {
				tx = tx.Where("tbl_reports.property_id IS NULL")
			}

			if err = tx.Count(&count).Error; err != nil {
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
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r Repository) CountUnbanPropertyRequests(ctx context.Context, reporterID int, propertyID *int) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountUnbanPropertyRequestsRepository", trace.WithAttributes(attribute.String("repository", "CountUnbanPropertyRequests")))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			tx := r.db.Model(&models.Report{}).
				Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
				Where("tbl_reports.reporter_id = ?", reporterID).
				Where("tbl_report_categories.type = ?", models.ReportTypeUnbanProperty).
				Where("tbl_reports.deleted_at IS NULL")

			if propertyID != nil {
				tx = tx.Where("tbl_reports.property_id = ?", *propertyID)
			} else {
				tx = tx.Where("tbl_reports.property_id IS NULL")
			}

			if err = tx.Count(&count).Error; err != nil {
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
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r Repository) GetBookingByID(ctx context.Context, id int) (models.Booking, error) {
	var booking models.Booking
	err := r.db.WithContext(ctx).First(&booking, id).Error
	if err == gorm.ErrRecordNotFound {
		return booking, exception.ErrRecordNotFound
	}
	return booking, err
}

func (r Repository) GetPropertyByID(ctx context.Context, id int) (models.Property, error) {
	var property models.Property
	err := r.db.WithContext(ctx).First(&property, id).Error
	if err == gorm.ErrRecordNotFound {
		return property, exception.ErrRecordNotFound
	}
	return property, err
}

func (r Repository) GetReportCategoryByID(ctx context.Context, id int) (models.ReportCategory, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetReportCategoryByIDRepository", trace.WithAttributes(attribute.String("repository", "GetReportCategoryByID"), attribute.Int64("id", int64(id))))
		category     models.ReportCategory
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.First(&category, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return category, err
	}

	return category, nil
}

func (r Repository) CreateReportImages(ctx context.Context, images []models.ReportImage) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateReportImagesRepository", trace.WithAttributes(attribute.String("repository", "CreateReportImages")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Create(&images).Error; err != nil {
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
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	return err
}

// resetUnbanRequestsByPropertyID soft-deletes all unban_property requests for a specific property
func resetUnbanRequestsByPropertyID(tx *gorm.DB, propertyID int) error {
	var ids []int
	if err := tx.Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_reports.property_id = ?", propertyID).
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.deleted_at IS NULL").
		Pluck("tbl_reports.id", &ids).Error; err != nil {
		return err
	}
	if len(ids) > 0 {
		return tx.Where("id IN ?", ids).Delete(&models.Report{}).Error
	}
	return nil
}

// resetUnbanRequestsByHostID soft-deletes all unban_property requests for a host (both single and all-property)
func resetUnbanRequestsByHostID(tx *gorm.DB, hostID int) error {
	var ids []int
	if err := tx.Table("tbl_reports").
		Joins("JOIN tbl_report_categories ON tbl_report_categories.id = tbl_reports.report_category_id").
		Where("tbl_reports.reporter_id = ?", hostID).
		Where("tbl_report_categories.type = ?", "unban_property").
		Where("tbl_reports.deleted_at IS NULL").
		Pluck("tbl_reports.id", &ids).Error; err != nil {
		return err
	}
	if len(ids) > 0 {
		return tx.Where("id IN ?", ids).Delete(&models.Report{}).Error
	}
	return nil
}

// cancelBookingsByPropertyIDs cancels all pending/confirmed bookings for the given property IDs
// and returns rooms to inventory inside the provided transaction.
func cancelBookingsByPropertyIDs(tx *gorm.DB, propertyIDs []int) error {
	today := time.Now().Truncate(24 * time.Hour)

	// build a set for quick lookup
	propertyIDSet := make(map[int]struct{}, len(propertyIDs))
	for _, pid := range propertyIDs {
		propertyIDSet[pid] = struct{}{}
	}

	var bookings []models.Booking
	if err := tx.
		Preload("BookingOrders").
		Joins("JOIN tbl_booking_orders bo ON bo.booking_id = tbl_bookings.id").
		Where("bo.property_id IN ?", propertyIDs).
		Where("tbl_bookings.booking_status IN ?", []string{
			string(models.BookingStatusPending),
			string(models.BookingStatusConfirmed),
		}).
		Group("tbl_bookings.id").
		Find(&bookings).Error; err != nil {
		return err
	}

	for _, booking := range bookings {
		// confirmed booking: skip if any affected order has check-in today
		if booking.BookingStatus == models.BookingStatusConfirmed {
			skip := false
			for _, order := range booking.BookingOrders {
				if order.PropertyID == nil {
					continue
				}
				if _, affected := propertyIDSet[*order.PropertyID]; !affected {
					continue
				}
				if order.CheckInDate.Truncate(24 * time.Hour).Equal(today) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		// Return rooms to inventory
		for _, order := range booking.BookingOrders {
			for d := order.CheckInDate; d.Before(order.CheckOutDate); d = d.AddDate(0, 0, 1) {
				var calendar models.RoomCategoryCalendar
				if err := tx.Where("room_category_id = ? AND date = ?", order.RoomCategoryID, d).First(&calendar).Error; err == nil {
					calendar.AvailableRooms += order.RoomsBooked
					calendar.BookedRooms -= order.RoomsBooked
					if err := tx.Save(&calendar).Error; err != nil {
						return err
					}
				}
			}
		}
		// Cancel booking
		if err := tx.Model(&booking).Update("booking_status", models.BookingStatusCancelled).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetActiveBookingGuestIDsByPropertyIDs returns guest IDs of pending/confirmed bookings for the given property IDs.
func (r Repository) GetActiveBookingGuestIDsByPropertyIDs(ctx context.Context, propertyIDs []int) ([]int, error) {
	var guestIDs []int
	err := r.db.WithContext(ctx).
		Model(&models.Booking{}).
		Select("DISTINCT tbl_bookings.guest_id").
		Joins("JOIN tbl_booking_orders bo ON bo.booking_id = tbl_bookings.id").
		Where("bo.property_id IN ?", propertyIDs).
		Where("tbl_bookings.booking_status IN ?", []string{
			string(models.BookingStatusPending),
			string(models.BookingStatusConfirmed),
		}).
		Pluck("guest_id", &guestIDs).Error
	return guestIDs, err
}

// GetPropertyIDsByHostID returns all property IDs owned by a host.
func (r Repository) GetPropertyIDsByHostID(ctx context.Context, hostID int) ([]int, error) {
	var propertyIDs []int
	err := r.db.WithContext(ctx).
		Model(&models.Property{}).
		Where("host_id = ?", hostID).
		Pluck("id", &propertyIDs).Error
	return propertyIDs, err
}

func (r Repository) ActionReport(ctx context.Context, report models.Report, action string, adminNote string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ActionReportRepository", trace.WithAttributes(attribute.String("repository", "ActionReport"), attribute.Int64("id", int64(report.ID))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Update report status and resolved action
				resolvedAction := models.ReportAction(action)
				status := models.ReportStatusResolved
				if resolvedAction == models.ReportActionDismissReport {
					status = models.ReportStatusDismissed
				}
				if err = tx.Model(&models.Report{}).Where("id = ?", report.ID).Updates(map[string]any{
					"status":          status,
					"resolved_action": resolvedAction,
					"admin_note":      adminNote,
				}).Error; err != nil {
					return err
				}

				switch resolvedAction {
				case models.ReportActionDeactivateProperty:
					if report.PropertyID == nil {
						return exception.ErrReportActionNotAllowed
					}
					if err = tx.Model(&models.Property{}).Where("id = ?", *report.PropertyID).Updates(map[string]any{
						"is_active":          false,
						"is_banned_by_admin": true,
					}).Error; err != nil {
						return err
					}
					if err = cancelBookingsByPropertyIDs(tx, []int{*report.PropertyID}); err != nil {
						return err
					}
					return resetUnbanRequestsByPropertyID(tx, *report.PropertyID)

				case models.ReportActionDeleteReview:
					if report.BookingID == nil {
						return exception.ErrReportActionNotAllowed
					}
					return tx.Model(&models.Booking{}).Where("id = ?", *report.BookingID).Updates(map[string]any{
						"rating":      nil,
						"comment":     "",
						"reviewed_at": nil,
					}).Error

				case models.ReportActionBanUserReview:
					if report.Booking == nil || report.Booking.GuestID == nil {
						return exception.ErrReportActionNotAllowed
					}
					return tx.Model(&models.User{}).Where("id = ?", *report.Booking.GuestID).Update("is_banned_review", true).Error

				case models.ReportActionDeactivateProperties:
					if report.Property == nil || report.Property.HostID == nil {
						return exception.ErrReportActionNotAllowed
					}
					// Mark properties that are already banned before this account ban
					if err = tx.Model(&models.Property{}).
						Where("host_id = ? AND is_banned_by_admin = ?", *report.Property.HostID, true).
						Update("is_banned_before_account_ban", true).Error; err != nil {
						return err
					}
					if err = tx.Model(&models.Property{}).Where("host_id = ?", *report.Property.HostID).Updates(map[string]any{
						"is_active":          false,
						"is_banned_by_admin": true,
					}).Error; err != nil {
						return err
					}
					if err = tx.Model(&models.User{}).Where("id = ?", *report.Property.HostID).Updates(map[string]any{"is_banned_all_properties": true, "is_verified_host": false, "banned_at": time.Now()}).Error; err != nil {
						return err
					}
					if err = tx.Model(&models.UserVerification{}).Where("user_id = ? AND status = ?", *report.Property.HostID, models.VerificationStatusApproved).Update("status", models.VerificationStatusRevoked).Error; err != nil {
						return err
					}
					var propertyIDs []int
					if err = tx.Model(&models.Property{}).Where("host_id = ?", *report.Property.HostID).Pluck("id", &propertyIDs).Error; err != nil {
						return err
					}
					if err = cancelBookingsByPropertyIDs(tx, propertyIDs); err != nil {
						return err
					}
					return resetUnbanRequestsByHostID(tx, *report.Property.HostID)

				case models.ReportActionUnbanProperty:
					if report.PropertyID == nil {
						return exception.ErrReportActionNotAllowed
					}
					return tx.Model(&models.Property{}).Where("id = ?", *report.PropertyID).Updates(map[string]any{
						"is_active":          false,
						"is_banned_by_admin": false,
					}).Error

				case models.ReportActionUnbanAllProperties:
					// Resolve hostID from property or reporter (unban-all reports have no property_id)
					var unbanHostID int
					if report.Property != nil && report.Property.HostID != nil {
						unbanHostID = *report.Property.HostID
					} else if report.ReporterID != nil {
						unbanHostID = *report.ReporterID
					} else {
						return exception.ErrReportActionNotAllowed
					}
					// Only unban properties that were banned due to account ban, not pre-existing bans
					if err = tx.Model(&models.Property{}).
						Where("host_id = ? AND is_banned_before_account_ban = ?", unbanHostID, false).
						Updates(map[string]any{
							"is_active":          false,
							"is_banned_by_admin": false,
						}).Error; err != nil {
						return err
					}
					// Reset the flag for all properties now that account ban is lifted
					if err = tx.Model(&models.Property{}).
						Where("host_id = ?", unbanHostID).
						Update("is_banned_before_account_ban", false).Error; err != nil {
						return err
					}
					return tx.Model(&models.User{}).Where("id = ?", unbanHostID).Update("is_banned_all_properties", false).Error

				case models.ReportActionUnbanReview:
					if report.ReporterID == nil {
						return exception.ErrReportActionNotAllowed
					}
					return tx.Model(&models.User{}).Where("id = ?", *report.ReporterID).Update("is_banned_review", false).Error
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == exception.ErrReportActionNotAllowed {
				// pass through
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
	return err
}

func (r Repository) GetMyReportsPaginate(ctx context.Context, userID int, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetMyReportsPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetMyReportsPaginate"), attribute.Int64("userID", int64(userID))))
		reports      = make([]models.Report, 0)
		err          error
	)

	reportType, _ := pagination.GetStringAttribute("type")

	tx := r.db.Model(&models.Report{}).Where("tbl_reports.reporter_id = ?", userID)
	if reportType != "" {
		tx = tx.Joins("JOIN tbl_report_categories rc ON rc.id = tbl_reports.report_category_id").
			Where("rc.type = ?", reportType)
	}

	utils.Block{
		Try: func() {
			if err = tx.
				Preload("ReportCategory").
				Preload("Reporter").
				Preload("Property").
				Preload("Booking.Guest").
				Scopes(paginator.Paginate(reports, pagination, tx)).
				Find(&reports).Error; err != nil {
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

	pagination.Data = reports

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}
	return pagination, nil
}

func (r Repository) GetMyReportByID(ctx context.Context, id int, userID int) (models.Report, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetMyReportByIDRepository", trace.WithAttributes(attribute.String("repository", "GetMyReportByID"), attribute.Int64("id", int64(id)), attribute.Int64("userID", int64(userID))))
		report       models.Report
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Preload("ReportCategory").
				Preload("Reporter").
				Preload("Property").
				Preload("Property.Host").
				Preload("Booking.Guest").
				Preload("Booking.BookingOrders.Property").
				Preload("Booking.BookingOrders.Property.Host").
				Preload("Images").
				Where("tbl_reports.reporter_id = ?", userID).
				First(&report, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return report, err
	}
	return report, nil
}

func (r Repository) GetUserByID(ctx context.Context, id int) (models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return user, exception.ErrRecordNotFound
	}
	return user, err
}

func (r Repository) GetReportByID(ctx context.Context, id int) (models.Report, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetReportByIDRepository", trace.WithAttributes(attribute.String("repository", "GetReportByID"), attribute.Int64("id", int64(id))))
		report       models.Report
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Preload("ReportCategory").
				Preload("Reporter").
				Preload("Property").
				Preload("Property.Host").
				Preload("Booking.Guest").
				Preload("Booking.BookingOrders.Property").
				Preload("Booking.BookingOrders.Property.Host").
				Preload("Images").
				First(&report, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return report, err
	}

	return report, nil
}
