package handlers

import (
	"fmt"
	"genrent/internal/middleware"
	"genrent/internal/models"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─── REQUEST TYPES ───────────────────────────────────────────────────────────

type HandoverUploadRequest struct {
	PhotoURLs []string               `json:"photo_urls" binding:"required,min=1"`
	Checklist map[string]interface{} `json:"checklist"`
	Notes     string                 `json:"notes"`
}

type DisputeRequest struct {
	Description   string   `json:"description" binding:"required,min=10"`
	ClaimedAmount float64  `json:"claimed_amount" binding:"required,min=0"`
	PhotoURLs     []string `json:"photo_urls"`
}

type ResolveDisputeRequest struct {
	AdminNotes string `json:"admin_notes" binding:"required"`
	Status     string `json:"status" binding:"required,oneof=resolved rejected"`
}

// ─── HELPERS ─────────────────────────────────────────────────────────────────

func generateReturnOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func getEquipmentVendorUserID(db *gorm.DB, booking models.Booking) uint {
	if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
		var eq models.Equipment
		if db.Preload("Vendor").First(&eq, booking.EquipmentID).Error == nil {
			return eq.Vendor.UserID
		}
	}
	if booking.GeneratorID != nil && *booking.GeneratorID != 0 {
		var gen models.Generator
		if db.Preload("Vendor").First(&gen, booking.GeneratorID).Error == nil {
			return gen.Vendor.UserID
		}
	}
	return 0
}

// ─── VENDOR: Upload Handover Photos ──────────────────────────────────────────

// UploadHandover — vendor uploads delivery or return photos + checklist
// POST /api/v1/bookings/:id/handover  (body: { type: "delivery"|"return", photo_urls, checklist })
func UploadHandover(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		bookingID := c.Param("id")

		handoverType := c.Query("type")
		if handoverType != "delivery" && handoverType != "return" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type query param must be 'delivery' or 'return'"})
			return
		}

		var req HandoverUploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var booking models.Booking
		if err := db.Preload("Equipment").Preload("Equipment.Vendor").
			Preload("Generator").Preload("Generator.Vendor").
			First(&booking, bookingID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}

		// Verify caller is the vendor of this booking
		vendorUserID := getEquipmentVendorUserID(db, booking)
		if vendorUserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the booking vendor can upload handover photos"})
			return
		}

		// For delivery handover: booking must be dispatched
		if handoverType == "delivery" && booking.Status != models.BookingDispatched {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("delivery handover can only be uploaded when booking is dispatched, current status: %s", booking.Status)})
			return
		}
		// For return handover: booking must be delivered
		if handoverType == "return" && booking.Status != models.BookingDelivered {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("return handover can only be uploaded when booking is delivered, current status: %s", booking.Status)})
			return
		}

		handover := models.BookingHandover{
			BookingID:  booking.ID,
			Type:       handoverType,
			PhotoURLs:  models.JSONBArray(req.PhotoURLs),
			Checklist:  models.JSONBMap(req.Checklist),
			Notes:      req.Notes,
			UploadedBy: &userID,
		}

		if err := db.Create(&handover).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save handover"})
			return
		}

		// Notify customer to verify
		createNotif(db, booking.CustomerID, booking.ID, models.NotifHandoverUploaded,
			"Equipment Photos Ready for Verification",
			fmt.Sprintf("The vendor has uploaded %d photos for your equipment %s handover. Please verify.", len(req.PhotoURLs), handoverType))

		c.JSON(http.StatusCreated, gin.H{"message": "Handover uploaded. Customer will be notified.", "handover": handover})
	}
}

// GetHandovers — get all handover records for a booking
// GET /api/v1/bookings/:id/handover
func GetHandovers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		role := middleware.GetRole(c)
		bookingID := c.Param("id")

		var booking models.Booking
		if err := db.First(&booking, bookingID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}

		// Access control: only customer, vendor, or admin
		if role == models.RoleCustomer && booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		var handovers []models.BookingHandover
		db.Where("booking_id = ?", bookingID).Order("created_at ASC").Find(&handovers)

		c.JSON(http.StatusOK, gin.H{"handovers": handovers})
	}
}

// ─── CUSTOMER: Verify Return / Initiate Return ────────────────────────────────

// InitiateReturn — vendor triggers return flow and sends a return OTP to customer
// POST /api/v1/bookings/:id/initiate-return
func InitiateReturn(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		bookingID := c.Param("id")

		var booking models.Booking
		if err := db.Preload("Equipment").Preload("Equipment.Vendor").
			Preload("Generator").Preload("Generator.Vendor").
			First(&booking, bookingID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}

		vendorUserID := getEquipmentVendorUserID(db, booking)
		if vendorUserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the booking vendor can initiate return"})
			return
		}

		if booking.Status != models.BookingDelivered {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("return can only be initiated when booking is in 'delivered' state, current: %s", booking.Status)})
			return
		}

		otp := generateReturnOTP()
		now := time.Now()
		if err := db.Model(&booking).Updates(map[string]interface{}{
			"return_otp":            otp,
			"return_initiated_at":   now,
			"return_otp_verified":   false,
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate return"})
			return
		}

		// Notify customer with OTP
		createNotif(db, booking.CustomerID, booking.ID, models.NotifReturnOTP,
			"Equipment Return OTP",
			fmt.Sprintf("The vendor is picking up your equipment. Return OTP: %s. Share this with the vendor to confirm pickup.", otp))

		c.JSON(http.StatusOK, gin.H{"message": "Return initiated. OTP sent to customer."})
	}
}

// ConfirmReturn — customer gives OTP to vendor who confirms return
// POST /api/v1/bookings/:id/confirm-return  body: { otp: "123456" }
func ConfirmReturn(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		bookingID := c.Param("id")

		var body struct {
			OTP string `json:"otp" binding:"required,len=6"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err)})
			return
		}

		var booking models.Booking
		if err := db.Preload("Equipment").Preload("Equipment.Vendor").
			Preload("Generator").Preload("Generator.Vendor").
			First(&booking, bookingID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}

		// Only the vendor confirms the return OTP
		vendorUserID := getEquipmentVendorUserID(db, booking)
		if vendorUserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "only the vendor can confirm return OTP"})
			return
		}

		if booking.ReturnInitiatedAt == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "return has not been initiated yet"})
			return
		}
		if booking.ReturnOTPVerified {
			c.JSON(http.StatusConflict, gin.H{"error": "return already confirmed"})
			return
		}
		if booking.ReturnOTP != body.OTP {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid return OTP"})
			return
		}

		now := time.Now()
		db.Model(&booking).Updates(map[string]interface{}{
			"return_otp_verified": true,
			"status":              models.BookingCompleted,
			"completed_at":        now,
		})

		// Restore equipment stock
		if booking.EquipmentID != nil && *booking.EquipmentID != 0 {
			var eq models.Equipment
			if db.First(&eq, booking.EquipmentID).Error == nil {
				newQty := eq.AvailableQuantity + 1
				restoreUpdates := map[string]interface{}{"available_quantity": newQty}
				if eq.AvailabilityStatus == models.StatusBooked {
					restoreUpdates["availability_status"] = models.StatusAvailable
				}
				db.Model(&eq).Updates(restoreUpdates)
			}
		}

		createNotif(db, booking.CustomerID, booking.ID, models.NotifCompleted,
			"Booking Completed",
			"Your equipment has been returned. Please rate the vendor!")

		c.JSON(http.StatusOK, gin.H{"message": "Return confirmed. Booking completed!", "status": models.BookingCompleted})
	}
}

// ─── CUSTOMER: Raise Damage Dispute ──────────────────────────────────────────

// RaiseDamageDispute — customer raises a damage claim within 48h of completion
// POST /api/v1/bookings/:id/dispute
func RaiseDamageDispute(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		bookingID := c.Param("id")

		var req DisputeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var booking models.Booking
		if err := db.First(&booking, bookingID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}

		if booking.CustomerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		if booking.Status != models.BookingCompleted && booking.Status != models.BookingDelivered {
			c.JSON(http.StatusConflict, gin.H{"error": "disputes can only be raised on delivered or completed bookings"})
			return
		}

		// Check 48h window after completion
		if booking.CompletedAt != nil && time.Since(*booking.CompletedAt) > 48*time.Hour {
			c.JSON(http.StatusConflict, gin.H{"error": "dispute window has expired (48 hours after completion)"})
			return
		}

		// Only one dispute per booking
		var existing int64
		db.Model(&models.DamageDispute{}).Where("booking_id = ?", booking.ID).Count(&existing)
		if existing > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "a dispute has already been raised for this booking"})
			return
		}

		dispute := models.DamageDispute{
			BookingID:     booking.ID,
			RaisedBy:      userID,
			Description:   req.Description,
			ClaimedAmount: req.ClaimedAmount,
			PhotoURLs:     models.JSONBArray(req.PhotoURLs),
			Status:        models.DisputeOpen,
		}

		if err := db.Create(&dispute).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dispute"})
			return
		}

		// Notify admin (user_id = 1 as convention, or all admins)
		auditLog(db, userID, "dispute_raised", "damage_disputes", dispute.ID, "", "open", c.ClientIP())

		c.JSON(http.StatusCreated, gin.H{"message": "Dispute raised. Admin team will review within 24-48 hours.", "dispute": dispute})
	}
}

// GetMyDisputes — customer gets their disputes
// GET /api/v1/disputes
func GetMyDisputes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var disputes []models.DamageDispute
		db.Where("raised_by = ?", userID).Preload("Booking").Order("created_at DESC").Find(&disputes)

		c.JSON(http.StatusOK, gin.H{"disputes": disputes})
	}
}

// ─── ADMIN: Manage Disputes ───────────────────────────────────────────────────

// AdminListDisputes — admin views all disputes
// GET /api/v1/admin/disputes
func AdminListDisputes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.DefaultQuery("status", "")
		page, _ := parseIntQuery(c, "page", 1)
		limit, _ := parseIntQuery(c, "limit", 20)
		offset := (page - 1) * limit

		query := db.Preload("Booking").Preload("User")
		if status != "" {
			query = query.Where("status = ?", status)
		}

		var disputes []models.DamageDispute
		var total int64
		query.Model(&models.DamageDispute{}).Count(&total)
		query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&disputes)

		c.JSON(http.StatusOK, gin.H{"disputes": disputes, "total": total, "page": page, "limit": limit})
	}
}

// AdminResolveDispute — admin resolves or rejects a dispute
// PUT /api/v1/admin/disputes/:id/resolve
func AdminResolveDispute(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		disputeID := c.Param("id")

		var req ResolveDisputeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		var dispute models.DamageDispute
		if err := db.First(&dispute, disputeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "dispute not found"})
			return
		}

		if dispute.Status != models.DisputeOpen {
			c.JSON(http.StatusConflict, gin.H{"error": "dispute is already closed"})
			return
		}

		now := time.Now()
		db.Model(&dispute).Updates(map[string]interface{}{
			"status":      models.DisputeStatus(req.Status),
			"admin_notes": req.AdminNotes,
			"resolved_at": now,
		})

		// Notify the customer who raised the dispute
		msg := fmt.Sprintf("Your damage dispute #%d has been %s. Admin notes: %s", dispute.ID, req.Status, req.AdminNotes)
		createNotif(db, dispute.RaisedBy, dispute.BookingID, models.NotifDisputeResolved, "Dispute "+req.Status, msg)

		c.JSON(http.StatusOK, gin.H{"message": "Dispute " + req.Status, "dispute": dispute})
	}
}

// parseIntQuery is a small helper to parse integer query params with a default
func parseIntQuery(c *gin.Context, key string, def int) (int, bool) {
	val := c.Query(key)
	if val == "" {
		return def, false
	}
	n := 0
	_, err := fmt.Sscanf(val, "%d", &n)
	if err != nil || n < 1 {
		return def, false
	}
	return n, true
}
