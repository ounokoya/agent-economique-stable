#!/bin/bash
# 🚀 Installation automatique Nomad Server + Client
# Server: 31.57.224.79 (Singapore)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

echo -e "${BLUE}🚀 Installation Nomad Server${NC}"
echo "========================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    log_error "Ce script doit être exécuté en tant que root"
    exit 1
fi

# Update system
log_info "Mise à jour système..."
apt update && apt upgrade -y
log_success "Système à jour"

# Install dependencies
log_info "Installation dépendances..."
apt install -y wget curl unzip gpg jq
log_success "Dépendances installées"

# Add HashiCorp repository
log_info "Ajout repository HashiCorp..."
wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg

echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list

apt update
log_success "Repository ajouté"

# Install Nomad
log_info "Installation Nomad..."
apt install -y nomad
log_success "Nomad installé"

# Verify installation
NOMAD_VERSION=$(nomad version | head -1)
log_success "Version: $NOMAD_VERSION"

# Create directories
log_info "Création dossiers..."
mkdir -p /etc/nomad.d
mkdir -p /opt/nomad/data
mkdir -p /root/data
log_success "Dossiers créés"

# Check if wg0 interface exists (WireGuard should be installed first)
if ip link show wg0 &> /dev/null; then
    log_success "Interface WireGuard (wg0) détectée"
    NETWORK_INTERFACE="wg0"
else
    log_warning "Interface WireGuard (wg0) non détectée, utilisation interface par défaut"
    NETWORK_INTERFACE=$(ip route | grep default | awk '{print $5}' | head -1)
    log_info "Interface réseau: $NETWORK_INTERFACE"
fi

# Create Nomad configuration
log_info "Création configuration Nomad..."
cat > /etc/nomad.d/nomad.hcl <<EOF
# Nomad Server + Client Configuration
# Server: 31.57.224.79 (Singapore)
# Datacenter: sg1

datacenter = "sg1"
data_dir   = "/opt/nomad/data"
bind_addr  = "0.0.0.0"

# Server mode
server {
  enabled          = true
  bootstrap_expect = 1
  
  server_join {
    retry_join = ["127.0.0.1"]
  }
}

# Client mode (jobs s'exécutent sur même serveur)
client {
  enabled = true
  
  # Network interface (WireGuard si disponible)
  network_interface = "$NETWORK_INTERFACE"
  
  # Options
  options = {
    "driver.raw_exec.enable" = "1"
    "docker.cleanup.image"   = "true"
  }
  
  # Host volume (path générique /root/data)
  host_volume "app-data" {
    path      = "/root/data"
    read_only = false
  }
}

# UI
ui {
  enabled = true
  
  # Accessible via VPN sur port 4646
}

# Telemetry
telemetry {
  publish_allocation_metrics = true
  publish_node_metrics       = true
  prometheus_metrics         = true
}

# ACL (désactivé pour l'instant)
acl {
  enabled = false
}
EOF
log_success "Configuration créée (interface: $NETWORK_INTERFACE)"

# Create systemd service
log_info "Configuration service systemd..."
cat > /etc/systemd/system/nomad.service <<'EOF'
[Unit]
Description=Nomad
Documentation=https://www.nomadproject.io/docs/
Wants=network-online.target
After=network-online.target

[Service]
Type=exec
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=/usr/bin/nomad agent -config=/etc/nomad.d
KillMode=process
KillSignal=SIGINT
LimitNOFILE=65536
LimitNPROC=infinity
Restart=on-failure
RestartSec=2
TasksMax=infinity

[Install]
WantedBy=multi-user.target
EOF
log_success "Service systemd créé"

# Configure firewall
log_info "Configuration firewall UFW..."
apt install -y ufw

# Allow SSH first (important!)
ufw allow 22/tcp

# Nomad ports
ufw allow 4646/tcp  # HTTP API + UI
ufw allow 4647/tcp  # RPC
ufw allow 4648/tcp  # Serf

# WireGuard
ufw allow 51820/udp

# Enable firewall
ufw --force enable
log_success "Firewall configuré"

# Start Nomad
log_info "Démarrage Nomad..."
systemctl daemon-reload
systemctl enable nomad
systemctl start nomad
sleep 5
log_success "Nomad démarré"

# Verify Nomad is running
if systemctl is-active --quiet nomad; then
    log_success "Nomad est opérationnel"
    
    # Wait for leader election
    log_info "Attente élection leader..."
    sleep 10
    
    # Check cluster
    nomad server members
    nomad node status
    
    echo ""
    log_success "Installation Nomad terminée!"
    echo ""
    echo "📊 Accès UI Nomad:"
    echo "   URL: http://31.57.224.79:4646"
    echo ""
    echo "🔧 Commandes utiles:"
    echo "   systemctl status nomad"
    echo "   nomad server members"
    echo "   nomad node status"
    echo "   journalctl -u nomad -f"
else
    log_error "Nomad n'a pas démarré correctement"
    log_info "Vérifier les logs: journalctl -u nomad -n 50"
    exit 1
fi
