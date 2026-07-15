package handlers

import (
	"fmt"
	"genrent/internal/apierr"
	"genrent/internal/middleware"
	"genrent/internal/models"
	"genrent/internal/services/email"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// emailDataFromBooking builds an email.EmailData from a preloaded booking.
func emailDataFromBooking(b models.Booking, otp string) email.EmailData {
	genName := ""
	vendorName := ""
	if b.Generator.ID != 0 {
		genName = b.Generator.Name
		if b.Generator.Vendor.ID != 0 {
			vendorName = b.Generator.Vendor.CompanyName
		}
	}
	customerName := ""
	if b.Customer.ID != 0 {
		customerName = b.Customer.Name
	}
	return email.EmailData{
		BookingID:     b.ID,
		GeneratorName: genName,
		VendorName:    vendorName,
		CustomerName:  customerName,
		Amount:    b.AdvanceAmount,
		StartDate: b.StartDate.Format("02 Jan 2006"),
		EndDate:   b.EndDate.Format("02 Jan 2006"),
		OTP:       otp,
	}
}

type CreateBookingRequest struct {
	GeneratorID uint   `json:"generator_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Notes       string `json:"notes"`
}

type ReviewRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Review string `json:"review"`
}

// CreateBooking — customer requests a booking; generator is reserved for 30 min
func CreateBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req CreateBookingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
			return
		}
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
			return
		}
		if !endDate.After(startDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after start_date"})
			return
		}

		// Abuse protection: max 5 active bookings per customer
		var activeBookings int64
		db.Model(&models.Booking{}).Where(
			"customer_id = ? AND status NOT IN ?", userID,
			[]string{string(models.BookingCompleted), string(models.BookingCancelled)},
		).Count(&activeBookings)
		if activeBookings >= 5 {
			apierr.Respond(c, http.StatusTooManyRequests, apierr.ErrBookingLimitReached, "maximum 5 active bookings allowed per customer")
			return
		}

		var booking models.Booking
		var generator models.Generator

		txErr := db.Transaction(func(tx *gorm.DB) error {
			// Lock the generator row to prevent concurrent reservations
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Preload("Vendor").First(&generator, req.GeneratorID).Error; err != nil {
				return fmt.Errorf("generator not found")
			}
			if generator.AvailabilityStatus != models.StatusAvailable {
				return fmt.Errorf("generator is not available for booking")
			}

			days := math.Ceil(endDate.Sub(startDate).Hours() / 24)
			totalPrice := days * generator.PricePerDay
			advanceAmount := math.Round(totalPrice*0.30*100) / 100

			// Lock generator for 30 minutes
			expiry := time.Now().Add(30 * time.Minute)
			if err := tx.Model(&generator).Updates(map[string]interface{}{
				"availability_status": models.StatusReserved,
				"reservation_expiry":  expiry,
			}).Error; err != nil {
				return err
			}

		booking = models.Booking{
			CustomerID:    userID,
			GeneratorID:   &req.GeneratorID,
			StartDate:     startDate,
			EndDate:       endDate,
			TotalPrice:    totalPrice,
			AdvanceAmount: advanceAmount,
			Address:       req.Address,
			Notes:         req.Notes,
			Status:        models.BookingRequested,
			}
			if err := tx.Create(&booking).Error; err != nil {
				return err
			}
			return nil
		})

		if txErr != nil {
			apierr.Conflict(c, apierr.ErrGeneratorUnavailable, txErr.Error())
			return
		}

		db.Model(&models.Vendor{}).Where("id = ?", generator.VendorID).
			UpdateColumn("total_bookings", gorm.Expr("total_bookings + 1"))

		db.Preload("Customer").Preload("Generator").Preload("Generator.Vendor").First(&booking, booking.ID)

		// Notify vendor of new booking request
		createNotif(db, generator.Vendor.UserID, booking.ID,
			models.NotifBookingRequested,
			"New Booking Request",
			fmt.Sprintf("Customer %s has requested your generator '%s'. Accept or reject within 2 hours.",
				booking.Customer.Name, generator.Name))
		sendBookingEmail(db, generator.Vendor.UserID, models.NotifBookingRequested, emailDataFromBooking(booking, ""))

		c.JSON(http.StatusCreated, gin.H{
			"booking":        booking,
			"advance_amount": booking.AdvanceAmount,
			"vendor_amount":  math.Round(booking.TotalPrice*0.15*100) / 100,
			"platform_fee":   math.Round(booking.TotalPrice*0.15*100) / 100,
			"message":        "Booking request sent to vendor. Waiting for vendor to accept.",
		})
	}
}

// VendorAcceptBooking — vendor accepts the request
func VendorAcceptBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		var booking models.Booking
		if result := db.Preload("Customer").Preload("Generator").Preload("Equipment").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		var isOwner bool
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			isOwner = booking.Generator.VendorID == vendor.ID
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			isOwner = booking.Equipment.VendorID == vendor.ID
		}
		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if !booking.CanTransitionTo(models.BookingAccepted) {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("cannot accept booking in '%s' state", booking.Status)})
			return
		}

		now := time.Now()
		// Extend reservation when vendor accepts
		extendedExpiry := now.Add(1 * time.Hour)
		db.Model(&booking).Updates(map[string]interface{}{
			"status":      models.BookingAccepted,
			"accepted_at": now,
		})
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			db.Model(&models.Generator{}).Where("id = ?", booking.GeneratorID).
				Update("reservation_expiry", extendedExpiry)
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			db.Model(&models.Equipment{}).Where("id = ?", booking.EquipmentID).
				Update("reservation_expiry", extendedExpiry)
		}

		auditLog(db, userID, "booking_accepted", "booking", booking.ID, string(models.BookingRequested), string(models.BookingAccepted), c.ClientIP())

		// Track vendor response time (minutes from booking creation to acceptance)
		responseMinutes := time.Since(booking.CreatedAt).Minutes()
		db.Model(&vendor).UpdateColumn("avg_response_minutes",
			gorm.Expr("CASE WHEN avg_response_minutes = 0 THEN ? ELSE (avg_response_minutes * 0.8 + ? * 0.2) END",
				responseMinutes, responseMinutes))

		var name string
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			name = booking.Generator.Name
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			name = booking.Equipment.Name
		}

		createNotif(db, booking.CustomerID, booking.ID,
			models.NotifBookingAccepted,
			"Booking Accepted!",
			fmt.Sprintf("Vendor has accepted your booking for '%s'. Pay the advance of ₹%.0f to confirm.", name, booking.AdvanceAmount))
		sendBookingEmail(db, booking.CustomerID, models.NotifBookingAccepted, emailDataFromBooking(booking, ""))

		c.JSON(http.StatusOK, gin.H{
			"message": "Booking accepted. Customer will now be prompted to pay the advance.",
			"status":  models.BookingAccepted,
		})
	}
}

// VendorRejectBooking — vendor rejects the request
func VendorRejectBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		var body struct {
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		var booking models.Booking
		if result := db.Preload("Generator").Preload("Equipment").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		var isOwner bool
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			isOwner = booking.Generator.VendorID == vendor.ID
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			isOwner = booking.Equipment.VendorID == vendor.ID
		}
		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if !booking.CanTransitionTo(models.BookingCancelled) {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("cannot reject booking in '%s' state", booking.Status)})
			return
		}
		// Only allow rejection of requested bookings
		if booking.Status != models.BookingRequested {
			c.JSON(http.StatusConflict, gin.H{"error": "can only reject a booking in requested state"})
			return
		}

		reason := body.Reason
		if reason == "" {
			reason = "Vendor rejected the request"
		}
		db.Model(&booking).Updates(map[string]interface{}{
			"status":        models.BookingCancelled,
			"cancel_reason": reason,
		})
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			db.Model(&models.Generator{}).Where("id = ?", booking.GeneratorID).
				Updates(map[string]interface{}{
					"availability_status": models.StatusAvailable,
					"reservation_expiry":  nil,
				})
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			db.Model(&models.Equipment{}).Where("id = ?", booking.EquipmentID).
				Updates(map[string]interface{}{
					"availability_status": models.StatusAvailable,
					"reservation_expiry":  nil,
				})
		}
		db.Model(&vendor).UpdateColumn("cancelled_bookings", gorm.Expr("cancelled_bookings + 1"))
		recalcReliabilityScore(db, vendor.ID)

		auditLog(db, userID, "booking_rejected", "booking", booking.ID, string(models.BookingRequested), string(models.BookingCancelled), c.ClientIP())

		// Increment vendor risk score slightly for each rejection (too many = flagged)
		db.Model(&vendor).UpdateColumn("risk_score", gorm.Expr("risk_score + 0.5"))

		createNotif(db, booking.CustomerID, booking.ID,
			models.NotifBookingRejected,
			"Booking Rejected",
			fmt.Sprintf("Your booking was rejected by the vendor. Reason: %s", reason))
		sendBookingEmail(db, booking.CustomerID, models.NotifBookingRejected, func() email.EmailData {
			d := emailDataFromBooking(booking, "")
			d.Message = reason
			return d
		}())

		c.JSON(http.StatusOK, gin.H{"message": "Booking rejected", "status": models.BookingCancelled})
	}
}

// VendorDispatchGenerator — vendor dispatches, OTP generated
func VendorDispatchGenerator(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		var booking models.Booking
		if result := db.Preload("Customer").Preload("Generator").Preload("Equipment").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		var isOwner bool
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			isOwner = booking.Generator.VendorID == vendor.ID
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			isOwner = booking.Equipment.VendorID == vendor.ID
		}
		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if !booking.CanTransitionTo(models.BookingDispatched) {
			apierr.Conflict(c, apierr.ErrBookingInvalidState, fmt.Sprintf("cannot dispatch booking in '%s' state — advance payment must be completed first", booking.Status))
			return
		}

		otp := fmt.Sprintf("%06d", rand.Intn(999999))
		now := time.Now()
		db.Model(&booking).Updates(map[string]interface{}{
			"status":        models.BookingDispatched,
			"delivery_otp":  otp,
			"dispatched_at": now,
		})

		var name string
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			name = booking.Generator.Name
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			name = booking.Equipment.Name
		}

		createNotif(db, booking.CustomerID, booking.ID,
			models.NotifDispatched,
			"Equipment Dispatched!",
			fmt.Sprintf("Your equipment '%s' is on the way! Delivery OTP: %s — share this with the delivery person to confirm arrival.", name, otp))
		sendBookingEmail(db, booking.CustomerID, models.NotifDispatched, emailDataFromBooking(booking, otp))

		c.JSON(http.StatusOK, gin.H{
			"message":      "Generator dispatched. OTP sent to customer for delivery confirmation.",
			"status":       models.BookingDispatched,
			"delivery_otp": otp, // In production: remove from response, send via SMS/email
		})
	}
}

// VendorResendOTP — resends the existing delivery OTP to the customer via email + in-app notification.
// Only the vendor of the booking can call this, and only when booking is in 'dispatched' state.
func VendorResendOTP(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		var booking models.Booking
		if result := db.Preload("Customer").Preload("Generator").Preload("Equipment").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		var isOwner bool
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			isOwner = booking.Generator.VendorID == vendor.ID
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			isOwner = booking.Equipment.VendorID == vendor.ID
		}
		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if booking.Status != models.BookingDispatched {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OTP can only be resent for dispatched bookings"})
			return
		}
		if booking.DeliveryOTP == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no OTP found for this booking"})
			return
		}

		// Re-send via in-app notification
		createNotif(db, booking.CustomerID, booking.ID,
			models.NotifDispatched,
			"Delivery OTP Resent",
			fmt.Sprintf("Your delivery OTP has been resent: %s — use this to confirm equipment arrival.", booking.DeliveryOTP))

		// Re-send via email (same dispatch template with OTP)
		sendBookingEmail(db, booking.CustomerID, models.NotifDispatched, emailDataFromBooking(booking, booking.DeliveryOTP))

		c.JSON(http.StatusOK, gin.H{
			"message": "OTP resent to customer via email and notification.",
		})
	}
}

// CustomerConfirmDelivery — customer enters OTP, escrow released
func CustomerConfirmDelivery(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		var body struct {
			OTP string `json:"otp" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "OTP is required"})
			return
		}

		var booking models.Booking
		if result := db.Preload("Generator").Preload("Equipment").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if !booking.CanTransitionTo(models.BookingDelivered) {
			apierr.Conflict(c, apierr.ErrBookingInvalidState, fmt.Sprintf("cannot confirm delivery for booking in '%s' state", booking.Status))
			return
		}
		if booking.DeliveryOTP != body.OTP {
			apierr.Respond(c, http.StatusUnauthorized, apierr.ErrOTPInvalid, "incorrect OTP — please check and try again")
			return
		}

		now := time.Now()
		db.Model(&booking).Updates(map[string]interface{}{
			"status":       models.BookingDelivered,
			"otp_verified": true,
			"delivered_at": now,
		})

		releaseEscrow(db, booking)

		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			db.Model(&models.Generator{}).Where("id = ?", booking.GeneratorID).
				Updates(map[string]interface{}{
					"availability_status": models.StatusBooked,
					"reservation_expiry":  nil,
				})
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			db.Model(&models.Equipment{}).Where("id = ?", booking.EquipmentID).
				Updates(map[string]interface{}{
					"availability_status": models.StatusBooked,
					"reservation_expiry":  nil,
				})
		}

		// Notify vendor that escrow was released
		var vendorUserID uint
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			var gen models.Generator
			db.Preload("Vendor").First(&gen, booking.GeneratorID)
			vendorUserID = gen.Vendor.UserID
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			var eq models.Equipment
			db.Preload("Vendor").First(&eq, booking.EquipmentID)
			vendorUserID = eq.Vendor.UserID
		}

		if vendorUserID > 0 {
			createNotif(db, vendorUserID, booking.ID,
				models.NotifDelivered,
				"Delivery Confirmed — Payment Released!",
				fmt.Sprintf("Customer confirmed delivery via OTP. ₹%.0f has been released from escrow to your wallet.", math.Round(booking.AdvanceAmount*0.5*100)/100))
			sendBookingEmail(db, vendorUserID, models.NotifDelivered, emailDataFromBooking(booking, ""))
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Delivery confirmed via OTP! Escrow released to vendor.",
			"status":  models.BookingDelivered,
		})
	}
}

// CustomerCompleteBooking — customer marks rental service as complete
func CustomerCompleteBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		var booking models.Booking
		if result := db.First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if !booking.CanTransitionTo(models.BookingCompleted) {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("cannot complete booking in '%s' state — delivery must be confirmed first", booking.Status)})
			return
		}

		now := time.Now()
		db.Model(&booking).Updates(map[string]interface{}{
			"status":       models.BookingCompleted,
			"completed_at": now,
		})
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			db.Model(&models.Generator{}).Where("id = ?", booking.GeneratorID).
				Updates(map[string]interface{}{
					"availability_status": models.StatusAvailable,
					"reservation_expiry":  nil,
				})
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			// Feature 4: restore stock atomically
			var eq models.Equipment
			if db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&eq, booking.EquipmentID).Error == nil {
				newQty := eq.AvailableQuantity + 1
				restoreUpdates := map[string]interface{}{"available_quantity": newQty}
				if eq.AvailabilityStatus == models.StatusBooked {
					restoreUpdates["availability_status"] = models.StatusAvailable
				}
				db.Model(&eq).Updates(restoreUpdates)
			}
		}

		var vendorID uint
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			var gen models.Generator
			if db.First(&gen, booking.GeneratorID).Error == nil {
				vendorID = gen.VendorID
			}
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			var eq models.Equipment
			if db.First(&eq, booking.EquipmentID).Error == nil {
				vendorID = eq.VendorID
			}
		}

		if vendorID > 0 {
			var vendor models.Vendor
			if db.First(&vendor, vendorID).Error == nil {
				db.Model(&vendor).UpdateColumn("successful_deliveries", gorm.Expr("successful_deliveries + 1"))
				recalcReliabilityScore(db, vendor.ID)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Booking completed! Please rate the vendor.", "status": models.BookingCompleted})
	}
}

// CancelBooking — customer or vendor cancels with reason
func CancelBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		role := middleware.GetRole(c)
		id := c.Param("id")

		var body struct {
			Reason string `json:"reason"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var booking models.Booking
		if result := db.Preload("Generator").Preload("Equipment").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}

		if role == models.RoleCustomer && booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if role == models.RoleVendor {
			var vendor models.Vendor
			db.Where("user_id = ?", userID).First(&vendor)
			var isOwner bool
			if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
				isOwner = booking.Generator.VendorID == vendor.ID
			} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
				isOwner = booking.Equipment.VendorID == vendor.ID
			}
			if !isOwner {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
		}

		if !booking.CanTransitionTo(models.BookingCancelled) {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("booking cannot be cancelled at '%s' stage", booking.Status)})
			return
		}

		reason := body.Reason
		if reason == "" {
			reason = "Cancelled by " + string(role)
		}
		db.Model(&booking).Updates(map[string]interface{}{
			"status":        models.BookingCancelled,
			"cancel_reason": reason,
		})

		auditLog(db, userID, "booking_cancelled", "booking", booking.ID, string(booking.Status), string(models.BookingCancelled), c.ClientIP())

		// Increment customer risk score when customer cancels a paid booking
		if booking.AdvancePaid && role == models.RoleCustomer {
			db.Model(&models.User{}).Where("id = ?", userID).
				UpdateColumn("risk_score", gorm.Expr("risk_score + 1"))
		}

		if booking.AdvancePaid {
			refundEscrow(db, booking)
		}

		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			db.Model(&models.Generator{}).Where("id = ?", booking.GeneratorID).
				Updates(map[string]interface{}{
					"availability_status": models.StatusAvailable,
					"reservation_expiry":  nil,
				})
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			// Feature 4: restore stock atomically
			var eq models.Equipment
			if db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&eq, booking.EquipmentID).Error == nil {
				newQty := eq.AvailableQuantity + 1
				restoreUpdates := map[string]interface{}{"available_quantity": newQty}
				if eq.AvailabilityStatus == models.StatusBooked {
					restoreUpdates["availability_status"] = models.StatusAvailable
				}
				db.Model(&eq).Updates(restoreUpdates)
			}
		}

		var vendorUserID uint
		var vendorID uint
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			var gen models.Generator
			if db.Preload("Vendor").First(&gen, booking.GeneratorID).Error == nil {
				vendorUserID = gen.Vendor.UserID
				vendorID = gen.VendorID
			}
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			var eq models.Equipment
			if db.Preload("Vendor").First(&eq, booking.EquipmentID).Error == nil {
				vendorUserID = eq.Vendor.UserID
				vendorID = eq.VendorID
			}
		}

		if role == models.RoleVendor {
			if vendorID > 0 {
				db.Model(&models.Vendor{}).Where("id = ?", vendorID).
					UpdateColumn("cancelled_bookings", gorm.Expr("cancelled_bookings + 1"))
				recalcReliabilityScore(db, vendorID)
			}
			createNotif(db, booking.CustomerID, booking.ID, models.NotifCancelled,
				"Booking Cancelled by Vendor",
				fmt.Sprintf("Your booking was cancelled by the vendor. Reason: %s", reason))
			cd := emailDataFromBooking(booking, "")
			cd.Message = reason
			sendBookingEmail(db, booking.CustomerID, models.NotifCancelled, cd)
		} else {
			// Notify vendor if customer cancels
			if vendorUserID > 0 {
				createNotif(db, vendorUserID, booking.ID, models.NotifCancelled,
					"Booking Cancelled by Customer",
					fmt.Sprintf("Customer cancelled the booking. Reason: %s", reason))
				cd := emailDataFromBooking(booking, "")
				cd.Message = reason
				sendBookingEmail(db, vendorUserID, models.NotifCancelled, cd)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled", "status": models.BookingCancelled})
	}
}

// SubmitReview — customer rates vendor after completion
func SubmitReview(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id := c.Param("id")

		var req ReviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var booking models.Booking
		if result := db.Preload("Generator").Preload("Equipment").First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if booking.Status != models.BookingCompleted {
			c.JSON(http.StatusConflict, gin.H{"error": "can only review completed bookings"})
			return
		}
		if booking.CustomerRating > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "already reviewed"})
			return
		}

		db.Model(&booking).Updates(map[string]interface{}{
			"customer_rating": req.Rating,
			"customer_review": req.Review,
		})

		var vendorID uint
		if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
			var gen models.Generator
			if db.First(&gen, booking.GeneratorID).Error == nil {
				vendorID = gen.VendorID
			}
		} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			var eq models.Equipment
			if db.First(&eq, booking.EquipmentID).Error == nil {
				vendorID = eq.VendorID
			}
		}

		if vendorID > 0 {
			var vendor models.Vendor
			if db.First(&vendor, vendorID).Error == nil {
				newTotal := vendor.TotalRatings + 1
				newAvg := math.Round(((vendor.AverageRating*float64(vendor.TotalRatings))+float64(req.Rating))/float64(newTotal)*100) / 100
				db.Model(&vendor).Updates(map[string]interface{}{
					"total_ratings":  newTotal,
					"average_rating": newAvg,
				})
				recalcReliabilityScore(db, vendor.ID)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Review submitted. Thank you!"})
	}
}

// GetMyBookings — vendor or customer
func GetMyBookings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		role := middleware.GetRole(c)

		var bookings []models.Booking
		if role == models.RoleVendor {
			var vendor models.Vendor
			db.Where("user_id = ?", userID).First(&vendor)
			db.Preload("Customer").Preload("Generator").Preload("Equipment").Preload("Equipment.Category").
				Joins("LEFT JOIN equipment ON equipment.id = bookings.equipment_id").
				Joins("LEFT JOIN generators ON generators.id = bookings.generator_id").
				Where("(equipment.vendor_id = ? OR generators.vendor_id = ?)", vendor.ID, vendor.ID).
				Order("bookings.created_at DESC").Find(&bookings)
		} else {
			db.Preload("Generator").Preload("Generator.Vendor").
				Preload("Equipment").Preload("Equipment.Vendor").Preload("Equipment.Category").
				Where("customer_id = ?", userID).
				Order("created_at DESC").Find(&bookings)
		}

		c.JSON(http.StatusOK, gin.H{"bookings": bookings})
	}
}

// GetBooking — single booking with permission check
func GetBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := middleware.GetUserID(c)
		role := middleware.GetRole(c)

		var booking models.Booking
		if result := db.Preload("Customer").
			Preload("Generator").Preload("Generator.Vendor").
			Preload("Equipment").Preload("Equipment.Vendor").
			First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}

		if role != models.RoleAdmin {
			if role == models.RoleCustomer && booking.CustomerID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
			if role == models.RoleVendor {
				var vendor models.Vendor
				db.Where("user_id = ?", userID).First(&vendor)
				var isOwner bool
				if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
					isOwner = booking.Generator.VendorID == vendor.ID
				} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
					isOwner = booking.Equipment.VendorID == vendor.ID
				}
				if !isOwner {
					c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
					return
				}
			}
		}

		c.JSON(http.StatusOK, booking)
	}
}

// GetBookingStatus — lightweight endpoint for frontend polling (returns only status + key timestamps)
func GetBookingStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := middleware.GetUserID(c)

		var booking models.Booking
		if result := db.Select("id, customer_id, generator_id, status, advance_paid, otp_verified, accepted_at, dispatched_at, delivered_at, completed_at, cancel_reason, updated_at").
			First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		if booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":            booking.ID,
			"status":        booking.Status,
			"advance_paid":  booking.AdvancePaid,
			"otp_verified":  booking.OTPVerified,
			"accepted_at":   booking.AcceptedAt,
			"dispatched_at": booking.DispatchedAt,
			"delivered_at":  booking.DeliveredAt,
			"completed_at":  booking.CompletedAt,
			"cancel_reason": booking.CancelReason,
			"updated_at":    booking.UpdatedAt,
		})
	}
}

// UpdateBookingStatus — admin only
func UpdateBookingStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Status models.BookingStatus `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}
		var booking models.Booking
		if result := db.First(&booking, id); result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
		db.Model(&booking).Update("status", req.Status)
		c.JSON(http.StatusOK, gin.H{"message": "status updated", "status": req.Status})
	}
}

// ---- private helpers ----

func releaseEscrow(db *gorm.DB, booking models.Booking) {
	vendorAmount := math.Round(booking.AdvanceAmount*0.5*100) / 100
	var vendorID uint
	if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
		var gen models.Generator
		if db.First(&gen, booking.GeneratorID).Error == nil {
			vendorID = gen.VendorID
		}
	} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
		var eq models.Equipment
		if db.First(&eq, booking.EquipmentID).Error == nil {
			vendorID = eq.VendorID
		}
	}
	if vendorID == 0 {
		return
	}
	var wallet models.VendorWallet
	if db.Where("vendor_id = ?", vendorID).First(&wallet).Error != nil {
		return
	}
	db.Model(&wallet).Updates(map[string]interface{}{
		"balance":      gorm.Expr("balance + ?", vendorAmount),
		"hold_balance": gorm.Expr("hold_balance - ?", vendorAmount),
	})
	db.Create(&models.WalletTransaction{
		WalletID:    wallet.ID,
		BookingID:   &booking.ID,
		Amount:      vendorAmount,
		Type:        models.WalletEscrowRelease,
		Description: fmt.Sprintf("Escrow released for booking #%d after OTP delivery confirmation", booking.ID),
	})
}

func refundEscrow(db *gorm.DB, booking models.Booking) {
	vendorAmount := math.Round(booking.AdvanceAmount*0.5*100) / 100
	var vendorID uint
	if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
		var gen models.Generator
		if db.First(&gen, booking.GeneratorID).Error == nil {
			vendorID = gen.VendorID
		}
	} else if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
		var eq models.Equipment
		if db.First(&eq, booking.EquipmentID).Error == nil {
			vendorID = eq.VendorID
		}
	}
	if vendorID == 0 {
		return
	}
	var wallet models.VendorWallet
	if db.Where("vendor_id = ?", vendorID).First(&wallet).Error != nil {
		return
	}
	db.Model(&wallet).UpdateColumn("hold_balance", gorm.Expr("hold_balance - ?", vendorAmount))
	db.Create(&models.WalletTransaction{
		WalletID:    wallet.ID,
		BookingID:   &booking.ID,
		Amount:      vendorAmount,
		Type:        models.WalletDebit,
		Description: fmt.Sprintf("Escrow refunded — booking #%d cancelled", booking.ID),
	})

	// Reverse platform revenue — every refund path goes through here
	var origPayment models.Payment
	if db.Where("booking_id = ? AND status = ?", booking.ID, models.PaymentCompleted).
		First(&origPayment).Error == nil {
		db.Create(&models.PlatformRevenue{
			PaymentID:   origPayment.ID,
			BookingID:   booking.ID,
			Amount:      -origPayment.PlatformFee,
			Type:        models.PlatformRefund,
			Description: fmt.Sprintf("Fee reversal — booking #%d refunded", booking.ID),
		})
		db.Model(&origPayment).Update("status", models.PaymentRefunded)
	}
}

func recalcReliabilityScore(db *gorm.DB, vendorID uint) {
	var vendor models.Vendor
	if db.First(&vendor, vendorID).Error != nil {
		return
	}
	if vendor.TotalBookings == 0 {
		return
	}
	deliveryRate := float64(vendor.SuccessfulDeliveries) / float64(vendor.TotalBookings) * 5.0
	var score float64
	if vendor.TotalRatings > 0 {
		score = deliveryRate*0.6 + vendor.AverageRating*0.4
	} else {
		score = deliveryRate
	}
	db.Model(&vendor).UpdateColumn("reliability_score", math.Round(score*100)/100)
}
