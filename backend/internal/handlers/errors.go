package handlers

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// fieldLabels maps JSON field names to human-readable labels.
var fieldLabels = map[string]string{
	// auth
	"name":     "Full name",
	"email":    "Email address",
	"phone":    "Phone number",
	"password": "Password",
	"role":     "Role",

	// vendor
	"company_name": "Company name",
	"city":         "City",
	"address":      "Address",

	// generator
	"capacity_kva":    "Generator capacity (KVA)",
	"price_per_day":   "Price per day",
	"price_per_month": "Price per month",
	"fuel_type":       "Fuel type",
	"location":        "Location",
	"image_url":       "Image URL",

	// booking
	"generator_id": "Generator",
	"start_date":   "Start date",
	"end_date":     "End date",
	"total_price":  "Total price",

	// payment
	"booking_id": "Booking",
	"amount":     "Amount",
	"method":     "Payment method",

	// misc
	"otp":    "OTP",
	"rating": "Rating",
	"review": "Review",
}

// tagMessages maps validator tag names to message templates.
// Use {field} as a placeholder for the field label.
var tagMessages = map[string]string{
	"required": "{field} is required",
	"email":    "{field} must be a valid email address",
	"min":      "{field} is too short (minimum {param} characters)",
	"max":      "{field} is too long (maximum {param} characters)",
	"gte":      "{field} must be at least {param}",
	"lte":      "{field} must be at most {param}",
	"gt":       "{field} must be greater than {param}",
	"lt":       "{field} must be less than {param}",
	"len":      "{field} must be exactly {param} characters",
	"oneof":    "{field} must be one of: {param}",
	"url":      "{field} must be a valid URL",
	"numeric":  "{field} must be a number",
	"alpha":    "{field} must contain only letters",
	"alphanum": "{field} must contain only letters and numbers",
}

// ValidationErrors converts a Go validator error into a list of
// human-readable error strings. Falls back gracefully for non-validation errors.
func ValidationErrors(err error) []string {
	var errs validator.ValidationErrors
	if !isValidationError(err, &errs) {
		return []string{err.Error()}
	}

	messages := make([]string, 0, len(errs))
	for _, fe := range errs {
		messages = append(messages, formatFieldError(fe))
	}
	return messages
}

// ValidationError returns a single human-readable string for the first error.
func ValidationError(err error) string {
	msgs := ValidationErrors(err)
	if len(msgs) == 0 {
		return "invalid request"
	}
	return msgs[0]
}

func isValidationError(err error, target *validator.ValidationErrors) bool {
	if errs, ok := err.(validator.ValidationErrors); ok {
		*target = errs
		return true
	}
	return false
}

func formatFieldError(fe validator.FieldError) string {
	// Derive the human-readable field name from the JSON tag (namespace uses struct field name)
	field := jsonFieldName(fe)
	label, ok := fieldLabels[field]
	if !ok {
		// Convert "CapacityKVA" → "Capacity KVA" as fallback
		label = splitCamelCase(fe.Field())
	}

	tmpl, ok := tagMessages[fe.Tag()]
	if !ok {
		return fmt.Sprintf("%s is invalid", label)
	}

	msg := strings.ReplaceAll(tmpl, "{field}", label)
	msg = strings.ReplaceAll(msg, "{param}", fe.Param())
	return msg
}

// jsonFieldName extracts the JSON key from the struct field's namespace.
// e.g. "CreateGeneratorRequest.CapacityKVA" → looks up json tag → "capacity_kva"
func jsonFieldName(fe validator.FieldError) string {
	// fe.Field() returns the struct field name; fe.Namespace() has the full path.
	// We use the lowercased field as a fallback key into fieldLabels.
	return strings.ToLower(fe.Field())
}

// splitCamelCase converts "CapacityKVA" → "Capacity KVA"
func splitCamelCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, ' ')
		}
		result = append(result, r)
	}
	return string(result)
}
