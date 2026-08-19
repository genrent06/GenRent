# 🆓 Oracle Cloud Free Tier Guide

Get a **FREE VPS** with Oracle Cloud to host your GenRent backend and PostgreSQL database.

---

## 💰 What You Get (FREE Forever!)

### Compute Instance (Always Free)
- **2 AMD-based CPUs** (or up to 4 OCPUs)
- **1 GB RAM**
- **50 GB Boot Volume**
- **Unlimited Bandwidth** (10 TB/month)

### Database (Always Free)
- **2 Autonomous Databases** (20 GB each)
- **4 TB Block Volume Storage**

### Other Free Services
- **Load Balancer** (1)
- **Network Monitoring** (500 million data points)
- **Email Delivery** (GB/month)

---

## 📋 Step-by-Step Setup

### Step 1: Create Oracle Cloud Account

1. Go to [oracle.com/cloud/free](https://www.oracle.com/cloud/free/)
2. Click **"Sign Up"** or **"Try Free"**
3. Enter your email address
4. Fill in your details:
   - **Country**: Select your country
   - **First Name / Last Name**
   - **Phone Number**: (must be valid for verification)
   - **Company**: (can be individual, enter your name)

### Step 2: Verify Email & Phone

1. Check your email for verification code
2. Enter the code to verify email
3. Wait for SMS/Call with phone verification code
4. Enter the code to verify phone

### Step 3: Account Details

1. Enter your **Home Region** (choose closest to your target audience)
   - Recommended regions:
     - **Mumbai** (ap-mumbai-1) - For India
     - **Singapore** (ap-singapore-1) - For Asia
     - **London** (eu-london-1) - For Europe
     - **Phoenix** (us-phoenix-1) - For US West

2. Click **"Create Account"**

### Step 4: Enter Credit Card

**Note**: Oracle requires a credit card for identity verification, but you won't be charged unless you upgrade.

1. Enter credit card details
2. Enter billing address
3. Click **"Verify Card"** (small temporary charge will be made and refunded)

### Step 5: Choose Tenancy Name

1. Enter a unique tenancy name (e.g., `genrent-production`)
2. Click **"Create Tenancy"**

---

## 🖥️ Creating Your Free VPS

### Step 1: Navigate to Compute

1. After login, click **"Menu"** (☰) in top left
2. Go to **"Compute"** → **"Instances"**
3. Click **"Create Instance"**

### Step 2: Configure Instance

#### Name & Compartment
```
Name: genrent-app
Create in Compartment: (your compartment)
```

#### Placement
```
Availability Domain: AD-1 (or any available)
Fault Domain: No Preference
```

#### Image & Shape
```
Image: Ubuntu 22.04 Minimal
   OR: Oracle Linux 8
Shape: VM.Standard.E2.1.Micro (Free Tier)
   - 1 OCPU, 1 GB RAM
```

#### Networking
```
Virtual Cloud Network: Create new VCN
Subnet: Create new public subnet
Assign Public IP: Yes (✓)
```

#### SSH Key
```
SSH Key: Paste your public SSH key
```

**To generate SSH key on your computer:**
```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/oracle_cloud
cat ~/.ssh/oracle_cloud.pub
# Copy and paste the output
```

### Step 3: Create Instance

1. Click **"Create"**
2. Wait 2-5 minutes for provisioning
3. Note down the **Public IP Address**

---

## 🔐 Connect to Your VPS

### First Time Connection

```bash
# SSH into your instance
ssh -i ~/.ssh/oracle_cloud ubuntu@YOUR_PUBLIC_IP

# Or if using root user
ssh -i ~/.ssh/oracle_cloud opc@YOUR_PUBLIC_IP
```

### Update System
```bash
sudo apt-get update -y
sudo apt-get upgrade -y
```

---

## 🚀 Run GenRent Setup Script

Once connected to your VPS:

```bash
# Download and run setup script
curl -fsSL https://your-repo/deploy/vps-setup.sh | bash

# Or if you have the script locally:
scp deploy/vps-setup.sh ubuntu@YOUR_IP:/home/ubuntu/
ssh ubuntu@YOUR_IP
chmod +x vps-setup.sh
./vps-setup.sh
```

---

## 🔥 Oracle Cloud Specifics

### Opening Ports in Security List

1. Go to **Networking** → **Virtual Cloud Networks**
2. Click your VCN
3. Go to **Security Lists**
4. Click **Default Security List**
5. Add Ingress Rules:

**For HTTP:**
```
Source: CIDR 0.0.0.0/0
IP Protocol: TCP
Destination Port: 80
```

**For HTTPS:**
```
Source: CIDR 0.0.0.0/0
IP Protocol: TCP
Destination Port: 443
```

**For SSH:**
```
Source: CIDR 0.0.0.0/0
IP Protocol: TCP
Destination Port: 22
```

### Always Free Limits

| Resource | Free Tier Limit | Current Usage |
|----------|----------------|----------------|
| OCPUs | 2 (4 for Ampere A1) | 1 |
| RAM | 24 GB | 1 GB |
| Storage | 50 GB | 50 GB |
| Bandwidth | 10 TB/month | ~1 GB/month |

---

## 📊 Monitoring Your Instance

### From Oracle Console:
1. Go to **Compute** → **Instances**
2. Click your instance
3. View **CPU**, **Memory**, **Network** usage

### From VPS:
```bash
# Check resources
free -h
df -h
top
```

---

## 🔄 Creating Resources for GenRent

### Option 1: Use Free Compute Instance (Recommended)
- Run Docker containers directly on VPS
- Use VPS resources efficiently
- Easy to manage

### Option 2: Use Autonomous Database
- Use Oracle's managed PostgreSQL-compatible database
- More reliable but complex setup
- Only if you want separate database server

---

## 🆘 Common Issues & Solutions

### Issue: "Out of capacity"
**Solution**: Try different availability domain or region

### Issue: "SSH connection refused"
**Solution**:
```bash
# Check security list rules
# Ensure port 22 is open
# Check instance is running
```

### Issue: "Instance not booting"
**Solution**: Check boot volume, recreate if needed

---

## 📝 Important Notes

### About the Free Tier:
- ✅ **Truly FREE** - No hidden costs
- ✅ **Forever** - Not just a trial
- ✅ **Production Ready** - Good for small apps
- ⚠️ **Credit card required** - For verification only

### Renewal:
- Free tier automatically renews monthly
- No action needed to maintain free tier
- Only pay if you upgrade resources

### Upgrading:
- You can upgrade anytime if needed
- Pay-as-you-go pricing available
- Scale up/down as needed

---

## 🎯 Next Steps After VPS Setup

1. ✅ **Get your VPS Public IP**
2. ✅ **Test SSH connection**
3. ✅ **Run setup script**
4. ✅ **Deploy GenRent**
5. ✅ **Configure Cloudflare DNS** with your VPS IP

---

## 📞 Support

- Oracle Cloud Community: [community.oracle.com](https://community.oracle.com)
- Documentation: [docs.oracle.com/en-us/iaas](https://docs.oracle.com/en-us/iaas)
- Always Free FAQ: [oracle.com/cloud/free/faq](https://www.oracle.com/cloud/free/faq)

---

**You now have a FREE VPS ready for GenRent!** 🎉
