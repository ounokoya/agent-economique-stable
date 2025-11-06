#!/bin/bash
# 🚀 Déploie tous les agents SOL (5m, 15m, 1h)

set -e

echo "🚀 Déploiement de tous les agents SOL"
echo "================================================"

# Arrêter l'ancien service si il existe
echo "🛑 Arrêt des anciens services..."
nomad job stop -address http://193.29.62.96:4646/ -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 agent-economique-notifications 2>/dev/null || true

sleep 2

# Déployer les 3 timeframes
echo ""
echo "📊 Déploiement SOL 5m..."
./deploy/deploy_agent.sh sol 5m

echo ""
echo "📊 Déploiement SOL 15m..."
./deploy/deploy_agent.sh sol 15m

echo ""
echo "📊 Déploiement SOL 1h..."
./deploy/deploy_agent.sh sol 1h

echo ""
echo "================================================"
echo "✅ Tous les agents SOL sont déployés !"
echo ""
echo "📱 Abonnez-vous à: notifications.koyad.com/notification-agent-eco"
echo "================================================"
