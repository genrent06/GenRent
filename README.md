# GenRent - Construction Equipment Rental Platform

A full-stack marketplace platform for renting construction equipment, connecting vendors with customers. Built with Go (Gin), PostgreSQL, and vanilla JavaScript.

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue)](https://golang.org)
[![Database](https://img.shields.io/badge/Database-PostgreSQL-336791)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Project Setup Guide](#project-setup-guide)
- [Environment Configuration](#environment-configuration)
- [Development](#development)
- [API Documentation](#api-documentation)
- [User Roles](#user-roles)
- [Deployment](#deployment)
- [Contributing](#contributing)

---

## Overview

GenRent is a construction equipment rental marketplace that enables:
- **Vendors** to list equipment (generators, excavators, cranes, etc.) for rent
- **Customers** to browse, book, and pay for equipment rentals
- **Admins** to manage vendors, bookings, payments, and disputes

The platform handles the complete rental lifecycle from listing → booking → payment → delivery → return → dispute resolution.

---

## Features

### Core Features
- **Equipment Listings**: Vendors can list equipment with photos, specifications, and pricing
- **Search & Filter**: Search equipment by category, location, price, and availability
- **Booking System**: Complete booking flow with OTP verification for delivery/return
- **Payment Processing**: Integrated payment handling with platform fees
- **Vendor Wallets**: Automatic fund management and withdrawal system
- **Ratings & Reviews**: Customer ratings for vendors and equipment
- **Dispute Resolution**: structured damage dispute handling with evidence
- **Audit Trail**: Complete audit logging for all actions

### Security Features
- JWT-based authentication
- Password hashing with bcrypt
- OTP verification for equipment delivery/return
- Role-based access control (Customer, Vendor, Admin)
- Rate limiting and request timeout protection
- Security headers (CSP, X-Frame-Options, etc.)
- CORS configuration

### User Features
- **Customers**: Browse equipment, place bookings, track orders, manage profile
- **Vendors**: Manage inventory, accept bookings, track earnings, request withdrawals
- **Admins**: Dashboard overview, manage vendors, handle disputes, view metrics

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                             │
│  (HTML/CSS/JS) - Static pages served by Gin                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Backend API                              │
│  (Go + Gin Framework)                                        │
│  ┌─────────────┬──────────────┬──────────────────────────┐  │
│  │  Handlers   │  Middleware  │       Services           │  │
│  │  - Auth     │  - CORS      │  - Email                 │  │
│  │  - Booking  │  - Auth      │  - Workers (expiry)      │  │
│  │  - Vendor   │  - Rate Lim  │                          │  │
│  │  - Payment  │  - Security  │                          │  │
│  │  - Equipment│  - Recovery  │                          │  │
│  └─────────────┴──────────────┴──────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     PostgreSQL                                │
│  - Users, Vendors, Equipment, Bookings, Payments           │
│  - Wallets, Transactions, Withdrawals, Disputes            │
└─────────────────────────────────────────────────────────────┘
```

---

## Tech Stack

### Backend
- **Language**: Go 1.25+
- **Framework**: Gin Web Framework
- **ORM**: GORM
- **Database**: PostgreSQL 15
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: bcrypt (golang.org/x/crypto)
- **Config**: godotenv

### Frontend
- **HTML5/CSS3/JavaScript** (Vanilla, no frameworks)
- **Static files served by Gin**

### Infrastructure
- **Docker** & **Docker Compose**
- **Caddy** (Reverse proxy with automatic HTTPS)

### DevOps
- **Make** for build automation
- **Air** for live reload (dev)

---

## Project Structure

```
genrent/
├── backend/
│   ├── cmd/
│   │   └── main.go              # Application entry point
│   ├── internal/
│   │   ├── apierr/              # API error handling
│   │   ├── config/              # Configuration management
│   │   ├── database/            # Database connection
│   │   ├── handlers/            # HTTP request handlers
│   │   │   ├── auth.go          # Authentication endpoints
│   │   │   ├── booking.go       # Booking management
│   │   │   ├── equipment.go     # Equipment CRUD
│   │   │   ├── vendor.go        # Vendor operations
│   │   │   ├── payment.go       # Payment processing
│   │   │   └── ...
│   │   ├── middleware/          # HTTP middleware
│   │   │   ├── auth.go          # JWT authentication
│   │   │   ├── cors.go          # CORS handling
│   │   │   ├── ratelimit.go     # Rate limiting
│   │   │   ├── security.go      # Security headers
│   │   │   └── ...
│   │   ├── migrate/             # Database migrations
│   │   │   ├── migrate.go       # Migration runner
│   │   │   └── *.sql            # SQL migration files
│   │   ├── models/              # GORM models
│   │   │   ├── user.go
│   │   │   ├── vendor.go
│   │   │   ├── equipment.go
│   │   │   ├── booking.go
│   │   │   └── ...
│   │   ├── services/
│   │   │   └── email/           # Email service
│   │   └── workers/             # Background workers
│   │       └── expiry.go        # Equipment expiry worker
│   ├── docker-compose.yml       # Docker services definition
│   ├── Dockerfile               # Container image
│   ├── Caddyfile                # Caddy reverse proxy config
│   ├── Makefile                 # Build/run commands
│   ├── .env                     # Environment variables (local)
│   ├── .env.example             # Environment template
│   └── go.mod                   # Go dependencies
├── frontend/
│   ├── index.html               # Home page
│   ├── login.html               # Login page
│   ├── register.html            # Registration
│   ├── vendor-dashboard.html    # Vendor dashboard
│   ├── admin-dashboard.html     # Admin dashboard
│   ├── my-bookings.html         # Customer bookings
│   ├── add-equipment.html       # Add equipment form
│   ├── forgot-password.html     # Password reset request
│   ├── reset-password.html      # Password reset form
│   ├── booking.html             # Booking details
│   ├── payment.html             # Payment page
│   ├── css/                     # Stylesheets
│   └── js/                      # JavaScript modules
├── Makefile                     # Root makefile (delegates to backend)
├── go.work                      # Go workspace
└── README.md                    # This file
```

---

## Project Setup Guide

This guide walks you through setting up the GenRent development environment from scratch.

### System Requirements

| Component | Minimum Version | Recommended |
|-----------|----------------|-------------|
| Go | 1.25 | Latest stable |
| PostgreSQL | 15 | 15+ |
| Docker | 20.10+ | Latest |
| Docker Compose | 2.0+ | Latest |
| Make | Any | Latest |
| RAM | 4 GB | 8 GB+ |
| Disk Space | 2 GB | 5 GB+ |

### Option 1: Docker Setup (Recommended for Beginners)

#### Step 1: Install Docker

**macOS:**
```bash
brew install --cask docker
# Or download from: https://www.docker.com/products/docker-desktop
```

**Ubuntu/Debian:**
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

**Windows:**
Download Docker Desktop from: https://www.docker.com/products/docker-desktop

#### Step 2: Clone the Repository

```bash
git clone <repository-url>
cd genrent
```

#### Step 3: Start PostgreSQL Container

```bash
make docker-up
```

Expected output:
```
PostgreSQL started on port 5432
Database: genrent | User: postgres | Password: postgres
```

#### Step 4: Configure Environment Variables

```bash
cd backend
cp .env.example .env
```

Edit `.env` with your preferred editor:
```bash
nano .env
# or
code .env
```

**Minimum Required Configuration:**
```env
DATABASE_URL=host=localhost user=postgres password=postgres dbname=genrent port=5433 sslmode=disable
JWT_SECRET=genrent-secret-key-change-this-in-production
PORT=8080
ENV=development
```

#### Step 5: Install Go Dependencies

```bash
go mod download
```

#### Step 6: Run the Application

```bash
make run
```

Expected output:
```
Running SQL migrations...
  Applying 001_schema...
  ✓ Applied 001_schema
✓ SQL migrations completed successfully

[GIN-debug] Listening and serving HTTP on :8080
```

#### Step 7: Verify Installation

Open your browser:
- **Application**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **API Docs**: http://localhost:8080/docs

You should see:
```json
{
  "status": "ok",
  "database": "connected",
  "version": "v1"
}
```

#### Step 8: Create Admin User

In a new terminal:
```bash
make seed-admin
```

Output:
```
Admin created: admin@genrent.in / admin123
```

#### Step 9: Login as Admin

1. Navigate to: http://localhost:8080/login.html
2. Enter credentials:
   - Email: `admin@genrent.in`
   - Password: `admin123`
3. You should be redirected to the admin dashboard

---

### Option 2: Local PostgreSQL Setup (For Advanced Users)

#### Step 1: Install Go

**macOS:**
```bash
brew install go
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install golang-go
```

**Windows:**
Download from: https://go.dev/dl/

Verify installation:
```bash
go version
# Should output: go version go1.25.x ...
```

#### Step 2: Install PostgreSQL

**macOS:**
```bash
brew install postgresql@15
brew services start postgresql@15
```

**Ubuntu/Debian:**
```bash
sudo apt install postgresql-15 postgresql-contrib-15
sudo systemctl start postgresql
```

**Windows:**
Download from: https://www.postgresql.org/download/windows/

#### Step 3: Create Database and User

```bash
# Switch to postgres user
sudo -u postgres psql

# In psql, run:
CREATE DATABASE genrent;
CREATE USER genrent WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE genrent TO genrent;
\q
```

#### Step 4: Clone and Configure

```bash
git clone <repository-url>
cd genrent/backend
cp .env.example .env
```

Edit `.env`:
```env
DATABASE_URL=host=localhost user=genrent password=your_password dbname=genrent port=5432 sslmode=disable
JWT_SECRET=your-secret-key-min-32-chars
PORT=8080
ENV=development
```

#### Step 5: Install Dependencies

```bash
go mod download
```

#### Step 6: Run the Application

```bash
make run
```

#### Step 7: Create Admin User

```bash
make seed-admin
```

---

### Option 3: Full Docker Stack (Production-like Setup)

This runs everything in containers - database, app, and reverse proxy.

#### Step 1: Clone Repository

```bash
git clone <repository-url>
cd genrent/backend
```

#### Step 2: Configure Environment

```bash
cp .env.example .env
nano .env
```

Set production-like values:
```env
ENV=development
DATABASE_URL=host=postgres user=postgres password=postgres dbname=genrent port=5432 sslmode=disable
JWT_SECRET=development-secret-key
ALLOWED_ORIGINS=http://localhost:8080
```

#### Step 3: Build and Start All Services

```bash
docker compose up -d
```

Expected output:
```
[+] Running 3/3
 ✔ Network genrent_default      Created
 ✔ Container genrent-db          Started
 ✔ Container genrent-app         Started
```

#### Step 4: View Logs

```bash
docker compose logs -f app
```

#### Step 5: Create Admin User

```bash
docker compose exec app make seed-admin
```

---

### Verify Your Setup

After completing any of the above options, verify everything is working:

#### 1. Check Database Connection

```bash
curl http://localhost:8080/health
```

Should return:
```json
{
  "status": "ok",
  "database": "connected",
  "version": "v1"
}
```

#### 2. Check API Documentation

```bash
curl http://localhost:8080/docs
```

#### 3. Test Registration

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","phone":"1234567890","password":"test123","role":"customer"}'
```

Should return:
```json
{
  "message": "User registered successfully"
}
```

#### 4. Test Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

Should return a JWT token.

---

### Development Tools Setup

#### Install Air (Live Reload)

```bash
go install github.com/cosmtrek/air@latest
```

Then run with live reload:
```bash
make dev
```

#### Install Make (if not available)

**macOS:**
```bash
brew install make
```

**Ubuntu/Debian:**
```bash
sudo apt install build-essential
```

---

### Common Setup Issues & Solutions

#### Issue 1: Port 8080 Already in Use

**Error:**
```
bind: address already in use
```

**Solution:**
```bash
# Find process using port 8080
lsof -ti:8080

# Kill it
kill -9 $(lsof -ti:8080)

# Or change port in .env
PORT=8081
```

#### Issue 2: Database Connection Refused

**Error:**
```
failed to connect to database: connection refused
```

**Solution:**
```bash
# Check PostgreSQL is running
docker ps | grep postgres
# or
sudo systemctl status postgresql

# Check port
nc -zv localhost 5432
```

#### Issue 3: Module Download Errors

**Error:**
```
go: module ... not found
```

**Solution:**
```bash
# Initialize go workspace
go work init

# Try again
go mod download
```

#### Issue 4: Migration Failures

**Error:**
```
migration 001_schema failed
```

**Solution:**
```bash
# Drop and recreate database
dropdb genrent
createdb genrent

# Restart application
make run
```

---

### Next Steps After Setup

1. **Explore the API Documentation**
   - Visit: http://localhost:8080/docs

2. **Register Test Users**
   - Create a customer account
   - Create a vendor account

3. **List Equipment** (as vendor)
   ```bash
   # First, get vendor JWT from login
   TOKEN="your-vendor-jwt"
   
   curl -X POST http://localhost:8080/api/v1/equipment \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Generator 50KVA",
       "category_id": 1,
       "daily_price": 1500,
       "city": "Delhi",
       "location": "Industrial Area"
     }'
   ```

4. **Test Booking Flow** (as customer)
   - Browse equipment
   - Create a booking
   - Verify OTP flow

5. **Check Admin Dashboard**
   - Visit: http://localhost:8080/admin-dashboard.html
   - View platform metrics
   - Manage vendors

---

---

## Environment Configuration

### Required Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DATABASE_URL` | PostgreSQL connection string | See below | `host=localhost user=postgres password=postgres dbname=genrent port=5432 sslmode=disable` |
| `JWT_SECRET` | Secret for JWT signing | ⚠️ Insecure default | `your-32-char-random-secret` |
| `PORT` | Server port | `8080` | `8080` |
| `ENV` | Environment | `development` | `production` |

### Optional Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `ALLOWED_ORIGINS` | CORS allowed origins | `*` | `https://genrent.in,https://app.genrent.in` |
| `SMTP_HOST` | SMTP server host | - | `smtp.gmail.com` |
| `SMTP_PORT` | SMTP server port | `25` | `587` |
| `SMTP_USER` | SMTP username | - | `your@gmail.com` |
| `SMTP_PASS` | SMTP password | - | `your-app-password` |
| `SMTP_FROM` | From email address | `noreply@genrent.com` | `noreply@genrent.com` |
| `SMTP_FROM_NAME` | From name | `GenRent` | `GenRent` |

### Environment Files

- `.env` - Local development (gitignored)
- `.env.example` - Template with all variables
- `.env.development` - Development defaults
- `.env.staging` - Staging environment
- `.env.production` - Production environment

**⚠️ Security Warning**: Never commit `.env` files with real credentials!

---

## Development

### Available Make Commands

```bash
# Run the server (migrations run automatically)
make run

# Build binary
make build

# Install/update dependencies
make tidy

# Start PostgreSQL (Docker)
make docker-up

# Start all services (PostgreSQL + App + Caddy)
make docker-all

# Stop Docker services
make docker-down

# Create admin user
make seed-admin

# Format code
make fmt

# Run with live reload (requires air)
make dev
```

### Development Mode with Live Reload

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with auto-reload
make dev
```

### Running Migrations

Migrations run automatically on application startup. The system:
1. Creates `schema_migrations` table if not exists
2. Reads all `.sql` files from `internal/migrate/`
3. Executes any new migrations in order
4. Records applied migrations

To manually reset migrations:
```sql
DELETE FROM schema_migrations WHERE version = '001_schema';
-- Restart application to re-run
```

### Database Schema

The database includes tables for:
- Users (customers, vendors, admins)
- Vendors (extended profiles)
- Equipment Categories (hierarchical)
- Equipment (rental items)
- Generators (legacy, being migrated)
- Bookings (rental orders)
- Payments (payment records)
- Vendor Wallets (vendor funds)
- Wallet Transactions (transaction history)
- Vendor Bank Accounts (withdrawal details)
- Withdrawal Requests (withdrawal management)
- Platform Revenues (platform fees)
- Booking Handovers (delivery/return docs)
- Damage Disputes (dispute management)
- Notifications (user notifications)
- Audit Logs (audit trail)
- Password Resets (reset tokens)

For detailed schema documentation, see [backend/internal/migrate/README.md](backend/internal/migrate/README.md)

---

## API Documentation

For complete API documentation including all endpoints, request/response examples, authentication, and error handling, see:

**[📘 Full API Documentation](API_DOCUMENTATION.md)**

### Quick Reference

#### Base URLs
- Development: `http://localhost:8080/api/v1`
- Production: `https://your-domain.com/api/v1`

#### Authentication
Most endpoints require JWT authentication:
```
Authorization: Bearer <your-jwt-token>
```

#### Key Endpoints

| Endpoint | Method | Description | Auth |
|----------|--------|-------------|------|
| `/health` | GET | Health check | No |
| `/docs` | GET | API documentation (JSON) | No |
| `/metrics` | GET | Platform statistics | No |
| `/auth/register` | POST | Register new user | No |
| `/auth/login` | POST | Login & get token | No |
| `/auth/forgot-password` | POST | Request password reset | No |
| `/auth/reset-password` | POST | Reset password | No |
| `/auth/profile` | GET | Get user profile | Yes |
| `/equipment` | GET | Search equipment | No |
| `/equipment/:id` | GET | Get equipment details | No |
| `/equipment` | POST | Create equipment | Vendor |
| `/bookings` | POST | Create booking | Customer |
| `/bookings` | GET | Get my bookings | Yes |
| `/bookings/:id/accept` | POST | Accept booking | Vendor |
| `/wallet` | GET | Get wallet balance | Vendor |
| `/wallet/withdraw` | POST | Request withdrawal | Vendor |
| `/admin/stats` | GET | Platform statistics | Admin |

#### Example Requests

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

**Create Booking:**
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

For detailed API documentation with all endpoints, schemas, and examples, refer to **[API_DOCUMENTATION.md](API_DOCUMENTATION.md)**.

---

## User Roles

### Customer
- Browse and search equipment
- Create bookings
- Make payments
- Track booking status
- Leave ratings and reviews
- Request password reset

### Vendor
- Register as vendor (requires verification)
- List equipment for rent
- Accept/reject booking requests
- View earnings and wallet balance
- Request withdrawals
- Manage bank accounts
- View ratings and reviews

### Admin
- Dashboard with platform metrics
- Verify vendor applications
- Handle withdrawal requests
- Manage disputes
- View audit logs
- Access all bookings and vendors

---

## Deployment

### Docker Deployment (Production)

The project includes a production-ready Docker Compose setup with:
- PostgreSQL database
- Go application
- Caddy reverse proxy (automatic HTTPS)

#### Deployment Steps

1. **Prepare Production Environment**
   ```bash
   # Copy production env template
   cp backend/.env.production backend/.env
   
   # Edit with production values
   nano backend/.env
   ```

2. **Deploy**
   ```bash
   make deploy
   ```

3. **Configure Caddy** (if needed)
   Edit `Caddyfile` with your domain, then:
   ```bash
   make caddy-reload
   ```

4. **View Logs**
   ```bash
   make logs        # Application logs
   make logs-db     # Database logs
   make logs-caddy  # Caddy logs
   ```

### Manual Deployment

1. **Build Binary**
   ```bash
   cd backend
   go build -o bin/genrent ./cmd/main.go
   ```

2. **Set Up Database**
   ```bash
   # Create database and user
   createdb genrent
   ```

3. **Configure Environment**
   ```bash
   export DATABASE_URL="host=localhost user=genrent password=secret dbname=genrent port=5432 sslmode=require"
   export JWT_SECRET="your-32-char-random-secret"
   export ENV="production"
   # ... other variables
   ```

4. **Run Application**
   ```bash
   ./bin/genrent
   ```

5. **Set Up Reverse Proxy** (nginx, Caddy, etc.)

---

## Testing

### Manual Testing Flow

A test script is included for manual testing:

```bash
./test_flow.sh
```

This tests:
1. User registration
2. Vendor registration
3. Login
4. Equipment listing
5. Booking creation
6. OTP verification
7. Payment processing

---

## Troubleshooting

### Database Connection Issues

```
Error: failed to connect to database: connection refused
```

**Solution**: Ensure PostgreSQL is running:
```bash
make docker-up
# or
docker ps | grep postgres
```

### Migration Failures

```
Error: migration 001_schema failed
```

**Solution**: Check the migration file for syntax errors, or manually reset:
```sql
DROP TABLE schema_migrations;
-- Restart application
```

### CORS Errors

**Solution**: Set `ALLOWED_ORIGINS` in `.env`:
```
ALLOWED_ORIGINS=http://localhost:8080,https://yourdomain.com
```

### JWT Issues

**Solution**: Ensure `JWT_SECRET` is set and consistent across restarts. In production, use a strong 32+ character secret.

---

## Security Considerations

### Before Deploying to Production

1. **Change JWT_SECRET**
   ```bash
   # Generate a secure secret
   openssl rand -base64 32
   ```

2. **Set Strong Database Password**
   ```bash
   # In .env
   DB_PASSWORD=<strong-password>
   ```

3. **Configure SMTP for Email**
   - Use Gmail App Passwords or a dedicated email service
   - Don't use your main Gmail password

4. **Review CORS Settings**
   ```bash
   ALLOWED_ORIGINS=https://yourdomain.com
   ```

5. **Enable HTTPS**
   - The Docker setup includes Caddy for automatic TLS
   - Ensure your domain DNS points to the server

6. **Database SSL**
   ```bash
   DATABASE_URL=host=... sslmode=require
   ```

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format code
- Write meaningful commit messages
- Add tests for new features

---

## License

This project is licensed under the MIT License.

---

## Contact & Support

For questions, issues, or contributions:
- Open an issue on GitHub
- Contact the development team

---

## Roadmap

### Planned Features

#### 🚀 High Priority
- [ ] **Payment Gateway Integration** 📘 [Detailed Implementation Plan](ROADMAP_PAYMENT_GATEWAY.md)
  - Razorpay integration for Indian market
  - Stripe integration for international payments
  - UPI payment support
  - Escrow system improvements
  - Auto-refund processing

- [ ] **Advanced Search & Filtering**
  - Full-text search with Elasticsearch
  - Advanced filters (price range, ratings, location radius)
  - Saved search queries
  - Search suggestions and autocomplete
  - Filter by availability dates

- [ ] **Real-time Notifications (WebSocket)**
  - Live booking status updates
  - Instant chat between vendors and customers
  - Real-time equipment availability
  - Push notifications for mobile
  - Notification preferences

#### 🎯 Medium Priority
- [ ] **Equipment Maintenance Tracking**
  - Maintenance schedule management
  - Service history tracking
  - Automated maintenance reminders
  - Downtime reporting
  - Maintenance cost tracking

- [ ] **Advanced Analytics Dashboard**
  - Vendor performance analytics
  - Customer booking patterns
  - Revenue trends and forecasts
  - Equipment utilization reports
  - Geographic heat maps
  - Export reports (PDF, Excel)

- [ ] **Vendor Verification Improvements**
  - Document upload (GST, PAN, Aadhaar)
  - Automated verification with government APIs
  - Video KYC integration
  - Business verification checks
  - Rating system improvements

#### 📱 Mobile & Platform
- [ ] **Mobile Applications**
  - React Native for iOS
  - React Native for Android
  - Offline mode support
  - Push notifications
  - Location-based equipment search

- [ ] **Multi-language Support**
  - Hindi, Tamil, Telugu support
  - Regional language for Indian states
  - Currency localization
  - Date/time format localization

#### 🛡️ Insurance & Safety
- [ ] **Equipment Insurance Options**
  - Integration with insurance providers
  - Automated insurance quotes
  - Damage coverage options
  - Claim processing workflow
  - Insurance policy management

#### ⚖️ Dispute & Resolution
- [ ] **Automated Dispute Resolution**
  - AI-powered dispute assessment
  - Automated evidence analysis
  - Suggested resolution recommendations
  - Escalation workflows
  - Mediation chat system

#### 🔄 Future Enhancements
- [ ] **Equipment Comparison Tool**
  - Side-by-side equipment comparison
  - Price comparison across vendors
  - Feature comparison matrix

- [ ] **Bulk Booking System**
  - Multiple equipment booking
  - Corporate rental plans
  - Long-term contract discounts

- [ ] **Vendor Subscription Plans**
  - Premium listing plans
  - Featured equipment options
  - Analytics access tiers

- [ ] **Customer Loyalty Program**
  - Points system for bookings
  - Referral bonuses
  - Discount coupons
  - Membership tiers

- [ ] **Equipment Rental Marketplace**
  - Peer-to-peer equipment sharing
  - Equipment rental bidding
  - Last-minute deals

### Technology Upgrades

#### Backend
- [ ] GraphQL API alternative
- [ ] Redis caching layer
- [ ] Message queue (RabbitMQ/Kafka)
- [ ] Microservices architecture
- [ ] API rate limiting per user

#### Frontend
- [ ] Migration to React/Vue.js
- [ ] Progressive Web App (PWA)
- [ ] Image optimization and CDN
- [ ] SEO optimization

#### DevOps
- [ ] CI/CD pipeline optimization
- [ ] Automated testing suite
- [ ] Docker swarm/Kubernetes
- [ ] Monitoring & alerting (Prometheus/Grafana)
- [ ] Log aggregation (ELK stack)

### Known Issues

See GitHub Issues for a list of known issues and feature requests.

### Contributing to Roadmap

We welcome community contributions! If you're interested in working on any of these features:

1. Check existing [GitHub Issues](https://github.com/your-repo/issues)
2. Comment on the issue you want to work on
3. Create a fork and submit a Pull Request
4. Join our community discussions

---

**Built with ❤️ for the construction industry**
