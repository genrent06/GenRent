# 📦 GenRent Deployment Package

Complete deployment solution for GenRent with **free Cloudflare + free VPS**.

---

## 🎯 Overview

This deployment package provides:
- ✅ Free VPS setup (Oracle Cloud)
- ✅ Cloudflare DNS and SSL configuration
- ✅ Automated deployment scripts
- ✅ Production-ready configuration
- ✅ Automated backups
- ✅ Health monitoring

---

## 📁 Files in This Directory

| File | Description |
|------|-------------|
| `vps-setup.sh` | One-command VPS initialization |
| `deploy.sh` | Application deployment script |
| `QUICKSTART.md` | 30-minute quick start guide |
| `CLOUDFLARE_SETUP.md` | Complete Cloudflare configuration |
| `ORACLE_CLOUD_GUIDE.md` | Oracle Cloud signup & setup |

---

## 🚀 Quick Deployment

### Option 1: Automated (Recommended)

```bash
# 1. Get VPS (Oracle Cloud Free Tier)
# See ORACLE_CLOUD_GUIDE.md

# 2. On your VPS, run setup:
curl -fsSL https://your-repo/deploy/vps-setup.sh | bash

# 3. Configure Cloudflare DNS (see CLOUDFLARE_SETUP.md)

# 4. Deploy application:
cd /opt/genrent
./deploy/deploy.sh
```

### Option 2: Manual

See `QUICKSTART.md` for detailed step-by-step instructions.

---

## 🔧 Scripts Overview

### vps-setup.sh

**What it does:**
- Updates system packages
- Installs Docker & Docker Compose
- Configures firewall
- Creates application directories
- Generates secure secrets (JWT, DB password)
- Optimizes system settings
- Sets up swap if low memory

**Usage:**
```bash
curl -fsSL https://your-repo/deploy/vps-setup.sh | bash
```

### deploy.sh

**What it does:**
- Creates production .env file
- Builds Docker images
- Starts all services (app, database, Caddy)
- Runs health checks
- Sets up automated backups
- Displays success message

**Usage:**
```bash
cd /opt/genrent
./deploy/deploy.sh
```

---

## 📋 Prerequisites

### For VPS:
- Ubuntu 20.04+ or Debian 11+
- At least 1GB RAM
- 10GB storage
- Public IP address

### For Domain:
- Domain name (e.g., genrent.in)
- Access to domain's DNS settings

### For Cloudflare:
- Cloudflare account (free)

---

## 🎯 Deployment Architecture

```
┌─────────────────┐
│  Cloudflare     │  (Free DNS, SSL, CDN)
│  (CDN/Proxy)    │
└────────┬────────┘
         │ HTTPS
         ▼
┌─────────────────┐
│  VPS (Oracle)   │  (Free Tier)
│  ┌───────────┐  │
│  │ Caddy     │  │  (Reverse Proxy)
│  │ (Port 443)│  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────┴─────┐  │
│  │ App       │  │  (Go Backend)
│  │ (Port 8080)│ │
│  └───────────┘  │
│        │        │
│  ┌─────┴─────┐  │
│  │PostgreSQL │  │  (Database)
│  │ (Port 5432)│ │
│  └───────────┘  │
└─────────────────┘
```

---

## 🔒 Security Features

- ✅ Generated secure secrets (JWT, DB passwords)
- ✅ Firewall configured (only 80, 443, 22)
- ✅ SSL/TLS encryption (Cloudflare)
- ✅ DDoS protection (Cloudflare)
- ✅ Rate limiting (application level)
- ✅ Docker isolation

---

## 📊 Monitoring

### Application Health:
```bash
curl https://genrent.in/health
```

### Application Logs:
```bash
ssh ubuntu@YOUR_IP
cd /opt/genrent/backend
docker-compose logs -f app
```

### Database Backups:
```bash
# Automated: Daily at 2 AM
# Manual: docker exec genrent-db pg_dump -U genrent genrent | gzip > backup.sql.gz
```

---

## 🔄 Updates & Maintenance

### Update Application:
```bash
ssh ubuntu@YOUR_IP
cd /opt/genrent
git pull
cd backend
docker-compose build --no-cache
docker-compose up -d
```

### View Logs:
```bash
ssh ubuntu@YOUR_IP
cd /opt/genrent/backend
docker-compose logs -f
```

### Restart Services:
```bash
ssh ubuntu@YOUR_IP
cd /opt/genrent/backend
docker-compose restart app
```

### Check Resources:
```bash
ssh ubuntu@YOUR_IP
free -h    # Memory
df -h      # Disk
docker ps  # Containers
```

---

## 🆘 Troubleshooting

### Application won't start:
```bash
# Check logs
docker-compose logs app

# Check environment variables
cat .env

# Rebuild
docker-compose build --no-cache
docker-compose up -d
```

### Database connection failed:
```bash
# Check postgres is running
docker-compose ps postgres

# Check postgres logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres psql -U genrent -d genrent
```

### Can't access website:
```bash
# Check VPS is running
ssh ubuntu@YOUR_IP

# Check Docker containers
docker ps

# Check Cloudflare DNS
dig genrent.in

# Check firewall
sudo ufw status
```

---

## 📈 Performance Optimization

The setup includes:
- ✅ Swap configuration for low-memory VPS
- ✅ File descriptor limits increased
- ✅ TCP optimizations
- ✅ Cloudflare CDN caching
- ✅ Docker resource limits
- ✅ Automated backups to prevent disk fill

---

## 💰 Cost Summary

| Item | Cost | Notes |
|------|------|-------|
| VPS (Oracle Cloud) | **FREE** | 1 CPU, 1GB RAM |
| Cloudflare DNS | **FREE** | Forever free |
| Cloudflare SSL | **FREE** | Automatic renewal |
| Cloudflare CDN | **FREE** | Global CDN |
| Domain Name | ~$10-15/year | From registrar |
| **TOTAL** | **~$10-15/year** | Domain only |

---

## 🎓 Documentation

| Document | Description |
|----------|-------------|
| QUICKSTART.md | Quick 30-minute deployment guide |
| CLOUDFLARE_SETUP.md | Complete Cloudflare configuration |
| ORACLE_CLOUD_GUIDE.md | Oracle Cloud signup and VPS creation |
| DEPLOYMENT_FIXES_SUMMARY.md | What was fixed for production |

---

## ✅ Deployment Checklist

- [ ] Oracle Cloud account created
- [ ] VPS instance created
- [ ] VPS accessible via SSH
- [ ] vps-setup.sh script run
- [ ] Code uploaded to VPS
- [ ] deploy.sh script run
- [ ] Application responding
- [ ] Health check passing
- [ ] Cloudflare account created
- [ ] Domain added to Cloudflare
- [ ] Nameservers updated
- [ ] DNS A record created
- [ ] DNS propagated
- [ ] HTTPS working
- [ ] Admin user created
- [ ] Payment gateway configured

---

## 🚀 Ready to Deploy?

Start here: [QUICKSTART.md](QUICKSTART.md)

---

**Version:** 1.0
**Last Updated:** 2026-08-19
