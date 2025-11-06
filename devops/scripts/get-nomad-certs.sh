#!/bin/bash
# 📥 Récupération Certificats Nomad Client
# Download certificates from server to local machine

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
REMOTE_CERT_PATH="/tmp/nomad-client-certs"
LOCAL_CERT_DIR="$HOME/.nomad-certs"

echo -e "${BLUE}📥 Récupération Certificats Nomad Client${NC}"
echo "=============================================="
echo "Serveur: $REMOTE_USER@$REMOTE_HOST"
echo "Local: $LOCAL_CERT_DIR"
echo "=============================================="
echo ""

# Test SSH connection
log_info "Test connexion SSH..."
if ! ssh -o ConnectTimeout=5 "$REMOTE_USER@$REMOTE_HOST" "echo 'OK'" &>/dev/null; then
    log_error "Connexion SSH échouée"
    exit 1
fi
log_success "Connexion SSH OK"

# Check if certificates exist on server
log_info "Vérification certificats sur serveur..."
if ! ssh "$REMOTE_USER@$REMOTE_HOST" "test -d $REMOTE_CERT_PATH"; then
    log_error "Certificats non trouvés sur serveur: $REMOTE_CERT_PATH"
    echo ""
    log_info "Les certificats doivent être générés d'abord sur le serveur:"
    echo "   ssh $REMOTE_USER@$REMOTE_HOST"
    echo "   cd /root/agent_economique_stable/devops/scripts"
    echo "   ./generate-nomad-certs.sh"
    exit 1
fi
log_success "Certificats trouvés sur serveur"

# List certificates on server
log_info "Certificats disponibles:"
ssh "$REMOTE_USER@$REMOTE_HOST" "ls -lh $REMOTE_CERT_PATH/"
echo ""

# Create local directory
log_info "Création dossier local: $LOCAL_CERT_DIR"
mkdir -p "$LOCAL_CERT_DIR"
log_success "Dossier créé"

# Download certificates
log_info "Téléchargement certificats..."
scp -r "$REMOTE_USER@$REMOTE_HOST:$REMOTE_CERT_PATH/*" "$LOCAL_CERT_DIR/"
log_success "Certificats téléchargés"

# Set permissions
log_info "Configuration permissions..."
chmod 600 "$LOCAL_CERT_DIR"/*.pem
log_success "Permissions configurées"

# Verify downloaded files
echo ""
log_success "Certificats récupérés avec succès!"
echo ""
echo "📁 Emplacement local: $LOCAL_CERT_DIR"
echo ""
echo "📋 Fichiers téléchargés:"
ls -lh "$LOCAL_CERT_DIR"/*.pem
echo ""

# Create environment file
ENV_FILE="$LOCAL_CERT_DIR/nomad-env.sh"
cat > "$ENV_FILE" <<EOF
#!/bin/bash
# Nomad Environment Variables (TLS)
# Source this file: source ~/.nomad-certs/nomad-env.sh

export NOMAD_ADDR="https://10.8.0.1:4646"
export NOMAD_CACERT="$LOCAL_CERT_DIR/ca.pem"
export NOMAD_CLIENT_CERT="$LOCAL_CERT_DIR/cli.pem"
export NOMAD_CLIENT_KEY="$LOCAL_CERT_DIR/cli-key.pem"

echo "✅ Nomad environment configured (TLS)"
echo "   NOMAD_ADDR: \$NOMAD_ADDR"
EOF

chmod +x "$ENV_FILE"
log_success "Fichier environnement créé: $ENV_FILE"

echo ""
echo "🎯 Configuration Nomad CLI:"
echo ""
echo "Option 1 - Temporaire (session actuelle):"
echo "   source $ENV_FILE"
echo ""
echo "Option 2 - Permanent (ajouter à ~/.bashrc):"
echo "   echo 'source $ENV_FILE' >> ~/.bashrc"
echo "   source ~/.bashrc"
echo ""
echo "Option 3 - Manuel (copier-coller):"
echo "   export NOMAD_ADDR=\"https://10.8.0.1:4646\""
echo "   export NOMAD_CACERT=\"$LOCAL_CERT_DIR/ca.pem\""
echo "   export NOMAD_CLIENT_CERT=\"$LOCAL_CERT_DIR/cli.pem\""
echo "   export NOMAD_CLIENT_KEY=\"$LOCAL_CERT_DIR/cli-key.pem\""
echo ""
echo "🔍 Tester la connexion:"
echo "   source $ENV_FILE"
echo "   ping 10.8.0.1                 # VPN actif?"
echo "   nomad server members          # Nomad accessible?"
echo ""

log_success "Récupération certificats terminée!"
