#!/bin/bash
# 📦 Déploiement binaire Scalping Live Bybit
# Compile et upload vers serveur Singapour

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
REMOTE_HOST="31.57.224.79"
REMOTE_USER="root"
REMOTE_PATH="/root/data/scalping-live-bybit"
BINARY_NAME="scalping_live_bybit"
APP_SOURCE="cmd/scalping_live_bybit"

echo -e "${BLUE}📦 Déploiement Binaire Scalping Live Bybit${NC}"
echo "=============================================="
echo "Serveur: $REMOTE_USER@$REMOTE_HOST"
echo "Path: $REMOTE_PATH"
echo "Binary: $BINARY_NAME"
echo "=============================================="
echo ""

# Get project root (2 levels up from scripts/)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

cd "$PROJECT_ROOT"

# Verify source exists
if [ ! -d "$APP_SOURCE" ]; then
    log_error "Source directory not found: $APP_SOURCE"
    exit 1
fi

log_info "Project root: $PROJECT_ROOT"

# Compile for Linux
log_info "Compilation pour Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "$BINARY_NAME" "./$APP_SOURCE"

if [ ! -f "$BINARY_NAME" ]; then
    log_error "Compilation échouée"
    exit 1
fi

BINARY_SIZE=$(du -h "$BINARY_NAME" | cut -f1)
log_success "Binaire compilé: $BINARY_SIZE"

# Test SSH connection
log_info "Test connexion SSH..."
if ! ssh -o ConnectTimeout=5 "$REMOTE_USER@$REMOTE_HOST" "echo 'SSH OK'" &>/dev/null; then
    log_error "Connexion SSH échouée"
    rm "$BINARY_NAME"
    exit 1
fi
log_success "Connexion SSH OK"

# Create remote directory
log_info "Création dossier distant..."
ssh "$REMOTE_USER@$REMOTE_HOST" "mkdir -p $REMOTE_PATH/{config,logs,state,data}"
log_success "Dossiers créés"

# Upload binary
log_info "Upload binaire vers serveur..."
scp "$BINARY_NAME" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"

# Set permissions
ssh "$REMOTE_USER@$REMOTE_HOST" "chmod +x $REMOTE_PATH/$BINARY_NAME"
log_success "Binaire uploadé et configuré"

# Verify binary on server
log_info "Vérification binaire distant..."
REMOTE_SIZE=$(ssh "$REMOTE_USER@$REMOTE_HOST" "du -h $REMOTE_PATH/$BINARY_NAME | cut -f1")
log_success "Binaire distant: $REMOTE_SIZE"

# Test binary
log_info "Test binaire (version check)..."
if ssh "$REMOTE_USER@$REMOTE_HOST" "$REMOTE_PATH/$BINARY_NAME -version" &>/dev/null; then
    log_success "Binaire fonctionnel"
else
    log_warning "Test version échoué (peut être normal si flag -version non supporté)"
fi

# Cleanup local binary
log_info "Nettoyage binaire local..."
rm "$BINARY_NAME"
log_success "Nettoyage effectué"

# Summary
echo ""
log_success "Déploiement binaire terminé!"
echo ""
echo "📊 Résumé:"
echo "   Binaire: $BINARY_NAME"
echo "   Taille: $REMOTE_SIZE"
echo "   Path: $REMOTE_PATH/$BINARY_NAME"
echo ""
echo "🔍 Vérification manuelle:"
echo "   ssh $REMOTE_USER@$REMOTE_HOST"
echo "   ls -lh $REMOTE_PATH/$BINARY_NAME"
echo ""
echo "🚀 Prochaine étape:"
echo "   ./deploy-nomad-job.sh"
