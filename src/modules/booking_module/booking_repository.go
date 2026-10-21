package booking_module

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/utils"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/booking_module/dtos"
)

// CreateBooking creates a new booking with booking orders
func (r Repository) CreateBooking(ctx context.Context, dto *dtos.CreateBooking) (*models.Booking, error) {
	ctx, span := r.tracer.TraceStart(ctx, "CreateBookingRepository", trace.WithAttributes(attribute.String("repository", "CreateBooking")))
	defer r.tracer.TraceEnd(span)

	// Start transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Prepare booking data - single loop for efficiency
	var totalPrice float64
	var bookingCheckIn, bookingCheckOut time.Time
	type parsedOrder struct {
		checkIn         time.Time
		checkOut        time.Time
		propertyID      *int
		roomCategory    models.RoomCategory
		orderTotalPrice float64
		pricePerRoom    float64
	}
	parsedOrders := make([]parsedOrder, len(dto.BookingOrders))

	for i, orderDTO := range dto.BookingOrders {
		// Parse dates once
		orderCheckIn, err := time.Parse("2006-01-02", orderDTO.CheckInDate)
		if err != nil {
			tx.Rollback()
			return nil, exception.ErrInvalidDateFormat
		}

		orderCheckOut, err := time.Parse("2006-01-02", orderDTO.CheckOutDate)
		if err != nil {
			tx.Rollback()
			return nil, exception.ErrInvalidDateFormat
		}

		// Validate: check-out date must be after check-in date
		if !orderCheckOut.After(orderCheckIn) {
			tx.Rollback()
			return nil, exception.ErrInvalidDateRange
		}

		// Get RoomCategory (includes PropertyID and BasePrice)
		var roomCategory models.RoomCategory
		if err := tx.First(&roomCategory, orderDTO.RoomCategoryID).Error; err != nil {
			tx.Rollback()
			return nil, exception.ErrRecordNotFound
		}

		// Calculate price for each date in the booking range
		var orderTotalPrice float64
		var totalDays int
		for d := orderCheckIn; d.Before(orderCheckOut); d = d.AddDate(0, 0, 1) {
			// Find RoomCategoryCalendar for this date
			var calendar models.RoomCategoryCalendar
			err := tx.Where("room_category_id = ? AND date = ?", orderDTO.RoomCategoryID, d).
				First(&calendar).Error

			if err != nil {
				tx.Rollback()
				return nil, exception.ErrCalendarDateNotAvailable
			}

			// Calculate final price for this date: BasePrice + PriceAdjustment
			finalPrice := roomCategory.BasePrice + calendar.PriceAdjustment
			orderTotalPrice += finalPrice * float64(orderDTO.RoomsBooked)
			totalDays++
		}

		// Calculate average price per room for this order
		pricePerRoom := 0.0
		if totalDays > 0 {
			pricePerRoom = orderTotalPrice / float64(orderDTO.RoomsBooked) / float64(totalDays)
		}

		// Store parsed data for reuse
		parsedOrders[i] = parsedOrder{
			checkIn:         orderCheckIn,
			checkOut:        orderCheckOut,
			propertyID:      roomCategory.PropertyID,
			roomCategory:    roomCategory,
			orderTotalPrice: orderTotalPrice,
			pricePerRoom:    pricePerRoom,
		}

		// Add to total price
		totalPrice += orderTotalPrice

		// Calculate booking date range
		if i == 0 || orderCheckIn.Before(bookingCheckIn) {
			bookingCheckIn = orderCheckIn
		}
		if i == 0 || orderCheckOut.After(bookingCheckOut) {
			bookingCheckOut = orderCheckOut
		}
	}

	// Override with provided dates if available
	if dto.CheckInDate != "" {
		if parsedCheckIn, err := time.Parse("2006-01-02", dto.CheckInDate); err == nil {
			bookingCheckIn = parsedCheckIn
		}
	}
	if dto.CheckOutDate != "" {
		if parsedCheckOut, err := time.Parse("2006-01-02", dto.CheckOutDate); err == nil {
			bookingCheckOut = parsedCheckOut
		}
	}

	// Create booking entity
	booking := models.Booking{
		CheckInDate:   bookingCheckIn,
		CheckOutDate:  bookingCheckOut,
		NumberOfPets:  *dto.NumberOfPets,
		TotalPrice:    totalPrice,
		BookingStatus: models.BookingStatusPending,
		GuestNote:     dto.GuestNote,
		GuestID:       dto.GuestID,
	}

	// Create booking
	if err := tx.Create(&booking).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Associate pet profiles if provided
	if len(dto.PetProfileIDs) > 0 {
		var petProfiles []models.PetProfile
		if err := tx.Where("id IN ?", dto.PetProfileIDs).Find(&petProfiles).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Model(&booking).Association("PetProfiles").Append(&petProfiles); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Create booking orders using parsed data
	bookingOrders := make([]models.BookingOrder, len(dto.BookingOrders))
	for i, orderDTO := range dto.BookingOrders {
		parsed := parsedOrders[i]

		bookingOrders[i] = models.BookingOrder{
			CheckInDate:    parsed.checkIn,
			CheckOutDate:   parsed.checkOut,
			RoomsBooked:    orderDTO.RoomsBooked,
			PricePerRoom:   parsed.pricePerRoom,
			TotalPrice:     parsed.orderTotalPrice,
			RoomCategoryID: orderDTO.RoomCategoryID,
			PropertyID:     parsed.propertyID,
			BookingID:      &booking.ID,
		}
	}

	if err := tx.Create(&bookingOrders).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update BookedRooms in RoomCategoryCalendar using parsed dates
	for i, orderDTO := range dto.BookingOrders {
		parsed := parsedOrders[i]

		// Loop through each date in the booking range
		for d := parsed.checkIn; d.Before(parsed.checkOut); d = d.AddDate(0, 0, 1) {
			// Find or create RoomCategoryCalendar for this date
			var calendar models.RoomCategoryCalendar
			err := tx.Set("gorm:query_option", "FOR UPDATE").
				Where("room_category_id = ? AND date = ?", orderDTO.RoomCategoryID, d).
				First(&calendar).Error

			if err != nil {
				// If calendar doesn't exist for this date, return specific error
				tx.Rollback()
				return nil, exception.ErrCalendarDateNotAvailable
			}

			// Check if enough rooms are available
			if calendar.AvailableRooms < orderDTO.RoomsBooked {
				tx.Rollback()
				return nil, exception.ErrInsufficientRooms
			}

			// Update available rooms (deduct booked rooms)
			calendar.AvailableRooms -= orderDTO.RoomsBooked
			// Update booked rooms (increment)
			calendar.BookedRooms += orderDTO.RoomsBooked

			if err := tx.Save(&calendar).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload booking with associations
	if err := r.db.WithContext(ctx).
		Preload("BookingOrders").
		Preload("Guest").
		Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
		First(&booking, booking.ID).Error; err != nil {
		return nil, err
	}

	return &booking, nil
}

// GetAllBookingsPaginated retrieves all bookings for a specific guest with pagination
func (r Repository) GetAllBookingsPaginated(ctx context.Context, guestID int, pagination *paginator.Pagination) ([]models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllBookingsPaginatedRepository", trace.WithAttributes(attribute.String("repository", "GetAllBookingsPaginated"), attribute.Int("guestId", guestID)))
		bookings     = make([]models.Booking, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			baseDB := r.db.WithContext(ctx).Where("guest_id = ?", guestID)
			if err = r.db.WithContext(ctx).
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("Guest").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				Where("guest_id = ?", guestID).
				Scopes(paginator.Paginate(&models.Booking{}, pagination, baseDB)).
				Find(&bookings).Error; err != nil {
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
		return nil, err
	}

	return bookings, nil
}

// GetAllBookings retrieves all bookings for a specific guest
func (r Repository) GetAllBookings(ctx context.Context, guestID int) ([]models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllBookingsRepository", trace.WithAttributes(attribute.String("repository", "GetAllBookings"), attribute.Int("guestId", guestID)))
		bookings     = make([]models.Booking, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.WithContext(ctx).
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("Guest").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				Where("guest_id = ?", guestID).
				Order("created_at DESC").
				Find(&bookings).Error; err != nil {
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
		return nil, err
	}

	return bookings, nil
}

// GetBookingByID retrieves a single booking by ID with all associations
func (r Repository) GetBookingByID(ctx context.Context, id int) (*models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetBookingByIDRepository", trace.WithAttributes(attribute.String("repository", "GetBookingByID"), attribute.Int("bookingId", id)))
		booking      models.Booking
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query
			if err = r.db.WithContext(ctx).
				Preload("Guest").
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				First(&booking, id).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found
			if err.Error() == "record not found" {
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

	return &booking, nil
}

// ConfirmBookingPayment updates booking status to confirmed and sets confirmedAt timestamp
func (r Repository) ConfirmBookingPayment(ctx context.Context, bookingID int) (*models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "ConfirmBookingPaymentRepository", trace.WithAttributes(attribute.String("repository", "ConfirmBookingPayment"), attribute.Int("bookingId", bookingID)))
		booking      models.Booking
		err          error
	)

	utils.Block{
		Try: func() {
			// First, get the booking to verify it exists and is in pending status
			if err = r.db.WithContext(ctx).First(&booking, bookingID).Error; err != nil {
				utils.Throw(err)
			}

			// Check if booking is already confirmed
			if booking.BookingStatus == models.BookingStatusConfirmed {
				utils.Throw(exception.ErrBookingAlreadyConfirmed)
			}

			// Check if booking is in pending status (only pending can be confirmed)
			if booking.BookingStatus != models.BookingStatusPending {
				utils.Throw(exception.ErrInvalidBookingStatus)
			}

			// Update booking status to confirmed and set confirmedAt timestamp
			now := time.Now()
			if err = r.db.WithContext(ctx).Model(&booking).Updates(map[string]interface{}{
				"booking_status": models.BookingStatusConfirmed,
				"confirmed_at":   now,
			}).Error; err != nil {
				utils.Throw(err)
			}

			// Reload booking with associations
			if err = r.db.WithContext(ctx).
				Preload("Guest").
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				First(&booking, bookingID).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found
			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else if err != exception.ErrInvalidBookingStatus && err != exception.ErrBookingAlreadyConfirmed {
				// Logging (don't log business logic errors)
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

	return &booking, nil
}

// CancelBooking updates booking status to cancelled and returns rooms to inventory
func (r Repository) CancelBooking(ctx context.Context, bookingID int) (*models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CancelBookingRepository", trace.WithAttributes(attribute.String("repository", "CancelBooking"), attribute.Int("bookingId", bookingID)))
		booking      models.Booking
		err          error
	)

	utils.Block{
		Try: func() {
			// Start transaction
			tx := r.db.WithContext(ctx).Begin()
			if tx.Error != nil {
				utils.Throw(tx.Error)
			}

			// Get the booking with booking orders
			if err = tx.Preload("BookingOrders").First(&booking, bookingID).Error; err != nil {
				tx.Rollback()
				utils.Throw(err)
			}

			// Check if booking is already cancelled
			if booking.BookingStatus == models.BookingStatusCancelled {
				tx.Rollback()
				utils.Throw(exception.ErrBookingAlreadyCancelled)
			}

			// Only pending or confirmed bookings can be cancelled
			if booking.BookingStatus != models.BookingStatusPending && booking.BookingStatus != models.BookingStatusConfirmed {
				tx.Rollback()
				utils.Throw(exception.ErrCannotCancelBooking)
			}

			// Return rooms to inventory (update RoomCategoryCalendar)
			for _, order := range booking.BookingOrders {
				// Loop through each date in the booking range
				for d := order.CheckInDate; d.Before(order.CheckOutDate); d = d.AddDate(0, 0, 1) {
					var calendar models.RoomCategoryCalendar
					err := tx.Where("room_category_id = ? AND date = ?", order.RoomCategoryID, d).
						First(&calendar).Error

					if err == nil {
						// Return the rooms (increase available, decrease booked)
						calendar.AvailableRooms += order.RoomsBooked
						calendar.BookedRooms -= order.RoomsBooked

						if err := tx.Save(&calendar).Error; err != nil {
							tx.Rollback()
							utils.Throw(err)
						}
					}
					// If calendar not found, skip (data might have been deleted)
				}
			}

			// Update booking status to cancelled
			if err = tx.Model(&booking).Update("booking_status", models.BookingStatusCancelled).Error; err != nil {
				tx.Rollback()
				utils.Throw(err)
			}

			// Commit transaction
			if err = tx.Commit().Error; err != nil {
				utils.Throw(err)
			}

			// Reload booking with associations
			if err = r.db.WithContext(ctx).
				Preload("Guest").
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				First(&booking, bookingID).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found
			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else if err != exception.ErrBookingAlreadyCancelled && err != exception.ErrCannotCancelBooking {
				// Logging (don't log business logic errors)
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

	return &booking, nil
}

// SuccessBooking updates booking status to successful and sets successedAt timestamp
func (r Repository) SuccessBooking(ctx context.Context, bookingID int) (*models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "SuccessBookingRepository", trace.WithAttributes(attribute.String("repository", "SuccessBooking"), attribute.Int("bookingId", bookingID)))
		booking      models.Booking
		err          error
	)

	utils.Block{
		Try: func() {
			// First, get the booking to verify it exists
			if err = r.db.WithContext(ctx).First(&booking, bookingID).Error; err != nil {
				utils.Throw(err)
			}

			// Check if booking is already successful
			if booking.BookingStatus == models.BookingStatusSuccessful {
				utils.Throw(exception.ErrBookingAlreadySuccessful)
			}

			// Only confirmed bookings can be marked as successful
			if booking.BookingStatus != models.BookingStatusConfirmed {
				utils.Throw(exception.ErrCannotSuccessBooking)
			}

			// Check if current time is after 12:00 on checkout day
			now := time.Now()
			checkoutDay12PM := time.Date(booking.CheckOutDate.Year(), booking.CheckOutDate.Month(), booking.CheckOutDate.Day(), 12, 0, 0, 0, now.Location())
			if now.Before(checkoutDay12PM) {
				utils.Throw(exception.ErrBookingSuccessNotAllowedYet)
			}
			if err = r.db.WithContext(ctx).Model(&booking).Updates(map[string]interface{}{
				"booking_status": models.BookingStatusSuccessful,
				"successed_at":   now,
			}).Error; err != nil {
				utils.Throw(err)
			}

			// Reload booking with associations
			if err = r.db.WithContext(ctx).
				Preload("Guest").
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				First(&booking, bookingID).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found
			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else if err != exception.ErrBookingAlreadySuccessful && err != exception.ErrCannotSuccessBooking {
				// Logging (don't log business logic errors)
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

	return &booking, nil
}

// GetAllBookingsByHostIDPaginated retrieves all bookings for properties owned by a specific host with pagination
func (r Repository) GetAllBookingsByHostIDPaginated(ctx context.Context, hostID int, pagination *paginator.Pagination) ([]models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllBookingsByHostIDPaginatedRepository", trace.WithAttributes(attribute.String("repository", "GetAllBookingsByHostIDPaginated"), attribute.Int("hostId", hostID)))
		bookings     = make([]models.Booking, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			baseDB := r.db.WithContext(ctx).
				Joins("JOIN tbl_booking_orders ON tbl_booking_orders.booking_id = tbl_bookings.id").
				Joins("JOIN tbl_properties ON tbl_properties.id = tbl_booking_orders.property_id").
				Where("tbl_properties.host_id = ?", hostID).
				Distinct("tbl_bookings.id")

			if err = r.db.WithContext(ctx).
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("Guest").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				Joins("JOIN tbl_booking_orders ON tbl_booking_orders.booking_id = tbl_bookings.id").
				Joins("JOIN tbl_properties ON tbl_properties.id = tbl_booking_orders.property_id").
				Where("tbl_properties.host_id = ?", hostID).
				Group("tbl_bookings.id").
				Scopes(paginator.Paginate(&models.Booking{}, pagination, baseDB)).
				Find(&bookings).Error; err != nil {
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
		return nil, err
	}

	return bookings, nil
}

// GetBookingsByHostID retrieves all bookings for properties owned by a specific host
func (r Repository) GetBookingsByHostID(ctx context.Context, hostID int) ([]models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetBookingsByHostIDRepository", trace.WithAttributes(attribute.String("repository", "GetBookingsByHostID"), attribute.Int("hostId", hostID)))
		bookings     = make([]models.Booking, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query - find bookings through booking_orders -> properties -> host
			// Using a subquery to find bookings that have at least one booking order for a property owned by this host
			if err = r.db.WithContext(ctx).
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("Guest").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				Joins("JOIN tbl_booking_orders ON tbl_booking_orders.booking_id = tbl_bookings.id").
				Joins("JOIN tbl_properties ON tbl_properties.id = tbl_booking_orders.property_id").
				Where("tbl_properties.host_id = ?", hostID).
				Group("tbl_bookings.id").
				Order("tbl_bookings.created_at DESC").
				Find(&bookings).Error; err != nil {
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
		return nil, err
	}

	return bookings, nil
}

// GetBookingByIDForHost retrieves a booking by ID for a specific host (validates ownership)
func (r Repository) GetBookingByIDForHost(ctx context.Context, bookingID int, hostID int) (*models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetBookingByIDForHostRepository", trace.WithAttributes(attribute.String("repository", "GetBookingByIDForHost"), attribute.Int("bookingId", bookingID), attribute.Int("hostId", hostID)))
		booking      models.Booking
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query - find booking and verify it belongs to a property owned by this host
			if err = r.db.WithContext(ctx).
				Preload("Guest").
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				Joins("JOIN tbl_booking_orders ON tbl_booking_orders.booking_id = tbl_bookings.id").
				Joins("JOIN tbl_properties ON tbl_properties.id = tbl_booking_orders.property_id").
				Where("tbl_bookings.id = ?", bookingID).
				Where("tbl_properties.host_id = ?", hostID).
				Group("tbl_bookings.id").
				First(&booking).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found
			if err.Error() == "record not found" {
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

	return &booking, nil
}

// CreateBookingReview creates a review for a completed booking
func (r Repository) CreateBookingReview(ctx context.Context, bookingID int, guestID int, dto *dtos.CreateBookingReview) (*models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateBookingReviewRepository", trace.WithAttributes(attribute.String("repository", "CreateBookingReview")))
		booking      models.Booking
		err          error
	)

	utils.Block{
		Try: func() {
			// Get booking by ID and verify ownership
			if err = r.db.
				Where("id = ?", bookingID).
				Where("guest_id = ?", guestID).
				First(&booking).Error; err != nil {
				utils.Throw(err)
			}

			// Check if booking status is "successful"
			if booking.BookingStatus != models.BookingStatusSuccessful {
				err = exception.ErrBookingNotSuccessful
				utils.Throw(err)
			}

			// Check if booking already has a review
			if booking.Rating != nil {
				err = exception.ErrBookingAlreadyReviewed
				utils.Throw(err)
			}

			// Update booking with review
			if err = r.db.
				Model(&booking).
				Updates(map[string]interface{}{
					"rating":  dto.Rating,
					"comment": dto.Comment,
				}).Error; err != nil {
				utils.Throw(err)
			}

			// Reload booking to get updated data
			if err = r.db.First(&booking, bookingID).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found or expected errors
			if err.Error() == "record not found" {
				err = exception.ErrRecordNotFound
			} else if err == exception.ErrBookingNotSuccessful || err == exception.ErrBookingAlreadyReviewed {
				// Keep the error as is (expected errors)
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
		return nil, err
	}

	return &booking, nil
}

func (r Repository) CreateBookingReviewImages(ctx context.Context, images []models.BookingReviewImage) error {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "CreateBookingReviewImagesRepository", trace.WithAttributes(attribute.String("repository", "CreateBookingReviewImages")))
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

	if err != nil {
		return err
	}
	return nil
}

// GetAllBookingsByPropertyIDPaginated retrieves all bookings for a specific property owned by a host with pagination
func (r Repository) GetAllBookingsByPropertyIDPaginated(ctx context.Context, propertyID int, hostID int, pagination *paginator.Pagination) ([]models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetAllBookingsByPropertyIDPaginatedRepository", trace.WithAttributes(attribute.String("repository", "GetAllBookingsByPropertyIDPaginated"), attribute.Int("propertyId", propertyID), attribute.Int("hostId", hostID)))
		bookings     = make([]models.Booking, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			baseDB := r.db.WithContext(ctx).
				Joins("JOIN tbl_booking_orders ON tbl_booking_orders.booking_id = tbl_bookings.id").
				Joins("JOIN tbl_properties ON tbl_properties.id = tbl_booking_orders.property_id").
				Where("tbl_booking_orders.property_id = ?", propertyID).
				Where("tbl_properties.host_id = ?", hostID).
				Distinct("tbl_bookings.id")

			if err = r.db.WithContext(ctx).
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("Guest").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				Joins("JOIN tbl_booking_orders ON tbl_booking_orders.booking_id = tbl_bookings.id").
				Joins("JOIN tbl_properties ON tbl_properties.id = tbl_booking_orders.property_id").
				Where("tbl_booking_orders.property_id = ?", propertyID).
				Where("tbl_properties.host_id = ?", hostID).
				Group("tbl_bookings.id").
				Scopes(paginator.Paginate(&models.Booking{}, pagination, baseDB)).
				Find(&bookings).Error; err != nil {
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
		return nil, err
	}

	return bookings, nil
}

// GetBookingsByPropertyID retrieves all bookings for a specific property owned by a host
func (r Repository) GetBookingsByPropertyID(ctx context.Context, propertyID int, hostID int) ([]models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetBookingsByPropertyIDRepository", trace.WithAttributes(attribute.String("repository", "GetBookingsByPropertyID"), attribute.Int("propertyId", propertyID), attribute.Int("hostId", hostID)))
		bookings     = make([]models.Booking, 0)
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query - find bookings for specific property and verify host owns it
			if err = r.db.WithContext(ctx).
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("Guest").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				Joins("JOIN tbl_booking_orders ON tbl_booking_orders.booking_id = tbl_bookings.id").
				Joins("JOIN tbl_properties ON tbl_properties.id = tbl_booking_orders.property_id").
				Where("tbl_booking_orders.property_id = ?", propertyID).
				Where("tbl_properties.host_id = ?", hostID).
				Group("tbl_bookings.id").
				Order("tbl_bookings.created_at DESC").
				Find(&bookings).Error; err != nil {
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
		return nil, err
	}

	return bookings, nil
}

// GetBookingByIDForProperty retrieves a booking by ID for a specific property (validates ownership)
func (r Repository) GetBookingByIDForProperty(ctx context.Context, bookingID int, propertyID int, hostID int) (*models.Booking, error) {
	var (
		_, childSpan = r.tracer.TraceStart(ctx, "GetBookingByIDForPropertyRepository", trace.WithAttributes(attribute.String("repository", "GetBookingByIDForProperty"), attribute.Int("bookingId", bookingID), attribute.Int("propertyId", propertyID), attribute.Int("hostId", hostID)))
		booking      models.Booking
		err          error
	)

	utils.Block{
		Try: func() {
			// Execute query - find booking for specific property and verify host owns it
			if err = r.db.WithContext(ctx).
				Preload("Guest").
				Preload("BookingOrders.Property").
				Preload("BookingOrders.RoomCategory").
				Preload("PetProfiles.Breed").Preload("PetProfiles.PetType").Preload("PetProfiles.PetEnergy").Preload("PetProfiles.PetSizeCategory").
				Joins("JOIN tbl_booking_orders ON tbl_booking_orders.booking_id = tbl_bookings.id").
				Joins("JOIN tbl_properties ON tbl_properties.id = tbl_booking_orders.property_id").
				Where("tbl_bookings.id = ?", bookingID).
				Where("tbl_booking_orders.property_id = ?", propertyID).
				Where("tbl_properties.host_id = ?", hostID).
				Group("tbl_bookings.id").
				First(&booking).Error; err != nil {
				utils.Throw(err)
			}
		},
		Catch: func(e utils.Exception) {
			err = e.(error)
			// Check if record not found
			if err.Error() == "record not found" {
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

	return &booking, nil
}
