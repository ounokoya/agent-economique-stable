#!/bin/bash
# 🚀 Installation complète serveur Singapour
# Ordre: WireGuard → Nomad → Caddy
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

echo -e "${BLUE}🚀 INSTALLATION COMPLÈTE SERVEUR PRODUCTION${NC}"
echo "================================================"
echo "Serveur: 31.57.224.79 (Singapore)"
echo "Stack: WireGuard → TLS Certs → Nomad → Caddy"
echo "================================================"
echo ""

# Check root
if [ "$EUID" -ne 0 ]; then 
    log_error "Ce script doit être exécuté en tant que root"
    exit 1
fi

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Confirm
read -p "⚠️  Continuer l'installation complète? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_warning "Installation annulée"
    exit 0
fi

echo ""
echo "================================================"
echo "📦 ÉTAPE 1/3: Installation WireGuard VPN"
echo "================================================"
echo ""

# Step 1: WireGuard
log_info "Installation WireGuard..."
if [ -f "$SCRIPT_DIR/setup-wireguard.sh" ]; then
    bash "$SCRIPT_DIR/setup-wireguard.sh" server
    log_success "WireGuard installé"
else
    log_error "Script setup-wireguard.sh introuvable"
    exit 1
fi

echo ""
log_info "Attente activation interface wg0..."
sleep 5

# Verify WireGuard is running
if ip link show wg0 &> /dev/null; then
    log_success "Interface wg0 active"
else
    log_error "Interface wg0 non active"
    log_warning "Vérifier: wg show"
    exit 1
fi

echo ""
echo "================================================"
echo "📦 ÉTAPE 2/4: Génération Certificats TLS Nomad"
echo "================================================"
echo ""

# Step 2: Generate TLS certificates
log_info "Génération certificats TLS..."
if [ -f "$SCRIPT_DIR/generate-nomad-certs.sh" ]; then
    bash "$SCRIPT_DIR/generate-nomad-certs.sh"
    log_success "Certificats TLS générés"
else
    log_error "Script generate-nomad-certs.sh introuvable"
    exit 1
fi

echo ""
echo "================================================"
echo "📦 ÉTAPE 3/4: Installation Nomad"
echo "================================================"
echo ""

# Step 3: Nomad (uses wg0 network interface + TLS certs)
log_info "Installation Nomad..."
if [ -f "$SCRIPT_DIR/install-nomad.sh" ]; then
    # Modify install-nomad.sh to use our config with wg0
    bash "$SCRIPT_DIR/install-nomad.sh"
    log_success "Nomad installé"
else
    log_error "Script install-nomad.sh introuvable"
    exit 1
fi

echo ""
log_info "Application configuration Nomad avec interface wg0..."

# Copy our config with wg0 network_interface
if [ -f "$SCRIPT_DIR/../configs/nomad-server.hcl" ]; then
    cp "$SCRIPT_DIR/../configs/nomad-server.hcl" /etc/nomad.d/nomad.hcl
    systemctl restart nomad
    sleep 5
    log_success "Configuration Nomad mise à jour"
else
    log_warning "Config nomad-server.hcl introuvable, utilise config par défaut"
fi

# Verify Nomad
if systemctl is-active --quiet nomad; then
    log_success "Nomad opérationnel"
else
    log_error "Nomad non actif"
    exit 1
fi

echo ""
echo "================================================"
echo "📦 ÉTAPE 4/4: Installation Caddy"
echo "================================================"
echo ""

# Step 4: Caddy
log_info "Installation Caddy..."
if [ -f "$SCRIPT_DIR/install-caddy.sh" ]; then
    bash "$SCRIPT_DIR/install-caddy.sh"
    log_success "Caddy installé"
else
    log_error "Script install-caddy.sh introuvable"
    exit 1
fi

echo ""
echo "================================================"
echo "✅ INSTALLATION TERMINÉE"
echo "================================================"
echo ""

# Summary
log_success "Infrastructure complète installée!"
echo ""
echo "📊 Services installés:"
echo "   ✅ WireGuard VPN    (wg0 - 10.8.0.1)"
echo "   ✅ TLS Certs        (/etc/nomad.d/certs/)"
echo "   ✅ Nomad Server     (https://10.8.0.1:4646)"
echo "   ✅ Caddy Proxy      (http://10.8.0.1:80)"
echo ""

echo "🔍 Vérifications:"
echo "   wg show"
echo "   systemctl status nomad"
echo "   systemctl status caddy"
echo "   nomad server members"
echo "   nomad node status"
echo ""

echo "🌐 Accès (via VPN):"
echo "   Nomad UI direct: https://10.8.0.1:4646 (TLS)"
echo "   Nomad via Caddy: http://10.8.0.1:80"
echo "   Health check:    http://10.8.0.1:8080/health"
echo ""

echo "🔐 Sécurité:"
echo "   Firewall UFW:    Active"
echo "   VPN Required:    Oui (WireGuard)"
echo "   TLS Enabled:     Oui (Nomad HTTPS)"
echo "   ACL Nomad:       Désactivé (à activer en prod)"
echo ""

echo "📝 Prochaines étapes:"
echo "   1. Récupérer certificats client:"
echo "      scp -r root@31.57.224.79:/tmp/nomad-client-certs ~/nomad-certs"
echo ""
echo "   2. Configurer client WireGuard sur machine locale"
echo ""
echo "   3. Tester connexion VPN:"
echo "      ping 10.8.0.1"
echo ""
echo "   4. Configurer Nomad CLI (avec TLS):"
echo "      export NOMAD_ADDR=\"https://10.8.0.1:4646\""
echo "      export NOMAD_CACERT=\"\$HOME/nomad-certs/ca.pem\""
echo "      export NOMAD_CLIENT_CERT=\"\$HOME/nomad-certs/cli.pem\""
echo "      export NOMAD_CLIENT_KEY=\"\$HOME/nomad-certs/cli-key.pem\""
echo ""
echo "   5. Tester Nomad:"
echo "      nomad server members"
echo ""
echo "   6. Déployer application Scalping Live"
echo ""

log_success "Setup serveur terminé!"
