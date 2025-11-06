#!/bin/bash
# 🎯 Déploiement Job Nomad Scalping Live Bybit
# Deploy via VPN WireGuard

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

# Configuration
export NOMAD_ADDR="${NOMAD_ADDR:-https://10.8.0.1:4646}"
export NOMAD_CACERT="${NOMAD_CACERT:-$HOME/nomad-certs/ca.pem}"
export NOMAD_CLIENT_CERT="${NOMAD_CLIENT_CERT:-$HOME/nomad-certs/cli.pem}"
export NOMAD_CLIENT_KEY="${NOMAD_CLIENT_KEY:-$HOME/nomad-certs/cli-key.pem}"
JOB_NAME="scalping-live-bybit"

echo -e "${BLUE}🎯 Déploiement Job Nomad${NC}"
echo "=============================================="
echo "Job: $JOB_NAME"
echo "Nomad: $NOMAD_ADDR"
echo "=============================================="
echo ""

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
JOB_FILE="$SCRIPT_DIR/../configs/scalping-live-bybit.nomad"

# Verify job file exists
if [ ! -f "$JOB_FILE" ]; then
    log_error "Job file not found: $JOB_FILE"
    exit 1
fi

log_info "Job file: $JOB_FILE"

# Test Nomad connection
log_info "Test connexion Nomad..."
if ! nomad status &>/dev/null; then
    log_error "Connexion Nomad échouée"
    echo ""
    log_warning "Vérifications:"
    echo "   1. VPN WireGuard actif? (wg show)"
    echo "   2. NOMAD_ADDR correct? (export NOMAD_ADDR=http://10.8.0.1:4646)"
    echo "   3. Nomad accessible? (ping 10.8.0.1)"
    exit 1
fi
log_success "Connexion Nomad OK"

# Check if job already exists
log_info "Vérification job existant..."
if nomad job status "$JOB_NAME" &>/dev/null; then
    log_warning "Job $JOB_NAME existe déjà"
    
    # Ask to stop
    read -p "⚠️  Arrêter et redéployer le job? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Arrêt job existant..."
        nomad job stop "$JOB_NAME"
        sleep 3
        log_success "Job arrêté"
    else
        log_warning "Déploiement annulé"
        exit 0
    fi
else
    log_info "Nouveau déploiement"
fi

# Validate job file
log_info "Validation job file..."
if nomad job validate "$JOB_FILE"; then
    log_success "Job file valide"
else
    log_error "Job file invalide"
    exit 1
fi

# Plan job (dry-run)
log_info "Planning job (dry-run)..."
nomad job plan "$JOB_FILE" || true
echo ""

# Deploy job
log_info "Déploiement job..."
nomad job run "$JOB_FILE"
sleep 2
log_success "Job déployé"

# Wait for allocation
log_info "Attente allocation..."
sleep 5

# Get allocation ID
ALLOC_ID=$(nomad job allocs "$JOB_NAME" -json | jq -r '.[0].ID' 2>/dev/null || echo "")

if [ -z "$ALLOC_ID" ]; then
    log_warning "Allocation ID introuvable"
    echo ""
    log_info "Vérifier manuellement:"
    echo "   nomad job status $JOB_NAME"
    exit 0
fi

log_success "Allocation ID: $ALLOC_ID"

# Check allocation status
log_info "Vérification status allocation..."
ALLOC_STATUS=$(nomad alloc status "$ALLOC_ID" | grep "Status" | head -1 | awk '{print $3}')
log_info "Status: $ALLOC_STATUS"

# Show job status
echo ""
log_info "Status job:"
nomad job status "$JOB_NAME"

echo ""
log_success "Déploiement job terminé!"
echo ""
echo "📊 Informations:"
echo "   Job: $JOB_NAME"
echo "   Allocation: $ALLOC_ID"
echo "   Status: $ALLOC_STATUS"
echo ""
echo "🔍 Commandes utiles:"
echo "   nomad job status $JOB_NAME"
echo "   nomad alloc status $ALLOC_ID"
echo "   nomad alloc logs -f $ALLOC_ID"
echo "   nomad alloc logs -stderr -f $ALLOC_ID"
echo ""
echo "🌐 Nomad UI:"
echo "   $NOMAD_ADDR"
echo ""

# Offer to follow logs
read -p "📋 Suivre les logs? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Logs en direct (Ctrl+C pour quitter)..."
    echo ""
    nomad alloc logs -f "$ALLOC_ID"
fi
