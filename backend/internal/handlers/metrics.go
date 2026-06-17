package handlers

import (
	"genrent/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetMetrics returns lightweight platform health metrics.
// This endpoint is public (no auth) — expose carefully in production
// or protect it behind a network-level firewall / internal IP allowlist.
func GetMetrics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var totalUsers, totalVendors, totalGenerators int64
		var activeBookings, completedBookings, totalBookings int64
		var totalRevenue, revenueToday, revenueMonth float64

		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		db.Model(&models.User{}).Count(&totalUsers)
		db.Model(&models.Vendor{}).Where("verified = ?", true).Count(&totalVendors)
		db.Model(&models.Generator{}).Where("availability_status = ?", models.StatusAvailable).Count(&totalGenerators)

		db.Model(&models.Booking{}).Where(
			"status NOT IN ?",
			[]string{string(models.BookingCompleted), string(models.BookingCancelled)},
		).Count(&activeBookings)
		db.Model(&models.Booking{}).Where("status = ?", models.BookingCompleted).Count(&completedBookings)
		db.Model(&models.Booking{}).Count(&totalBookings)

		db.Model(&models.Booking{}).Where("status = ?", models.BookingCompleted).
			Select("COALESCE(SUM(total_price), 0)").Scan(&totalRevenue)
		db.Model(&models.Booking{}).Where("status = ? AND completed_at >= ?", models.BookingCompleted, todayStart).
			Select("COALESCE(SUM(total_price), 0)").Scan(&revenueToday)
		db.Model(&models.Booking{}).Where("status = ? AND completed_at >= ?", models.BookingCompleted, monthStart).
			Select("COALESCE(SUM(total_price), 0)").Scan(&revenueMonth)

		c.JSON(http.StatusOK, gin.H{
			"total_users":         totalUsers,
			"verified_vendors":    totalVendors,
			"available_generators": totalGenerators,
			"active_bookings":     activeBookings,
			"completed_bookings":  completedBookings,
			"total_bookings":      totalBookings,
			"platform_revenue": gin.H{
				"all_time": totalRevenue,
				"today":    revenueToday,
				"month":    revenueMonth,
			},
			"generated_at": now.UTC().Format(time.RFC3339),
		})
	}
}
