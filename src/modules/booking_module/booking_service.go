package booking_module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"joodpakpaw/internal/exception"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/booking_module/dtos"
	"joodpakpaw/src/modules/booking_module/responses"
	"joodpakpaw/src/modules/notification_module"
)

// CreateBooking creates a new booking
func (s Service) CreateBooking(ctx context.Context, dto *dtos.CreateBooking) (*responses.CreateBookingResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "CreateBookingService", trace.WithAttributes(attribute.String("service", "CreateBooking")))
	defer s.tracer.TraceEnd(span)

	// Validate guest exists
	_, err := s.userRepository().GetUserByID(ctx, *dto.GuestID)
	if err != nil {
		return nil, exception.ErrRecordNotFound
	}

	// Validate mutual exclusivity: numberOfPets and petProfileIds cannot both be provided
	if dto.NumberOfPets != nil && len(dto.PetProfileIDs) > 0 {
		return nil, exception.ErrNumberOfPetsConflict
	}

	// Auto-fill numberOfPets from petProfileIds length if only petProfileIds is provided
	if dto.NumberOfPets == nil && len(dto.PetProfileIDs) > 0 {
		count := len(dto.PetProfileIDs)
		dto.NumberOfPets = &count
	}

	// Default to 0 if neither numberOfPets nor petProfileIds provided
	if dto.NumberOfPets == nil {
		zero := 0
		dto.NumberOfPets = &zero
	}

	// Validate pet profiles belong to the guest
	if len(dto.PetProfileIDs) > 0 {
		var count int64
		if err := s.bookingRepository().db.
			Model(&models.PetProfile{}).
			Where("id IN ? AND guest_id = ?", dto.PetProfileIDs, *dto.GuestID).
			Count(&count).Error; err != nil {
			return nil, exception.ErrInvalidPetProfile
		}
		if int(count) != len(dto.PetProfileIDs) {
			return nil, exception.ErrInvalidPetProfile
		}

		// Find the maximum pet size code among all pets in the booking
		// small=0, medium=1, large=2
		var petProfiles []models.PetProfile
		if err := s.bookingRepository().db.
			Preload("PetSizeCategory").
			Where("id IN ?", dto.PetProfileIDs).
			Find(&petProfiles).Error; err != nil {
			return nil, exception.ErrInvalidPetProfile
		}
		maxPetSizeCode := -1
		for _, p := range petProfiles {
			if p.PetSizeCategory != nil && p.PetSizeCategory.Code > maxPetSizeCode {
				maxPetSizeCode = p.PetSizeCategory.Code
			}
		}

		// Validate each room category supports the largest pet size
		if maxPetSizeCode >= 0 {
			for _, order := range dto.BookingOrders {
				var roomCat models.RoomCategory
				if err := s.bookingRepository().db.Preload("PetSizeCategory").First(&roomCat, order.RoomCategoryID).Error; err != nil {
					return nil, exception.ErrRecordNotFound
				}
				// Room must support pet size >= maxPetSizeCode
				if roomCat.PetSizeCategory == nil || roomCat.PetSizeCategory.Code < maxPetSizeCode {
					return nil, exception.ErrPetSizeNotCompatible
				}
			}
		}
	}

	// Validate all room categories in booking orders
	totalPetCapacity := 0
	for _, order := range dto.BookingOrders {
		// Validate room category exists (and this will validate property too since RoomCategory has PropertyID)
		var roomCategory models.RoomCategory
		if err := s.bookingRepository().db.Preload("Property").First(&roomCategory, order.RoomCategoryID).Error; err != nil {
			return nil, exception.ErrRecordNotFound
		}

		// Accumulate total pet capacity across all orders
		totalPetCapacity += roomCategory.NumberOfPets * order.RoomsBooked

		// Validate that the room category belongs to an existing property
		if roomCategory.PropertyID == nil || roomCategory.Property == nil {
			return nil, exception.ErrRecordNotFound
		}

		// Validate that the property is not banned by admin
		if roomCategory.Property.IsBannedByAdmin {
			return nil, exception.ErrRecordNotFound
		}

		// Validate that the guest is not the host of the property
		if roomCategory.Property.HostID != nil && *roomCategory.Property.HostID == *dto.GuestID {
			return nil, exception.ErrCannotBookOwnProperty
		}
	}

	// Validate that the number of pets does not exceed the total pet capacity
	if *dto.NumberOfPets > totalPetCapacity {
		return nil, exception.ErrInsufficientPetCapacity
	}

	// Create booking through repository
	booking, err := s.bookingRepository().CreateBooking(ctx, dto)
	if err != nil {
		return nil, err
	}

	// Send notification to guest
	go s.sendBookingCreatedNotification(ctx, booking, dto)

	// Build response
	response := new(responses.CreateBookingResponse).Make(*booking)

	return response, nil
}

func (s Service) sendBookingCreatedNotification(ctx context.Context, booking *models.Booking, dto *dtos.CreateBooking) {
	if booking.GuestID == nil || len(dto.BookingOrders) == 0 {
		return
	}

	// Get property name from first booking order's room category
	var roomCat models.RoomCategory
	if err := s.bookingRepository().db.Preload("Property").First(&roomCat, dto.BookingOrders[0].RoomCategoryID).Error; err != nil {
		return
	}
	if roomCat.Property == nil {
		return
	}

	checkIn := booking.CheckInDate.Format("2 Jan 06")
	checkOut := booking.CheckOutDate.Format("2 Jan 06")
	propertyName := roomCat.Property.Title
	frontendURL := os.Getenv("APP_FRONTEND_URL")

	// Notify guest
	guestTitle := fmt.Sprintf("Your booking %s - %s at %s is booked", checkIn, checkOut, propertyName)
	notification_module.NewService().SendToUser(ctx, *booking.GuestID, "guest", "booking", guestTitle, "", "bookings", &booking.ID)

	// Send email to guest
	var guest models.User
	if err := s.bookingRepository().db.First(&guest, *booking.GuestID).Error; err == nil {
		guestURL := fmt.Sprintf("%s/profile/bookings/%d/?from=notifications", frontendURL, booking.ID)
		guestName := fmt.Sprintf("%s %s", guest.FirstName, guest.LastName)
		if err := s.mailService.SendBookingCreatedEmail(guest.Email, guestName, propertyName, checkIn, checkOut, booking.NumberOfPets, guestURL, false); err != nil {
			s.logger.Error("Failed to send booking email to guest", zap.Error(err), zap.String("email", guest.Email))
		}
	}

	// Notify host
	if roomCat.Property.HostID != nil {
		hostTitle := fmt.Sprintf("New booking at %s | %s - %s | %d pets", propertyName, checkIn, checkOut, booking.NumberOfPets)
		notification_module.NewService().SendToUser(ctx, *roomCat.Property.HostID, "host", "booking", hostTitle, "", "bookings", &booking.ID)

		// Send email to host
		var host models.User
		if err := s.bookingRepository().db.First(&host, *roomCat.Property.HostID).Error; err == nil {
			hostURL := fmt.Sprintf("%s/profile/host/booking/%d/?from=notifications", frontendURL, booking.ID)
			hostName := fmt.Sprintf("%s %s", host.FirstName, host.LastName)
			if err := s.mailService.SendBookingCreatedEmail(host.Email, hostName, propertyName, checkIn, checkOut, booking.NumberOfPets, hostURL, true); err != nil {
				s.logger.Error("Failed to send booking email to host", zap.Error(err), zap.String("email", host.Email))
			}
		}
	}
}

// GetAllBookingsPaginated retrieves all bookings for a specific guest with pagination
func (s Service) GetAllBookingsPaginated(ctx context.Context, guestID int, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetAllBookingsPaginatedService", trace.WithAttributes(attribute.String("service", "GetAllBookingsPaginated")))
	defer s.tracer.TraceEnd(span)

	bookings, err := s.bookingRepository().GetAllBookingsPaginated(ctx, guestID, pagination)
	if err != nil {
		return nil, err
	}

	pagination.Data = new(responses.GetBookingsResponse).Make(bookings).Bookings
	return pagination, nil
}

// GetAllBookings retrieves all bookings for a specific guest
func (s Service) GetAllBookings(ctx context.Context, guestID int) (*responses.GetBookingsResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetAllBookingsService", trace.WithAttributes(attribute.String("service", "GetAllBookings")))
	defer s.tracer.TraceEnd(span)

	bookings, err := s.bookingRepository().GetAllBookings(ctx, guestID)
	if err != nil {
		return nil, err
	}

	response := new(responses.GetBookingsResponse).Make(bookings)
	return response, nil
}

// GetBookingByID retrieves a booking by ID
func (s Service) GetBookingByID(ctx context.Context, id int) (*responses.GetBookingByIDResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetBookingByIDService", trace.WithAttributes(attribute.String("service", "GetBookingByID")))
	defer s.tracer.TraceEnd(span)

	booking, err := s.bookingRepository().GetBookingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := new(responses.GetBookingByIDResponse).Make(*booking)
	return response, nil
}

// GetPropertyByID retrieves a property by ID (helper for validation)
func (r Repository) GetPropertyByID(ctx context.Context, id int) (*models.Property, error) {
	var property models.Property
	if err := r.db.WithContext(ctx).First(&property, id).Error; err != nil {
		return nil, err
	}
	return &property, nil
}

// GetUserByID retrieves a user by ID (helper for validation)
func (r Repository) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ConfirmBookingPayment confirms payment for a booking and updates status to confirmed
func (s Service) ConfirmBookingPayment(ctx context.Context, bookingID int) (*responses.GetBookingByIDResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "ConfirmBookingPaymentService", trace.WithAttributes(attribute.String("service", "ConfirmBookingPayment")))
	defer s.tracer.TraceEnd(span)

	// Call repository to confirm payment
	booking, err := s.bookingRepository().ConfirmBookingPayment(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	// Send payment confirmed notifications
	go s.sendPaymentConfirmedNotification(ctx, booking)

	// Build response
	response := new(responses.GetBookingByIDResponse).Make(*booking)
	return response, nil
}

func (s Service) sendPaymentConfirmedNotification(ctx context.Context, booking *models.Booking) {
	if len(booking.BookingOrders) == 0 {
		return
	}

	order := booking.BookingOrders[0]
	if order.Property == nil {
		return
	}

	checkIn := booking.CheckInDate.Format("2 Jan 06")
	checkOut := booking.CheckOutDate.Format("2 Jan 06")
	propertyName := order.Property.Title
	frontendURL := os.Getenv("APP_FRONTEND_URL")

	// Notify guest
	if booking.GuestID != nil {
		guestTitle := fmt.Sprintf("Payment confirmed for your booking at %s | %s - %s", propertyName, checkIn, checkOut)
		notification_module.NewService().SendToUser(ctx, *booking.GuestID, "guest", "payment", guestTitle, "", "bookings", &booking.ID)

		// Send email to guest
		var guest models.User
		if err := s.bookingRepository().db.First(&guest, *booking.GuestID).Error; err == nil {
			guestURL := fmt.Sprintf("%s/profile/bookings/%d/?from=notifications", frontendURL, booking.ID)
			guestName := fmt.Sprintf("%s %s", guest.FirstName, guest.LastName)
			if err := s.mailService.SendPaymentConfirmedEmail(guest.Email, guestName, propertyName, checkIn, checkOut, booking.NumberOfPets, guestURL, false); err != nil {
				s.logger.Error("Failed to send payment confirmed email to guest", zap.Error(err), zap.String("email", guest.Email))
			}
		}
	}

	// Notify host
	if order.Property.HostID != nil {
		hostTitle := fmt.Sprintf("Payment received for booking at %s | %s - %s | %d pets", propertyName, checkIn, checkOut, booking.NumberOfPets)
		notification_module.NewService().SendToUser(ctx, *order.Property.HostID, "host", "payment", hostTitle, "", "bookings", &booking.ID)

		// Send email to host
		var host models.User
		if err := s.bookingRepository().db.First(&host, *order.Property.HostID).Error; err == nil {
			hostURL := fmt.Sprintf("%s/profile/host/booking/%d/?from=notifications", frontendURL, booking.ID)
			hostName := fmt.Sprintf("%s %s", host.FirstName, host.LastName)
			if err := s.mailService.SendPaymentConfirmedEmail(host.Email, hostName, propertyName, checkIn, checkOut, booking.NumberOfPets, hostURL, true); err != nil {
				s.logger.Error("Failed to send payment confirmed email to host", zap.Error(err), zap.String("email", host.Email))
			}
		}
	}
}

// CancelBooking cancels a booking and returns rooms to inventory
func (s Service) CancelBooking(ctx context.Context, bookingID int) (*responses.GetBookingByIDResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "CancelBookingService", trace.WithAttributes(attribute.String("service", "CancelBooking")))
	defer s.tracer.TraceEnd(span)

	// Call repository to cancel booking
	booking, err := s.bookingRepository().CancelBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	// Send cancellation notifications
	go s.sendCancelBookingNotification(ctx, booking)

	// Build response
	response := new(responses.GetBookingByIDResponse).Make(*booking)
	return response, nil
}

func (s Service) sendCancelBookingNotification(ctx context.Context, booking *models.Booking) {
	if len(booking.BookingOrders) == 0 {
		return
	}

	order := booking.BookingOrders[0]
	if order.Property == nil {
		return
	}

	checkIn := booking.CheckInDate.Format("2 Jan 06")
	checkOut := booking.CheckOutDate.Format("2 Jan 06")
	propertyName := order.Property.Title
	frontendURL := os.Getenv("APP_FRONTEND_URL")

	// Notify guest
	if booking.GuestID != nil {
		guestTitle := fmt.Sprintf("Your booking at %s | %s - %s has been cancelled", propertyName, checkIn, checkOut)
		notification_module.NewService().SendToUser(ctx, *booking.GuestID, "guest", "cancel", guestTitle, "", "bookings", &booking.ID)

		// Send email to guest
		var guest models.User
		if err := s.bookingRepository().db.First(&guest, *booking.GuestID).Error; err == nil {
			guestURL := fmt.Sprintf("%s/profile/bookings/%d/?from=notifications", frontendURL, booking.ID)
			guestName := fmt.Sprintf("%s %s", guest.FirstName, guest.LastName)
			if err := s.mailService.SendBookingCancelledEmail(guest.Email, guestName, propertyName, checkIn, checkOut, booking.NumberOfPets, guestURL, false); err != nil {
				s.logger.Error("Failed to send booking cancelled email to guest", zap.Error(err), zap.String("email", guest.Email))
			}
		}
	}

	// Notify host
	if order.Property.HostID != nil {
		hostTitle := fmt.Sprintf("Booking at %s | %s - %s | %d pets has been cancelled", propertyName, checkIn, checkOut, booking.NumberOfPets)
		notification_module.NewService().SendToUser(ctx, *order.Property.HostID, "host", "cancel", hostTitle, "", "bookings", &booking.ID)

		// Send email to host
		var host models.User
		if err := s.bookingRepository().db.First(&host, *order.Property.HostID).Error; err == nil {
			hostURL := fmt.Sprintf("%s/profile/host/booking/%d/?from=notifications", frontendURL, booking.ID)
			hostName := fmt.Sprintf("%s %s", host.FirstName, host.LastName)
			if err := s.mailService.SendBookingCancelledEmail(host.Email, hostName, propertyName, checkIn, checkOut, booking.NumberOfPets, hostURL, true); err != nil {
				s.logger.Error("Failed to send booking cancelled email to host", zap.Error(err), zap.String("email", host.Email))
			}
		}
	}
}

// SuccessBooking marks a booking as completed
func (s Service) SuccessBooking(ctx context.Context, bookingID int) (*responses.GetBookingByIDResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "SuccessBookingService", trace.WithAttributes(attribute.String("service", "SuccessBooking")))
	defer s.tracer.TraceEnd(span)

	// Call repository to mark booking as successful
	booking, err := s.bookingRepository().SuccessBooking(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	// Send stay completed notifications
	go s.sendStayCompletedNotification(ctx, booking)

	// Build response
	response := new(responses.GetBookingByIDResponse).Make(*booking)
	return response, nil
}

func (s Service) sendStayCompletedNotification(ctx context.Context, booking *models.Booking) {
	if len(booking.BookingOrders) == 0 {
		return
	}

	order := booking.BookingOrders[0]
	if order.Property == nil {
		return
	}

	propertyName := order.Property.Title
	frontendURL := os.Getenv("APP_FRONTEND_URL")

	// Guest → action: review (prompt to leave a review)
	if booking.GuestID != nil {
		guestTitle := fmt.Sprintf("Your pets' stay at %s is complete! Leave a review to share your experience", propertyName)
		notification_module.NewService().SendToUser(ctx, *booking.GuestID, "guest", "review", guestTitle, "", "bookings", &booking.ID)

		// Send stay completed email to guest
		var guest models.User
		if err := s.bookingRepository().db.First(&guest, *booking.GuestID).Error; err == nil {
			guestURL := fmt.Sprintf("%s/profile/bookings/%d/?from=notifications", frontendURL, booking.ID)
			guestName := fmt.Sprintf("%s %s", guest.FirstName, guest.LastName)
			if err := s.mailService.SendStayCompletedEmail(guest.Email, guestName, propertyName, guestURL, false); err != nil {
				s.logger.Error("Failed to send stay completed email to guest", zap.Error(err), zap.String("email", guest.Email))
			}
		}
	}

	// Host → action: booking (stay completed)
	if order.Property.HostID != nil {
		hostTitle := fmt.Sprintf("Check-out completed at %s", propertyName)
		notification_module.NewService().SendToUser(ctx, *order.Property.HostID, "host", "review", hostTitle, "", "bookings", &booking.ID)

		// Send stay completed email to host
		var host models.User
		if err := s.bookingRepository().db.First(&host, *order.Property.HostID).Error; err == nil {
			hostURL := fmt.Sprintf("%s/profile/host/booking/%d/?from=notifications", frontendURL, booking.ID)
			hostName := fmt.Sprintf("%s %s", host.FirstName, host.LastName)
			if err := s.mailService.SendStayCompletedEmail(host.Email, hostName, propertyName, hostURL, true); err != nil {
				s.logger.Error("Failed to send stay completed email to host", zap.Error(err), zap.String("email", host.Email))
			}
		}
	}
}

// GetBookingsByHostIDPaginated retrieves all bookings for properties owned by a specific host with pagination
func (s Service) GetBookingsByHostIDPaginated(ctx context.Context, hostID int, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetBookingsByHostIDPaginatedService", trace.WithAttributes(attribute.String("service", "GetBookingsByHostIDPaginated")))
	defer s.tracer.TraceEnd(span)

	bookings, err := s.bookingRepository().GetAllBookingsByHostIDPaginated(ctx, hostID, pagination)
	if err != nil {
		return nil, err
	}

	pagination.Data = new(responses.GetBookingsForHostResponse).Make(bookings).Bookings
	return pagination, nil
}

// GetBookingsByHostID retrieves all bookings for properties owned by a specific host
func (s Service) GetBookingsByHostID(ctx context.Context, hostID int) (*responses.GetBookingsForHostResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetBookingsByHostIDService", trace.WithAttributes(attribute.String("service", "GetBookingsByHostID")))
	defer s.tracer.TraceEnd(span)

	bookings, err := s.bookingRepository().GetBookingsByHostID(ctx, hostID)
	if err != nil {
		return nil, err
	}

	response := new(responses.GetBookingsForHostResponse).Make(bookings)
	return response, nil
}

// GetBookingByIDForHost retrieves a booking by ID for a specific host
func (s Service) GetBookingByIDForHost(ctx context.Context, bookingID int, hostID int) (*responses.GetBookingByIDForHostResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetBookingByIDForHostService", trace.WithAttributes(attribute.String("service", "GetBookingByIDForHost")))
	defer s.tracer.TraceEnd(span)

	booking, err := s.bookingRepository().GetBookingByIDForHost(ctx, bookingID, hostID)
	if err != nil {
		return nil, err
	}

	response := new(responses.GetBookingByIDForHostResponse).Make(*booking)
	return response, nil
}

// CreateBookingReview creates a review for a completed booking
func (s Service) CreateBookingReview(ctx context.Context, bookingID int, guestID int, dto *dtos.CreateBookingReview) (*responses.CreateBookingReviewResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "CreateBookingReviewService", trace.WithAttributes(attribute.String("service", "CreateBookingReview")))
	defer s.tracer.TraceEnd(span)

	// Check if user is banned from writing reviews
	user, err := s.userRepository().GetUserByID(ctx, guestID)
	if err != nil {
		return nil, exception.ErrRecordNotFound
	}
	if user.IsBannedReview {
		return nil, exception.ErrUserBannedFromReview
	}

	// Call repository to create review
	booking, err := s.bookingRepository().CreateBookingReview(ctx, bookingID, guestID, dto)
	if err != nil {
		return nil, err
	}

	// Handle image uploads if images are provided
	if len(dto.Images) > 0 {
		reviewImages := make([]models.BookingReviewImage, 0, len(dto.Images))
		basePath := "booking-reviews"

		for index, fileHeader := range dto.Images {
			timestamp := time.Now().UnixNano()
			ext := filepath.Ext(fileHeader.Filename)
			fileName := fmt.Sprintf("%d_%d_%d%s", booking.ID, timestamp, index, ext)

			if err := s.fileSystem.Put(ctx, basePath, fileName, fileHeader); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to upload review image: %v", err))
				continue
			}

			publicURL := s.fileSystem.GetPublicURL(basePath, fileName)

			reviewImages = append(reviewImages, models.BookingReviewImage{
				FileName:          fileHeader.Filename,
				FilePath:          publicURL,
				FileContentType:   fileHeader.Header.Get("Content-Type"),
				FileExtensionName: strings.TrimPrefix(ext, "."),
				FileSize:          int(fileHeader.Size),
				BookingID:         &booking.ID,
			})
		}

		if len(reviewImages) > 0 {
			if err = s.bookingRepository().CreateBookingReviewImages(ctx, reviewImages); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to save review images: %v", err))
			} else {
				booking.ReviewImages = reviewImages
			}
		}
	}

	// Build response
	response := new(responses.CreateBookingReviewResponse).Make(booking)
	return response, nil
}

// GetBookingsByPropertyIDPaginated retrieves all bookings for a specific property with pagination
func (s Service) GetBookingsByPropertyIDPaginated(ctx context.Context, propertyID int, hostID int, pagination *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetBookingsByPropertyIDPaginatedService", trace.WithAttributes(attribute.String("service", "GetBookingsByPropertyIDPaginated")))
	defer s.tracer.TraceEnd(span)

	bookings, err := s.bookingRepository().GetAllBookingsByPropertyIDPaginated(ctx, propertyID, hostID, pagination)
	if err != nil {
		return nil, err
	}

	pagination.Data = new(responses.GetBookingsForPropertyResponse).Make(bookings).Bookings
	return pagination, nil
}

// GetBookingsByPropertyID retrieves all bookings for a specific property
func (s Service) GetBookingsByPropertyID(ctx context.Context, propertyID int, hostID int) (*responses.GetBookingsForPropertyResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetBookingsByPropertyIDService", trace.WithAttributes(attribute.String("service", "GetBookingsByPropertyID")))
	defer s.tracer.TraceEnd(span)

	bookings, err := s.bookingRepository().GetBookingsByPropertyID(ctx, propertyID, hostID)
	if err != nil {
		return nil, err
	}

	response := new(responses.GetBookingsForPropertyResponse).Make(bookings)
	return response, nil
}

// GetBookingByIDForProperty retrieves a booking by ID for a specific property
func (s Service) GetBookingByIDForProperty(ctx context.Context, bookingID int, propertyID int, hostID int) (*responses.GetBookingByIDForPropertyResponse, error) {
	ctx, span := s.tracer.TraceStart(ctx, "GetBookingByIDForPropertyService", trace.WithAttributes(attribute.String("service", "GetBookingByIDForProperty")))
	defer s.tracer.TraceEnd(span)

	booking, err := s.bookingRepository().GetBookingByIDForProperty(ctx, bookingID, propertyID, hostID)
	if err != nil {
		return nil, err
	}

	response := new(responses.GetBookingByIDForPropertyResponse).Make(*booking)
	return response, nil
}
