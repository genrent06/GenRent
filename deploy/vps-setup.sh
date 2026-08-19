#!/bin/bash
# ==============================================================================
# GenRent VPS Setup Script
# ==============================================================================
# This script sets up a fresh Ubuntu/Debian VPS for running GenRent
# Usage: curl -fsSL https://your-repo/raw/main/deploy/vps-setup.sh | bash
# ==============================================================================

set -e

echo "🚀 Setting up GenRent VPS..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ==============================================================================
# 1. System Updates
# ==============================================================================
echo -e "${YELLOW}📦 Updating system packages...${NC}"
sudo apt-get update -y
sudo apt-get upgrade -y

# ==============================================================================
# 2. Install Docker and Docker Compose
# ==============================================================================
echo -e "${YELLOW}🐳 Installing Docker...${NC}"
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    rm -f get-docker.sh
    echo -e "${GREEN}✅ Docker installed${NC}"
else
    echo -e "${GREEN}✅ Docker already installed${NC}"
fi

echo -e "${YELLOW}📦 Installing Docker Compose...${NC}"
if ! command -v docker-compose &> /dev/null; then
    sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    echo -e "${GREEN}✅ Docker Compose installed${NC}"
else
    echo -e "${GREEN}✅ Docker Compose already installed${NC}"
fi

# ==============================================================================
# 3. Install Required Tools
# ==============================================================================
echo -e "${YELLOW}🔧 Installing additional tools...${NC}"
sudo apt-get install -y git curl wget nginx certbot python3-pip

# ==============================================================================
# 4. Configure Firewall
# ==============================================================================
echo -e "${YELLOW}🔒 Configuring firewall...${NC}"
if command -v ufw &> /dev/null; then
    sudo ufw allow 22/tcp
    sudo ufw allow 80/tcp
    sudo ufw allow 443/tcp
    sudo ufw --force enable
    echo -e "${GREEN}✅ Firewall configured${NC}"
else
    echo -e "${YELLOW}⚠️  UFW not found, skipping firewall setup${NC}"
fi

# ==============================================================================
# 5. Create Application Directory
# ==============================================================================
echo -e "${YELLOW}📁 Creating application directory...${NC}"
sudo mkdir -p /opt/genrent
sudo chown -R $USER:$USER /opt/genrent
echo -e "${GREEN}✅ Directory /opt/genrent created${NC}"

# ==============================================================================
# 6. Setup Log Directory
# ==============================================================================
echo -e "${YELLOW}📋 Setting up log directories...${NC}"
sudo mkdir -p /var/log/caddy
sudo mkdir -p /var/backups/genrent
sudo chown -R $USER:$USER /var/log/caddy
sudo chown -R $USER:$USER /var/backups/genrent
echo -e "${GREEN}✅ Log directories created${NC}"

# ==============================================================================
# 7. Install Monitoring Tools
# ==============================================================================
echo -e "${YELLOW}📊 Installing monitoring tools...${NC}"
curl -sSL https://raw.githubusercontent.com/sleuthkit/sleuthkit/master/install.sh | bash 2>/dev/null || true

# ==============================================================================
# 8. Generate Random Secrets
# ==============================================================================
echo -e "${YELLOW}🔐 Generating secure secrets...${NC}"
JWT_SECRET=$(openssl rand -base64 32 | tr -d '/+=' | head -c 32)
DB_PASSWORD=$(openssl rand -base64 16 | tr -d '/+=' | head -c 16)

echo -e "${GREEN}✅ Secrets generated${NC}"
echo -e "${YELLOW}⚠️  SAVE THESE CREDENTIALS:${NC}"
echo "================================"
echo "JWT_SECRET: $JWT_SECRET"
echo "DB_PASSWORD: $DB_PASSWORD"
echo "================================"

# Save to file for later use
cat > /opt/genrent/.secrets <<EOF
JWT_SECRET=$JWT_SECRET
DB_PASSWORD=$DB_PASSWORD
EOF

chmod 600 /opt/genrent/.secrets

# ==============================================================================
# 9. Optimize System for Production
# ==============================================================================
echo -e "${YELLOW}⚡ Optimizing system...${NC}"

# Increase file descriptor limits
cat <<EOF | sudo tee -a /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536
EOF

# Optimize TCP settings
cat <<EOF | sudo tee -a /etc/sysctl.conf
net.core.somaxconn = 1024
net.ipv4.tcp_max_syn_backlog = 2048
net.ipv4.tcp_fin_timeout = 30
EOF

sudo sysctl -p > /dev/null 2>&1 || true

echo -e "${GREEN}✅ System optimized${NC}"

# ==============================================================================
# 10. Setup Swap (if low memory)
# ==============================================================================
echo -e "${YELLOW}💾 Checking memory...${NC}"
TOTAL_MEM=$(free -m | awk '/Mem:/ {print $2}')
if [ "$TOTAL_MEM" -lt 1024 ]; then
    echo -e "${YELLOW}⚠️  Low memory detected, setting up swap...${NC}"
    sudo fallocate -l 2G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
    echo -e "${GREEN}✅ Swap configured (2GB)${NC}"
else
    echo -e "${GREEN}✅ Sufficient memory available${NC}"
fi

# ==============================================================================
# DONE!
# ==============================================================================
echo -e "${GREEN}"
echo "============================================"
echo "✅ VPS Setup Complete!"
echo "============================================"
echo -e "${NC}"
echo "Next steps:"
echo "1. Upload your code: scp -r genrent/ user@VPS_IP:/opt/genrent"
echo "2. SSH into VPS: ssh user@VPS_IP"
echo "3. Run deploy script: cd /opt/genrent && ./deploy.sh"
echo ""
echo "Your credentials are saved in: /opt/genrent/.secrets"
echo ""
echo -e "${YELLOW}⚠️  IMPORTANT: Copy your secrets now!${NC}"
cat /opt/genrent/.secrets
echo ""
