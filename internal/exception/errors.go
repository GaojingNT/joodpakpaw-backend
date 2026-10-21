package exception

import "github.com/rotisserie/eris"

// ERROR
var (
	// Sql error message
	SqlErrorMessage string
	// Server errors
	ErrInternalServerError            = eris.New("internal server error")
	ErrUnauthorized                   = eris.New("unauthorized")
	ErrForbidden                      = eris.New("forbidden")
	ErrNotFound                       = eris.New("not found")
	ErrTypeConversionFailed           = eris.New("type conversion failed")
	ErrInvalidRequestBody             = eris.New("invalid request body")
	ErrFileTypeNotSupported           = eris.New("file type not supported")
	ErrFileSizeLimitExceeded          = eris.New("file size limit exceeded")
	ErrInvalidFileFormat              = eris.New("invalid file format for avatar")
	ErrAvatarFileSizeTooLarge         = eris.New("avatar file size too large")
	ErrInvalidCharactersInName        = eris.New("invalid characters in name")
	ErrRequestContentTypeNotSupported = eris.New("request content type not supported")
	// Database errors
	ErrDbQueryStatement = eris.New("database query statement error")
	ErrRecordNotFound   = eris.New("record not found")
	// Auth errors
	ErrInvalidLoginCredential      = eris.New("invalid login credential")
	ErrInvalidResetPasswordToken   = eris.New("invalid reset password token")
	ErrResetPasswordTokenExpired   = eris.New("reset password token expired")
	ErrResetPasswordTokenUsed      = eris.New("reset password token already used")
	ErrPasswordNotMatch            = eris.New("password not match")
	ErrNewPasswordSameAsOld        = eris.New("new password same as old password")
	ErrUserAlreadyExists           = eris.New("email already exists")
	ErrPhoneAlreadyExists          = eris.New("phone number already exists")
	ErrInvalidOTP                  = eris.New("invalid OTP code")
	ErrOTPExpired                  = eris.New("OTP expired")
	ErrEmailAlreadyVerified        = eris.New("email already verified")
	ErrAccountNotVerified          = eris.New("account not verified")
	ErrInvalidGoogleToken          = eris.New("invalid google id token")
	ErrGoogleSignInNotConfigured   = eris.New("google sign-in is not configured")
	// token errors
	ErrInvalidToken              = eris.New("invalid token")
	ErrTokenMissing              = eris.New("authorization token is required")
	ErrInvalidAuthorizationToken = eris.New("invalid authorization token")
	ErrExpiredAuthorizationToken = eris.New("expired authorization token")
	ErrInvalidRequestParameter   = eris.New("invalid request parameter")
	ErrInvalidUserData           = eris.New("invalid user data")
	ErrInvalidTokenClaim         = eris.New("invalid token claim")
	ErrInvalidTokenClaimData     = eris.New("invalid token claim data")
	ErrTokenExpired              = eris.New("token expired")
	// Property errors
	ErrPropertyRequiresRoomCategory = eris.New("property requires at least one room category to be active")
	// Favorite Folder errors
	ErrPropertyAlreadyInFolder     = eris.New("property already exists in another folder")
	ErrPropertyAlreadyInSameFolder = eris.New("property already exists in this folder")
	// Room Category errors
	ErrMaxRoomCategoriesExceeded = eris.New("maximum room categories limit exceeded")
	// Booking errors
	ErrInsufficientRooms       = eris.New("insufficient rooms available")
	ErrInsufficientPetCapacity = eris.New("insufficient pet capacity for the number of pets")
	ErrInvalidDateFormat        = eris.New("invalid date format")
	ErrFutureBirthDate          = eris.New("birth date cannot be in the future")
	ErrInvalidDateRange         = eris.New("check-out date must be after check-in date")
	ErrCalendarDateNotAvailable = eris.New("calendar date not available for booking")
	ErrInvalidBookingStatus     = eris.New("invalid booking status for this operation")
	ErrBookingAlreadyConfirmed  = eris.New("booking payment already confirmed")
	ErrBookingAlreadyCancelled  = eris.New("booking already cancelled")
	ErrBookingAlreadySuccessful = eris.New("booking already completed")
	ErrCannotCancelBooking      = eris.New("cannot cancel booking in current status")
	ErrCannotSuccessBooking     = eris.New("cannot complete booking, must be confirmed first")
	ErrBookingNotSuccessful     = eris.New("can only review completed checkouts")
	ErrBookingAlreadyReviewed        = eris.New("booking already has a review")
	ErrUserBannedFromReview          = eris.New("you are banned from writing reviews")
	ErrBookingSuccessNotAllowedYet = eris.New("cannot complete booking before 12:00 on checkout day")
	ErrCannotBookOwnProperty       = eris.New("cannot book your own property")
	ErrCannotReportOwnProperty     = eris.New("cannot report your own property")
	ErrCannotReportOwnReview       = eris.New("cannot report your own review")
	ErrInvalidPetProfile           = eris.New("one or more pet profiles are invalid or do not belong to the guest")
	ErrNumberOfPetsConflict        = eris.New("cannot provide both numberOfPets and petProfileIds")
	ErrPetSizeNotCompatible        = eris.New("pet size is not compatible with the selected room category")
	// Property errors
	ErrHostBannedAllProperties = eris.New("you are banned from creating properties")
	// Role errors
	ErrRoleAlreadyAssigned = eris.New("role already assigned to user")
	// Report action errors
	ErrReportActionNotAllowed        = eris.New("action not allowed for this report type")
	ErrReportAlreadyActioned         = eris.New("report has already been actioned")
	ErrUnbanRequestAlreadyPending    = eris.New("an unban request is already pending, please wait for it to be reviewed")
	ErrUnbanRequestLimitExceeded     = eris.New("unban request limit exceeded, maximum 2 requests allowed")
	ErrMustUnbanAllProperties        = eris.New("you are banned from all properties, please submit an unban all properties request instead")
	// Verification errors
	ErrVerificationAlreadyPending   = eris.New("verification already pending")
	ErrVerificationAlreadyApproved  = eris.New("verification already approved")
	ErrVerificationNotPending       = eris.New("verification is not in pending status")
	ErrVerificationFileSizeTooLarge = eris.New("verification file size too large")
)

// MESSAGE RESPONSE
var (
	// Server error response messages
	NotFoundResponseError              = &ErrorResponse{Code: "S-404", Message: "not found"}
	ErrorResponseInternalServerError   = &ErrorResponse{Code: "S-500", Message: "internal server error"}
	InvalidFileFormatResponseError     = &ErrorResponse{Code: "S-400", Message: "invalid file format for avatar"}
	AvatarFileSizeTooLargeResponseError = &ErrorResponse{Code: "S-413", Message: "avatar file size too large, maximum 10MB"}
	InvalidCharactersInNameResponseError = &ErrorResponse{Code: "S-400-2", Message: "invalid characters in name"}
	// Database error response messages
	DbQueryStatementResponseError = &ErrorResponse{Code: "DB-1001", Message: "database query statement error"}
	RecordNotFoundResponseError   = &ErrorResponse{Code: "DB-1002", Message: "record not found"}
	// Auth error response messages
	InvalidLoginCredentialResponseError      = &ErrorResponse{Code: "A-3001", Message: "invalid username or password"}
	InvalidResetPasswordTokenResponseError   = &ErrorResponse{Code: "A-3002", Message: "invalid reset password token"}
	UserAlreadyExistsResponseError           = &ErrorResponse{Code: "A-3003", Message: "email already exists"}
	PhoneAlreadyExistsResponseError          = &ErrorResponse{Code: "A-3004", Message: "phone number already exists"}
	InvalidOTPResponseError                  = &ErrorResponse{Code: "A-3005", Message: "invalid OTP code"}
	OTPExpiredResponseError                  = &ErrorResponse{Code: "A-3006", Message: "OTP expired"}
	EmailAlreadyVerifiedResponseError        = &ErrorResponse{Code: "A-3007", Message: "email already verified"}
	AccountNotVerifiedResponseError          = &ErrorResponse{Code: "A-3008", Message: "account not verified"}
	ResetPasswordTokenExpiredResponseError   = &ErrorResponse{Code: "A-3009", Message: "reset password token expired"}
	ResetPasswordTokenUsedResponseError      = &ErrorResponse{Code: "A-3010", Message: "reset password token already used"}
	PasswordNotMatchResponseError            = &ErrorResponse{Code: "A-3011", Message: "password not match"}
	NewPasswordSameAsOldResponseError        = &ErrorResponse{Code: "A-3012", Message: "new password same as old password"}
	InvalidGoogleTokenResponseError          = &ErrorResponse{Code: "A-3013", Message: "invalid google sign-in token"}
	GoogleSignInNotConfiguredResponseError   = &ErrorResponse{Code: "A-3014", Message: "google sign-in is not configured"}
	// token error response messages
	UnauthorizedResponseError              = &ErrorResponse{Code: "T-1001", Message: "unauthorized"}
	ForbiddenResponseError                 = &ErrorResponse{Code: "T-1002", Message: "forbidden"}
	InvalidTokenResponseError              = &ErrorResponse{Code: "T-2003", Message: "invalid token"}
	TokenMissingResponseError              = &ErrorResponse{Code: "T-2004", Message: "authorization token is required"}
	InvalidAuthorizationTokenResponseError = &ErrorResponse{Code: "T-2005", Message: "invalid authorization token"}
	ExpiredAuthorizationTokenResponseError = &ErrorResponse{Code: "T-2006", Message: "expired authorization token"}
	InvalidRequestParameterResponseError   = &ErrorResponse{Code: "T-2007", Message: "invalid request parameter"}
	InvalidUserDataResponseError           = &ErrorResponse{Code: "T-2008", Message: "invalid user data"}
	InvalidTokenClaimResponseError         = &ErrorResponse{Code: "T-2009", Message: "invalid token claim"}
	InvalidTokenClaimDataResponseError     = &ErrorResponse{Code: "T-2010", Message: "invalid token claim data"}
	TokenExpiredResponseError              = &ErrorResponse{Code: "T-2011", Message: "token expired"}
	// Property error response messages
	PropertyRequiresRoomCategoryResponseError = &ErrorResponse{Code: "P-4001", Message: "property requires at least one room category to be active"}
	HostBannedAllPropertiesResponseError      = &ErrorResponse{Code: "P-4002", Message: "you are banned from creating properties"}
	// Favorite Folder error response messages
	PropertyAlreadyInFolderResponseError     = &ErrorResponse{Code: "FF-4101", Message: "this property is already in another favorite folder"}
	PropertyAlreadyInSameFolderResponseError = &ErrorResponse{Code: "FF-4102", Message: "this property already exists in this folder"}
	// Room Category error response messages
	MaxRoomCategoriesExceededResponseError = &ErrorResponse{Code: "RC-4501", Message: "cannot create more than 10 room categories per property"}
	// Booking error response messages
	InsufficientRoomsResponseError       = &ErrorResponse{Code: "B-5001", Message: "insufficient rooms available for the selected dates"}
	InsufficientPetCapacityResponseError = &ErrorResponse{Code: "B-5002", Message: "insufficient pet capacity, please book more rooms"}
	InvalidDateFormatResponseError        = &ErrorResponse{Code: "B-5003", Message: "invalid date format, expected YYYY-MM-DD"}
	FutureBirthDateResponseError          = &ErrorResponse{Code: "B-5013", Message: "birth date cannot be in the future"}
	InvalidDateRangeResponseError         = &ErrorResponse{Code: "B-5004", Message: "check-out date must be after check-in date"}
	CalendarDateNotAvailableResponseError = &ErrorResponse{Code: "B-5005", Message: "selected date is not available for booking"}
	InvalidBookingStatusResponseError     = &ErrorResponse{Code: "B-5006", Message: "booking status must be pending to confirm payment"}
	BookingAlreadyConfirmedResponseError  = &ErrorResponse{Code: "B-5007", Message: "booking payment already confirmed"}
	BookingAlreadyCancelledResponseError  = &ErrorResponse{Code: "B-5008", Message: "booking already cancelled"}
	BookingAlreadySuccessfulResponseError = &ErrorResponse{Code: "B-5009", Message: "booking already completed"}
	CannotCancelBookingResponseError      = &ErrorResponse{Code: "B-5010", Message: "cannot cancel booking, only pending or confirmed bookings can be cancelled"}
	CannotSuccessBookingResponseError     = &ErrorResponse{Code: "B-5011", Message: "cannot complete booking, booking must be confirmed first"}
	BookingNotSuccessfulResponseError     = &ErrorResponse{Code: "B-5012", Message: "can only review completed checkouts"}
	BookingAlreadyReviewedResponseError        = &ErrorResponse{Code: "B-5013", Message: "this booking already has a review"}
	UserBannedFromReviewResponseError          = &ErrorResponse{Code: "B-5019", Message: "you are banned from writing reviews"}
	BookingSuccessNotAllowedYetResponseError = &ErrorResponse{Code: "B-5014", Message: "cannot complete booking before 12:00 on checkout day"}
	CannotBookOwnPropertyResponseError       = &ErrorResponse{Code: "B-5015", Message: "cannot book your own property"}
	CannotReportOwnPropertyResponseError     = &ErrorResponse{Code: "P-4003", Message: "cannot report your own property"}
	CannotReportOwnReviewResponseError       = &ErrorResponse{Code: "P-4004", Message: "cannot report your own review"}
	InvalidPetProfileResponseError           = &ErrorResponse{Code: "B-5016", Message: "one or more pet profiles are invalid or do not belong to the guest"}
	NumberOfPetsConflictResponseError        = &ErrorResponse{Code: "B-5017", Message: "cannot provide both numberOfPets and petProfileIds"}
	PetSizeNotCompatibleResponseError        = &ErrorResponse{Code: "B-5018", Message: "pet size is not compatible with the selected room category"}
	// Role response messages
	RoleAlreadyAssignedResponseError = &ErrorResponse{Code: "R-7001", Message: "role already assigned to user"}
	// Report action response messages
	ReportActionNotAllowedResponseError   = &ErrorResponse{Code: "RA-8001", Message: "action not allowed for this report type"}
	ReportAlreadyActionedResponseError    = &ErrorResponse{Code: "RA-8002", Message: "report has already been actioned"}
	UnbanRequestAlreadyPendingResponseError  = &ErrorResponse{Code: "RA-8003", Message: "an unban request is already pending, please wait for it to be reviewed"}
	UnbanRequestLimitExceededResponseError   = &ErrorResponse{Code: "RA-8004", Message: "unban request limit exceeded, maximum 2 requests allowed"}
	MustUnbanAllPropertiesResponseError      = &ErrorResponse{Code: "RA-8005", Message: "you are banned from all properties, please submit an unban all properties request instead"}
	// Verification response messages
	VerificationAlreadyPendingResponseError   = &ErrorResponse{Code: "V-6001", Message: "verification request is already pending"}
	VerificationAlreadyApprovedResponseError  = &ErrorResponse{Code: "V-6002", Message: "verification is already approved"}
	VerificationNotPendingResponseError       = &ErrorResponse{Code: "V-6003", Message: "verification is not in pending status"}
	VerificationFileSizeTooLargeResponseError = &ErrorResponse{Code: "V-6004", Message: "verification file size too large, maximum 10MB"}
)
