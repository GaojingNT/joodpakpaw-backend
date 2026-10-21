package validator

import (
	"github.com/go-playground/validator"
	"joodpakpaw/internal/exception"
)

func Validate(data interface{}) []exception.ParameterError {
	var (
		errParams []exception.ParameterError
		validate  = validator.New()
		err       error
	)

	// ลงทะเบียน custom validators สำหรับ password strength
	RegisterPasswordValidators(validate)

	err = validate.Struct(data)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element exception.ParameterError

			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()

			errParams = append(errParams, element)
		}
	}

	return errParams
}
