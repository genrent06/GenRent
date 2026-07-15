package models

import "time"

type NotificationType string

const (
	NotifBookingRequested NotificationType = "booking_requested"
	NotifBookingAccepted  NotificationType = "booking_accepted"
	NotifBookingRejected  NotificationType = "booking_rejected"
	NotifAdvancePaid      NotificationType = "advance_paid"
	NotifDispatched       NotificationType = "dispatched"
	NotifDelivered        NotificationType = "delivered"
	NotifCompleted        NotificationType = "completed"
	NotifCancelled        NotificationType = "booking_cancelled"
	NotifVendorPenalized    NotificationType = "vendor_penalized"
	NotifWithdrawalOTPSent  NotificationType = "withdrawal_otp_sent"
	NotifWithdrawalPending  NotificationType = "withdrawal_pending"
	NotifWithdrawalApproved NotificationType = "withdrawal_approved"
	NotifWithdrawalRejected NotificationType = "withdrawal_rejected"
	// Feature 5
	NotifHandoverUploaded NotificationType = "handover_uploaded"
	NotifReturnOTP        NotificationType = "return_otp"
	NotifDisputeRaised    NotificationType = "dispute_raised"
	NotifDisputeResolved  NotificationType = "dispute_resolved"
)

type Notification struct {
	ID        uint             `json:"id" gorm:"primaryKey"`
	UserID    uint             `json:"user_id" gorm:"not null;index"`
	BookingID *uint            `json:"booking_id"`
	Type      NotificationType `json:"type" gorm:"type:varchar(30);index"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Read      bool             `json:"read" gorm:"default:false;index"`
	CreatedAt time.Time        `json:"created_at"`
}
