package dtos

import "mime/multipart"

type (
	SubmitVerification struct {
		IdPhotoFile      *multipart.FileHeader `json:"-" form:"-" validate:"-"`
		SelfieWithIdFile *multipart.FileHeader `json:"-" form:"-" validate:"-"`
	}
	RejectVerification struct {
		RejectDescription string `json:"rejectDescription" form:"rejectDescription" validate:"required,max=500"`
	}
)
