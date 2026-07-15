package payment

import "fmt"

// PaymentError represents a payment-related error
type PaymentError struct {
	Code    string
	Message string
	Err     error
}

func (e *PaymentError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *PaymentError) Unwrap() error {
	return e.Err
}

// Error codes
const (
	ErrCodeInvalidAmount      = "INVALID_AMOUNT"
	ErrCodeInvalidCurrency    = "INVALID_CURRENCY"
	ErrCodePaymentFailed      = "PAYMENT_FAILED"
	ErrCodePaymentNotFound    = "PAYMENT_NOT_FOUND"
	ErrCodeRefundFailed       = "REFUND_FAILED"
	ErrCodeOrderCreationFailed = "ORDER_CREATION_FAILED"
	ErrCodeVerificationFailed = "VERIFICATION_FAILED"
	ErrCodeWebhookInvalid     = "WEBHOOK_INVALID"
	ErrCodeAPIError           = "API_ERROR"
	ErrCodeTimeout            = "TIMEOUT"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
)

// NewPaymentError creates a new payment error
func NewPaymentError(code, message string, err error) *PaymentError {
	return &PaymentError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Predefined error constructors
func ErrInvalidAmount(amount float64) *PaymentError {
	return NewPaymentError(
		ErrCodeInvalidAmount,
		fmt.Sprintf("Invalid amount: %.2f", amount),
		nil,
	)
}

func ErrInvalidCurrency(currency string) *PaymentError {
	return NewPaymentError(
		ErrCodeInvalidCurrency,
		fmt.Sprintf("Invalid currency: %s", currency),
		nil,
	)
}

func ErrPaymentFailed(paymentID string, err error) *PaymentError {
	return NewPaymentError(
		ErrCodePaymentFailed,
		fmt.Sprintf("Payment failed: %s", paymentID),
		err,
	)
}

func ErrPaymentNotFound(paymentID string) *PaymentError {
	return NewPaymentError(
		ErrCodePaymentNotFound,
		fmt.Sprintf("Payment not found: %s", paymentID),
		nil,
	)
}

func ErrRefundFailed(paymentID string, err error) *PaymentError {
	return NewPaymentError(
		ErrCodeRefundFailed,
		fmt.Sprintf("Refund failed for payment: %s", paymentID),
		err,
	)
}

func ErrOrderCreationFailed(err error) *PaymentError {
	return NewPaymentError(
		ErrCodeOrderCreationFailed,
		"Failed to create payment order",
		err,
	)
}

func ErrVerificationFailed(paymentID string, err error) *PaymentError {
	return NewPaymentError(
		ErrCodeVerificationFailed,
		fmt.Sprintf("Payment verification failed: %s", paymentID),
		err,
	)
}

func ErrWebhookInvalid(reason string) *PaymentError {
	return NewPaymentError(
		ErrCodeWebhookInvalid,
		fmt.Sprintf("Invalid webhook: %s", reason),
		nil,
	)
}

func ErrAPIError(operation string, err error) *PaymentError {
	return NewPaymentError(
		ErrCodeAPIError,
		fmt.Sprintf("API error during %s", operation),
		err,
	)
}

// IsPaymentError checks if an error is a PaymentError
func IsPaymentError(err error) bool {
	_, ok := err.(*PaymentError)
	return ok
}

// GetPaymentErrorCode extracts the error code from a PaymentError
func GetPaymentErrorCode(err error) string {
	if paymentErr, ok := err.(*PaymentError); ok {
		return paymentErr.Code
	}
	return ""
}
