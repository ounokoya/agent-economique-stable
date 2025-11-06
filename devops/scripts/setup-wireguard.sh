#!/bin/bash
# 🔒 Setup WireGuard VPN
# Usage: ./setup-wireguard.sh [server|client]

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

# Check argument
if [ $# -eq 0 ]; then
    log_error "Usage: $0 [server|client]"
    exit 1
fi

MODE=$1

if [ "$MODE" != "server" ] && [ "$MODE" != "client" ]; then
    log_error "Mode invalide. Utiliser: server ou client"
    exit 1
fi

# Check root
if [ "$EUID" -ne 0 ]; then 
    log_error "Ce script doit être exécuté en tant que root"
    exit 1
fi

echo -e "${BLUE}🔒 Setup WireGuard VPN - Mode: $MODE${NC}"
echo "============================================="

# Install WireGuard
log_info "Installation WireGuard..."
apt update
apt install -y wireguard wireguard-tools
log_success "WireGuard installé"

# Enable IP forwarding (server only)
if [ "$MODE" = "server" ]; then
    log_info "Activation IP forwarding..."
    echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
    echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf
    sysctl -p
    log_success "IP forwarding activé"
fi

# Generate keys
log_info "Génération clés WireGuard..."
cd /etc/wireguard
wg genkey | tee ${MODE}_private.key | wg pubkey > ${MODE}_public.key
chmod 600 ${MODE}_private.key
log_success "Clés générées"

# Display keys
PRIVATE_KEY=$(cat ${MODE}_private.key)
PUBLIC_KEY=$(cat ${MODE}_public.key)

echo ""
log_success "Clés générées avec succès!"
echo ""
echo "=========================================="
echo "🔑 PRIVATE KEY (à garder secret):"
echo "$PRIVATE_KEY"
echo ""
echo "🔑 PUBLIC KEY (à partager):"
echo "$PUBLIC_KEY"
echo "=========================================="
echo ""

# Create config based on mode
if [ "$MODE" = "server" ]; then
    log_info "Création configuration serveur..."
    
    # Get network interface
    INTERFACE=$(ip route | grep default | awk '{print $5}' | head -1)
    log_info "Interface réseau détectée: $INTERFACE"
    
    # Ask for client public key
    echo ""
    read -p "📋 Entrer la PUBLIC KEY du CLIENT: " CLIENT_PUBLIC_KEY
    
    cat > /etc/wireguard/wg0.conf <<EOF
[Interface]
# Server VPN IP
Address = 10.8.0.1/24

# Server private key
PrivateKey = $PRIVATE_KEY

# WireGuard port
ListenPort = 51820

# NAT rules
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o $INTERFACE -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o $INTERFACE -j MASQUERADE

# Peer: Client
[Peer]
# Client public key
PublicKey = $CLIENT_PUBLIC_KEY

# Client VPN IP
AllowedIPs = 10.8.0.2/32

# Keep alive
PersistentKeepalive = 25
EOF
    
    log_success "Configuration serveur créée"
    
else
    log_info "Création configuration client..."
    
    # Ask for server details
    echo ""
    read -p "📋 Entrer la PUBLIC KEY du SERVEUR: " SERVER_PUBLIC_KEY
    read -p "📋 Entrer l'IP PUBLIQUE du serveur (ex: 31.57.224.79): " SERVER_IP
    
    cat > /etc/wireguard/wg0.conf <<EOF
[Interface]
# Client VPN IP
Address = 10.8.0.2/24

# Client private key
PrivateKey = $PRIVATE_KEY

# DNS (optional)
DNS = 8.8.8.8

[Peer]
# Server public key
PublicKey = $SERVER_PUBLIC_KEY

# Server endpoint
Endpoint = $SERVER_IP:51820

# Traffic to route (VPN only, not full tunnel)
AllowedIPs = 10.8.0.0/24

# Keep alive
PersistentKeepalive = 25
EOF
    
    log_success "Configuration client créée"
fi

# Secure config
chmod 600 /etc/wireguard/wg0.conf
log_success "Configuration sécurisée"

# Start WireGuard
log_info "Démarrage WireGuard..."
wg-quick up wg0
systemctl enable wg-quick@wg0
log_success "WireGuard démarré"

# Show status
echo ""
log_success "WireGuard opérationnel!"
echo ""
wg show
echo ""

if [ "$MODE" = "server" ]; then
    log_info "Test de connexion:"
    echo "   Sur CLIENT: ping 10.8.0.1"
else
    log_info "Test de connexion:"
    echo "   ping 10.8.0.1"
    echo ""
    log_info "Test Nomad via VPN:"
    echo "   export NOMAD_ADDR=\"http://10.8.0.1:4646\""
    echo "   curl http://10.8.0.1:4646/v1/status/leader"
fi

echo ""
log_success "Setup WireGuard terminé!"
