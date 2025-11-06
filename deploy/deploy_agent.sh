#!/bin/bash
# 🚀 Script de déploiement Agent Economique pour une paire/TF spécifique

set -e

# 🔧 CONFIGURATION NOMAD
NOMAD_ADDR="http://193.29.62.96:4646/"
NOMAD_TOKEN="1fc424de-5992-f4a5-c90e-cccabd7ef5d9"
CERTS_DIR="certs"

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

# Usage
if [ $# -ne 2 ]; then
    echo "Usage: $0 <symbol> <timeframe>"
    echo ""
    echo "Examples:"
    echo "  $0 sol 5m"
    echo "  $0 sol 15m"
    echo "  $0 sol 1h"
    echo ""
    exit 1
fi

SYMBOL=$(echo "$1" | tr '[:upper:]' '[:lower:]')
TIMEFRAME="$2"
JOB_FILE="deploy/agent-economique-${SYMBOL}-${TIMEFRAME}.nomad"
JOB_NAME="agent-economique-${SYMBOL}-${TIMEFRAME}"

echo -e "${BLUE}🚀 Déploiement Agent Economique${NC}"
echo "================================================"
echo "• Symbol: ${SYMBOL^^}USDT"
echo "• Timeframe: $TIMEFRAME"
echo "• Job: $JOB_NAME"
echo "================================================"

# Vérifier le fichier job
if [ ! -f "$JOB_FILE" ]; then
    log_error "Fichier job manquant: $JOB_FILE"
    echo ""
    echo "Fichiers disponibles:"
    ls -1 deploy/agent-economique-*.nomad | sed 's/deploy\//  • /'
    exit 1
fi

# Vérifier les certificats TLS
if [ ! -d "$CERTS_DIR" ] || [ ! -f "$CERTS_DIR/ca.pem" ]; then
    log_error "Certificats TLS manquants dans $CERTS_DIR/"
    exit 1
fi

# Déploiement avec CLI Nomad local
log_info "Déploiement sur Nomad..."
nomad job run \
    -token "$NOMAD_TOKEN" \
    -address "$NOMAD_ADDR" \
    -ca-cert="$CERTS_DIR/ca.pem" \
    -client-cert="$CERTS_DIR/client.pem" \
    -client-key="$CERTS_DIR/client-key.pem" \
    "$JOB_FILE"

if [ $? -eq 0 ]; then
    log_success "Job déployé avec succès"
else
    log_error "Échec du déploiement"
    exit 1
fi

echo ""
echo "================================================"
log_success "Déploiement terminé!"
echo ""
echo -e "${BLUE}📊 Commandes utiles:${NC}"
echo "   Status: nomad job status -address $NOMAD_ADDR -token $NOMAD_TOKEN $JOB_NAME"
echo "   Logs: nomad alloc logs -address $NOMAD_ADDR -token $NOMAD_TOKEN -f -task agent-${SYMBOL}-${TIMEFRAME} \$(nomad job allocs -address $NOMAD_ADDR -token $NOMAD_TOKEN $JOB_NAME -json | jq -r '.[0].ID')"
echo "   Stop: nomad job stop -address $NOMAD_ADDR -token $NOMAD_TOKEN $JOB_NAME"
echo "================================================"
