# 🚀 Workflow Déploiement Scalping Live Bybit

**Serveur:** 31.57.224.79 (Singapour)  
**VPN:** 10.8.0.1  
**Application:** scalping_live_bybit  
**Exchange:** Bybit USDT Perpetual

---

## 📋 Prérequis

- ✅ Nomad Server opérationnel (doc 01)
- ✅ WireGuard VPN configuré (doc 02)
- ✅ Connexion VPN active
- ✅ Accès Bybit API vérifié

---

## 🔧 Variables d'Environnement

### **Configuration Locale**

```bash
# .env or export in ~/.bashrc
export NOMAD_ADDR="http://10.8.0.1:4646"
export REMOTE_HOST="10.8.0.1"
export REMOTE_SSH="root@31.57.224.79"  # Public IP for SSH
```

---

## 📦 Structure Déploiement

```
devops/
├── scripts/
│   ├── deploy-binary.sh           # Upload binaire
│   ├── deploy-nomad-job.sh        # Déployer job Nomad
│   └── full-deploy.sh             # Déploiement complet
├── configs/
│   └── scalping-live-bybit.nomad  # Job definition
└── docs/
    └── 03-deployment-workflow.md  # Cette doc
```

---

## 🚀 Étape 1 : Compilation & Upload Binaire

### **Script : `devops/scripts/deploy-binary.sh`**

```bash
#!/bin/bash
set -e

echo "🔨 Compilation Scalping Live Bybit..."

# Compile for Linux
GOOS=linux GOARCH=amd64 go build -o scalping_live_bybit ./cmd/scalping_live_bybit

echo "✅ Binaire compilé"

# Upload to server
echo "📤 Upload vers serveur Singapour..."
scp scalping_live_bybit root@31.57.224.79:/root/data/scalping-live-bybit/

# Set permissions
ssh root@31.57.224.79 "chmod +x /root/data/scalping-live-bybit/scalping_live_bybit"

echo "✅ Binaire déployé"

# Cleanup local binary
rm scalping_live_bybit
```

### **Exécution**

```bash
cd /root/projects/trading_space/windsurf_space/harmonie_60_space/agent_economique_stable
./devops/scripts/deploy-binary.sh
```

---

## 📝 Étape 2 : Déployer Configuration

### **Upload config.yaml**

```bash
# Upload config
scp config/config.yaml root@31.57.224.79:/root/data/scalping-live-bybit/config/

# Verify
ssh root@31.57.224.79 "cat /root/data/scalping-live-bybit/config/config.yaml | head -20"
```

---

## 🎯 Étape 3 : Déployer Job Nomad

### **Script : `devops/scripts/deploy-nomad-job.sh`**

```bash
#!/bin/bash
set -e

export NOMAD_ADDR="http://10.8.0.1:4646"

echo "🚀 Déploiement Job Nomad Scalping Live Bybit..."

# Check if job exists
if nomad job status scalping-live-bybit &>/dev/null; then
    echo "⚠️  Job existant détecté, arrêt..."
    nomad job stop -purge scalping-live-bybit
    sleep 3
fi

# Deploy job
echo "📤 Déploiement job..."
nomad job run devops/configs/scalping-live-bybit.nomad

# Monitor deployment
echo "📊 Monitoring déploiement..."
nomad job status scalping-live-bybit

# Get allocation ID
ALLOC_ID=$(nomad job allocs scalping-live-bybit -json | jq -r '.[0].ID')

echo ""
echo "✅ Job déployé!"
echo "📋 Allocation ID: $ALLOC_ID"
echo ""
echo "🔍 Commandes utiles:"
echo "  nomad alloc logs -f $ALLOC_ID"
echo "  nomad alloc status $ALLOC_ID"
```

### **Exécution**

```bash
./devops/scripts/deploy-nomad-job.sh
```

---

## 🔄 Déploiement Complet (One-Shot)

### **Script : `devops/scripts/full-deploy.sh`**

```bash
#!/bin/bash
set -e

echo "🚀 DÉPLOIEMENT COMPLET SCALPING LIVE BYBIT"
echo "=========================================="

# 1. Compile & upload binary
echo ""
echo "📦 Étape 1/3: Compilation et upload binaire..."
./devops/scripts/deploy-binary.sh

# 2. Upload config
echo ""
echo "⚙️  Étape 2/3: Upload configuration..."
scp config/config.yaml root@31.57.224.79:/root/data/scalping-live-bybit/config/

# 3. Deploy Nomad job
echo ""
echo "🎯 Étape 3/3: Déploiement job Nomad..."
./devops/scripts/deploy-nomad-job.sh

echo ""
echo "✅ DÉPLOIEMENT TERMINÉ!"
```

### **Exécution**

```bash
./devops/scripts/full-deploy.sh
```

---

## 📊 Monitoring & Logs

### **Status Job**

```bash
export NOMAD_ADDR="http://10.8.0.1:4646"

# Status général
nomad job status scalping-live-bybit

# Détails allocation
nomad alloc status <ALLOC_ID>
```

### **Logs en Direct**

```bash
# Get allocation ID
ALLOC_ID=$(nomad job allocs scalping-live-bybit -json | jq -r '.[0].ID')

# Follow logs
nomad alloc logs -f $ALLOC_ID

# Stderr
nomad alloc logs -stderr -f $ALLOC_ID
```

### **Logs SSH Direct**

```bash
ssh root@31.57.224.79 "tail -f /root/data/scalping-live-bybit/logs/scalping.log"
```

---

## 🔧 Gestion Job

### **Arrêter Job**

```bash
nomad job stop scalping-live-bybit
```

### **Redémarrer Job**

```bash
nomad job stop scalping-live-bybit
sleep 3
nomad job run devops/configs/scalping-live-bybit.nomad
```

### **Purge Job**

```bash
nomad job stop -purge scalping-live-bybit
```

---

## 🆘 Troubleshooting

### **Job ne démarre pas**

```bash
# 1. Vérifier binaire existe
ssh root@31.57.224.79 "ls -lh /root/data/scalping-live-bybit/scalping_live_bybit"

# 2. Tester binaire manuellement
ssh root@31.57.224.79 "/root/data/scalping-live-bybit/scalping_live_bybit -config /root/data/scalping-live-bybit/config/config.yaml"

# 3. Vérifier logs allocation
nomad alloc logs <ALLOC_ID>
```

### **Erreur API Bybit**

```bash
# Test API depuis serveur
ssh root@31.57.224.79 'curl -s "https://api.bybit.com/v5/market/kline?category=linear&symbol=SOLUSDT&interval=5&limit=2"'

# Should return JSON with retCode: 0
```

### **Allocation Unhealthy**

```bash
# Vérifier health checks
nomad alloc status <ALLOC_ID>

# Regarder events
nomad alloc status <ALLOC_ID> | grep -A 10 "Recent Events"
```

---

## 📱 Notifications

### **Test Notification**

```bash
# Send test notification to verify ntfy
curl -d "Test notification from Scalping Live Bybit" \
  -H "Title: 🧪 Test Notification" \
  -H "Tags: test" \
  https://notifications.koyad.com/scalping-live-bybit
```

---

## 🔄 Workflow Mise à Jour

### **Update Code Only**

```bash
# 1. Recompile
GOOS=linux GOARCH=amd64 go build -o scalping_live_bybit ./cmd/scalping_live_bybit

# 2. Upload
scp scalping_live_bybit root@31.57.224.79:/root/data/scalping-live-bybit/

# 3. Restart job
nomad job stop scalping-live-bybit
sleep 2
nomad job run devops/configs/scalping-live-bybit.nomad

# 4. Cleanup
rm scalping_live_bybit
```

### **Update Config Only**

```bash
# 1. Upload new config
scp config/config.yaml root@31.57.224.79:/root/data/scalping-live-bybit/config/

# 2. Restart job
nomad job stop scalping-live-bybit
sleep 2
nomad job run devops/configs/scalping-live-bybit.nomad
```

---

## 📈 Accès UI Nomad

### **Via VPN**

```
URL: http://10.8.0.1:4646
```

**Jobs → scalping-live-bybit**
- Status
- Allocations
- Logs en direct
- Metrics

---

## 🔐 Backup & Recovery

### **Backup State**

```bash
# Backup trading state
ssh root@31.57.224.79 "tar -czf /tmp/scalping-backup-$(date +%Y%m%d).tar.gz /root/data/scalping-live-bybit/state/"

# Download backup
scp root@31.57.224.79:/tmp/scalping-backup-*.tar.gz ./backups/
```

### **Restore State**

```bash
# Upload backup
scp ./backups/scalping-backup-20251106.tar.gz root@31.57.224.79:/tmp/

# Extract
ssh root@31.57.224.79 "tar -xzf /tmp/scalping-backup-20251106.tar.gz -C /"
```

---

## 📚 Commandes Rapides

```bash
# Status rapide
nomad job status scalping-live-bybit

# Logs live
nomad alloc logs -f $(nomad job allocs scalping-live-bybit -json | jq -r '.[0].ID')

# Restart
nomad job stop scalping-live-bybit && sleep 2 && nomad job run devops/configs/scalping-live-bybit.nomad

# Check Bybit API
ssh root@31.57.224.79 'curl -s https://api.bybit.com/v5/market/kline?category=linear\&symbol=SOLUSDT\&interval=5\&limit=1 | grep retCode'
```

---

## 🎯 Checklist Déploiement

- [ ] VPN WireGuard actif (`wg show`)
- [ ] Nomad accessible (`nomad server members`)
- [ ] Binaire compilé et uploadé
- [ ] Config uploadée et vérifiée
- [ ] Job Nomad déployé
- [ ] Allocation healthy
- [ ] Logs montrent connexion Bybit OK
- [ ] Notifications fonctionnelles
- [ ] Monitoring actif

---

## 📝 Notes Production

1. **Timezone:** Serveur en UTC (vérifier avec `timedatectl`)
2. **Logs rotation:** Configurer logrotate pour `/root/data/scalping-live-bybit/logs/`
3. **Monitoring:** Surveiller allocation memory/CPU usage
4. **Alertes:** Configurer alertes ntfy pour erreurs critiques

---

## 🔄 Prochaines Étapes

- ✅ Infrastructure complète
- ✅ Déploiement automatisé
- ⏳ Monitoring avancé (Prometheus/Grafana)
- ⏳ Auto-scaling (si nécessaire)
