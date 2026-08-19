# 🚀 GenRent Production Deployment Guide

This guide covers deploying GenRent to a production server using Docker Compose.

---

## ✅ Pre-Deployment Checklist

### 1. Server Requirements
- **OS**: Ubuntu 20.04+ or Debian 11+
- **RAM**: Minimum 2GB (4GB+ recommended)
- **CPU**: 2+ cores
- **Disk**: 20GB+ SSD
- **Ports**: 80 (HTTP), 443 (HTTPS) must be open

### 2. Domain Configuration
- Point your domain (e.g., `genrent.in`) to your server's IP address
- Wait for DNS propagation (can take up to 24 hours)
- Verify with: `dig genrent.in`

### 3. Prerequisites on Server
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

---

## 📋 Deployment Steps

### Step 1: Clone Repository
```bash
git clone <your-repo-url> /opt/genrent
cd /opt/genrent
```

### Step 2: Configure Environment
```bash
cd backend
cp .env.production .env

# IMPORTANT: Update these values in .env:
# - JWT_SECRET (already generated)
# - DB_PASSWORD (already generated)
# - RAZORPAY_KEY_ID (get from Razorpay dashboard)
# - RAZORPAY_KEY_SECRET (get from Razorpay dashboard)
# - DOMAIN (confirm your domain)
# - BASE_URL (confirm your domain with https://)
```

### Step 3: Update Payment Gateway Credentials
1. Log into [Razorpay Dashboard](https://dashboard.razorpay.com/)
2. Go to Settings → API Keys
3. Generate new key pair for production
4. Update `.env` with:
   ```
   RAZORPAY_KEY_ID=rzp_live_XXXXX
   RAZORPAY_KEY_SECRET=your_actual_live_secret
   RAZORPAY_WEBHOOK_SECRET=your_webhook_secret
   ```

### Step 4: Create Required Directories
```bash
sudo mkdir -p /var/log/caddy
sudo mkdir -p /var/backups/genrent
sudo chown -R $USER:$USER /var/log/caddy
sudo chown -R $USER:$USER /var/backups/genrent
```

### Step 5: Build and Deploy
```bash
cd /opt/genrent/backend

# Build with production env
docker-compose --env-file .env build --no-cache

# Start all services
docker-compose --env-file .env up -d

# Check status
docker-compose ps
```

### Step 6: Verify Deployment
```bash
# Check logs
docker-compose logs -f app

# Check health
curl https://genrent.in/health

# Expected response:
# {"status":"ok","database":"connected","version":"v1"}
```

### Step 7: Create Admin User
```bash
curl -X POST https://genrent.in/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name":"Admin",
    "email":"admin@genrent.in",
    "phone":"9999999999",
    "password":"YOUR_STRONG_PASSWORD",
    "role":"admin"
  }'
```

---

## 🔒 Security Checklist

- ✅ JWT_SECRET is 32+ characters (generated)
- ✅ DB_PASSWORD is strong (generated)
- ✅ BASE_URL uses HTTPS
- ✅ ALLOWED_ORIGINS is specific to your domain
- ✅ SSL mode is `require` for production database
- ⚠️ **IMPORTANT**: Never commit `.env` to git!

---

## 🔧 Maintenance

### View Logs
```bash
# Application logs
docker-compose logs -f app

# Caddy (HTTPS) logs
docker-compose logs -f caddy

# Database logs
docker-compose logs -f postgres
```

### Database Backup
```bash
# Manual backup
docker-compose exec postgres pg_dump -U genrent genrent | gzip > backup.sql.gz

# Restore backup
gunzip < backup.sql.gz | docker-compose exec -T postgres psql -U genrent genrent
```

### Update Application
```bash
cd /opt/genrent
git pull
docker-compose --env-file .env build --no-cache
docker-compose --env-file .env up -d
```

### Restart Services
```bash
docker-compose restart app
```

### Zero-Downtime Deployment
```bash
# Using the Makefile target
make deploy
```

---

## 📊 Monitoring

### Health Check Endpoint
```
https://genrent.in/health
```

### Application Metrics
```
https://genrent.in/metrics
```

---

## 🔍 Troubleshooting

### Issue: Caddy can't get SSL certificate
**Solution**: Ensure DNS is fully propagated and port 80/443 are open.

### Issue: Database connection fails
**Solution**: Check if postgres container is healthy:
```bash
docker-compose ps postgres
docker-compose logs postgres
```

### Issue: Emails not sending
**Solution**: Verify SMTP credentials and check Gmail security settings.

### Issue: "Port already in use"
**Solution**: Check what's using the port:
```bash
sudo lsof -i :80
sudo lsof -i :443
```

---

## 🎯 What Was Fixed

Before deployment, the following issues were resolved:

1. ✅ **Hardcoded localhost URLs in email templates** - Now uses dynamic BASE_URL
2. ✅ **Secure JWT_SECRET generated** - 32-character random secret
3. ✅ **Secure DB_PASSWORD generated** - Strong random password
4. ✅ **BASE_URL environment variable added** - For email links
5. ✅ **Updated .env.production with all required values**
6. ✅ **Updated .env.example for reference**

---

## 📝 Post-Deployment Tasks

1. **Setup database backup cron job**:
   ```bash
   # Edit crontab
   crontab -e
   
   # Add daily backup at 2 AM
   0 2 * * * cd /opt/genrent/backend && bash scripts/backup.sh >> /var/log/genrent-backup.log 2>&1
   ```

2. **Configure firewall**:
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

3. **Monitor disk space**:
   ```bash
   df -h
   docker system prune -a --volumes  # Clean up unused Docker resources
   ```

4. **Update webhook secret in Razorpay**:
   - Go to Razorpay Dashboard → Settings → Webhooks
   - Add webhook URL: `https://genrent.in/api/v1/webhooks/payment/razorpay`
   - Copy the webhook secret to your `.env` file

---

## 🆘 Emergency Recovery

### Rollback to Previous Version
```bash
git log --oneline  # Find previous commit hash
git checkout <commit-hash>
docker-compose --env-file .env build --no-cache
docker-compose --env-file .env up -d
```

### Full Reset (Last Resort)
```bash
# Stop all services
docker-compose down -v

# This deletes all data! Only do this if you have backups
# Restart
docker-compose --env-file .env up -d
```

---

**Last Updated**: 2026-08-19
**Version**: 1.0
