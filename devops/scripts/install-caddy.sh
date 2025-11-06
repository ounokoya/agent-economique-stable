#!/bin/bash
# 🌐 Installation et configuration Caddy
# Reverse proxy pour services Nomad

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

echo -e "${BLUE}🌐 Installation Caddy${NC}"
echo "========================================"

# Check root
if [ "$EUID" -ne 0 ]; then 
    log_error "Ce script doit être exécuté en tant que root"
    exit 1
fi

# Install dependencies
log_info "Installation dépendances..."
apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
log_success "Dépendances installées"

# Add Caddy repository
log_info "Ajout repository Caddy..."
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg

curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list

apt update
log_success "Repository ajouté"

# Install Caddy
log_info "Installation Caddy..."
apt install -y caddy
log_success "Caddy installé"

# Verify installation
CADDY_VERSION=$(caddy version)
log_success "Version: $CADDY_VERSION"

# Create Caddy config directory
log_info "Création configuration..."
mkdir -p /etc/caddy

# Create Caddyfile
cat > /etc/caddy/Caddyfile <<'EOF'
# Caddy Configuration
# Reverse proxy pour services Nomad

# Global options
{
    # Disable automatic HTTPS for local/VPN usage
    auto_https off
    
    # Admin API (local only)
    admin localhost:2019
}

# Nomad UI (via VPN)
http://10.8.0.1:80 {
    reverse_proxy localhost:4646
    
    log {
        output file /var/log/caddy/nomad.log
    }
}

# Health check endpoint
http://10.8.0.1:8080 {
    respond /health 200
    respond /ready 200
}

# Future: Add other services here
# Example:
# http://10.8.0.1:8081 {
#     reverse_proxy localhost:8500  # Consul UI
# }
EOF

log_success "Caddyfile créé"

# Create log directory
mkdir -p /var/log/caddy
chown caddy:caddy /var/log/caddy

# Test configuration
log_info "Test configuration Caddy..."
if caddy validate --config /etc/caddy/Caddyfile; then
    log_success "Configuration valide"
else
    log_error "Configuration invalide"
    exit 1
fi

# Start Caddy
log_info "Démarrage Caddy..."
systemctl enable caddy
systemctl restart caddy
sleep 2

# Check status
if systemctl is-active --quiet caddy; then
    log_success "Caddy est opérationnel"
else
    log_error "Caddy n'a pas démarré"
    log_info "Vérifier logs: journalctl -u caddy -n 50"
    exit 1
fi

# Configure firewall
log_info "Configuration firewall..."
ufw allow 80/tcp
ufw allow 8080/tcp
log_success "Firewall configuré"

echo ""
log_success "Installation Caddy terminée!"
echo ""
echo "🌐 Endpoints disponibles (via VPN):"
echo "   http://10.8.0.1:80    → Nomad UI"
echo "   http://10.8.0.1:8080  → Health check"
echo ""
echo "🔧 Commandes utiles:"
echo "   systemctl status caddy"
echo "   caddy validate --config /etc/caddy/Caddyfile"
echo "   journalctl -u caddy -f"
echo "   curl http://10.8.0.1:8080/health"
