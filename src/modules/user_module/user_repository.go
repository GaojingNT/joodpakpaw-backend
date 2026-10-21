package user_module

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

func (r Repository) GetUserPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserPaginateRepository", trace.WithAttributes(attribute.String("repository", "GetUserPaginate")))
		users        = make([]models.User, 0)
		err          error
	)

	// Get attributes
	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")

	// Set tracing attributes
	r.tracer.SetAttributes(childSpan, attribute.String("search", searchAttribute))
	r.tracer.SetAttributes(childSpan, attribute.String("search_by", searchByAttribute))

	// Create pagination query
	tx := r.db

	utils.Block{
		Try: func() {
			// Execute query
			if err = tx.
				Scopes(models.SearchingScope(models.UserSearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(users, pagination, tx)).
				Find(&users).Error; err != nil {
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

	// Set data
	pagination.Data = users

	r.tracer.TraceEnd(childSpan)

	// Check error
	if err != nil {
		return nil, err
	}

	return pagination, nil
}

func (r Repository) AdminGetUsersPaginate(ctx context.Context, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AdminGetUsersPaginateRepository", trace.WithAttributes(attribute.String("repository", "AdminGetUsersPaginate")))
		users        = make([]models.User, 0)
		err          error
	)

	searchAttribute, _ := pagination.GetStringAttribute("search")
	searchByAttribute, _ := pagination.GetStringAttribute("search_by")

	tx := r.db

	utils.Block{
		Try: func() {
			if err = tx.
				Preload("Roles").
				Scopes(models.SearchingScope(models.UserSearchable(), searchAttribute, searchByAttribute)).
				Scopes(paginator.Paginate(users, pagination, tx)).
				Find(&users).Error; err != nil {
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

	pagination.Data = users
	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}
	return pagination, nil
}

func (r Repository) AdminGetUserByID(ctx context.Context, id int) (models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AdminGetUserByIDRepository", trace.WithAttributes(attribute.String("repository", "AdminGetUserByID"), attribute.Int64("id", int64(id))))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Preload("Roles").
				First(&user, id).Error; err != nil {
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
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return user, err
	}
	return user, nil
}

func (r Repository) CountUserViolations(ctx context.Context, userID int) (int64, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CountUserViolationsRepository", trace.WithAttributes(attribute.String("repository", "CountUserViolations"), attribute.Int64("userID", int64(userID))))
		count        int64
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.
				Table("action_logs").
				Where("target_id = ? AND action IN ? AND deleted_at IS NULL", userID, []string{"deactivate_property", "deactivate_all_properties"}).
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
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)

	if err != nil {
		return 0, err
	}
	return count, nil
}

func cancelUserBookingsByPropertyIDs(tx *gorm.DB, propertyIDs []int) error {
	today := time.Now().Truncate(24 * time.Hour)

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
		if err := tx.Model(&booking).Update("booking_status", models.BookingStatusCancelled).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r Repository) GetPropertyIDsByUserID(ctx context.Context, userID int) ([]int, error) {
	var propertyIDs []int
	err := r.db.WithContext(ctx).
		Model(&models.Property{}).
		Where("host_id = ?", userID).
		Pluck("id", &propertyIDs).Error
	return propertyIDs, err
}

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

func (r Repository) AdminDeactivateAllProperties(ctx context.Context, userID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AdminDeactivateAllPropertiesRepository", trace.WithAttributes(attribute.String("repository", "AdminDeactivateAllProperties"), attribute.Int64("userID", int64(userID))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				var user models.User
				if err = tx.First(&user, userID).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return exception.ErrRecordNotFound
					}
					return err
				}

				if err = tx.Model(&models.Property{}).Where("host_id = ?", userID).Updates(map[string]any{
					"is_active":          false,
					"is_banned_by_admin": true,
				}).Error; err != nil {
					return err
				}

				if err = tx.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{"is_banned_all_properties": true, "is_verified_host": false, "banned_at": time.Now()}).Error; err != nil {
					return err
				}

				if err = tx.Model(&models.UserVerification{}).Where("user_id = ? AND status = ?", userID, models.VerificationStatusApproved).Update("status", models.VerificationStatusRevoked).Error; err != nil {
					return err
				}

				var propertyIDs []int
				if err = tx.Model(&models.Property{}).Where("host_id = ?", userID).Pluck("id", &propertyIDs).Error; err != nil {
					return err
				}
				if len(propertyIDs) > 0 {
					if err = cancelUserBookingsByPropertyIDs(tx, propertyIDs); err != nil {
						return err
					}
				}
				return resetUnbanRequestsByHostID(tx, userID)
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == exception.ErrRecordNotFound {
				// pass through
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)
	return err
}

func (r Repository) AdminUnbanAllProperties(ctx context.Context, userID int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AdminUnbanAllPropertiesRepository", trace.WithAttributes(attribute.String("repository", "AdminUnbanAllProperties"), attribute.Int64("userID", int64(userID))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				var user models.User
				if err = tx.First(&user, userID).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return exception.ErrRecordNotFound
					}
					return err
				}

				if err = tx.Model(&models.Property{}).Where("host_id = ?", userID).Updates(map[string]any{
					"is_banned_by_admin": false,
				}).Error; err != nil {
					return err
				}

				return tx.Model(&models.User{}).Where("id = ?", userID).Update("is_banned_all_properties", false).Error
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == exception.ErrRecordNotFound {
				// pass through
			} else {
				err = e.(error)
				r.logger.Error(err.Error())
				sentry.CaptureException(err)
				exception.SqlErrorMessage = err.Error()
				err = exception.ErrDbQueryStatement
			}
		},
		Finally: nil,
	}.Do()

	r.tracer.TraceEnd(childSpan)
	return err
}

func (r Repository) GetUserByID(ctx context.Context, id int) (models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserByIDRepository", trace.WithAttributes(attribute.String("repository", "GetUserByID"), attribute.Int64("id", int64(id))))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.
				Preload("Roles.Permissions").
				First(&user, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
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
		return user, err
	}

	return user, nil
}

func (r Repository) CreateUser(ctx context.Context, user *models.User) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateUserRepository", trace.WithAttributes(attribute.String("repository", "CreateUser")))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.Create(&user).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
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

func (r Repository) UpdateUser(ctx context.Context, id int, user *models.User) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpdateUserRepository", trace.WithAttributes(attribute.String("repository", "UpdateUser"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// First, check if user exists
				var existingUser models.User
				if err = tx.First(&existingUser, id).Error; err != nil {
					return err
				}

				// Build update map dynamically to only update non-empty fields
				updateMap := make(map[string]interface{})

				if user.FirstName != "" {
					updateMap["first_name"] = user.FirstName
				}
				if user.LastName != "" {
					updateMap["last_name"] = user.LastName
				}
				if user.Avatar != "" {
					updateMap["avatar"] = user.Avatar
				}

				// Execute query only if there are fields to update
				if len(updateMap) > 0 {
					if err = tx.
						Model(&models.User{}).
						Where("id = ?", id).
						Updates(updateMap).Error; err != nil {
						return err
					}
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
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
		return err
	}

	r.tracer.TraceEnd(childSpan)

	return nil
}

func (r Repository) DeleteUser(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "DeleteUserRepository", trace.WithAttributes(attribute.String("repository", "DeleteUser"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// First, check if user exists and is not deleted
				var user models.User
				if err = tx.First(&user, id).Error; err != nil {
					return err
				}

				// Execute soft delete
				if err = tx.Delete(&user).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
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
		return err
	}

	return nil
}

func (r Repository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetUserByEmailRepository", trace.WithAttributes(attribute.String("repository", "GetUserByEmail"), attribute.String("email", email)))
		user         models.User
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.
				Where("email = ?", email).
				First(&user).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
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
		return user, err
	}

	return user, nil
}

func (r Repository) VerifyUserOTP(ctx context.Context, email string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "VerifyUserOTPRepository", trace.WithAttributes(attribute.String("repository", "VerifyUserOTP"), attribute.String("email", email)))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.
					Model(&models.User{}).
					Where("email = ?", email).
					Updates(
						map[string]interface{}{
							"is_verified":    true,
							"otp":            "",
							"otp_expired_at": nil,
						},
					).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
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

func (r Repository) ChangeUserPassword(ctx context.Context, id int, newPassword string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ChangeUserPasswordRepository", trace.WithAttributes(attribute.String("repository", "ChangeUserPassword"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// Execute query
				if err = r.db.
					Model(&models.User{}).
					Where("id = ?", id).
					Update("password", newPassword).Error; err != nil {
					return err
				}

				return nil
			}); err != nil {
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

func (r Repository) AddRoleToUser(ctx context.Context, id int, roleName string) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "AddRoleToUserRepository", trace.WithAttributes(attribute.String("repository", "AddRoleToUser"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				var user models.User
				if err = tx.Preload("Roles").First(&user, id).Error; err != nil {
					return err
				}

				// Check if user already has this role
				for _, role := range user.Roles {
					if role.Name == roleName {
						return exception.ErrRoleAlreadyAssigned
					}
				}

				// Find role by name
				var role models.Role
				if err = tx.Where("name = ?", roleName).First(&role).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return exception.ErrRecordNotFound
					}
					return err
				}

				// Assign role to user
				return tx.Model(&user).Association("Roles").Append(&role)
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == exception.ErrRoleAlreadyAssigned || err == exception.ErrRecordNotFound {
				// pass through
			} else if err == gorm.ErrRecordNotFound {
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

func (r Repository) UpgradeUserToHost(ctx context.Context, id int) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "UpgradeUserToHostRepository", trace.WithAttributes(attribute.String("repository", "UpgradeUserToHost"), attribute.Int64("id", int64(id))))
		err          error
	)

	utils.Block{
		Try: func() {
			if err = r.db.Transaction(func(tx *gorm.DB) error {
				// First, check if user exists
				var user models.User
				if err = tx.First(&user, id).Error; err != nil {
					return err
				}

				// Check if user already has host role
				if err = tx.Preload("Roles").First(&user, id).Error; err != nil {
					return err
				}

				// Check if user already is a host
				for _, role := range user.Roles {
					if role.Name == "host" {
						// User already has host role, just update IsBecomeToHost
						if err = tx.Model(&models.User{}).
							Where("id = ?", id).
							Update("is_become_to_host", true).Error; err != nil {
							return err
						}
						return nil
					}
				}

				// Find host role
				var hostRole models.Role
				if err = tx.Where("name = ?", "host").First(&hostRole).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						return exception.ErrRecordNotFound
					}
					return err
				}

				// Update IsBecomeToHost to true
				if err = tx.Model(&models.User{}).
					Where("id = ?", id).
					Update("is_become_to_host", true).Error; err != nil {
					return err
				}

				// Assign host role to user
				if err = tx.Model(&user).Association("Roles").Append(&hostRole); err != nil {
					return err
				}

				return nil
			}); err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			if err == gorm.ErrRecordNotFound {
				err = exception.ErrRecordNotFound
			} else {
				err = e.(error)
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
		return err
	}

	return nil
}
