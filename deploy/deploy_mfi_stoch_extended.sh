#!/bin/bash
# 🚀 Deploy MFI+Stoch Extended Strategy on 1m, 5m, 15m

set -e

# 🔧 CONFIGURATION NOMAD
NOMAD_ADDR="http://193.29.62.96:4646/"
NOMAD_TOKEN="1fc424de-5992-f4a5-c90e-cccabd7ef5d9"
CERTS_DIR="certs"
BINARY_PATH="/root/data/backtest-optimizer/mfi_stoch_live_notifications"

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

echo -e "${BLUE}🚀 Déploiement MFI+Stoch Extended Strategy${NC}"
echo "============================================="
echo "• Stratégie: MFI+Stoch (Double Confluence)"
echo "• MM: Extended avec re-entry"
echo "• Timeframes: 1m, 5m, 15m"
echo "• Symbol: SOLUSDT"
echo "============================================="
echo ""

# 🔍 VÉRIFIER CERTIFICATS
log_info "Vérification certificats TLS..."
if [ ! -d "$CERTS_DIR" ] || [ ! -f "$CERTS_DIR/ca.pem" ]; then
    log_error "Certificats TLS manquants dans $CERTS_DIR/"
    exit 1
fi
log_success "Certificats TLS trouvés"

# 📦 DÉPLOIEMENT DU BINAIRE SUR LE SERVEUR DISTANT
log_info "Déploiement du binaire sur le serveur distant..."
./deploy/deploy_mfi_stoch_binary.sh

# Vérifier que le binaire distant existe
log_info "Vérification du binaire distant..."
REMOTE_HOST="193.29.62.96"
REMOTE_USER="root"
if ! ssh $REMOTE_USER@$REMOTE_HOST "test -f $BINARY_PATH && test -x $BINARY_PATH"; then
    log_error "Binaire distant non trouvé ou non exécutable: $BINARY_PATH"
    exit 1
fi
log_success "Binaire distant validé: $BINARY_PATH"
echo ""

# 🛑 ARRÊTER LES ANCIENS JOBS
log_info "Arrêt des anciens jobs Agent Economique..."

OLD_JOBS=("agent-economique-notifications" "agent-economique-sol-5m" "agent-economique-sol-15m" "agent-economique-sol-1h")
for job in "${OLD_JOBS[@]}"; do
    if nomad job status -address "$NOMAD_ADDR" -token "$NOMAD_TOKEN" "$job" &>/dev/null; then
        log_warning "Arrêt de $job..."
        nomad job stop -address "$NOMAD_ADDR" -token "$NOMAD_TOKEN" "$job" 2>/dev/null || true
        sleep 1
    fi
done

log_success "Anciens jobs arrêtés"
sleep 2
echo ""

# 🚀 DÉPLOIEMENT DES NOUVEAUX JOBS
log_info "Déploiement MFI+Stoch Extended 15m (RECOMMANDÉ)..."
nomad job run \
    -token "$NOMAD_TOKEN" \
    -address "$NOMAD_ADDR" \
    -ca-cert="$CERTS_DIR/ca.pem" \
    -client-cert="$CERTS_DIR/client.pem" \
    -client-key="$CERTS_DIR/client-key.pem" \
    deploy/mfi-stoch-extended-15m.nomad
log_success "15m déployé"
echo ""

log_info "Déploiement MFI+Stoch Extended 5m..."
nomad job run \
    -token "$NOMAD_TOKEN" \
    -address "$NOMAD_ADDR" \
    -ca-cert="$CERTS_DIR/ca.pem" \
    -client-cert="$CERTS_DIR/client.pem" \
    -client-key="$CERTS_DIR/client-key.pem" \
    deploy/mfi-stoch-extended-5m.nomad
log_success "5m déployé"
echo ""

log_info "Déploiement MFI+Stoch Extended 1m (HAUTE FRÉQUENCE)..."
nomad job run \
    -token "$NOMAD_TOKEN" \
    -address "$NOMAD_ADDR" \
    -ca-cert="$CERTS_DIR/ca.pem" \
    -client-cert="$CERTS_DIR/client.pem" \
    -client-key="$CERTS_DIR/client-key.pem" \
    deploy/mfi-stoch-extended-1m.nomad
log_success "1m déployé"
echo ""

# ✅ RÉSUMÉ FINAL
echo "============================================="
log_success "Déploiement MFI+Stoch Extended terminé!"
echo ""
echo -e "${BLUE}📱 Notifications ntfy:${NC}"
echo "   • 15m (🏆 MEILLEUR): notifications.koyad.com/mfi-stoch-15m"
echo "   • 5m: notifications.koyad.com/mfi-stoch-5m"
echo "   • 1m (⚡ ACTIF): notifications.koyad.com/mfi-stoch-1m"
echo ""
echo -e "${BLUE}📊 Performances backtestées (2024):${NC}"
echo "   • 15m: +6049% (4,764 trades, WR 85.4%) 🥇"
echo "   • 1m:  +3622% (52,378 trades, WR 77.2%)"
echo "   • 5m:  Non testé ⚠️"
echo ""
echo -e "${BLUE}🔧 Commandes utiles:${NC}"
echo "   Status: nomad job status -address $NOMAD_ADDR -token $NOMAD_TOKEN mfi-stoch-extended-15m"
echo "   Arrêt: nomad job stop -address $NOMAD_ADDR -token $NOMAD_TOKEN mfi-stoch-extended-15m"
echo "============================================="
