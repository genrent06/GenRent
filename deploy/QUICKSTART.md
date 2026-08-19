# 🚀 GenRent Quick Start Deployment Guide

Complete guide to deploy GenRent with **Cloudflare (Free)** + **Oracle Cloud Free Tier VPS**.

---

## ⏱️ Total Time: ~30 Minutes

1. Oracle Cloud Setup: 10 minutes
2. VPS Configuration: 10 minutes
3. Cloudflare DNS: 5 minutes
4. GenRent Deployment: 5 minutes

---

## 📋 Prerequisites

- ✅ Domain name (e.g., genrent.in)
- ✅ Credit card (for verification only, not charged)
- ✅ SSH client on your computer
- ✅ Basic terminal knowledge

---

## 🎯 Quick Start Steps

### 1️⃣ Get Free VPS (Oracle Cloud)

```bash
# Go to: https://oracle.com/cloud/free
# Sign up, verify email/phone
# Create account (takes 2-5 minutes)
```

See full guide: [ORACLE_CLOUD_GUIDE.md](ORACLE_CLOUD_GUIDE.md)

### 2️⃣ Create VPS Instance

Once logged in:
1. Menu → Compute → Instances → Create Instance
2. Configure:
   - Name: `genrent-app`
   - Image: `Ubuntu 22.04`
   - Shape: `VM.Standard.E2.1.Micro` (Free)
3. Click Create
4. Note the **Public IP** when ready

### 3️⃣ Connect to VPS

```bash
# Generate SSH key
ssh-keygen -t rsa -b 4096 -f ~/.ssh/oracle_cloud

# Connect
ssh -i ~/.ssh/oracle_cloud ubuntu@YOUR_PUBLIC_IP
```

### 4️⃣ Run Setup Script

```bash
# On your VPS:
curl -fsSL https://raw.githubusercontent.com/your-repo/deploy/vps-setup.sh | bash
```

This installs:
- Docker
- Docker Compose
- Firewall
- Required tools

### 5️⃣ Upload Your Code

```bash
# From your computer:
scp -r backend/ ubuntu@YOUR_PUBLIC_IP:/opt/genrent/

# Or clone directly on VPS:
ssh ubuntu@YOUR_PUBLIC_IP
cd /opt/genrent
git clone YOUR_REPO_URL .
```

### 6️⃣ Deploy GenRent

```bash
# On your VPS:
cd /opt/genrent
chmod +x deploy/deploy.sh
./deploy/deploy.sh
```

### 7️⃣ Configure Cloudflare

1. Log in to [cloudflare.com](https://dash.cloudflare.com/)
2. Add your domain
3. Update nameservers at your domain registrar
4. Add DNS A record pointing to your VPS IP:
   ```
   Type: A
   Name: @
   IPv4: YOUR_VPS_IP
   Proxy: ON (☁️)
   ```

See full guide: [CLOUDFLARE_SETUP.md](CLOUDFLARE_SETUP.md)

### 8️⃣ Verify Deployment

```bash
# Health check
curl https://genrent.in/health

# Expected response:
# {"status":"ok","database":"connected","version":"v1"}
```

### 9️⃣ Create Admin User

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

## 🔧 Quick Commands Reference

### SSH into VPS:
```bash
ssh -i ~/.ssh/oracle_cloud ubuntu@YOUR_PUBLIC_IP
```

### View Application Logs:
```bash
ssh ubuntu@YOUR_IP
cd /opt/genrent/backend
docker-compose logs -f app
```

### Restart Application:
```bash
ssh ubuntu@YOUR_IP
cd /opt/genrent/backend
docker-compose restart app
```

### Update Application:
```bash
ssh ubuntu@YOUR_IP
cd /opt/genrent
git pull
cd backend
docker-compose build --no-cache
docker-compose up -d
```

### Check Services Status:
```bash
ssh ubuntu@YOUR_IP
cd /opt/genrent/backend
docker-compose ps
```

---

## 📊 URL Access After Deployment

| Service | URL |
|---------|-----|
| Frontend | `https://genrent.in` |
| API | `https://genrent.in/api/v1` |
| Health Check | `https://genrent.in/health` |
| Admin Dashboard | `https://genrent.in/admin-dashboard` |
| Vendor Dashboard | `https://genrent.in/vendor-dashboard` |

---

## 🔒 Security Checklist

- [ ] VPS firewall configured (ports 22, 80, 443)
- [ ] Strong admin password created
- [ ] JWT_SECRET is secure (auto-generated)
- [ ] DB_PASSWORD is strong (auto-generated)
- [ ] Cloudflare proxy enabled (☁️ orange cloud)
- [ ] SSL/TLS configured in Cloudflare
- [ ] Razorpay credentials updated in .env

---

## 🆘 Troubleshooting

### Can't SSH into VPS:
```bash
# Check security list in Oracle Cloud Console
# Ensure port 22 is open from 0.0.0.0/0
```

### Website not loading:
```bash
# Check Cloudflare DNS propagated
dig genrent.in

# Check if VPS is running
ssh ubuntu@YOUR_IP
docker-compose ps
```

### Database connection failed:
```bash
# Check postgres container
docker-compose logs postgres

# Check DATABASE_URL in .env
cat /opt/genrent/backend/.env | grep DATABASE_URL
```

---

## 💰 Cost Breakdown

| Service | Cost |
|---------|------|
| Oracle Cloud VPS | **FREE** |
| Cloudflare DNS | **FREE** |
| Cloudflare SSL | **FREE** |
| Cloudflare CDN | **FREE** |
| Domain (yearly) | ~$10-15/year |
| **Total** | **~$10-15/year** |

---

## 📈 What's Included

### Infrastructure:
- ✅ 1 CPU, 1GB RAM VPS
- ✅ 50GB storage
- ✅ 10TB/month bandwidth
- ✅ Docker containers
- ✅ PostgreSQL database

### Cloudflare:
- ✅ Global CDN
- ✅ DDoS protection
- ✅ SSL certificates
- ✅ DNS management
- ✅ Caching
- ✅ Analytics

### Application:
- ✅ Go backend
- ✅ PostgreSQL database
- ✅ Email service
- ✅ Payment gateway integration
- ✅ WebSocket support
- ✅ Admin dashboard
- ✅ Vendor dashboard

---

## 🎓 Next Steps

1. **Customize your branding** - Update logo, colors
2. **Setup payment gateway** - Add Razorpay credentials
3. **Configure email** - Test email sending
4. **Add content** - Create categories, equipment
5. **Test thoroughly** - Test all features
6. **Launch to users** - Share your domain!

---

## 📞 Need Help?

- Oracle Cloud Support: [community.oracle.com](https://community.oracle.com)
- Cloudflare Support: [community.cloudflare.com](https://community.cloudflare.com)
- GitHub Issues: Create an issue in the repository

---

**🎉 Congratulations! Your GenRent platform is now live!**

**Your site is accessible at:** `https://genrent.in`

**Admin Dashboard:** `https://genrent.in/admin-dashboard`

---

**Last Updated:** 2026-08-19
