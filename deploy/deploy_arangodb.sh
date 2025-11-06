#!/bin/bash
# 🗄️ Déploiement ArangoDB sur Nomad
# Prérequis: Volume créé avec deploy_arango_volume.sh

set -e

# 🔧 CONFIGURATION NOMAD
NOMAD_ADDR="http://193.29.62.96:4646/"
NOMAD_TOKEN="1fc424de-5992-f4a5-c90e-cccabd7ef5d9"
CERTS_DIR="certs"
JOB_FILE="deploy/arangodb.nomad"
JOB_NAME="arangodb-agent-economique"

echo "🗄️ Déploiement ArangoDB sur Nomad"
echo "================================================"

# Vérifier le fichier job
if [ ! -f "$JOB_FILE" ]; then
    echo "❌ Fichier job manquant: $JOB_FILE"
    exit 1
fi

# Vérifier les certificats TLS
if [ ! -d "$CERTS_DIR" ] || [ ! -f "$CERTS_DIR/ca.pem" ]; then
    echo "❌ Certificats TLS manquants dans $CERTS_DIR/"
    exit 1
fi

# Déploiement avec CLI Nomad local
echo "🚀 Déploiement sur Nomad..."
nomad job run \
    -token "$NOMAD_TOKEN" \
    -address "$NOMAD_ADDR" \
    -ca-cert="$CERTS_DIR/ca.pem" \
    -client-cert="$CERTS_DIR/client.pem" \
    -client-key="$CERTS_DIR/client-key.pem" \
    "$JOB_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Job déployé avec succès"
else
    echo "❌ Échec du déploiement"
    exit 1
fi

echo ""
echo "📊 Commandes utiles:"
echo "   Status: nomad job status -address $NOMAD_ADDR -token $NOMAD_TOKEN $JOB_NAME"
echo "   Logs: nomad alloc logs -address $NOMAD_ADDR -token $NOMAD_TOKEN -f <alloc-id>"
echo "================================================"
