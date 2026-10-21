package utils

import (
	"encoding/json"
	"strings"

	"joodpakpaw/internal/exception"
)

// ParseFormDataArray parses a JSON array string from form-data into a slice of integers
// Example input: "[1,2,3,4,5]" -> Output: []int{1,2,3,4,5}
//
// Parameters:
//   - formValue: The string value from form-data (e.g., "[1,2,3]")
//   - fieldName: The name of the field for error reporting (e.g., "facilities")
//
// Returns:
//   - []int: The parsed integer array
//   - exception.ParameterError: Error with field information if parsing fails
func ParseFormDataArray(formValue string, fieldName string) ([]int, *exception.ParameterError) {
	// Return empty array if input is empty
	if formValue == "" {
		return []int{}, nil
	}

	// Remove whitespace
	formValue = strings.TrimSpace(formValue)

	// Validate JSON array format
	if !strings.HasPrefix(formValue, "[") || !strings.HasSuffix(formValue, "]") {
		return nil, &exception.ParameterError{
			FailedField: fieldName,
			Tag:         "invalid_format",
			Value:       "Expected JSON array format like [1,2,3]",
		}
	}

	// Parse JSON array
	var result []int
	if err := json.Unmarshal([]byte(formValue), &result); err != nil {
		return nil, &exception.ParameterError{
			FailedField: fieldName,
			Tag:         "invalid_json",
			Value:       "Failed to parse JSON array: " + err.Error(),
		}
	}

	return result, nil
}
