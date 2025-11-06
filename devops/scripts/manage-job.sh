#!/bin/bash
# 🔧 Gestion Job Nomad Scalping Live Bybit
# Utilitaire pour logs, status, restart, etc.

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

# Functions
show_help() {
    echo -e "${BLUE}🔧 Gestion Job Nomad${NC}"
    echo ""
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  status      Afficher status du job"
    echo "  logs        Suivre les logs (stdout)"
    echo "  errors      Suivre les erreurs (stderr)"
    echo "  restart     Redémarrer le job"
    echo "  stop        Arrêter le job"
    echo "  info        Informations détaillées"
    echo "  ui          Ouvrir Nomad UI (affiche URL)"
    echo "  help        Afficher cette aide"
    echo ""
    echo "Environment:"
    echo "  NOMAD_ADDR=$NOMAD_ADDR"
    echo ""
}

get_alloc_id() {
    nomad job allocs "$JOB_NAME" -json 2>/dev/null | jq -r '.[0].ID' 2>/dev/null || echo ""
}

cmd_status() {
    log_info "Status job $JOB_NAME..."
    echo ""
    nomad job status "$JOB_NAME"
}

cmd_logs() {
    ALLOC_ID=$(get_alloc_id)
    if [ -z "$ALLOC_ID" ]; then
        log_error "Allocation introuvable"
        exit 1
    fi
    
    log_info "Logs (stdout) - Allocation: $ALLOC_ID"
    log_info "Ctrl+C pour quitter"
    echo ""
    nomad alloc logs -f "$ALLOC_ID"
}

cmd_errors() {
    ALLOC_ID=$(get_alloc_id)
    if [ -z "$ALLOC_ID" ]; then
        log_error "Allocation introuvable"
        exit 1
    fi
    
    log_info "Erreurs (stderr) - Allocation: $ALLOC_ID"
    log_info "Ctrl+C pour quitter"
    echo ""
    nomad alloc logs -stderr -f "$ALLOC_ID"
}

cmd_restart() {
    log_warning "Redémarrage du job $JOB_NAME..."
    
    read -p "⚠️  Confirmer redémarrage? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_warning "Redémarrage annulé"
        exit 0
    fi
    
    log_info "Arrêt job..."
    nomad job stop "$JOB_NAME"
    sleep 3
    
    log_info "Redémarrage job..."
    SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    JOB_FILE="$SCRIPT_DIR/../configs/scalping-live-bybit.nomad"
    
    if [ ! -f "$JOB_FILE" ]; then
        log_error "Job file introuvable: $JOB_FILE"
        exit 1
    fi
    
    nomad job run "$JOB_FILE"
    sleep 2
    
    log_success "Job redémarré"
    echo ""
    cmd_status
}

cmd_stop() {
    log_warning "Arrêt du job $JOB_NAME..."
    
    read -p "⚠️  Confirmer arrêt? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_warning "Arrêt annulé"
        exit 0
    fi
    
    nomad job stop "$JOB_NAME"
    log_success "Job arrêté"
}

cmd_info() {
    log_info "Informations détaillées..."
    echo ""
    
    # Job info
    echo -e "${BLUE}📋 Job Info${NC}"
    nomad job status "$JOB_NAME" | head -20
    echo ""
    
    # Allocation info
    ALLOC_ID=$(get_alloc_id)
    if [ -n "$ALLOC_ID" ]; then
        echo -e "${BLUE}📦 Allocation Info${NC}"
        nomad alloc status "$ALLOC_ID" | head -30
        echo ""
    fi
    
    # Recent events
    if [ -n "$ALLOC_ID" ]; then
        echo -e "${BLUE}📊 Recent Events${NC}"
        nomad alloc status "$ALLOC_ID" | grep -A 10 "Recent Events" || true
        echo ""
    fi
    
    # Resource usage
    if [ -n "$ALLOC_ID" ]; then
        echo -e "${BLUE}💻 Resource Usage${NC}"
        nomad alloc status "$ALLOC_ID" | grep -A 5 "Allocated Resources" || true
    fi
}

cmd_ui() {
    echo -e "${BLUE}🌐 Nomad UI${NC}"
    echo ""
    echo "URL: $NOMAD_ADDR"
    echo "Job: $NOMAD_ADDR/ui/jobs/$JOB_NAME"
    echo ""
    log_info "Ouvrir dans navigateur (via VPN)"
}

# Main
if [ $# -eq 0 ]; then
    show_help
    exit 0
fi

COMMAND=$1

# Test Nomad connection
if ! nomad status &>/dev/null; then
    log_error "Connexion Nomad échouée"
    echo ""
    log_warning "Vérifications:"
    echo "   1. VPN WireGuard actif? (wg show)"
    echo "   2. NOMAD_ADDR correct? (echo \$NOMAD_ADDR)"
    echo "   3. Nomad accessible? (ping 10.8.0.1)"
    exit 1
fi

case $COMMAND in
    status)
        cmd_status
        ;;
    logs)
        cmd_logs
        ;;
    errors)
        cmd_errors
        ;;
    restart)
        cmd_restart
        ;;
    stop)
        cmd_stop
        ;;
    info)
        cmd_info
        ;;
    ui)
        cmd_ui
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Commande inconnue: $COMMAND"
        echo ""
        show_help
        exit 1
        ;;
esac
