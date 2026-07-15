// Package apierr provides standardised API error codes and a response helper.
// Every error response has the shape:
//
//	{"error_code": "BOOKING_INVALID_STATE", "message": "human-readable explanation"}
//
// Frontend code should switch on error_code for programmatic handling,
// and display message to the user.
package apierr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ---- Error codes ----

const (
	// Auth
	ErrUnauthorized     = "UNAUTHORIZED"
	ErrForbidden        = "FORBIDDEN"
	ErrInvalidToken     = "INVALID_TOKEN"
	ErrTokenExpired     = "TOKEN_EXPIRED"

	// Validation
	ErrBadRequest       = "BAD_REQUEST"
	ErrMissingField     = "MISSING_FIELD"
	ErrInvalidFormat    = "INVALID_FORMAT"
	ErrRequestTooLarge  = "REQUEST_TOO_LARGE"

	// Resources
	ErrNotFound         = "NOT_FOUND"
	ErrConflict         = "CONFLICT"

	// Booking
	ErrBookingInvalidState  = "BOOKING_INVALID_STATE"
	ErrBookingAlreadyPaid   = "BOOKING_ALREADY_PAID"
	ErrBookingNotAvailable  = "BOOKING_NOT_AVAILABLE"
	ErrBookingLimitReached  = "BOOKING_LIMIT_REACHED"
	ErrOTPInvalid           = "OTP_INVALID"

	// Generator
	ErrGeneratorUnavailable  = "GENERATOR_UNAVAILABLE"
	ErrGeneratorLimitReached = "GENERATOR_LIMIT_REACHED"

	// Payment
	ErrPaymentFailed        = "PAYMENT_FAILED"
	ErrPaymentAlreadyDone   = "PAYMENT_ALREADY_DONE"

	// Server
	ErrInternal             = "INTERNAL_ERROR"
	ErrDatabaseError        = "DATABASE_ERROR"
	ErrRateLimited          = "RATE_LIMITED"
)

// APIError is the canonical error response body
type APIError struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

// Respond writes a standardised JSON error and aborts the request.
func Respond(c *gin.Context, status int, code string, message string) {
	c.AbortWithStatusJSON(status, APIError{
		ErrorCode: code,
		Message:   message,
	})
}

// Convenience shortcuts

func BadRequest(c *gin.Context, code string, message string) {
	Respond(c, http.StatusBadRequest, code, message)
}

func NotFound(c *gin.Context, entity string) {
	Respond(c, http.StatusNotFound, ErrNotFound, entity+" not found")
}

func Forbidden(c *gin.Context) {
	Respond(c, http.StatusForbidden, ErrForbidden, "access denied")
}

func Conflict(c *gin.Context, code string, message string) {
	Respond(c, http.StatusConflict, code, message)
}

func Internal(c *gin.Context) {
	Respond(c, http.StatusInternalServerError, ErrInternal, "internal server error")
}

func TooManyRequests(c *gin.Context, message string) {
	Respond(c, http.StatusTooManyRequests, ErrRateLimited, message)
}
