package models

import (
	"time"
)

// AdminDashboard represents dashboard analytics and metrics
type AdminDashboard struct {
	// User Metrics
	TotalUsers          int64 `json:"total_users"`
	ActiveUsers         int64 `json:"active_users"`
	NewUsersToday       int64 `json:"new_users_today"`
	NewUsersWeek        int64 `json:"new_users_week"`
	NewUsersMonth       int64 `json:"new_users_month"`
	VerifiedUsers       int64 `json:"verified_users"`

	// Vendor Metrics
	TotalVendors        int64 `json:"total_vendors"`
	ActiveVendors       int64 `json:"active_vendors"`
	PendingVendors      int64 `json:"pending_vendors"`

	// Equipment Metrics
	TotalEquipment      int64 `json:"total_equipment"`
	ActiveEquipment     int64 `json:"active_equipment"`
	PendingEquipment    int64 `json:"pending_equipment"`

	// Booking Metrics
	TotalBookings       int64 `json:"total_bookings"`
	ActiveBookings      int64 `json:"active_bookings"`
	PendingBookings     int64 `json:"pending_bookings"`
	CompletedBookings   int64 `json:"completed_bookings"`
	CancelledBookings   int64 `json:"cancelled_bookings"`

	// Revenue Metrics
	TotalRevenue        float64 `json:"total_revenue"`
	RevenueToday        float64 `json:"revenue_today"`
	RevenueWeek         float64 `json:"revenue_week"`
	RevenueMonth        float64 `json:"revenue_month"`

	// Review Metrics
	TotalReviews        int64 `json:"total_reviews"`
	PendingReviews      int64 `json:"pending_reviews"`
	FlaggedReviews      int64 `json:"flagged_reviews"`
	AverageRating       float64 `json:"average_rating"`

	// Payment Metrics
	TotalTransactions   int64 `json:"total_transactions"`
	PendingPayouts      int64 `json:"pending_payouts"`
	CompletedPayouts    int64 `json:"completed_payouts"`

	// Support Metrics
	OpenTickets         int64 `json:"open_tickets"`
	PendingResponses    int64 `json:"pending_responses"`
	CriticalTickets     int64 `json:"critical_tickets"`
}

// AdminUser represents user management data for admin
type AdminUser struct {
	ID              uint64    `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Role            string    `json:"role"`
	UserType        string    `json:"user_type"`
	IsVerified      bool      `json:"is_verified"`
	IsActive        bool      `json:"is_active"`
	RiskScore       float64   `json:"risk_score"`
	CreatedAt       time.Time `json:"created_at"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	TotalBookings   int       `json:"total_bookings"`
	TotalSpent      float64   `json:"total_spent"`
	TotalRevenue    float64   `json:"total_revenue"` // For vendors
	Status          string    `json:"status"`
}

// AdminUserFilterOptions represents filtering options for admin user management
type AdminUserFilterOptions struct {
	Role         string `json:"role,omitempty"`
	UserType     string `json:"user_type,omitempty"`
	Status       string `json:"status,omitempty"`
	IsVerified   *bool  `json:"is_verified,omitempty"`
	SearchQuery  string `json:"search_query,omitempty"`
	CreatedAfter string `json:"created_after,omitempty"`
	CreatedBefore string `json:"created_before,omitempty"`
	SortBy       string `json:"sort_by,omitempty"`       // created_at, name, email, bookings, revenue
	SortOrder    string `json:"sort_order,omitempty"`    // asc, desc
	Page         int    `json:"page,omitempty"`
	PageSize     int    `json:"page_size,omitempty"`
}

// AdminActivity represents admin activity log
type AdminActivity struct {
	ID          uint64 `json:"id"`
	AdminID     uint64 `json:"admin_id"`
	AdminName   string `json:"admin_name"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
	ResourceID  uint64 `json:"resource_id,omitempty"`
	Details     string `json:"details,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ContentModeration represents content moderation queue items
type ContentModeration struct {
	ID            uint64 `json:"id"`
	ContentType   string `json:"content_type"` // review, equipment, user, comment
	ContentID     uint64 `json:"content_id"`
	ContentTitle  string `json:"content_title"`
	ContentData   string `json:"content_data,omitempty"`
	ReportedAt    time.Time `json:"reported_at"`
	ReportedBy    uint64 `json:"reported_by"`
	ReportCount   int    `json:"report_count"`
	Reason        string `json:"reason"`
	Status        string `json:"status"` // pending, reviewing, approved, rejected, actioned
	Priority      string `json:"priority"` // low, medium, high, critical
	AssignedTo    *uint64 `json:"assigned_to,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy    *uint64 `json:"reviewed_by,omitempty"`
	Action        string `json:"action,omitempty"` // approved, rejected, deleted, hidden, warning
	ActionDetails string `json:"action_details,omitempty"`
	Notes         string `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	Timestamp           time.Time `json:"timestamp"`
	CPUUsage           float64   `json:"cpu_usage"`
	MemoryUsage        float64   `json:"memory_usage"`
	MemoryTotal        float64   `json:"memory_total"`
	MemoryAvailable    float64   `json:"memory_available"`
	DiskUsage          float64   `json:"disk_usage"`
	DiskTotal          float64   `json:"disk_total"`
	DiskAvailable      float64   `json:"disk_available"`
	ActiveConnections  int       `json:"active_connections"`
	RequestsPerSecond  float64   `json:"requests_per_second"`
	AverageResponseTime float64  `json:"average_response_time"`
	ErrorRate          float64   `json:"error_rate"`
	DatabaseStatus     string    `json:"database_status"`
	CacheStatus        string    `json:"cache_status"`
	QueueStatus         string    `json:"queue_status"`
}

// AnalyticsData represents detailed analytics data
type AnalyticsData struct {
	Period            string                   `json:"period"` // daily, weekly, monthly
	StartDate         time.Time                `json:"start_date"`
	EndDate           time.Time                `json:"end_date"`
	UserGrowth        []AnalyticsDataPoint     `json:"user_growth"`
	BookingTrends     []AnalyticsDataPoint     `json:"booking_trends"`
	RevenueTrends     []AnalyticsDataPoint     `json:"revenue_trends"`
	EquipmentStats     EquipmentAnalytics       `json:"equipment_stats"`
	CategoryBreakdown  []CategoryAnalytics      `json:"category_breakdown"`
	UserDemographics   UserDemographics         `json:"user_demographics"`
	GeographicData     []GeoAnalytics           `json:"geographic_data"`
}

// AnalyticsDataPoint represents a single data point in analytics
type AnalyticsDataPoint struct {
	Date    time.Time `json:"date"`
	Value   float64   `json:"value"`
	Count   int64     `json:"count"`
	Change  float64   `json:"change"` // percentage change from previous period
	Label   string    `json:"label,omitempty"`
}

// EquipmentAnalytics represents equipment-related analytics
type EquipmentAnalytics struct {
	TotalEquipment     int64   `json:"total_equipment"`
	ActiveEquipment    int64   `json:"active_equipment"`
	MostViewed         []EquipmentStats `json:"most_viewed"`
	MostRented         []EquipmentStats `json:"most_rented"`
	HighestRated       []EquipmentStats `json:"highest_rated"`
	MostRevenue        []EquipmentStats `json:"most_revenue"`
	ByCategory         []CategoryEquipmentStats `json:"by_category"`
}

// EquipmentStats represents equipment statistics
type EquipmentStats struct {
	EquipmentID  uint64  `json:"equipment_id"`
	Name         string  `json:"name"`
	Views        int64   `json:"views"`
	Rentals      int64   `json:"rentals"`
	Rating       float64 `json:"rating"`
	Revenue      float64 `json:"revenue"`
	CategoryName string  `json:"category_name"`
}

// CategoryEquipmentStats represents equipment stats by category
type CategoryEquipmentStats struct {
	CategoryID   uint64  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	TotalCount   int64   `json:"total_count"`
	ActiveCount  int64   `json:"active_count"`
	TotalRentals int64   `json:"total_rentals"`
	AverageRating float64 `json:"average_rating"`
	TotalRevenue float64 `json:"total_revenue"`
}

// CategoryAnalytics represents category analytics
type CategoryAnalytics struct {
	CategoryID    uint64  `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	EquipmentCount int64   `json:"equipment_count"`
	BookingCount  int64   `json:"booking_count"`
	Revenue       float64 `json:"revenue"`
	Growth        float64 `json:"growth"`
}

// UserDemographics represents user demographic data
type UserDemographics struct {
	ByRole          map[string]int64 `json:"by_role"`
	ByUserType      map[string]int64 `json:"by_user_type"`
	ByCountry       map[string]int64 `json:"by_country"`
	ByCity          map[string]int64 `json:"by_city"`
	ByAgeGroup      map[string]int64 `json:"by_age_group"`
	ByVerification  map[string]int64 `json:"by_verification"`
}

// GeoAnalytics represents geographic analytics data
type GeoAnalytics struct {
	Country     string  `json:"country"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	UserCount   int64   `json:"user_count"`
	BookingCount int64 `json:"booking_count"`
	Revenue     float64 `json:"revenue"`
}

// AdminAction represents admin action permissions
type AdminAction struct {
	ID          uint64 `json:"id"`
	AdminID     uint64 `json:"admin_id"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
	GrantedAt   time.Time `json:"granted_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	GrantedBy   uint64 `json:"granted_by"`
	Reason      string `json:"reason,omitempty"`
}

// SystemAlert represents system alerts and notifications
type SystemAlert struct {
	ID          uint64 `json:"id"`
	AlertType   string `json:"alert_type"` // security, performance, revenue, content, system
	Severity    string `json:"severity"`    // info, warning, critical
	Title       string `json:"title"`
	Message     string `json:"message"`
	Metadata    string `json:"metadata,omitempty"` // JSON data
	IsResolved  bool   `json:"is_resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy  *uint64 `json:"resolved_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BulkAction represents bulk operations for admin
type BulkAction struct {
	ID           uint64 `json:"id"`
	AdminID      uint64 `json:"admin_id"`
	ActionType   string `json:"action_type"` // approve, reject, delete, verify, suspend
	TargetType   string `json:"target_type"` // users, equipment, bookings, reviews
	TargetIDs    string `json:"target_ids"` // JSON array
	Status       string `json:"status"` // pending, processing, completed, failed
	TotalTargets int    `json:"total_targets"`
	Processed    int    `json:"processed"`
	Successful   int    `json:"successful"`
	Failed       int    `json:"failed"`
	ErrorDetails string `json:"error_details,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ScheduledTask represents scheduled background tasks
type ScheduledTask struct {
	ID          uint64 `json:"id"`
	TaskName    string `json:"task_name"`
	Description string `json:"description"`
	TaskType    string `json:"task_type"` // report, cleanup, maintenance, analytics
	Schedule    string `json:"schedule"` // cron expression
	IsActive    bool   `json:"is_active"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	NextRunAt   *time.Time `json:"next_run_at,omitempty"`
	RunCount    int    `json:"run_count"`
	FailureCount int   `json:"failure_count"`
	LastStatus  string `json:"last_status"` // success, error, running
	LastError   string `json:"last_error,omitempty"`
	Config      string `json:"config,omitempty"` // JSON config
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReportDefinition represents custom report definitions
type ReportDefinition struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ReportType  string `json:"report_type"` // analytics, financial, user, equipment
	Query       string `json:"query"` // SQL query or template
	Parameters  string `json:"parameters,omitempty"` // JSON parameters definition
	Format      string `json:"format"` // json, csv, pdf, excel
	Schedule    string `json:"schedule,omitempty"` // cron expression
	Recipients  string `json:"recipients,omitempty"` // JSON array of emails
	CreatedBy   uint64 `json:"created_by"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReportExecution represents report execution history
type ReportExecution struct {
	ID              uint64 `json:"id"`
	ReportID        uint64 `json:"report_id"`
	ExecutedBy      uint64 `json:"executed_by"`
	Parameters      string `json:"parameters,omitempty"` // JSON
	Status          string `json:"status"` // pending, running, completed, failed
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	FileURL        string `json:"file_url,omitempty"`
	FileSize       int64 `json:"file_size,omitempty"`
	RowCount       int64 `json:"row_count,omitempty"`
	Error          string `json:"error,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// AdminRole represents admin role and permissions
type AdminRole struct {
	ID          uint64 `json:"id"`
	RoleName    string `json:"role_name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Permissions string `json:"permissions"` // JSON array of permission strings
	IsSystem    bool   `json:"is_system"` // System roles cannot be deleted
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AdminPermission represents granular admin permissions
type AdminPermission struct {
	ID          uint64 `json:"id"`
	Permission  string `json:"permission"`
	Category    string `json:"category"` // users, equipment, bookings, payments, content, system
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationTemplate represents notification templates
type NotificationTemplate struct {
	ID          uint64 `json:"id"`
	TemplateKey string `json:"template_key"`
	TemplateType string `json:"template_type"` // email, push, sms
	Subject     string `json:"subject,omitempty"`
	Body        string `json:"body"`
	Variables   string `json:"variables,omitempty"` // JSON array of variable names
	Language    string `json:"language"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
