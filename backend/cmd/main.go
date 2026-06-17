package main

import (
	"context"
	"genrent/internal/config"
	"genrent/internal/database"
	"genrent/internal/handlers"
	"genrent/internal/middleware"
	"genrent/internal/services/email"
	"genrent/internal/workers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// Start background expiry worker
	workers.StartExpiryWorker(db)

	// Init email service
	handlers.InitEmail(email.Config{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Pass:     cfg.SMTPPass,
		From:     cfg.SMTPFrom,
		FromName: cfg.SMTPFromName,
		Enabled:  cfg.EmailEnabled,
	})

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})
	r.Use(middleware.PanicRecovery())           // structured JSON panic recovery
	r.Use(middleware.RequestID())               // X-Request-ID on every request
	r.Use(middleware.RequestMonitor())          // structured JSON request logging
	r.Use(middleware.SecurityHeaders())         // CSP, X-Frame-Options, etc.
	r.Use(middleware.CORS(cfg.AllowedOrigins))  // configurable per-environment CORS
	r.Use(middleware.BodyLimit(2 << 20))        // 2 MB max request body
	r.Use(middleware.RequestTimeout(30 * time.Second)) // abort stuck requests after 30s
	r.MaxMultipartMemory = 8 << 20              // 8 MB max multipart (file uploads)

	// Health check — no auth, no versioning (used by load balancers / uptime monitors)
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "unreachable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
			"version":  "v1",
		})
	})

	// Metrics — lightweight platform stats (protect behind IP allowlist in production)
	r.GET("/metrics", handlers.GetMetrics(db))

	// API docs — structured endpoint reference
	r.GET("/docs", handlers.GetAPIDocs)

	// Serve frontend static files from ../frontend
	frontendPath := "../frontend"
	r.Static("/static", frontendPath)
	r.StaticFile("/", frontendPath+"/index.html")
	r.StaticFile("/login", frontendPath+"/login.html")
	r.StaticFile("/register", frontendPath+"/register.html")
	r.StaticFile("/vendor-dashboard", frontendPath+"/vendor-dashboard.html")
	r.StaticFile("/add-equipment", frontendPath+"/add-equipment.html")
	r.StaticFile("/admin-dashboard", frontendPath+"/admin-dashboard.html")
	r.StaticFile("/booking", frontendPath+"/booking.html")
	r.StaticFile("/payment", frontendPath+"/payment.html")
	r.StaticFile("/my-bookings", frontendPath+"/my-bookings.html")

	// Helper to register category & equipment routes
	registerEquipmentRoutes := func(rg *gin.RouterGroup) {
		// Category routes
		rg.GET("/categories", handlers.GetCategories(db))
		rg.GET("/categories/hierarchy", handlers.GetCategoryHierarchy(db))
		rg.GET("/categories/popular", handlers.PopularCategories(db))
		rg.GET("/categories/:id", handlers.GetCategory(db))
		rg.GET("/categories/:id/equipment", handlers.GetCategoryEquipment(db))

		// Equipment routes
		rg.GET("/equipment/search", handlers.SearchEquipment(db))
		rg.GET("/equipment/:id", handlers.GetEquipment(db))

		// Protected equipment routes (vendor only)
		prot := rg.Group("/")
		prot.Use(middleware.Auth(cfg.JWTSecret))
		{
			prot.POST("/equipment", middleware.VendorOnly(), handlers.CreateEquipment(db))
			prot.PUT("/equipment/:id", middleware.VendorOnly(), handlers.UpdateEquipment(db))
			prot.DELETE("/equipment/:id", middleware.VendorOnly(), handlers.DeleteEquipment(db))
			prot.GET("/equipment/mine", middleware.VendorOnly(), handlers.GetMyEquipment(db))
			prot.PUT("/equipment/:id/status", middleware.VendorOnly(), handlers.UpdateEquipmentStatus(db))
			prot.GET("/equipment-stats", middleware.VendorOnly(), handlers.EquipmentBookingStats(db))
		}
	}

	// Legacy /api group to support the /api prefix used by some frontend files
	apiLegacy := r.Group("/api")
	registerEquipmentRoutes(apiLegacy)

	// All API routes are versioned under /api/v1
	api := r.Group("/api/v1")
	registerEquipmentRoutes(api)

	// Payment gateway webhooks — public but rate-limited (gateway posts here)
	api.POST("/webhooks/payment", middleware.RateLimit(60, 60), handlers.HandlePaymentWebhook(db))

	// Public routes — rate limited
	auth := api.Group("/auth")
	auth.Use(middleware.RateLimit(10, 60)) // 10 requests per minute
	auth.POST("/register", handlers.Register(db))
	auth.POST("/login", handlers.Login(db, cfg.JWTSecret))

	// Public read routes
	api.GET("/generators", handlers.SearchGenerators(db))
	api.GET("/generators/:id", handlers.GetGenerator(db))
	api.GET("/vendors", handlers.ListVendors(db))
	api.GET("/vendors/:id", handlers.GetVendorByID(db))

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.Auth(cfg.JWTSecret))
	{
		protected.GET("/auth/profile", handlers.GetProfile(db))

		// Vendor management
		protected.POST("/vendors", handlers.CreateVendor(db))
		protected.GET("/vendors/me", handlers.GetMyVendorProfile(db))
		protected.PUT("/vendors/me", handlers.UpdateVendorProfile(db))

		// Generator management (vendor only)
		protected.POST("/generators", middleware.VendorOnly(), handlers.CreateGenerator(db))
		protected.PUT("/generators/:id", middleware.VendorOnly(), handlers.UpdateGenerator(db))
		protected.DELETE("/generators/:id", middleware.VendorOnly(), handlers.DeleteGenerator(db))
		protected.GET("/generators/mine", middleware.VendorOnly(), handlers.GetMyGenerators(db))

		// Bookings — customer (rate limited)
		bookings := protected.Group("/bookings")
		bookings.Use(middleware.RateLimit(30, 60)) // 30 per minute
		bookings.POST("", handlers.CreateBookingV2(db))
		bookings.GET("", handlers.GetMyBookingsV2(db))
		bookings.GET("/:id", handlers.GetBooking(db))
		bookings.GET("/:id/status", handlers.GetBookingStatus(db)) // lightweight polling endpoint
		bookings.POST("/:id/confirm-delivery", handlers.CustomerConfirmDelivery(db))
		bookings.POST("/:id/complete", handlers.CustomerCompleteBooking(db))
		bookings.POST("/:id/cancel", handlers.CancelBooking(db))
		bookings.POST("/:id/review", handlers.SubmitReview(db))

		// Bookings — vendor actions
		protected.POST("/bookings/:id/accept", middleware.VendorOnly(), handlers.VendorAcceptBooking(db))
		protected.POST("/bookings/:id/reject", middleware.VendorOnly(), handlers.VendorRejectBooking(db))
		protected.POST("/bookings/:id/dispatch", middleware.VendorOnly(), handlers.VendorDispatchGenerator(db))
		protected.POST("/bookings/:id/resend-otp", middleware.VendorOnly(), handlers.VendorResendOTP(db))
		// Feature 5: Handover photos
		bookings.POST("/:id/handover", handlers.UploadHandover(db))    // vendor uploads photos
		bookings.GET("/:id/handover", handlers.GetHandovers(db))        // view handover records
		// Feature 5: Return flow
		protected.POST("/bookings/:id/initiate-return", middleware.VendorOnly(), handlers.InitiateReturn(db))
		protected.POST("/bookings/:id/confirm-return", middleware.VendorOnly(), handlers.ConfirmReturn(db))
		// Feature 5: Damage disputes
		bookings.POST("/:id/dispute", handlers.RaiseDamageDispute(db))
		protected.GET("/disputes", handlers.GetMyDisputes(db))

		// Admin booking status override
		protected.PUT("/bookings/:id/status", handlers.UpdateBookingStatus(db))

		// Payments — rate limited
		payments := protected.Group("/payments")
		payments.Use(middleware.RateLimit(10, 60))
		payments.GET("/booking/:booking_id", handlers.GetPaymentDetails(db))
		payments.POST("", handlers.ProcessPayment(db))

		// Vendor Wallet & Withdrawals
		protected.GET("/wallet", middleware.VendorOnly(), handlers.GetVendorWallet(db))
		protected.POST("/wallet/withdraw", middleware.VendorOnly(), handlers.RequestWithdrawal(db))
		protected.POST("/wallet/withdraw/:id/confirm", middleware.VendorOnly(), handlers.ConfirmWithdrawalOTP(db))
		protected.GET("/wallet/withdrawals", middleware.VendorOnly(), handlers.GetWithdrawals(db))
		protected.GET("/wallet/bank-accounts", middleware.VendorOnly(), handlers.GetBankAccounts(db))
		protected.POST("/wallet/bank-accounts", middleware.VendorOnly(), handlers.SaveBankAccount(db))
		protected.DELETE("/wallet/bank-accounts/:id", middleware.VendorOnly(), handlers.DeleteBankAccount(db))

		// Notifications
		protected.GET("/notifications", handlers.GetNotifications(db))
		protected.POST("/notifications/:id/read", handlers.MarkNotificationRead(db))
		protected.POST("/notifications/read-all", handlers.MarkAllNotificationsRead(db))

		// My audit trail
		protected.GET("/my-activity", handlers.GetMyAuditLogs(db))

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/vendors", handlers.AdminListVendors(db))
			admin.PUT("/vendors/:id/verify", handlers.AdminVerifyVendor(db))
			admin.PUT("/vendors/:id/reject", handlers.AdminRejectVendor(db))
			admin.PUT("/vendors/:id/penalize", handlers.AdminPenalizeVendor(db))
			admin.GET("/generators", handlers.AdminListGenerators(db))
			admin.PUT("/generators/:id/status", handlers.AdminUpdateGeneratorStatus(db))
			admin.GET("/bookings", handlers.AdminListBookings(db))
			admin.POST("/bookings/:id/force-cancel", handlers.AdminForceCancel(db))
			admin.POST("/bookings/:id/release-escrow", handlers.AdminReleaseEscrow(db))
			admin.POST("/bookings/:id/refund", handlers.AdminRefundCustomer(db))
			admin.GET("/stats", handlers.AdminGetStats(db))
			admin.GET("/audit-logs", handlers.GetAuditLogs(db))
			admin.GET("/withdrawals", handlers.AdminListWithdrawals(db))
			admin.POST("/withdrawals/:id/approve", handlers.AdminApproveWithdrawal(db))
			admin.POST("/withdrawals/:id/reject", handlers.AdminRejectWithdrawal(db))
			// Feature 5: Admin dispute management
			admin.GET("/disputes", handlers.AdminListDisputes(db))
			admin.PUT("/disputes/:id/resolve", handlers.AdminResolveDispute(db))
		}
	}

	// ---- Graceful Shutdown ----
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start server in background
	go func() {
		log.Printf("GenRent server starting on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for OS interrupt signal (Ctrl+C, SIGTERM from systemd/Docker)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server — waiting up to 10s for in-flight requests...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Forced shutdown:", err)
	}
	log.Println("Server exited cleanly")
}
