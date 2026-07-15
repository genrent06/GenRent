package handlers

import (
	"genrent/internal/middleware"
	"genrent/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// auditLog records a system-wide audit event
func auditLog(db *gorm.DB, userID uint, action, entityType string, entityID uint, oldValue, newValue, ip string) {
	db.Create(&models.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   oldValue,
		NewValue:   newValue,
		IPAddress:  ip,
	})
}

// GetAuditLogs — admin only: retrieve audit trail
func GetAuditLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset := (page - 1) * limit
		entityType := c.Query("entity_type")
		entityID := c.Query("entity_id")
		action := c.Query("action")

		query := db.Model(&models.AuditLog{})
		if entityType != "" {
			query = query.Where("entity_type = ?", entityType)
		}
		if entityID != "" {
			query = query.Where("entity_id = ?", entityID)
		}
		if action != "" {
			query = query.Where("action = ?", action)
		}

		var logs []models.AuditLog
		var total int64
		query.Count(&total)
		query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs)

		c.JSON(http.StatusOK, gin.H{"logs": logs, "total": total, "page": page})
	}
}

// GetMyAuditLogs — returns audit trail for the logged-in user's own actions
func GetMyAuditLogs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		var logs []models.AuditLog
		db.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&logs)
		c.JSON(http.StatusOK, gin.H{"logs": logs})
	}
}
