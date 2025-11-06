#!/bin/bash
# 🚀 Script de déploiement du job Nomad Agent Economique Live

set -e

# 🔧 CONFIGURATION NOMAD
NOMAD_ADDR="http://193.29.62.96:4646/"
NOMAD_TOKEN="1fc424de-5992-f4a5-c90e-cccabd7ef5d9"
CERTS_DIR="certs"
JOB_FILE="deploy/agent-economique-notifications.nomad"
JOB_NAME="agent-economique-notifications"

# 🎨 COULEURS
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

echo -e "${BLUE}🚀 Déploiement Job Nomad Agent Economique${NC}"
echo "============================================="
echo "• Job: $JOB_NAME"
echo "• Fichier: $JOB_FILE"
echo "• Cluster: $NOMAD_ADDR"
echo "============================================="

# 🔍 VÉRIFICATIONS PRÉALABLES
log_info "Vérification des prérequis..."

# Vérifier le fichier job Nomad
if [ ! -f "$JOB_FILE" ]; then
    log_error "Fichier job manquant: $JOB_FILE"
    exit 1
fi

# Vérifier les certificats TLS
if [ ! -d "$CERTS_DIR" ] || [ ! -f "$CERTS_DIR/ca.pem" ]; then
    log_error "Certificats TLS manquants dans $CERTS_DIR/"
    log_error "Copier les certificats depuis backtest_optimizer/certs/"
    exit 1
fi

log_success "Certificats TLS trouvés"

# Vérifier Nomad CLI
if ! command -v nomad &> /dev/null; then
    log_warning "Nomad CLI non installé. Installation..."
    curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
    sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
    sudo apt-get update && sudo apt-get install nomad
    log_success "Nomad CLI installé"
fi

log_success "Prérequis validés"

# 📊 AFFICHAGE CONFIGURATION
log_info "Configuration du job (définie dans $JOB_FILE):"
echo "  • Symbol: SUIUSDT"
echo "  • Timeframe: 15m"
echo "  • Stratégie: Agent Economique (CCI+MFI+Stoch)"
echo "  • MM: Paliers progressifs (25/50/75/100%)"
echo "  • Notifications: notifications.koyad.com/notification-agent-eco"

# 🛑 ARRÊT DU JOB EXISTANT
log_info "Vérification job existant..."

if nomad job status -address "$NOMAD_ADDR" -token "$NOMAD_TOKEN" "$JOB_NAME" &> /dev/null; then
    log_warning "Job existant trouvé. Arrêt en cours..."
    nomad job stop -address "$NOMAD_ADDR" -token "$NOMAD_TOKEN" "$JOB_NAME" || true
    sleep 3
    log_success "Job existant arrêté"
else
    log_info "Aucun job existant trouvé"
fi

# 🚀 DÉPLOIEMENT
log_info "Déploiement du job Nomad avec certificats TLS..."

nomad job run \
    -token "$NOMAD_TOKEN" \
    -address "$NOMAD_ADDR" \
    -ca-cert="$CERTS_DIR/ca.pem" \
    -client-cert="$CERTS_DIR/client.pem" \
    -client-key="$CERTS_DIR/client-key.pem" \
    "$JOB_FILE"

log_success "Job déployé avec succès"

# ⏳ ATTENTE ET VÉRIFICATION
log_info "Attente du démarrage..."
sleep 5

echo ""
log_info "Statut du job:"
nomad job status -address "$NOMAD_ADDR" -token "$NOMAD_TOKEN" "$JOB_NAME"

echo ""
log_info "Allocations:"
nomad job allocs -address "$NOMAD_ADDR" -token "$NOMAD_TOKEN" "$JOB_NAME"

# 📋 LOGS RÉCENTS
echo ""
log_info "Logs récents (10 dernières lignes):"
ALLOC_ID=$(nomad job allocs -address "$NOMAD_ADDR" -token "$NOMAD_TOKEN" "$JOB_NAME" -json | jq -r '.[0].ID' 2>/dev/null || echo "")

if [ -n "$ALLOC_ID" ] && [ "$ALLOC_ID" != "null" ]; then
    timeout 10s nomad alloc logs -address "$NOMAD_ADDR" -token "$NOMAD_TOKEN" -tail -n 10 "$ALLOC_ID" 2>/dev/null || log_warning "Logs pas encore disponibles"
else
    log_warning "Allocation ID non trouvée"
fi

# ✅ RÉSUMÉ FINAL
echo ""
echo "============================================="
log_success "Déploiement Nomad terminé!"
echo ""
echo -e "${BLUE}📱 Pour recevoir les notifications:${NC}"
echo "   1. Installer l'app ntfy"
echo "   2. S'abonner à: notifications.koyad.com/notification-agent-eco"
echo ""
echo -e "${BLUE}🔧 Pour modifier la config:${NC}"
echo "   1. Éditer: $JOB_FILE (templates)"
echo "   2. Relancer: ./deploy/deploy_nomad.sh"
echo ""
echo -e "${BLUE}📊 Commandes utiles:${NC}"
echo "   • Logs: nomad alloc logs -address $NOMAD_ADDR -token $NOMAD_TOKEN -f $ALLOC_ID"
echo "   • Statut: nomad job status -address $NOMAD_ADDR -token $NOMAD_TOKEN $JOB_NAME"
echo "   • Arrêt: nomad job stop -address $NOMAD_ADDR -token $NOMAD_TOKEN $JOB_NAME"
echo "============================================="
