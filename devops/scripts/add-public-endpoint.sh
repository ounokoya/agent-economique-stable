#!/bin/bash
# Script pour ajouter un endpoint public à Caddy
# Usage: ./add-public-endpoint.sh <port> <backend-url>

set -e

PORT=${1:-3000}
BACKEND=${2:-localhost:8000}

echo "📝 Ajout endpoint public: :$PORT → $BACKEND"

# Backup Caddyfile
cp /etc/caddy/Caddyfile /etc/caddy/Caddyfile.backup

# Ajouter endpoint public
cat >> /etc/caddy/Caddyfile << EOF

# Endpoint public ajouté le $(date)
:$PORT {
    bind 0.0.0.0  # Public
    
    reverse_proxy $BACKEND
    
    log {
        output file /var/log/caddy/public-$PORT.log
    }
}
EOF

# Ouvrir port firewall
ufw allow $PORT/tcp

# Recharger Caddy
caddy reload --config /etc/caddy/Caddyfile

echo "✅ Endpoint public configuré:"
echo "   URL: http://31.57.224.79:$PORT"
echo "   Backend: $BACKEND"
echo ""
echo "⚠️  Nomad UI reste privé (VPN uniquement)"
