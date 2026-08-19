# ☁️ Cloudflare Setup Guide for GenRent

This guide shows you how to configure Cloudflare with your VPS for free DNS, SSL, and DDoS protection.

---

## 🎯 What Cloudflare Will Provide (All Free!)

- ✅ **DNS Management** - Manage your domain's DNS records
- ✅ **SSL/TLS Certificates** - Automatic HTTPS encryption
- ✅ **DDoS Protection** - Protect your app from attacks
- ✅ **CDN** - Fast content delivery worldwide
- ✅ **Caching** - Speed up your application
- ✅ **Page Rules** - Redirects and custom rules

---

## 📋 Step-by-Step Setup

### Step 1: Create Cloudflare Account

1. Go to [cloudflare.com](https://www.cloudflare.com/)
2. Click **"Sign Up"**
3. Use your email (or sign up with Google)
4. Verify your email address

### Step 2: Add Your Domain

1. After logging in, click **"Add a Site"**
2. Enter your domain (e.g., `genrent.in`)
3. Click **"Add Site"**

### Step 3: Choose Plan

1. Select **"Free"** plan
2. Click **"Continue"**

### Step 4: Update Nameservers

Cloudflare will show you two nameservers like:
```
alice.ns.cloudflare.com
bob.ns.cloudflare.com
```

#### At Your Domain Registrar (where you bought the domain):

**If you bought from GoDaddy:**
1. Log in to GoDaddy
2. Go to **My Products** → **Domains**
3. Click on your domain
4. Click **"Manage Nameservers"**
5. Select **"Custom"**
6. Replace existing nameservers with Cloudflare's nameservers
7. Click **"Save"**

**If you bought from Namecheap:**
1. Log in to Namecheap
2. Go to **Domain List** → click **"Manage"** next to your domain
3. Go to **"Nameservers"** tab
4. Select **"Custom DNS"**
5. Enter Cloudflare's nameservers
6. Click **"Save Changes"**

**Wait for DNS propagation** (can take 2-48 hours)

### Step 5: Configure DNS Records

1. Go to **DNS** → **Records**
2. Delete any existing records
3. Add new records:

#### Record 1: Root Domain
```
Type: A
Name: @ (or genrent.in)
IPv4 address: YOUR_VPS_PUBLIC_IP
Proxy status: Proxied (☁️ orange cloud)
TTL: Auto
```

#### Record 2: WWW
```
Type: CNAME
Name: www
Target: genrent.in
Proxy status: Proxied (☁️ orange cloud)
TTL: Auto
```

#### Record 3: For email (optional)
```
Type: MX
Name: @
Priority: 10
Target: alt1.aspmx.l.google.com
```

### Step 6: SSL/TLS Configuration

1. Go to **SSL/TLS** → **Overview**
2. Set to **"Flexible"** (if your VPS doesn't have SSL)
3. Or set to **"Full"** (if your VPS has SSL certificate)
4. Set **"Always Use HTTPS"** to **ON**

### Step 7: Security Settings

#### Enable Bot Protection:
1. Go to **Security** → **Bots**
2. Set to **"Medium"**

#### Enable Challenge Passage:
1. Go to **Security** → **Settings**
2. Turn on **"Challenge Passage"** (for trusted users)

### Step 8: Cache Configuration

1. Go to **Caching** → **Configuration**
2. Set **"Caching Level"** to **Standard**
3. Set **"Browser Cache TTL"** to **Respect Existing Headers**

---

## 🔧 Cloudflare Settings for GenRent

### Page Rules (Optional Redirects)

1. Go to **Rules** → **Page Rules**
2. Create rules:

**Rule 1: Redirect HTTP to HTTPS**
```
URL: http://genrent.in/*
Settings: Forwarding URL (301) -> https://genrent.in/$1
```

**Rule 2: Redirect www to non-www**
```
URL: http://www.genrent.in/*
Settings: Forwarding URL (301) -> https://genrent.in/$1
```

### Firewall Rules (Optional)

Go to **Security** → **WAF** → **Custom Rules**

**Rate Limiting for API:**
```
Expression: (http.request.uri.path contains "/api/")
Action: Rate Limit (10 requests per minute)
```

---

## 🧪 Verify Cloudflare is Working

### Check 1: DNS Resolution
```bash
dig genrent.in
# Should show Cloudflare IPs
```

### Check 2: SSL Certificate
Visit `https://genrent.in` and check the certificate - it should be issued by Cloudflare

### Check 3: Performance
Use [pagespeed.web.dev](https://pagespeed.web.dev) to check performance improvements

---

## 📊 Monitor with Cloudflare

### Real-time Analytics:
- Go to **Analytics** → **Overview**
- See traffic, threats blocked, bandwidth saved

### Security Events:
- Go to **Security** → **Events**
- See blocked threats and challenges

---

## ⚡ Cloudflare Features to Enable

### 1. Auto Minify
- Go to **Caching** → **Configuration**
- Enable **Auto Minify** for CSS, JS, HTML

### 2. Brotli Compression
- Go to **Network** → **Optimization**
- Enable **Brotli**

### 3. HTTP/2
- Go to **Network** → **HTTP/2**
- Ensure it's **ON**

### 4. 0-RTT Connection Resumption
- Go to **Network** → **0-RTT**
- Enable for faster connections

---

## 🔍 Troubleshooting

### Issue: "ERR_TOO_MANY_REDIRECTS"
**Solution**: Go to SSL/TLS and change from "Flexible" to "Full"

### Issue: DNS not propagating
**Solution**: Use `dig` command to check:
```bash
dig genrent.in @alice.ns.cloudflare.com
```

### Issue: Site not loading
**Solution**: Check if VPS firewall allows Cloudflare IPs
```bash
sudo ufw allow from 173.245.48.0/20
sudo ufw allow from 103.21.244.0/22
sudo ufw allow from 103.22.200.0/22
sudo ufw allow from 103.31.4.0/22
```

---

## 📱 Cloudflare Mobile App

Download the Cloudflare app to:
- Get push notifications for security events
- Manage DNS on the go
- View analytics anytime

---

## 🎯 Final Checklist

- [ ] Domain added to Cloudflare
- [ ] Nameservers updated at registrar
- [ ] DNS propagated (check with `dig`)
- [ ] A record created pointing to VPS IP
- [ ] SSL/TLS configured
- [ ] Always HTTPS enabled
- [ ] Firewall rules configured (optional)
- [ ] Caching optimized
- [ ] Page rules created (optional)
- [ ] Tested site loads correctly

---

## 🚀 Once Cloudflare is Configured

Your site will be accessible at:
- `https://genrent.in`
- `https://www.genrent.in`

Both will be protected by Cloudflare's CDN and DDoS protection!

---

**Need Help?**
- Cloudflare Docs: [developers.cloudflare.com](https://developers.cloudflare.com)
- Community: [community.cloudflare.com](https://community.cloudflare.com)
