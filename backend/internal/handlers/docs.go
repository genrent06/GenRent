package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAPIDocs returns a structured API reference for all GenRent endpoints.
// Visit GET /docs for a human-readable JSON overview.
func GetAPIDocs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":   "GenRent API",
		"version": "v1",
		"base":    "/api/v1",
		"health":  "GET /health",
		"docs":    "GET /docs",
		"endpoints": gin.H{
			"auth": []gin.H{
				{
					"method": "POST", "path": "/api/v1/auth/register",
					"auth": false, "body": gin.H{"name": "string", "email": "string", "phone": "string", "password": "string", "role": "customer|vendor|admin"},
					"desc": "Register a new user",
				},
				{
					"method": "POST", "path": "/api/v1/auth/login",
					"auth": false, "body": gin.H{"email": "string", "password": "string"},
					"desc": "Login and receive JWT token",
				},
				{
					"method": "GET", "path": "/api/v1/auth/profile",
					"auth": true, "desc": "Get current user profile",
				},
			},
			"generators": []gin.H{
				{
					"method": "GET", "path": "/api/v1/generators",
					"auth": false,
					"query": gin.H{"city": "string", "lat": "float", "lng": "float", "radius": "float(km, default 5, max 25)", "capacity": "int", "min_price": "float", "max_price": "float", "page": "int", "limit": "int"},
					"desc": "Search available generators with geo-radius filter",
				},
				{"method": "GET", "path": "/api/v1/generators/:id", "auth": false, "desc": "Get generator details"},
				{"method": "POST", "path": "/api/v1/generators", "auth": true, "role": "vendor", "desc": "Create generator (max 50 per vendor)"},
				{"method": "PUT", "path": "/api/v1/generators/:id", "auth": true, "role": "vendor", "desc": "Update your generator"},
				{"method": "DELETE", "path": "/api/v1/generators/:id", "auth": true, "role": "vendor", "desc": "Soft-delete your generator"},
				{"method": "GET", "path": "/api/v1/generators/mine", "auth": true, "role": "vendor", "desc": "List your generators"},
			},
			"bookings": []gin.H{
				{
					"method": "POST", "path": "/api/v1/bookings",
					"auth": true, "role": "customer",
					"body": gin.H{"generator_id": "uint", "start_date": "YYYY-MM-DD", "end_date": "YYYY-MM-DD", "address": "string", "notes": "string"},
					"desc": "Create booking request (max 5 active per customer); locks generator 30 min",
				},
				{"method": "GET", "path": "/api/v1/bookings", "auth": true, "desc": "List my bookings"},
				{"method": "GET", "path": "/api/v1/bookings/:id", "auth": true, "desc": "Get booking detail"},
				{"method": "GET", "path": "/api/v1/bookings/:id/status", "auth": true, "desc": "Lightweight status poll (call every 10s)"},
				{"method": "POST", "path": "/api/v1/bookings/:id/accept", "auth": true, "role": "vendor", "desc": "Vendor accepts booking"},
				{"method": "POST", "path": "/api/v1/bookings/:id/reject", "auth": true, "role": "vendor", "desc": "Vendor rejects booking"},
				{"method": "POST", "path": "/api/v1/bookings/:id/dispatch", "auth": true, "role": "vendor", "desc": "Vendor dispatches generator; OTP sent to customer"},
				{"method": "POST", "path": "/api/v1/bookings/:id/confirm-delivery", "auth": true, "role": "customer", "body": gin.H{"otp": "string"}, "desc": "Customer confirms delivery with OTP; releases escrow"},
				{"method": "POST", "path": "/api/v1/bookings/:id/complete", "auth": true, "role": "customer", "desc": "Customer marks booking complete"},
				{"method": "POST", "path": "/api/v1/bookings/:id/cancel", "auth": true, "desc": "Cancel booking"},
				{"method": "POST", "path": "/api/v1/bookings/:id/review", "auth": true, "role": "customer", "body": gin.H{"rating": "1-5", "review": "string"}, "desc": "Submit vendor review"},
			},
			"payments": []gin.H{
				{"method": "GET", "path": "/api/v1/payments/booking/:booking_id", "auth": true, "desc": "Get payment breakdown (advance + escrow details)"},
				{
					"method": "POST", "path": "/api/v1/payments",
					"auth": true, "role": "customer",
					"body": gin.H{"booking_id": "uint", "method": "upi|card|netbanking|wallet|cash"},
					"desc": "Pay advance (30%); 15% to vendor escrow, 15% platform fee",
				},
			},
			"wallet": []gin.H{
				{"method": "GET", "path": "/api/v1/wallet", "auth": true, "role": "vendor", "desc": "Get vendor wallet: balance + hold_balance (escrow) + last 50 transactions"},
			},
			"notifications": []gin.H{
				{"method": "GET", "path": "/api/v1/notifications", "auth": true, "desc": "Get my notifications (unread first)"},
				{"method": "POST", "path": "/api/v1/notifications/:id/read", "auth": true, "desc": "Mark notification as read"},
				{"method": "POST", "path": "/api/v1/notifications/read-all", "auth": true, "desc": "Mark all notifications as read"},
			},
			"webhooks": []gin.H{
				{
					"method": "POST", "path": "/api/v1/webhooks/payment",
					"auth": false,
					"body": gin.H{"event": "payment.captured|payment.failed|refund.processed", "transaction_id": "string", "booking_id": "uint", "amount": "float", "status": "string"},
					"desc": "Payment gateway async callback (HMAC verification TODO for production)",
				},
			},
			"admin": []gin.H{
				{"method": "GET", "path": "/api/v1/admin/stats", "auth": true, "role": "admin", "desc": "Platform statistics with revenue_today + revenue_month"},
				{"method": "GET", "path": "/api/v1/admin/vendors", "auth": true, "role": "admin", "query": gin.H{"verified": "true|false", "page": "int", "limit": "int"}, "desc": "List all vendors"},
				{"method": "PUT", "path": "/api/v1/admin/vendors/:id/verify", "auth": true, "role": "admin", "desc": "Verify vendor"},
				{"method": "PUT", "path": "/api/v1/admin/vendors/:id/reject", "auth": true, "role": "admin", "desc": "Reject vendor"},
				{"method": "PUT", "path": "/api/v1/admin/vendors/:id/penalize", "auth": true, "role": "admin", "body": gin.H{"amount": "float", "reason": "string"}, "desc": "Deduct penalty from vendor wallet"},
				{"method": "GET", "path": "/api/v1/admin/generators", "auth": true, "role": "admin", "desc": "List all generators"},
				{"method": "PUT", "path": "/api/v1/admin/generators/:id/status", "auth": true, "role": "admin", "body": gin.H{"status": "available|reserved|booked|maintenance"}, "desc": "Override generator status"},
				{"method": "GET", "path": "/api/v1/admin/bookings", "auth": true, "role": "admin", "query": gin.H{"status": "string", "page": "int"}, "desc": "List all bookings"},
				{"method": "POST", "path": "/api/v1/admin/bookings/:id/force-cancel", "auth": true, "role": "admin", "body": gin.H{"reason": "string"}, "desc": "Force cancel any booking"},
				{"method": "POST", "path": "/api/v1/admin/bookings/:id/release-escrow", "auth": true, "role": "admin", "desc": "Release escrow to vendor (dispute: vendor wins)"},
				{"method": "POST", "path": "/api/v1/admin/bookings/:id/refund", "auth": true, "role": "admin", "body": gin.H{"reason": "string"}, "desc": "Refund customer (dispute: customer wins)"},
				{"method": "GET", "path": "/api/v1/admin/audit-logs", "auth": true, "role": "admin", "query": gin.H{"entity_type": "string", "entity_id": "uint", "action": "string"}, "desc": "Full audit trail"},
			},
			"metrics": []gin.H{
				{"method": "GET", "path": "/metrics", "auth": false, "desc": "Platform metrics: users, vendors, bookings, revenue (protect in production)"},
			},
		},
		"booking_flow": []string{
			"1. Customer creates booking → status=requested, generator locked 30min",
			"2. Vendor accepts → status=accepted",
			"3. Customer pays 30% advance → status=advance_paid, 15% in vendor escrow",
			"4. Vendor dispatches → status=dispatched, OTP generated",
			"5. Customer enters OTP → status=delivered, escrow released to vendor",
			"6. Customer marks complete → status=completed, review optional",
		},
	})
}
