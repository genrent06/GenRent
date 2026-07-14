# GenRent API Documentation

Complete API reference for the GenRent Construction Equipment Rental Platform.

**Base URL:** `http://localhost:8080/api/v1` (development) | `https://your-domain.com/api/v1` (production)

**API Version:** v1

---

## Table of Contents

- [Authentication](#authentication)
- [Public Endpoints](#public-endpoints)
- [Equipment & Categories](#equipment--categories)
- [Generators](#generators)
- [Vendors](#vendors)
- [Bookings](#bookings)
- [Payments](#payments)
- [Vendor Wallet](#vendor-wallet)
- [Notifications](#notifications)
- [Disputes](#disputes)
- [Admin Endpoints](#admin-endpoints)
- [Webhooks](#webhooks)
- [Error Responses](#error-responses)

---

## Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Login
Get JWT token by providing credentials.

```http
POST /api/v1/auth/login
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "user@example.com",
    "role": "customer"
  }
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "invalid credentials"
}
```

---

### Register
Create a new user account.

```http
POST /api/v1/auth/register
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "1234567890",
  "password": "SecurePassword123!",
  "role": "customer"
}
```

**Roles:** `customer` | `vendor` | `admin`

**Response (201 Created):**
```json
{
  "message": "registration successful",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "customer"
  }
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": "email already exists"
}
```

---

### Get Profile
Get current user profile.

```http
GET /api/v1/auth/profile
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "1234567890",
  "role": "customer",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### Forgot Password
Request password reset email.

```http
POST /api/v1/auth/forgot-password
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response (200 OK):**
```json
{
  "message": "if the email exists, a reset link has been sent"
}
```

---

### Reset Password
Reset password using token from email.

```http
POST /api/v1/auth/reset-password
Content-Type: application/json
```

**Request Body:**
```json
{
  "token": "reset-token-from-email",
  "newPassword": "NewSecurePassword123!"
}
```

**Response (200 OK):**
```json
{
  "message": "password reset successfully"
}
```

---

## Public Endpoints

### Health Check
Check API and database health.

```http
GET /health
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "database": "connected",
  "version": "v1"
}
```

---

### API Documentation
Get structured API documentation (this page).

```http
GET /docs
```

---

### Platform Metrics
Get platform statistics (public but rate-limited).

```http
GET /metrics
```

**Response (200 OK):**
```json
{
  "total_users": 1250,
  "total_vendors": 85,
  "total_equipment": 450,
  "active_bookings": 45,
  "total_revenue": 450000.00
}
```

---

## Equipment & Categories

### Get All Categories
List all equipment categories.

```http
GET /api/v1/categories
```

**Response (200 OK):**
```json
{
  "categories": [
    {
      "id": 1,
      "name": "Generators",
      "parent_category_id": null,
      "description": "Power generators for construction",
      "icon_url": "/static/icons/generator.png"
    },
    {
      "id": 2,
      "name": "Excavators",
      "parent_category_id": null,
      "description": "Heavy excavation equipment"
    }
  ]
}
```

---

### Get Category Hierarchy
Get categories with nested subcategories.

```http
GET /api/v1/categories/hierarchy
```

**Response (200 OK):**
```json
{
  "categories": [
    {
      "id": 1,
      "name": "Generators",
      "children": [
        {
          "id": 11,
          "name": "Diesel Generators"
        },
        {
          "id": 12,
          "name": "Gas Generators"
        }
      ]
    }
  ]
}
```

---

### Get Popular Categories
Get categories with most equipment.

```http
GET /api/v1/categories/popular
```

---

### Get Category Details
Get specific category information.

```http
GET /api/v1/categories/:id
```

**Response (200 OK):**
```json
{
  "id": 1,
  "name": "Generators",
  "description": "Power generators for construction sites",
  "equipment_count": 45
}
```

---

### Get Category Equipment
List equipment in a category.

```http
GET /api/v1/categories/:id/equipment?city=Delhi
```

**Query Parameters:**
- `city` (optional): Filter by city
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response (200 OK):**
```json
{
  "equipment": [
    {
      "id": 101,
      "name": "Generator 50KVA",
      "daily_price": 1500,
      "vendor": {
        "id": 5,
        "company_name": "Power Rentals Co",
        "city": "Delhi"
      }
    }
  ],
  "total": 45,
  "page": 1
}
```

---

### Search Equipment
Search equipment with filters.

```http
GET /api/v1/equipment/search?city=Delhi&category=1&min_price=1000&max_price=5000
```

**Query Parameters:**
- `city` (optional): Filter by city
- `category` (optional): Category ID
- `min_price` (optional): Minimum daily price
- `max_price` (optional): Maximum daily price
- `available` (optional): Show only available (true/false)
- `page` (optional): Page number
- `limit` (optional): Items per page

**Response (200 OK):**
```json
{
  "equipment": [
    {
      "id": 101,
      "name": "Generator 50KVA Diesel",
      "brand": "Cummins",
      "model": "50KVA-QS",
      "description": "Heavy duty generator",
      "daily_price": 1500,
      "weekly_price": 9000,
      "monthly_price": 35000,
      "mobilization_fee": 500,
      "demobilization_fee": 500,
      "total_quantity": 5,
      "available_quantity": 3,
      "location": "Industrial Area",
      "city": "Delhi",
      "latitude": 28.7041,
      "longitude": 77.1025,
      "availability_status": "available",
      "image_url": "/static/images/gen50kva.jpg",
      "specs": {
        "capacity_kva": 50,
        "fuel_type": "diesel",
        "fuel_consumption": "5 L/hr"
      },
      "vendor": {
        "id": 5,
        "company_name": "Power Rentals Co",
        "average_rating": 4.5
      }
    }
  ],
  "total": 23,
  "page": 1
}
```

---

### Get Equipment Details
Get full details of specific equipment.

```http
GET /api/v1/equipment/:id
```

**Response (200 OK):**
```json
{
  "id": 101,
  "name": "Generator 50KVA Diesel",
  "brand": "Cummins",
  "model": "50KVA-QS",
  "description": "Heavy duty generator for construction sites",
  "daily_price": 1500,
  "weekly_price": 9000,
  "monthly_price": 35000,
  "mobilization_fee": 500,
  "demobilization_fee": 500,
  "total_quantity": 5,
  "available_quantity": 3,
  "location": "Industrial Area, Sector 5",
  "city": "Delhi",
  "latitude": 28.7041,
  "longitude": 77.1025,
  "availability_status": "available",
  "image_url": "/static/images/gen50kva.jpg",
  "specs": {
    "capacity_kva": 50,
    "fuel_type": "diesel",
    "fuel_consumption": "5 L/hr",
    "power_output": "50 KVA",
    "phase": "3 Phase"
  },
  "category": {
    "id": 1,
    "name": "Generators"
  },
  "vendor": {
    "id": 5,
    "company_name": "Power Rentals Co",
    "verified": true,
    "reliability_score": 4.8,
    "total_ratings": 45,
    "average_rating": 4.5
  },
  "created_at": "2024-01-10T10:00:00Z"
}
```

---

### Create Equipment (Vendor Only)
Create new equipment listing.

```http
POST /api/v1/equipment
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "category_id": 1,
  "name": "Generator 100KVA",
  "brand": "Cummins",
  "model": "100KVA-PG",
  "description": "Heavy duty generator",
  "daily_price": 2500,
  "weekly_price": 15000,
  "monthly_price": 60000,
  "mobilization_fee": 1000,
  "demobilization_fee": 1000,
  "total_quantity": 3,
  "available_quantity": 3,
  "location": "Industrial Area",
  "city": "Mumbai",
  "latitude": 19.076,
  "longitude": 72.8777,
  "image_url": "/static/images/gen100kva.jpg",
  "specs": {
    "capacity_kva": 100,
    "fuel_type": "diesel",
    "power_output": "100 KVA"
  }
}
```

**Response (201 Created):**
```json
{
  "id": 102,
  "message": "equipment created successfully"
}
```

---

### Update Equipment (Vendor Only)
Update existing equipment.

```http
PUT /api/v1/equipment/:id
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:** Same as Create Equipment

**Response (200 OK):**
```json
{
  "message": "equipment updated successfully"
}
```

---

### Delete Equipment (Vendor Only)
Delete equipment listing.

```http
DELETE /api/v1/equipment/:id
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "message": "equipment deleted successfully"
}
```

---

### Get My Equipment (Vendor Only)
Get all equipment listed by current vendor.

```http
GET /api/v1/equipment/mine
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "equipment": [
    {
      "id": 101,
      "name": "Generator 50KVA",
      "availability_status": "available",
      "total_bookings": 12
    }
  ]
}
```

---

### Update Equipment Status (Vendor Only)
Update equipment availability status.

```http
PUT /api/v1/equipment/:id/status
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "status": "maintenance"
}
```

**Status values:** `available` | `rented` | `maintenance` | `reserved`

**Response (200 OK):**
```json
{
  "message": "status updated successfully"
}
```

---

### Get Equipment Stats (Vendor Only)
Get booking statistics for vendor's equipment.

```http
GET /api/v1/equipment-stats
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "total_equipment": 15,
  "available": 12,
  "rented": 3,
  "total_bookings": 45,
  "this_month_bookings": 8,
  "revenue_this_month": 45000.00
}
```

---

## Generators

**Note:** Generators are being migrated to the Equipment system. These endpoints are for legacy support.

### Search Generators
Search generators with filters.

```http
GET /api/v1/generators?city=Delhi&min_capacity=10
```

**Query Parameters:**
- `city` (optional): Filter by city
- `min_capacity` (optional): Minimum KVA
- `max_capacity` (optional): Maximum KVA
- `fuel_type` (optional): diesel | petrol | gas
- `available` (optional): Show only available

---

### Get Generator Details
```http
GET /api/v1/generators/:id
```

---

### Create Generator (Vendor Only)
```http
POST /api/v1/generators
Authorization: Bearer <vendor-token>
```

---

### Update Generator (Vendor Only)
```http
PUT /api/v1/generators/:id
Authorization: Bearer <vendor-token>
```

---

### Delete Generator (Vendor Only)
```http
DELETE /api/v1/generators/:id
Authorization: Bearer <vendor-token>
```

---

### Get My Generators (Vendor Only)
```http
GET /api/v1/generators/mine
Authorization: Bearer <vendor-token>
```

---

## Vendors

### List Vendors
List all verified vendors.

```http
GET /api/v1/vendors?city=Delhi
```

**Query Parameters:**
- `city` (optional): Filter by city
- `verified` (optional): Show only verified
- `page` (optional): Page number

**Response (200 OK):**
```json
{
  "vendors": [
    {
      "id": 5,
      "company_name": "Power Rentals Co",
      "city": "Delhi",
      "verified": true,
      "reliability_score": 4.8,
      "total_equipment": 15,
      "average_rating": 4.5
    }
  ],
  "total": 45
}
```

---

### Get Vendor by ID
Get vendor profile details.

```http
GET /api/v1/vendors/:id
```

**Response (200 OK):**
```json
{
  "id": 5,
  "company_name": "Power Rentals Co",
  "description": "Leading generator rental company",
  "city": "Delhi",
  "verified": true,
  "reliability_score": 4.8,
  "total_bookings": 150,
  "successful_deliveries": 145,
  "cancelled_bookings": 5,
  "average_rating": 4.5,
  "total_ratings": 89,
  "total_equipment": 15
}
```

---

### Create Vendor Profile
Create vendor profile (after registering as vendor).

```http
POST /api/v1/vendors
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "company_name": "Power Rentals Co",
  "address": "123 Industrial Area",
  "city": "Delhi",
  "latitude": 28.7041,
  "longitude": 77.1025,
  "description": "Leading generator rental company",
  "phone": "+919876543210"
}
```

**Response (201 Created):**
```json
{
  "id": 5,
  "message": "vendor profile created successfully"
}
```

---

### Get My Vendor Profile
Get current vendor's profile.

```http
GET /api/v1/vendors/me
Authorization: Bearer <vendor-token>
```

---

### Update Vendor Profile
Update vendor profile.

```http
PUT /api/v1/vendors/me
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:** Same as Create Vendor Profile

**Response (200 OK):**
```json
{
  "message": "vendor profile updated successfully"
}
```

---

## Bookings

### Create Booking
Create a new booking request.

```http
POST /api/v1/bookings
Authorization: Bearer <customer-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "equipment_id": 101,
  "start_date": "2024-02-01T10:00:00Z",
  "end_date": "2024-02-10T10:00:00Z",
  "address": "Construction Site, Sector 62",
  "notes": "Need generator for 10 days",
  "mobilization_required": true
}
```

**Response (201 Created):**
```json
{
  "id": 501,
  "status": "requested",
  "total_price": 19000,
  "rental_price": 15000,
  "mobilization_fee": 1000,
  "demobilization_fee": 1000,
  "advance_amount": 5000,
  "equipment": {
    "id": 101,
    "name": "Generator 50KVA"
  },
  "vendor": {
    "id": 5,
    "company_name": "Power Rentals Co"
  }
}
```

---

### Get My Bookings
Get current user's bookings.

```http
GET /api/v1/bookings
Authorization: Bearer <token>
```

**Query Parameters:**
- `status` (optional): Filter by status
- `page` (optional): Page number

**Response (200 OK):**
```json
{
  "bookings": [
    {
      "id": 501,
      "status": "confirmed",
      "start_date": "2024-02-01T10:00:00Z",
      "end_date": "2024-02-10T10:00:00Z",
      "total_price": 19000,
      "equipment": {
        "id": 101,
        "name": "Generator 50KVA",
        "image_url": "/static/images/gen50kva.jpg"
      },
      "vendor": {
        "id": 5,
        "company_name": "Power Rentals Co",
        "phone": "+919876543210"
      }
    }
  ],
  "total": 5
}
```

---

### Get Booking Details
Get specific booking details.

```http
GET /api/v1/bookings/:id
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": 501,
  "status": "confirmed",
  "start_date": "2024-02-01T10:00:00Z",
  "end_date": "2024-02-10T10:00:00Z",
  "total_price": 19000,
  "rental_price": 15000,
  "mobilization_fee": 1000,
  "demobilization_fee": 1000,
  "advance_amount": 5000,
  "advance_paid": true,
  "address": "Construction Site, Sector 62",
  "notes": "Need generator for 10 days",
  "delivery_otp": "123456",
  "otp_verified": false,
  "created_at": "2024-01-25T10:00:00Z",
  "equipment": {
    "id": 101,
    "name": "Generator 50KVA",
    "brand": "Cummins"
  },
  "vendor": {
    "id": 5,
    "company_name": "Power Rentals Co",
    "phone": "+919876543210"
  }
}
```

---

### Get Booking Status
Lightweight endpoint for polling booking status.

```http
GET /api/v1/bookings/:id/status
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": 501,
  "status": "dispatched",
  "updated_at": "2024-01-26T14:30:00Z"
}
```

---

### Accept Booking (Vendor Only)
Vendor accepts a booking request.

```http
POST /api/v1/bookings/:id/accept
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "message": "booking accepted",
  "booking": {
    "id": 501,
    "status": "confirmed",
    "accepted_at": "2024-01-26T10:00:00Z"
  }
}
```

---

### Reject Booking (Vendor Only)
Vendor rejects a booking request.

```http
POST /api/v1/bookings/:id/reject
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "reason": "Equipment not available for these dates"
}
```

---

### Dispatch Equipment (Vendor Only)
Vendor dispatches equipment with OTP.

```http
POST /api/v1/bookings/:id/dispatch
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "message": "equipment dispatched",
  "otp": "123456",
  "booking": {
    "id": 501,
    "status": "dispatched",
    "dispatched_at": "2024-02-01T08:00:00Z"
  }
}
```

---

### Resend OTP (Vendor Only)
Resend delivery OTP.

```http
POST /api/v1/bookings/:id/resend-otp
Authorization: Bearer <vendor-token>
```

---

### Confirm Delivery (Customer)
Customer confirms delivery with OTP.

```http
POST /api/v1/bookings/:id/confirm-delivery
Authorization: Bearer <customer-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "otp": "123456"
}
```

**Response (200 OK):**
```json
{
  "message": "delivery confirmed",
  "booking": {
    "id": 501,
    "status": "delivered",
    "delivered_at": "2024-02-01T10:00:00Z"
  }
}
```

---

### Initiate Return (Vendor Only)
Vendor initiates equipment return.

```http
POST /api/v1/bookings/:id/initiate-return
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "message": "return initiated",
  "return_otp": "654321",
  "booking": {
    "id": 501,
    "status": "return_initiated",
    "return_initiated_at": "2024-02-10T16:00:00Z"
  }
}
```

---

### Confirm Return (Vendor Only)
Vendor confirms equipment return with OTP.

```http
POST /api/v1/bookings/:id/confirm-return
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "otp": "654321"
}
```

**Response (200 OK):**
```json
{
  "message": "return confirmed",
  "booking": {
    "id": 501,
    "status": "completed",
    "completed_at": "2024-02-10T17:00:00Z"
  }
}
```

---

### Complete Booking (Customer)
Customer marks booking as complete (after return).

```http
POST /api/v1/bookings/:id/complete
Authorization: Bearer <customer-token>
```

---

### Cancel Booking
Cancel a booking.

```http
POST /api/v1/bookings/:id/cancel
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "reason": "No longer needed"
}
```

**Response (200 OK):**
```json
{
  "message": "booking cancelled",
  "refund_amount": 5000
}
```

---

### Submit Review
Submit rating and review for completed booking.

```http
POST /api/v1/bookings/:id/review
Authorization: Bearer <customer-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "rating": 5,
  "review": "Excellent service, equipment was in good condition"
}
```

**Response (200 OK):**
```json
{
  "message": "review submitted successfully"
}
```

---

### Upload Handover Photos
Upload handover/checklist photos.

```http
POST /api/v1/bookings/:id/handover
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**Form Data:**
- `type`: delivery | return
- `photos`: Photo files (multiple)
- `notes`: Optional notes

**Response (200 OK):**
```json
{
  "message": "handover photos uploaded",
  "id": 789
}
```

---

### Get Handovers
Get handover records for booking.

```http
GET /api/v1/bookings/:id/handover
Authorization: Bearer <token>
```

---

### Raise Damage Dispute
Raise a damage dispute for equipment.

```http
POST /api/v1/bookings/:id/dispute
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "description": "Equipment returned with damaged fuel tank",
  "claimed_amount": 5000,
  "photo_urls": [
    "/static/uploads/damage1.jpg",
    "/static/uploads/damage2.jpg"
  ]
}
```

**Response (201 Created):**
```json
{
  "id": 301,
  "message": "dispute raised successfully",
  "status": "open"
}
```

---

### Get My Disputes
Get current user's disputes.

```http
GET /api/v1/disputes
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "disputes": [
    {
      "id": 301,
      "booking_id": 501,
      "status": "investigating",
      "claimed_amount": 5000,
      "created_at": "2024-02-12T10:00:00Z"
    }
  ]
}
```

---

## Payments

### Get Payment Details
Get payment details for a booking.

```http
GET /api/v1/payments/booking/:booking_id
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": 601,
  "booking_id": 501,
  "total_amount": 19000,
  "advance_amount": 5000,
  "vendor_amount": 17100,
  "platform_fee": 1900,
  "status": "completed",
  "paid_at": "2024-01-26T10:00:00Z",
  "method": "upi"
}
```

---

### Process Payment
Process payment for a booking.

```http
POST /api/v1/payments
Authorization: Bearer <customer-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "booking_id": 501,
  "amount": 5000,
  "method": "upi",
  "transaction_id": "TXN123456789"
}
```

**Response (200 OK):**
```json
{
  "message": "payment processed successfully",
  "payment": {
    "id": 601,
    "status": "completed"
  }
}
```

---

## Vendor Wallet

### Get Wallet Balance
Get vendor wallet details.

```http
GET /api/v1/wallet
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "balance": 45000.00,
  "hold_balance": 5000.00,
  "withdrawal_hold_balance": 0,
  "available_for_withdrawal": 40000.00,
  "currency": "INR"
}
```

---

### Request Withdrawal
Request withdrawal of funds.

```http
POST /api/v1/wallet/withdraw
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "amount": 10000,
  "bank_account_id": 5
}
```

**Response (200 OK):**
```json
{
  "id": 801,
  "message": "withdrawal requested",
  "otp": "999888",
  "expires_at": "2024-02-12T11:30:00Z"
}
```

---

### Confirm Withdrawal OTP
Confirm withdrawal with OTP.

```http
POST /api/v1/wallet/withdraw/:id/confirm
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "otp": "999888"
}
```

**Response (200 OK):**
```json
{
  "message": "withdrawal confirmed and submitted for approval"
}
```

---

### Get Withdrawals
Get vendor's withdrawal requests.

```http
GET /api/v1/wallet/withdrawals
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "withdrawals": [
    {
      "id": 801,
      "amount": 10000,
      "status": "approved",
      "created_at": "2024-02-12T10:00:00Z",
      "processed_at": "2024-02-12T14:00:00Z"
    }
  ]
}
```

---

### Get Bank Accounts
Get vendor's saved bank accounts.

```http
GET /api/v1/wallet/bank-accounts
Authorization: Bearer <vendor-token>
```

**Response (200 OK):**
```json
{
  "bank_accounts": [
    {
      "id": 5,
      "bank_name": "HDFC Bank",
      "account_no": "XXXXXXXX1234",
      "ifsc": "HDFC0001234",
      "account_name": "Power Rentals Co",
      "is_primary": true
    }
  ]
}
```

---

### Save Bank Account
Add or update bank account.

```http
POST /api/v1/wallet/bank-accounts
Authorization: Bearer <vendor-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "bank_name": "HDFC Bank",
  "account_no": "50100234567890",
  "ifsc": "HDFC0001234",
  "account_name": "Power Rentals Co",
  "is_primary": true
}
```

**Response (201 Created):**
```json
{
  "id": 5,
  "message": "bank account saved"
}
```

---

### Delete Bank Account
Delete saved bank account.

```http
DELETE /api/v1/wallet/bank-accounts/:id
Authorization: Bearer <vendor-token>
```

---

## Notifications

### Get Notifications
Get user's notifications.

```http
GET /api/v1/notifications
Authorization: Bearer <token>
```

**Query Parameters:**
- `read` (optional): Filter by read status (true/false)
- `limit` (optional): Number of notifications

**Response (200 OK):**
```json
{
  "notifications": [
    {
      "id": 901,
      "type": "booking_accepted",
      "title": "Booking Accepted",
      "message": "Your booking #501 has been accepted by Power Rentals Co",
      "read": false,
      "created_at": "2024-01-26T10:00:00Z",
      "booking_id": 501
    }
  ]
}
```

---

### Mark Notification Read
Mark specific notification as read.

```http
POST /api/v1/notifications/:id/read
Authorization: Bearer <token>
```

---

### Mark All Notifications Read
Mark all notifications as read.

```http
POST /api/v1/notifications/read-all
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "message": "all notifications marked as read"
}
```

---

## Audit & Activity

### Get My Activity
Get current user's audit trail.

```http
GET /api/v1/my-activity
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "activities": [
    {
      "id": 1001,
      "action": "booking_created",
      "entity_type": "booking",
      "entity_id": 501,
      "created_at": "2024-01-25T10:00:00Z",
      "ip_address": "192.168.1.100"
    }
  ]
}
```

---

## Admin Endpoints

All admin endpoints require `admin` role.

### List All Vendors
```http
GET /api/v1/admin/vendors
Authorization: Bearer <admin-token>
```

---

### Verify Vendor
```http
PUT /api/v1/admin/vendors/:id/verify
Authorization: Bearer <admin-token>
```

**Request Body:**
```json
{
  "verified": true
}
```

---

### Reject Vendor Application
```http
PUT /api/v1/admin/vendors/:id/reject
Authorization: Bearer <admin-token>
```

---

### Penalize Vendor
```http
PUT /api/v1/admin/vendors/:id/penalize
Authorization: Bearer <admin-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "reason": "Late delivery",
  "penalty_score": 0.5
}
```

---

### List All Generators
```http
GET /api/v1/admin/generators
Authorization: Bearer <admin-token>
```

---

### Update Generator Status
```http
PUT /api/v1/admin/generators/:id/status
Authorization: Bearer <admin-token>
```

---

### List All Bookings
```http
GET /api/v1/admin/bookings
Authorization: Bearer <admin-token>
```

**Query Parameters:**
- `status` (optional): Filter by status
- `vendor_id` (optional): Filter by vendor
- `customer_id` (optional): Filter by customer

---

### Force Cancel Booking
```http
POST /api/v1/admin/bookings/:id/force-cancel
Authorization: Bearer <admin-token>
```

---

### Release Escrow
Release vendor funds from escrow.

```http
POST /api/v1/admin/bookings/:id/release-escrow
Authorization: Bearer <admin-token>
```

---

### Refund Customer
Process refund for customer.

```http
POST /api/v1/admin/bookings/:id/refund
Authorization: Bearer <admin-token>
```

---

### Update Booking Status
Override booking status.

```http
PUT /api/v1/admin/bookings/:id/status
Authorization: Bearer <admin-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "status": "completed"
}
```

---

### Get Platform Stats
```http
GET /api/v1/admin/stats
Authorization: Bearer <admin-token>
```

**Response (200 OK):**
```json
{
  "total_users": 1250,
  "total_vendors": 85,
  "verified_vendors": 72,
  "total_equipment": 450,
  "active_bookings": 45,
  "completed_bookings": 523,
  "total_revenue": 450000.00,
  "pending_withdrawals": 5,
  "pending_disputes": 3
}
```

---

### Get Audit Logs
Get system audit logs.

```http
GET /api/v1/admin/audit-logs
Authorization: Bearer <admin-token>
```

**Query Parameters:**
- `user_id` (optional): Filter by user
- `action` (optional): Filter by action
- `entity_type` (optional): Filter by entity type
- `start_date` (optional): Filter from date
- `end_date` (optional): Filter to date

---

### List Withdrawals
Get all withdrawal requests.

```http
GET /api/v1/admin/withdrawals
Authorization: Bearer <admin-token>
```

**Query Parameters:**
- `status` (optional): pending | approved | rejected | processing | completed

---

### Approve Withdrawal
Approve vendor withdrawal request.

```http
POST /api/v1/admin/withdrawals/:id/approve
Authorization: Bearer <admin-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "note": "Verified and approved"
}
```

---

### Reject Withdrawal
Reject vendor withdrawal request.

```http
POST /api/v1/admin/withdrawals/:id/reject
Authorization: Bearer <admin-token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "reason": "Insufficient funds or verification pending"
}
```

---

## Webhooks

### Payment Webhook
Handle payment gateway notifications.

```http
POST /api/v1/webhooks/payment
Content-Type: application/json
```

**Request Body:**
```json
{
  "event": "payment.success",
  "data": {
    "transaction_id": "TXN123456789",
    "amount": 5000,
    "booking_id": 501,
    "status": "completed"
  }
}
```

**Response (200 OK):**
```json
{
  "message": "webhook processed successfully"
}
```

---

## Error Responses

All endpoints return consistent error responses.

### Error Response Format

```json
{
  "error": "error message description"
}
```

### Common HTTP Status Codes

| Status | Description |
|--------|-------------|
| 200 OK | Request successful |
| 201 Created | Resource created successfully |
| 400 Bad Request | Invalid request parameters |
| 401 Unauthorized | Missing or invalid authentication |
| 403 Forbidden | Insufficient permissions |
| 404 Not Found | Resource not found |
| 409 Conflict | Resource conflict (e.g., duplicate) |
| 422 Unprocessable Entity | Validation error |
| 429 Too Many Requests | Rate limit exceeded |
| 500 Internal Server Error | Server error |

### Common Error Messages

| Error | Description | Solution |
|-------|-------------|-----------|
| `invalid credentials` | Wrong email or password | Check login details |
| `missing authorization header` | No auth token provided | Add Authorization header |
| `invalid token` | JWT token is invalid | Login again to get new token |
| `token expired` | JWT token has expired | Login again to get new token |
| `insufficient permissions` | User lacks required role | Check user role |
| `email already exists` | Email already registered | Use different email |
| `equipment not found` | Equipment doesn't exist | Check equipment ID |
| `booking not found` | Booking doesn't exist | Check booking ID |
| `insufficient balance` | Not enough wallet balance | Add funds |
| `rate limit exceeded` | Too many requests | Wait and retry |

---

## Rate Limiting

Some endpoints have rate limiting applied:

| Endpoint | Limit | Period |
|----------|-------|--------|
| `/api/v1/auth/*` | 10 requests | 60 seconds |
| `/api/v1/bookings/*` | 30 requests | 60 seconds |
| `/api/v1/payments/*` | 10 requests | 60 seconds |
| `/api/v1/webhooks/payment` | 60 requests | 60 seconds |

Rate limit response (429):
```json
{
  "error": "rate limit exceeded",
  "retry_after": 30
}
```

---

## Pagination

List endpoints support pagination:

**Query Parameters:**
- `page` (default: 1) - Page number
- `limit` (default: 20) - Items per page (max: 100)

**Response includes:**
```json
{
  "items": [...],
  "total": 150,
  "page": 1,
  "limit": 20,
  "pages": 8
}
```

---

## Testing with cURL

### Example: Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Example: Create Booking
```bash
curl -X POST http://localhost:8080/api/v1/bookings \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "equipment_id": 101,
    "start_date": "2024-02-01T10:00:00Z",
    "end_date": "2024-02-10T10:00:00Z",
    "address": "Construction Site"
  }'
```

---

## SDK Integration Examples

### JavaScript/TypeScript
```typescript
const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

// Get my bookings
const { data } = await api.get('/bookings');
```

### Go
```go
client := &http.Client{
    BaseURL: "http://localhost:8080/api/v1",
}

req, _ := client.NewRequest("GET", "/bookings", nil)
req.Header.Set("Authorization", "Bearer " + token)
```

---

**Last Updated:** January 2024
**API Version:** v1.0
