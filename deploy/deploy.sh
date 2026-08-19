#!/bin/bash
# ==============================================================================
# GenRent Deployment Script
# ==============================================================================
# This script deploys GenRent to production
# Usage: ./deploy.sh
# ==============================================================================

set -e

echo "🚀 Deploying GenRent..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
APP_DIR="/opt/genrent"
BACKEND_DIR="$APP_DIR/backend"
SECRETS_FILE="$APP_DIR/.secrets"

# ==============================================================================
# 1. Check Prerequisites
# ==============================================================================
echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"

# Check if secrets file exists
if [ ! -f "$SECRETS_FILE" ]; then
    echo -e "${RED}❌ Secrets file not found! Run vps-setup.sh first${NC}"
    exit 1
fi

# Load secrets
source "$SECRETS_FILE"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker not installed. Run vps-setup.sh first${NC}"
    exit 1
fi

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose not installed. Run vps-setup.sh first${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites OK${NC}"

# ==============================================================================
# 2. Create Production .env File
# ==============================================================================
echo -e "${YELLOW}📝 Creating production environment file...${NC}"

# Check if user provided domain
if [ -z "$DOMAIN" ]; then
    echo -e "${YELLOW}⚠️  No DOMAIN set, using default (change this later!)${NC}"
    DOMAIN="genrent.in"
fi

# Check if user provided email
if [ -z "$SMTP_USER" ]; then
    SMTP_USER="your@gmail.com"
fi

# Check if user provided SMTP password
if [ -z "$SMTP_PASS" ]; then
    SMTP_PASS="your_app_password"
fi

cat > "$BACKEND_DIR/.env" <<EOF
# ── App ──────────────────────────────────────────────
ENV=production
APP_PORT=8080
JWT_SECRET=${JWT_SECRET}
ALLOWED_ORIGINS=https://${DOMAIN},https://www.${DOMAIN}
BASE_URL=https://${DOMAIN}

# ── Database ─────────────────────────────────────────
DB_USER=genrent
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=genrent
DATABASE_URL=host=postgres user=genrent password=${DB_PASSWORD} dbname=genrent port=5432 sslmode=disable

# ── Email ───────────────────────────────────────────
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=${SMTP_USER}
SMTP_PASS=${SMTP_PASS}
SMTP_FROM=${SMTP_USER}
SMTP_FROM_NAME=GenRent

# ── Payment Gateway ──────────────────────────────────
PAYMENT_GATEWAY=razorpay
RAZORPAY_KEY_ID=\${RAZORPAY_KEY_ID}
RAZORPAY_KEY_SECRET=\${RAZORPAY_KEY_SECRET}
RAZORPAY_WEBHOOK_SECRET=\${RAZORPAY_WEBHOOK_SECRET}

# ── Domain ───────────────────────────────────────────
DOMAIN=${DOMAIN}

# ── Backup ───────────────────────────────────────────
BACKUP_DIR=/var/backups/genrent
RETAIN_DAYS=7

# ── Platform Fees ───────────────────────────────────
PLATFORM_FEE_PERCENT=10.0
REFUND_AUTO_PROCESS=true
PAYMENT_TIMEOUT=900
EOF

chmod 600 "$BACKEND_DIR/.env"
echo -e "${GREEN}✅ Environment file created${NC}"

# ==============================================================================
# 3. Build and Deploy
# ==============================================================================
echo -e "${YELLOW}🐳 Building Docker images...${NC}"

cd "$BACKEND_DIR"

# Stop existing containers if running
docker-compose down 2>/dev/null || true

# Build with no cache for fresh build
docker-compose build --no-cache

echo -e "${GREEN}✅ Docker images built${NC}"

# ==============================================================================
# 4. Start Services
# ==============================================================================
echo -e "${YELLOW}🚀 Starting services...${NC}"

# Start in detached mode
docker-compose up -d

# Wait for services to be healthy
echo -e "${YELLOW}⏳ Waiting for services to start...${NC}"
sleep 10

# Check if containers are running
if docker-compose ps | grep -q "Exit"; then
    echo -e "${RED}❌ Some containers failed to start${NC}"
    docker-compose ps
    exit 1
fi

echo -e "${GREEN}✅ Services started${NC}"

# ==============================================================================
# 5. Health Check
# ==============================================================================
echo -e "${YELLOW}🏥 Running health check...${NC}"

# Wait for app to be ready
MAX_ATTEMPTS=30
ATTEMPT=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Health check passed${NC}"
        HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
        echo "Health Response: $HEALTH_RESPONSE"
        break
    fi
    ATTEMPT=$((ATTEMPT + 1))
    echo "Waiting for app to be ready... ($ATTEMPT/$MAX_ATTEMPTS)"
    sleep 2
done

if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
    echo -e "${RED}❌ Health check failed${NC}"
    echo "Check logs with: docker-compose logs app"
    exit 1
fi

# ==============================================================================
# 6. Setup Backup Cron Job
# ==============================================================================
echo -e "${YELLOW}💾 Setting up automated backups...${NC}"

# Create backup script
cat > "$BACKUP_DIR/scripts/auto-backup.sh" <<'EOF'
#!/bin/bash
BACKUP_DIR="/var/backups/genrent"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
mkdir -p "$BACKUP_DIR"

docker exec genrent-db pg_dump -U genrent genrent | gzip > "$BACKUP_DIR/genrent_$TIMESTAMP.sql.gz"
find "$BACKUP_DIR" -name "genrent_*.sql.gz" -mtime +7 -delete
echo "[$(date)] Backup completed: genrent_$TIMESTAMP.sql.gz"
EOF

chmod +x "$BACKEND_DIR/scripts/auto-backup.sh"

# Add to cron
(crontab -l 2>/dev/null | grep -v "auto-backup"; echo "0 2 * * * $BACKEND_DIR/scripts/auto-backup.sh >> /var/log/genrent-backup.log 2>&1") | crontab -

echo -e "${GREEN}✅ Automated backups configured (daily at 2 AM)${NC}"

# ==============================================================================
# 7. Display Success Message
# ==============================================================================
echo ""
echo -e "${GREEN}"
echo "============================================"
echo "✅ Deployment Successful!"
echo "============================================"
echo -e "${NC}"
echo "Application Status:"
docker-compose ps
echo ""
echo "View logs:"
echo "  docker-compose logs -f app"
echo ""
echo "View Caddy logs:"
echo "  docker-compose logs -f caddy"
echo ""
echo "View database logs:"
echo "  docker-compose logs -f postgres"
echo ""
echo "Health check:"
echo "  curl http://localhost:8080/health"
echo ""
echo "📝 Next Steps:"
echo "1. Configure Cloudflare DNS to point to this VPS"
echo "2. Update payment gateway credentials in .env"
echo "3. Create admin user using:"
echo "   curl -X POST http://localhost:8080/api/v1/auth/register \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"name\":\"Admin\",\"email\":\"admin@${DOMAIN}\",\"phone\":\"9999999999\",\"password\":\"YOUR_PASSWORD\",\"role\":\"admin\"}'"
echo ""
echo -e "${YELLOW}⚠️  IMPORTANT: Update these in $BACKEND_DIR/.env:${NC}"
echo "  - RAZORPAY_KEY_ID"
echo "  - RAZORPAY_KEY_SECRET"
echo "  - RAZORPAY_WEBHOOK_SECRET"
echo "  - SMTP_USER (if using different email)"
echo "  - SMTP_PASS (Gmail app password)"
echo ""
