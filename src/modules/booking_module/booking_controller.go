package booking_module

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"joodpakpaw/internal/exception"
	"joodpakpaw/internal/validator"
	"joodpakpaw/pkg/cache"
	"joodpakpaw/pkg/database/paginator"
	"joodpakpaw/src/models"
	"joodpakpaw/src/modules/booking_module/dtos"
	"joodpakpaw/src/modules/booking_module/responses"
)

// CreateBooking creates a new booking
//
//	@Summary		Create booking
//	@Description	Create a new booking with booking orders (guestId is automatically set from authenticated user)
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dtos.CreateBooking	true	"Booking data"
//	@Success		201		{object}	responses.CreateBookingResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/bookings [post]
func (c Controller) CreateBooking(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateBookingController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateBooking")))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Create data transfer object
	dto := new(dtos.CreateBooking)

	// Parse JSON body
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Set guest ID from authenticated user
	dto.GuestID = &authUser.ID

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	booking, err := c.bookingService().CreateBooking(ctx, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrInvalidDateFormat:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidDateFormatResponseError, err)
		case exception.ErrInvalidDateRange:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidDateRangeResponseError, err)
		case exception.ErrInsufficientRooms:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InsufficientRoomsResponseError, err)
		case exception.ErrInsufficientPetCapacity:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InsufficientPetCapacityResponseError, err)
		case exception.ErrCalendarDateNotAvailable:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.CalendarDateNotAvailableResponseError, err)
		case exception.ErrCannotBookOwnProperty:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.CannotBookOwnPropertyResponseError, err)
		case exception.ErrInvalidPetProfile:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidPetProfileResponseError, err)
		case exception.ErrNumberOfPetsConflict:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.NumberOfPetsConflictResponseError, err)
		case exception.ErrPetSizeNotCompatible:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.PetSizeNotCompatibleResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear booking cache
	c.m.cacher.Tag("bookings").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusCreated).JSON(booking)
}

// GetAllBookingsPaginated lists all bookings for authenticated user with pagination
//
//	@Summary		List all bookings (paginated)
//	@Description	Get all bookings for authenticated user with pagination support
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int		false	"Page number"		default(1)
//	@Param			limit	query		int		false	"Items per page"	default(20)
//	@Param			orderBy	query		string	false	"Order by field"	default(created_at)
//	@Param			sort	query		string	false	"Sort direction (asc/desc)"	default(desc)
//	@Success		200	{object}	paginator.Pagination
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/bookings [get]
func (c Controller) GetAllBookingsPaginated(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetAllBookingsPaginatedController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllBookingsPaginated")))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Parse pagination query params
	pagination := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 20)),
		paginator.WithOrderBy(f.Query("orderBy", "created_at")),
		paginator.WithSort(f.Query("sort", "desc")),
	)

	responseData, err := c.bookingService().GetAllBookingsPaginated(ctx, authUser.ID, pagination)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetAllBookings lists all bookings for authenticated user
//
//	@Summary		List all bookings
//	@Description	Get all bookings for authenticated user
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	responses.GetBookingsResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/bookings/all [get]
func (c Controller) GetAllBookings(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllBookingsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllBookings")))
		responseData *responses.GetBookingsResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing - include user ID in cache key
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"bookings"}, "GetAllBookings", authUser.ID)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, authUser.ID, c.bookingService().GetAllBookings)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetBooking get existing booking by ID for authenticated user
//
//	@Summary		Get booking
//	@Description	Get booking by ID with full details (user can only access their own bookings)
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			booking_id	path		int	true	"booking id"
//	@Success		200			{object}	responses.GetBookingByIDResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		403			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/bookings/{booking_id} [get]
func (c Controller) GetBooking(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetBookingController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetBooking"), attribute.Int("id", id)))
		responseData *responses.GetBookingByIDResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing - include user ID in cache key
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"bookings"}, "GetBooking", authUser.ID, id)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, id, c.bookingService().GetBookingByID)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Verify that the booking belongs to the authenticated user
	if responseData.GuestID == nil || *responseData.GuestID != authUser.ID {
		return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, exception.ErrRecordNotFound)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// ConfirmBookingPayment confirms payment for a booking
//
//	@Summary		Confirm booking payment
//	@Description	Confirm payment for a booking and change status from pending to confirmed
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			booking_id	path		int	true	"booking id"
//	@Success		200			{object}	responses.GetBookingByIDResponse
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		403			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/bookings/{booking_id}/confirm-payment [patch]
func (c Controller) ConfirmBookingPayment(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "ConfirmBookingPaymentController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ConfirmBookingPayment"), attribute.Int("id", id)))
		responseData *responses.GetBookingByIDResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// First, verify the booking exists and belongs to the user
	booking, err := c.bookingService().GetBookingByID(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Verify that the booking belongs to the authenticated user
	if booking.GuestID == nil || *booking.GuestID != authUser.ID {
		return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.RecordNotFoundResponseError, exception.ErrRecordNotFound)
	}

	// Confirm payment
	responseData, err = c.bookingService().ConfirmBookingPayment(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrBookingAlreadyConfirmed:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.BookingAlreadyConfirmedResponseError, err)
		case exception.ErrInvalidBookingStatus:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidBookingStatusResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear booking cache
	c.m.cacher.Tag("bookings").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CancelBooking cancels a booking
//
//	@Summary		Cancel booking
//	@Description	Cancel a booking and return rooms to inventory (only pending or confirmed bookings can be cancelled)
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			booking_id	path		int	true	"booking id"
//	@Success		200			{object}	responses.GetBookingByIDResponse
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		403			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/bookings/{booking_id}/cancel [patch]
func (c Controller) CancelBooking(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "CancelBookingController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CancelBooking"), attribute.Int("id", id)))
		responseData *responses.GetBookingByIDResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// First, verify the booking exists and belongs to the user
	booking, err := c.bookingService().GetBookingByID(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Verify that the booking belongs to the authenticated user
	if booking.GuestID == nil || *booking.GuestID != authUser.ID {
		return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.RecordNotFoundResponseError, exception.ErrRecordNotFound)
	}

	// Cancel booking
	responseData, err = c.bookingService().CancelBooking(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrBookingAlreadyCancelled:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.BookingAlreadyCancelledResponseError, err)
		case exception.ErrCannotCancelBooking:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.CannotCancelBookingResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear booking cache
	c.m.cacher.Tag("bookings").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// SuccessBooking marks a booking as completed
//
//	@Summary		Mark booking as completed
//	@Description	Mark a confirmed booking as successfully completed (only confirmed bookings can be marked as completed)
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			booking_id	path		int	true	"booking id"
//	@Success		200			{object}	responses.GetBookingByIDResponse
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		403			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/bookings/{booking_id}/success [patch]
func (c Controller) SuccessBooking(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "SuccessBookingController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "SuccessBooking"), attribute.Int("id", id)))
		responseData *responses.GetBookingByIDResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Verify the booking exists and belongs to the host's property
	_, err = c.bookingService().GetBookingByIDForHost(ctx, id, authUser.ID)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Mark booking as successful
	responseData, err = c.bookingService().SuccessBooking(ctx, id)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrBookingAlreadySuccessful:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.BookingAlreadySuccessfulResponseError, err)
		case exception.ErrCannotSuccessBooking:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.CannotSuccessBookingResponseError, err)
		case exception.ErrBookingSuccessNotAllowedYet:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.BookingSuccessNotAllowedYetResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear booking cache
	c.m.cacher.Tag("bookings").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetAllBookingsForHostPaginated lists all bookings for properties owned by the authenticated host with pagination
//
//	@Summary		List all bookings for host (paginated)
//	@Description	Get all bookings for properties owned by authenticated host with pagination support
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int		false	"Page number"		default(1)
//	@Param			limit	query		int		false	"Items per page"	default(20)
//	@Param			orderBy	query		string	false	"Order by field"	default(created_at)
//	@Param			sort	query		string	false	"Sort direction (asc/desc)"	default(desc)
//	@Success		200	{object}	paginator.Pagination
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/bookings/paginate [get]
func (c Controller) GetAllBookingsForHostPaginated(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetAllBookingsForHostPaginatedController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllBookingsForHostPaginated")))
		err       error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Parse pagination query params
	pagination := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 20)),
		paginator.WithOrderBy(f.Query("orderBy", "created_at")),
		paginator.WithSort(f.Query("sort", "desc")),
	)

	responseData, err := c.bookingService().GetBookingsByHostIDPaginated(ctx, authUser.ID, pagination)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetAllBookingsForHost lists all bookings for properties owned by the authenticated host
//
//	@Summary		List all bookings for host
//	@Description	Get all bookings for properties owned by authenticated host
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	responses.GetBookingsForHostResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/bookings/all [get]
func (c Controller) GetAllBookingsForHost(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetAllBookingsForHostController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllBookingsForHost")))
		responseData *responses.GetBookingsForHostResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing - include user ID in cache key
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"bookings"}, "GetAllBookingsForHost", authUser.ID)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, authUser.ID, c.bookingService().GetBookingsByHostID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetBookingForHost gets a specific booking by ID for the authenticated host
//
//	@Summary		Get booking for host
//	@Description	Get booking by ID for authenticated host (only bookings for host's properties)
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			booking_id	path		int	true	"booking id"
//	@Success		200			{object}	responses.GetBookingByIDForHostResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		403			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/bookings/{booking_id} [get]
func (c Controller) GetBookingForHost(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetBookingForHostController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetBookingForHost"), attribute.Int("id", id)))
		responseData *responses.GetBookingByIDForHostResponse
		err          error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing - include user ID and booking ID in cache key
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"bookings"}, "GetBookingForHost", authUser.ID, id)

	// Define cache function that passes both bookingID and hostID
	cacheFunc := func(ctx context.Context, bookingID int) (*responses.GetBookingByIDForHostResponse, error) {
		return c.bookingService().GetBookingByIDForHost(ctx, bookingID, authUser.ID)
	}

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, id, cacheFunc)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateBookingReview creates a review for a completed booking
//
//	@Summary		Create booking review
//	@Description	Create a review for a completed booking (rating 1-5, only for successful bookings)
//	@Tags			Booking Module (Version 1)
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int		true	"Booking ID"
//	@Param			rating	formData	int		true	"Rating (1-5)"
//	@Param			comment	formData	string	false	"Review comment (max 500 chars)"
//	@Param			images	formData	file	false	"Review images (multiple allowed)"
//	@Success		200		{object}	responses.CreateBookingReviewResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		409		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/bookings/{id}/review [post]
func (c Controller) CreateBookingReview(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateBookingReviewController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateBookingReview")))
		err       error
		bookingID int
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Get booking ID from URL parameter
	bookingID, err = f.ParamsInt("id")
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Create data transfer object
	dto := new(dtos.CreateBookingReview)

	// Parse multipart form data
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Parse images from multipart form
	if form, formErr := f.MultipartForm(); formErr == nil && form.File["images"] != nil {
		dto.Images = form.File["images"]
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	review, err := c.bookingService().CreateBookingReview(ctx, bookingID, authUser.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrBookingNotSuccessful:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.BookingNotSuccessfulResponseError, err)
		case exception.ErrBookingAlreadyReviewed:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.BookingAlreadyReviewedResponseError, err)
		case exception.ErrUserBannedFromReview:
			return exception.HttpErrorResponseMapping(f, fiber.StatusForbidden, exception.UserBannedFromReviewResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	// Clear booking cache
	c.m.cacher.Tag("bookings").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(review)
}
// GetAllBookingsForPropertyPaginated lists all bookings for a specific property with pagination
//
//	@Summary		List all bookings for property (paginated)
//	@Description	Get all bookings for a specific property owned by authenticated host with pagination support
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			property_id	path		int		true	"property id"
//	@Param			page		query		int		false	"Page number"		default(1)
//	@Param			limit		query		int		false	"Items per page"	default(20)
//	@Param			orderBy		query		string	false	"Order by field"	default(created_at)
//	@Param			sort		query		string	false	"Sort direction (asc/desc)"	default(desc)
//	@Success		200			{object}	paginator.Pagination
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id}/bookings/paginate [get]
func (c Controller) GetAllBookingsForPropertyPaginated(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "GetAllBookingsForPropertyPaginatedController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllBookingsForPropertyPaginated"), attribute.Int("propertyId", propertyID)))
		err           error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Parse pagination query params
	pagination := paginator.NewPagination(
		paginator.WithPage(f.QueryInt("page", 1)),
		paginator.WithLimit(f.QueryInt("limit", 20)),
		paginator.WithOrderBy(f.Query("orderBy", "created_at")),
		paginator.WithSort(f.Query("sort", "desc")),
	)

	responseData, err := c.bookingService().GetBookingsByPropertyIDPaginated(ctx, propertyID, authUser.ID, pagination)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetAllBookingsForProperty lists all bookings for a specific property
//
//	@Summary		List all bookings for property
//	@Description	Get all bookings for a specific property owned by authenticated host
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			property_id	path		int	true	"property id"
//	@Success		200			{object}	responses.GetBookingsForPropertyResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id}/bookings/all [get]
func (c Controller) GetAllBookingsForProperty(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "GetAllBookingsForPropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetAllBookingsForProperty"), attribute.Int("propertyId", propertyID)))
		responseData  *responses.GetBookingsForPropertyResponse
		err           error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing - include user ID and property ID in cache key
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"bookings"}, "GetAllBookingsForProperty", authUser.ID, propertyID)

	// Define cache function that passes both propertyID and hostID
	cacheFunc := func(ctx context.Context, propID int) (*responses.GetBookingsForPropertyResponse, error) {
		return c.bookingService().GetBookingsByPropertyID(ctx, propID, authUser.ID)
	}

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, propertyID, cacheFunc)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetBookingForProperty gets a specific booking by ID for a specific property
//
//	@Summary		Get booking for property
//	@Description	Get booking by ID for a specific property owned by authenticated host
//	@Tags			Booking Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			property_id	path		int	true	"property id"
//	@Param			booking_id	path		int	true	"booking id"
//	@Success		200			{object}	responses.GetBookingByIDForPropertyResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		403			{object}	exception.ErrorResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/hosts/properties/{property_id}/bookings/{booking_id} [get]
func (c Controller) GetBookingForProperty(f *fiber.Ctx) error {
	var (
		propertyID, _ = f.ParamsInt("property_id")
		bookingID, _  = f.ParamsInt("booking_id")
		ctx, span     = c.m.tracer.TraceStart(f.Context(), "GetBookingForPropertyController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetBookingForProperty"), attribute.Int("propertyId", propertyID), attribute.Int("bookingId", bookingID)))
		responseData  *responses.GetBookingByIDForPropertyResponse
		err           error
	)

	// Get authenticated user from context
	authUser, ok := f.Locals("authUser").(*models.User)
	if !ok || authUser == nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
	}

	// Make cache indexing - include user ID, property ID and booking ID in cache key
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"bookings"}, "GetBookingForProperty", authUser.ID, propertyID, bookingID)

	// Define cache function that passes bookingID, propertyID and hostID
	cacheFunc := func(ctx context.Context, bID int) (*responses.GetBookingByIDForPropertyResponse, error) {
		return c.bookingService().GetBookingByIDForProperty(ctx, bID, propertyID, authUser.ID)
	}

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, bookingID, cacheFunc)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
