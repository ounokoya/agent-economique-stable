#!/bin/bash
# 🔨 Script de compilation et déploiement du binaire MFI+Stoch Live Notifications

set -e

# 🔧 CONFIGURATION
REMOTE_HOST="193.29.62.96"
REMOTE_USER="root"
REMOTE_BASE_DIR="/root/data/backtest-optimizer"
BINARY_NAME="mfi_stoch_live_notifications"

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

echo -e "${BLUE}🔨 Compilation et Déploiement du Binaire MFI+Stoch${NC}"
echo "================================================"
echo "• Binaire: $BINARY_NAME"
echo "• Serveur: $REMOTE_USER@$REMOTE_HOST"
echo "• Destination: $REMOTE_BASE_DIR/"
echo "================================================"

# 🔍 VÉRIFICATIONS PRÉALABLES
log_info "Vérification des prérequis..."

# Vérifier que Go est installé
if ! command -v go &> /dev/null; then
    log_error "Go n'est pas installé"
    exit 1
fi

# Vérifier que le code source existe
if [ ! -f "backend/cmd/mfi_stoch_live_notifications/main.go" ]; then
    log_error "Code source manquant: backend/cmd/mfi_stoch_live_notifications/main.go"
    exit 1
fi

log_success "Prérequis validés"

# 🔨 COMPILATION
log_info "Compilation du binaire MFI+Stoch..."

cd backend
if ! go build -o $BINARY_NAME cmd/mfi_stoch_live_notifications/main.go; then
    log_error "Échec de la compilation"
    exit 1
fi
cd ..

log_success "Binaire compilé: backend/$BINARY_NAME"

# 🔗 TEST DE CONNECTIVITÉ SSH
log_info "Test de connectivité SSH..."

if ! ssh -o ConnectTimeout=10 $REMOTE_USER@$REMOTE_HOST "echo 'SSH OK'" > /dev/null 2>&1; then
    log_error "Impossible de se connecter à $REMOTE_USER@$REMOTE_HOST"
    log_warning "Vérifiez les clés SSH et la connectivité réseau"
    exit 1
fi

log_success "Connectivité SSH validée"

# 🏗️ CRÉATION DES DOSSIERS DISTANTS
log_info "Création/vérification des dossiers distants..."

FOLDERS=("out" "state" "configs" "data" "logs")
for folder in "${FOLDERS[@]}"; do
    ssh $REMOTE_USER@$REMOTE_HOST "mkdir -p $REMOTE_BASE_DIR/$folder"
    log_info "  ✓ $REMOTE_BASE_DIR/$folder"
done

log_success "Arborescence distante préparée"

# 📤 UPLOAD DU BINAIRE
log_info "Upload du binaire sur le serveur distant..."

if ! scp backend/$BINARY_NAME $REMOTE_USER@$REMOTE_HOST:$REMOTE_BASE_DIR/; then
    log_error "Échec de l'upload du binaire"
    exit 1
fi

log_success "Binaire uploadé"

# 🔐 PERMISSIONS
log_info "Configuration des permissions..."

ssh $REMOTE_USER@$REMOTE_HOST "chmod +x $REMOTE_BASE_DIR/$BINARY_NAME"

log_success "Permissions configurées"

# 🧪 TEST DU BINAIRE DISTANT
log_info "Test du binaire sur le serveur distant..."

if ssh $REMOTE_USER@$REMOTE_HOST "$REMOTE_BASE_DIR/$BINARY_NAME -h" 2>&1 | grep -q "config"; then
    log_success "Binaire fonctionnel sur le serveur distant"
else
    log_warning "Le binaire ne répond pas normalement (peut nécessiter des dépendances)"
fi

# 🧹 NETTOYAGE LOCAL
log_info "Nettoyage du binaire local..."
rm -f backend/$BINARY_NAME

# 📋 VÉRIFICATION FINALE
log_info "Vérification finale..."

echo ""
echo "📁 Fichiers distants:"
ssh $REMOTE_USER@$REMOTE_HOST "ls -la $REMOTE_BASE_DIR/ | grep -E 'mfi_stoch|agent_economique'"

echo ""
echo "================================================"
log_success "Déploiement du binaire MFI+Stoch terminé!"
echo -e "${BLUE}📍 Binaire déployé:${NC}"
echo "   $REMOTE_BASE_DIR/$BINARY_NAME"
echo ""
echo -e "${BLUE}🚀 Prochaine étape:${NC}"
echo "   ./deploy/deploy_mfi_stoch_extended.sh"
echo "================================================"
