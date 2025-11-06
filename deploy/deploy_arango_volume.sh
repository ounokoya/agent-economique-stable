#!/bin/bash
# 📁 Création du volume pour ArangoDB sur le serveur distant

set -e

REMOTE_HOST="193.29.62.96"
REMOTE_USER="root"
VOLUME_PATH="/opt/arangodb_data"

echo "📁 Création du volume ArangoDB sur $REMOTE_HOST"
echo "================================================"

# Créer le dossier à distance avec permissions complètes
ssh $REMOTE_USER@$REMOTE_HOST "
    mkdir -p $VOLUME_PATH
    chmod 777 $VOLUME_PATH
    rm -rf $VOLUME_PATH/*
    echo '✅ Volume créé et vidé: $VOLUME_PATH'
    echo '✅ Permissions: 777 (lecture/écriture pour tous)'
"

if [ $? -eq 0 ]; then
    echo "✅ Volume prêt"
else
    echo "❌ Échec préparation volume"
    exit 1
fi

echo ""
echo "⚠️  N'oubliez pas de configurer le host volume dans Nomad:"
echo ""
echo "client {"
echo "  host_volume \"arangodb_data\" {"
echo "    path      = \"$VOLUME_PATH\""
echo "    read_only = false"
echo "  }"
echo "}"
echo ""
echo "Puis: sudo systemctl restart nomad"
echo "================================================"
