package exception

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"joodpakpaw/config"
)

func HttpErrorResponseMapping(c *fiber.Ctx, code int, errResponse *ErrorResponse, err error, errParams ...ParameterError) error {
	var (
		isResponseOK        bool
		fiberResponseStatus int
		fiberResponseBody   *ErrorResponse
	)

	// Error mapping
	switch err {
	// Database errors
	case ErrDbQueryStatement:
		errorMessage := &ErrorResponse{
			Code:    DbQueryStatementResponseError.Code,
			Message: fmt.Sprintf("%s, %s", DbQueryStatementResponseError.Message, SqlErrorMessage),
		}
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = errorMessage
	// Auth errors
	case ErrInvalidLoginCredential:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusUnauthorized
		fiberResponseBody = InvalidLoginCredentialResponseError
	case ErrUserAlreadyExists:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusConflict
		fiberResponseBody = UserAlreadyExistsResponseError
	case ErrPhoneAlreadyExists:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusConflict
		fiberResponseBody = PhoneAlreadyExistsResponseError
	case ErrInvalidOTP:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = InvalidOTPResponseError
	case ErrOTPExpired:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusUnauthorized
		fiberResponseBody = OTPExpiredResponseError
	case ErrEmailAlreadyVerified:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusConflict
		fiberResponseBody = EmailAlreadyVerifiedResponseError
	case ErrAccountNotVerified:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusForbidden
		fiberResponseBody = AccountNotVerifiedResponseError
	case ErrResetPasswordTokenExpired:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusUnauthorized
		fiberResponseBody = ResetPasswordTokenExpiredResponseError
	case ErrResetPasswordTokenUsed:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusUnauthorized
		fiberResponseBody = ResetPasswordTokenUsedResponseError
	case ErrPasswordNotMatch:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = PasswordNotMatchResponseError
	case ErrNewPasswordSameAsOld:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = NewPasswordSameAsOldResponseError
	case ErrTokenMissing:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusUnauthorized
		fiberResponseBody = TokenMissingResponseError
	case ErrInvalidAuthorizationToken:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusUnauthorized
		fiberResponseBody = InvalidAuthorizationTokenResponseError
	case ErrExpiredAuthorizationToken:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusUnauthorized
		fiberResponseBody = ExpiredAuthorizationTokenResponseError
	case ErrInvalidFileFormat:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = InvalidFileFormatResponseError
	case ErrAvatarFileSizeTooLarge:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = AvatarFileSizeTooLargeResponseError
	case ErrInvalidCharactersInName:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = InvalidCharactersInNameResponseError
	// Property errors
	case ErrPropertyRequiresRoomCategory:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = PropertyRequiresRoomCategoryResponseError
	// Favorite Folder errors
	case ErrPropertyAlreadyInFolder:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusConflict
		fiberResponseBody = PropertyAlreadyInFolderResponseError
	case ErrPropertyAlreadyInSameFolder:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusConflict
		fiberResponseBody = PropertyAlreadyInSameFolderResponseError
	// Booking errors
	case ErrCalendarDateNotAvailable:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = CalendarDateNotAvailableResponseError
	case ErrBookingNotSuccessful:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = BookingNotSuccessfulResponseError
	case ErrBookingAlreadyReviewed:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusConflict
		fiberResponseBody = BookingAlreadyReviewedResponseError
	// Verification errors
	case ErrVerificationAlreadyPending:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusConflict
		fiberResponseBody = VerificationAlreadyPendingResponseError
	case ErrVerificationAlreadyApproved:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusConflict
		fiberResponseBody = VerificationAlreadyApprovedResponseError
	case ErrVerificationNotPending:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = VerificationNotPendingResponseError
	case ErrVerificationFileSizeTooLarge:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = VerificationFileSizeTooLargeResponseError
	// Application errors
	case ErrInvalidRequestParameter:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		InvalidRequestParameterResponseError.Errors = errParams
		fiberResponseBody = InvalidRequestParameterResponseError
	default:
		// Default error response
		if errResponse != nil {
			isResponseOK = true
			fiberResponseStatus = code
			fiberResponseBody = errResponse
		} else {
			isResponseOK = false
			fiberResponseStatus = fiber.StatusInternalServerError
			fiberResponseBody = &ErrorResponse{
				Code:    "ER",
				Message: err.Error(),
			}
		}
	}

	// For development
	if !config.IsProduction {
		if isResponseOK {
			if fiberResponseBody != nil {
				return c.Status(fiberResponseStatus).JSON(fiberResponseBody)
			} else {
				return c.SendStatus(fiberResponseStatus)
			}
		}

		// Out of mapping
		return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Code: "ER", Message: err.Error()})
	}

	// -------------------------------------------------------------------------------------------

	// WARNING! this is for production HTTP response
	// Note: Hide response message for prevent vulnerability attack
	if isResponseOK {
		if fiberResponseBody != nil {
			fiberResponseBody.Message = ""
			return c.Status(fiberResponseStatus).JSON(fiberResponseBody)
		} else {
			return c.SendStatus(fiberResponseStatus)
		}
	}

	// Out of mapping
	return fiber.ErrInternalServerError
}
