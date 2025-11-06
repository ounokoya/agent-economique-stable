#!/bin/bash
# 🔒 Génération Certificats TLS pour Nomad
# CA + Server + Client certificates

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

echo -e "${BLUE}🔒 Génération Certificats TLS Nomad${NC}"
echo "=========================================="

# Configuration
SERVER_IP="31.57.224.79"
SERVER_VPN_IP="10.8.0.1"
CERT_DIR="/etc/nomad.d/certs"
VALIDITY_DAYS=3650  # 10 ans

# Check root
if [ "$EUID" -ne 0 ]; then 
    log_error "Ce script doit être exécuté en tant que root"
    exit 1
fi

# Install cfssl if not present
if ! command -v cfssl &> /dev/null; then
    log_info "Installation cfssl..."
    apt update
    apt install -y golang-cfssl
    log_success "cfssl installé"
fi

# Create cert directory
log_info "Création dossier certificats..."
mkdir -p "$CERT_DIR"
cd "$CERT_DIR"
log_success "Dossier: $CERT_DIR"

# 1. Generate CA
log_info "Génération Certificate Authority (CA)..."
cat > ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "87600h",
      "usages": [
        "signing",
        "key encipherment",
        "server auth",
        "client auth"
      ]
    }
  }
}
EOF

cat > ca-csr.json <<EOF
{
  "CN": "Nomad CA",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "SG",
      "L": "Singapore",
      "O": "Nomad",
      "OU": "CA",
      "ST": "Singapore"
    }
  ]
}
EOF

cfssl gencert -initca ca-csr.json | cfssljson -bare ca
log_success "CA générée: ca.pem, ca-key.pem"

# 2. Generate Server Certificate
log_info "Génération certificat serveur..."
cat > server-csr.json <<EOF
{
  "CN": "server.global.nomad",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "SG",
      "L": "Singapore",
      "O": "Nomad",
      "OU": "Server",
      "ST": "Singapore"
    }
  ]
}
EOF

# Add SANs (Subject Alternative Names)
cat > server-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "87600h",
      "usages": [
        "signing",
        "key encipherment",
        "server auth",
        "client auth"
      ]
    }
  }
}
EOF

cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=server-config.json \
  -hostname="server.global.nomad,localhost,127.0.0.1,$SERVER_IP,$SERVER_VPN_IP" \
  server-csr.json | cfssljson -bare server

log_success "Certificat serveur: server.pem, server-key.pem"

# 3. Generate Client Certificate
log_info "Génération certificat client..."
cat > client-csr.json <<EOF
{
  "CN": "client.global.nomad",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "SG",
      "L": "Singapore",
      "O": "Nomad",
      "OU": "Client",
      "ST": "Singapore"
    }
  ]
}
EOF

cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -hostname="client.global.nomad,localhost,127.0.0.1" \
  client-csr.json | cfssljson -bare client

log_success "Certificat client: client.pem, client-key.pem"

# 4. Generate CLI Certificate (for nomad CLI)
log_info "Génération certificat CLI..."
cat > cli-csr.json <<EOF
{
  "CN": "cli.global.nomad",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "SG",
      "L": "Singapore",
      "O": "Nomad",
      "OU": "CLI",
      "ST": "Singapore"
    }
  ]
}
EOF

cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -hostname="cli.global.nomad,localhost,127.0.0.1" \
  cli-csr.json | cfssljson -bare cli

log_success "Certificat CLI: cli.pem, cli-key.pem"

# Set permissions
log_info "Configuration permissions..."
chmod 600 *-key.pem
chmod 644 *.pem
chown -R root:root "$CERT_DIR"
log_success "Permissions configurées"

# Cleanup JSON files
rm -f *.json *.csr

# Summary
echo ""
log_success "Certificats générés avec succès!"
echo ""
echo "📁 Emplacement: $CERT_DIR"
echo ""
echo "📋 Fichiers générés:"
ls -lh "$CERT_DIR"/*.pem
echo ""
echo "🔐 Certificats:"
echo "   CA:     ca.pem, ca-key.pem"
echo "   Server: server.pem, server-key.pem"
echo "   Client: client.pem, client-key.pem"
echo "   CLI:    cli.pem, cli-key.pem"
echo ""
echo "📤 À copier sur machine locale:"
echo "   ca.pem"
echo "   cli.pem"
echo "   cli-key.pem"
echo ""
echo "💾 Sauvegarde locale recommandée:"
echo "   mkdir -p ~/nomad-certs-backup"
echo "   cp $CERT_DIR/*.pem ~/nomad-certs-backup/"
echo ""

# Create export package for local machine
EXPORT_DIR="/tmp/nomad-client-certs"
mkdir -p "$EXPORT_DIR"
cp ca.pem cli.pem cli-key.pem "$EXPORT_DIR/"

log_info "Package client créé: $EXPORT_DIR"
echo ""
echo "🚀 Pour utiliser depuis machine locale:"
echo "   scp -r root@$SERVER_IP:$EXPORT_DIR ~/nomad-certs"
echo ""
echo "   export NOMAD_ADDR=\"https://10.8.0.1:4646\""
echo "   export NOMAD_CACERT=\"\$HOME/nomad-certs/ca.pem\""
echo "   export NOMAD_CLIENT_CERT=\"\$HOME/nomad-certs/cli.pem\""
echo "   export NOMAD_CLIENT_KEY=\"\$HOME/nomad-certs/cli-key.pem\""
echo ""

log_success "Génération certificats terminée!"
