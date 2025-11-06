#!/bin/bash
# 🚀 Déploiement Complet Scalping Live Bybit
# Binaire + Configuration + Job Nomad

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

echo -e "${BLUE}🚀 DÉPLOIEMENT COMPLET SCALPING LIVE BYBIT${NC}"
echo "=============================================="
echo "Étapes:"
echo "  1. Upload binaire"
echo "  2. Upload configuration"
echo "  3. Déploiement job Nomad"
echo "=============================================="
echo ""

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

# Configuration
REMOTE_HOST="31.57.224.79"
REMOTE_USER="root"
REMOTE_PATH="/root/data/scalping-live-bybit"
export NOMAD_ADDR="${NOMAD_ADDR:-https://10.8.0.1:4646}"
export NOMAD_CACERT="${NOMAD_CACERT:-$HOME/nomad-certs/ca.pem}"
export NOMAD_CLIENT_CERT="${NOMAD_CLIENT_CERT:-$HOME/nomad-certs/cli.pem}"
export NOMAD_CLIENT_KEY="${NOMAD_CLIENT_KEY:-$HOME/nomad-certs/cli-key.pem}"

# Verify prerequisites
log_info "Vérification prérequis..."

# Check if in project root
if [ ! -d "$PROJECT_ROOT/cmd/scalping_live_bybit" ]; then
    log_error "Répertoire projet invalide"
    exit 1
fi

# Check config file
if [ ! -f "$PROJECT_ROOT/config/config.yaml" ]; then
    log_warning "config/config.yaml non trouvé (sera généré par Nomad)"
fi

# Check VPN
if ! ping -c 1 -W 2 10.8.0.1 &>/dev/null; then
    log_warning "VPN WireGuard semble inactif (ping 10.8.0.1 échoue)"
    log_info "Les commandes Nomad peuvent échouer"
fi

log_success "Prérequis OK"
echo ""

# Confirm deployment
read -p "🚀 Lancer le déploiement complet? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_warning "Déploiement annulé"
    exit 0
fi

echo ""
echo "================================================"
echo "📦 ÉTAPE 1/3: Déploiement Binaire"
echo "================================================"
echo ""

# Deploy binary
if [ -f "$SCRIPT_DIR/deploy-binary.sh" ]; then
    bash "$SCRIPT_DIR/deploy-binary.sh"
else
    log_error "Script deploy-binary.sh introuvable"
    exit 1
fi

echo ""
echo "================================================"
echo "⚙️  ÉTAPE 2/3: Upload Configuration"
echo "================================================"
echo ""

# Upload config if exists
if [ -f "$PROJECT_ROOT/config/config.yaml" ]; then
    log_info "Upload config.yaml..."
    scp "$PROJECT_ROOT/config/config.yaml" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/config/"
    log_success "Configuration uploadée"
else
    log_info "Pas de config locale, utilisation config Nomad template"
fi

echo ""
echo "================================================"
echo "🎯 ÉTAPE 3/3: Déploiement Job Nomad"
echo "================================================"
echo ""

# Deploy Nomad job
if [ -f "$SCRIPT_DIR/deploy-nomad-job.sh" ]; then
    bash "$SCRIPT_DIR/deploy-nomad-job.sh"
else
    log_error "Script deploy-nomad-job.sh introuvable"
    exit 1
fi

echo ""
echo "================================================"
echo "✅ DÉPLOIEMENT COMPLET TERMINÉ"
echo "================================================"
echo ""

# Get allocation ID
ALLOC_ID=$(nomad job allocs scalping-live-bybit -json 2>/dev/null | jq -r '.[0].ID' 2>/dev/null || echo "")

log_success "Application déployée avec succès!"
echo ""
echo "📊 Informations:"
echo "   Serveur: $REMOTE_HOST"
echo "   Path: $REMOTE_PATH"
echo "   Nomad: $NOMAD_ADDR"
if [ -n "$ALLOC_ID" ]; then
    echo "   Allocation: $ALLOC_ID"
fi
echo ""
echo "🔍 Vérification:"
echo "   nomad job status scalping-live-bybit"
if [ -n "$ALLOC_ID" ]; then
    echo "   nomad alloc logs -f $ALLOC_ID"
fi
echo "   ssh $REMOTE_USER@$REMOTE_HOST 'ls -lh $REMOTE_PATH'"
echo ""
echo "🌐 Accès:"
echo "   Nomad UI: $NOMAD_ADDR"
echo "   Logs: ssh $REMOTE_USER@$REMOTE_HOST 'tail -f $REMOTE_PATH/logs/scalping.log'"
echo ""
echo "📱 Notifications:"
echo "   Topic: scalping-live-bybit"
echo "   URL: https://notifications.koyad.com/scalping-live-bybit"
echo ""

log_success "Déploiement terminé!"
