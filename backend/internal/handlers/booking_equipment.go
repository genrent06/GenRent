package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateBookingRequestV2 supports both generator and equipment bookings
type CreateBookingRequestV2 struct {
	GeneratorID *uint  `json:"generator_id"` // optional (for backward compatibility)
	EquipmentID *uint  `json:"equipment_id"` // optional (new)
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Notes       string `json:"notes"`
}

// CreateBookingV2 handles both generator and equipment bookings
func CreateBookingV2(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var req CreateBookingRequestV2
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": ValidationError(err), "errors": ValidationErrors(err)})
			return
		}

		// Ensure either generator_id or equipment_id is provided
		if (req.GeneratorID == nil || *req.GeneratorID == 0) &&
			(req.EquipmentID == nil || *req.EquipmentID == 0) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "either generator_id or equipment_id must be provided"})
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
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "maximum 5 active bookings allowed per customer"})
			return
		}

		var booking models.Booking
		var equipment *models.Equipment
		var generator *models.Generator
		var categoryID *uint

		txErr := db.Transaction(func(tx *gorm.DB) error {
			// Handle equipment booking
			if req.EquipmentID != nil && *req.EquipmentID != 0 {
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					First(&equipment, *req.EquipmentID).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "equipment not found"})
					return err
				}

				// Feature 4: Multi-unit inventory check
				if equipment.AvailableQuantity <= 0 || equipment.AvailabilityStatus == models.StatusBooked {
					c.JSON(http.StatusConflict, gin.H{"error": "equipment is not available"})
					return gorm.ErrRecordNotFound
				}

				categoryID = &equipment.CategoryID
				booking.EquipmentID = &equipment.ID
				booking.CategoryID = categoryID

				// Feature 4: Atomically decrement available stock
				newQty := equipment.AvailableQuantity - 1
				updates := map[string]interface{}{
					"available_quantity": newQty,
				}
				if newQty == 0 {
					updates["availability_status"] = models.StatusBooked
				}
				tx.Model(&equipment).Updates(updates)
			} else if req.GeneratorID != nil && *req.GeneratorID != 0 {
				// Handle generator booking (backward compatibility)
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					First(&generator, *req.GeneratorID).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "generator not found"})
					return err
				}

				if generator.AvailabilityStatus != models.StatusAvailable {
					c.JSON(http.StatusConflict, gin.H{"error": "generator is not available"})
					return gorm.ErrRecordNotFound
				}

				booking.GeneratorID = &generator.ID

				// Reserve for 30 minutes
				expiry := time.Now().Add(30 * time.Minute)
				tx.Model(&generator).Updates(map[string]interface{}{
					"availability_status": models.StatusReserved,
					"reservation_expiry":  expiry,
				})
			}

			// Calculate pricing (Feature 3: include mobilization fees)
			days := int(endDate.Sub(startDate).Hours() / 24)
			if days < 1 {
				days = 1
			}
			var rentalPrice float64
			var mobFee, demobFee float64
			if req.EquipmentID != nil && *req.EquipmentID != 0 {
				rentalPrice = CalculateTieredPrice(days, equipment.DailyPrice, equipment.WeeklyPrice, equipment.MonthlyPrice)
				mobFee = equipment.MobilizationFee
				demobFee = equipment.DemobilizationFee
			} else if req.GeneratorID != nil && *req.GeneratorID != 0 {
				rentalPrice = CalculateTieredPrice(days, generator.PricePerDay, 0, generator.PricePerMonth)
			}
			totalPrice := rentalPrice + mobFee + demobFee
			advanceAmount := totalPrice * 0.30 // 30% advance

			booking.CustomerID = userID
			booking.StartDate = startDate
			booking.EndDate = endDate
			booking.RentalPrice = rentalPrice
			booking.MobilizationFee = mobFee
			booking.DemobilizationFee = demobFee
			booking.TotalPrice = totalPrice
			booking.AdvanceAmount = advanceAmount
			booking.Address = req.Address
			booking.Notes = req.Notes
			booking.Status = models.BookingRequested

			if err := tx.Create(&booking).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create booking"})
				return err
			}

			return nil
		})

		if txErr != nil {
			return
		}

		// Preload relationships
		db.Preload("Customer").
			Preload("Equipment").
			Preload("Equipment.Vendor").
			Preload("Equipment.Category").
			Preload("Generator").
			Preload("Generator.Vendor").
			First(&booking)

		c.JSON(http.StatusCreated, booking)
	}
}

// GetMyBookingsV2 returns all bookings for the current user (customer or vendor) with equipment/generator details
func GetMyBookingsV2(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		role := middleware.GetRole(c)

		var bookings []models.Booking
		if role == models.RoleVendor {
			var vendor models.Vendor
			if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "vendor profile not found"})
				return
			}

			db.Preload("Customer").
				Preload("Equipment").
				Preload("Equipment.Category").
				Preload("Generator").
				Joins("LEFT JOIN equipment ON bookings.equipment_id = equipment.id").
				Joins("LEFT JOIN generators ON bookings.generator_id = generators.id").
				Where("(equipment.vendor_id = ? OR generators.vendor_id = ?)", vendor.ID, vendor.ID).
				Order("bookings.created_at DESC").
				Find(&bookings)
		} else {
			db.Preload("Customer").
				Preload("Equipment").
				Preload("Equipment.Vendor").
				Preload("Equipment.Category").
				Preload("Generator").
				Preload("Generator.Vendor").
				Where("customer_id = ?", userID).
				Order("created_at DESC").
				Find(&bookings)
		}

		c.JSON(http.StatusOK, gin.H{"bookings": bookings})
	}
}

// GetVendorBookingsV2 returns all bookings for the current vendor (equipment)
func GetVendorBookingsV2(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		var bookings []models.Booking

		// Get bookings for equipment owned by this vendor
		db.
			Joins("LEFT JOIN equipment ON bookings.equipment_id = equipment.id").
			Joins("LEFT JOIN generators ON bookings.generator_id = generators.id").
			Where("(equipment.vendor_id = ? OR generators.vendor_id = ?)", vendor.ID, vendor.ID).
			Preload("Customer").
			Preload("Equipment").
			Preload("Equipment.Category").
			Preload("Generator").
			Order("bookings.created_at DESC").
			Find(&bookings)

		c.JSON(http.StatusOK, gin.H{"bookings": bookings})
	}
}

// EquipmentBookingStats returns booking statistics for equipment
func EquipmentBookingStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		var vendor models.Vendor
		if result := db.Where("user_id = ?", userID).First(&vendor); result.Error != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "vendor profile not found"})
			return
		}

		// Get all equipment for vendor with booking and revenue stats
		var stats []gin.H
		db.
			Select(`equipment.id, equipment.name, equipment.category_id, 
				COUNT(CASE WHEN bookings.status != 'cancelled' THEN 1 END) as total_bookings,
				SUM(CASE WHEN bookings.status = 'completed' THEN bookings.total_price ELSE 0 END) as total_revenue`).
			Table("equipment").
			Joins("LEFT JOIN bookings ON equipment.id = bookings.equipment_id").
			Where("equipment.vendor_id = ?", vendor.ID).
			Group("equipment.id").
			Scan(&stats)

		c.JSON(http.StatusOK, gin.H{"stats": stats})
	}
}

// CalculateTieredPrice calculates pricing using daily, weekly, and monthly rates
func CalculateTieredPrice(days int, daily, weekly, monthly float64) float64 {
	if days <= 0 {
		return 0
	}

	// Fallback if no weekly or monthly rates are set
	if weekly <= 0 && monthly <= 0 {
		return float64(days) * daily
	}

	var total float64

	if monthly > 0 {
		months := days / 30
		remainingDays := days % 30

		total += float64(months) * monthly

		// Calculate cost of remaining days
		var remainingCost float64
		if weekly > 0 {
			weeks := remainingDays / 7
			leftoverDays := remainingDays % 7

			weekCost := float64(weeks) * weekly
			dayCost := float64(leftoverDays) * daily
			if dayCost > weekly {
				dayCost = weekly
			}
			remainingCost = weekCost + dayCost
		} else {
			remainingCost = float64(remainingDays) * daily
		}

		// Cap remaining cost at monthly price
		if remainingCost > monthly {
			remainingCost = monthly
		}
		total += remainingCost
	} else {
		// Only weekly and daily are available
		weeks := days / 7
		leftoverDays := days % 7

		total += float64(weeks) * weekly
		dayCost := float64(leftoverDays) * daily
		if dayCost > weekly {
			dayCost = weekly
		}
		total += dayCost
	}

	return total
}
