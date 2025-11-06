# ⚡ Quick Start - Déploiement Scalping Live Bybit

Guide rapide pour déployer l'application de trading sur serveur Singapour.

---

## 🎯 Objectif

Déployer `scalping_live_bybit` sur serveur Singapour (31.57.224.79) avec :
- ✅ Infrastructure sécurisée (VPN WireGuard)
- ✅ Orchestration Nomad
- ✅ Reverse proxy Caddy
- ✅ Accès Exchange Bybit (pas de geo-restriction)

---

## 📋 Prérequis

### **Serveur Production (31.57.224.79)**
- Ubuntu 24.04 LTS
- Root access
- 2GB RAM minimum

### **Machine Locale (Développement)**
- SSH access au serveur
- Go 1.21+ installé
- WireGuard installé
- jq installé

---

## 🚀 Installation (One-Time Setup)

### **Étape 1 : Sur Serveur Singapour**

```bash
# Connexion SSH
ssh root@31.57.224.79

# Clone repository
git clone <votre-repo> /root/agent_economique_stable
cd /root/agent_economique_stable/devops/scripts

# Installation complète (WireGuard + Nomad + Caddy)
./setup-server.sh
```

**Durée:** ~15-20 minutes  
**Note:** Le script va demander la clé publique du client VPN.

---

### **Étape 2 : Sur Machine Locale (Client VPN)**

```bash
# Navigate to project
cd /root/projects/.../agent_economique_stable/devops/scripts

# Setup WireGuard client
sudo ./setup-wireguard.sh client
```

**Note:** Utiliser la clé publique du serveur affichée à l'étape 1.

---

### **Étape 3 : Récupération Certificats TLS**

```bash
# Sur machine locale
cd /root/projects/.../agent_economique_stable/devops/scripts

# Récupérer certificats Nomad
./get-nomad-certs.sh
```

**Ce script télécharge les certificats et crée `~/.nomad-certs/nomad-env.sh`**

---

### **Étape 4 : Configuration et Test**

```bash
# Charger environnement Nomad (avec TLS)
source ~/.nomad-certs/nomad-env.sh

# Test ping via VPN
ping 10.8.0.1

# Test Nomad (avec TLS)
nomad server members

# Expected: 1 server alive
```

---

## 🎯 Déploiement Application

### **Option A : Déploiement Complet (Recommandé)**

```bash
# Sur machine locale (via VPN)
cd devops/scripts

# Déploiement complet
./full-deploy.sh
```

**Ce script fait :**
1. Compile le binaire pour Linux
2. Upload binaire sur serveur
3. Upload configuration
4. Déploie job Nomad
5. Affiche status et logs

---

### **Option B : Déploiement Manuel (Step-by-Step)**

```bash
# 1. Deploy binary
./deploy-binary.sh

# 2. Deploy Nomad job
./deploy-nomad-job.sh
```

---

## 📊 Gestion Application

### **Utilitaire de Gestion**

```bash
# Voir status
./manage-job.sh status

# Suivre logs
./manage-job.sh logs

# Voir erreurs
./manage-job.sh errors

# Redémarrer
./manage-job.sh restart

# Arrêter
./manage-job.sh stop

# Infos détaillées
./manage-job.sh info

# Ouvrir UI
./manage-job.sh ui
```

---

## 🔍 Vérifications

### **Infrastructure**

```bash
# Sur serveur (31.57.224.79)
ssh root@31.57.224.79

# VPN actif?
wg show
# Expected: interface wg0, peer connected

# Nomad actif?
systemctl status nomad
nomad server members
nomad node status

# Caddy actif?
systemctl status caddy
```

### **Application**

```bash
# Sur machine locale (via VPN)
export NOMAD_ADDR="http://10.8.0.1:4646"

# Job status
nomad job status scalping-live-bybit

# Allocation status
ALLOC_ID=$(nomad job allocs scalping-live-bybit -json | jq -r '.[0].ID')
nomad alloc status $ALLOC_ID

# Logs live
nomad alloc logs -f $ALLOC_ID
```

---

## 🌐 Accès

### **Via VPN**

```
Nomad UI:      http://10.8.0.1:4646
Nomad (Caddy): http://10.8.0.1:80
Health Check:  http://10.8.0.1:8080/health
```

### **Notifications**

```
Topic: scalping-live-bybit
URL:   https://notifications.koyad.com/scalping-live-bybit
```

---

## 🔄 Workflow Mise à Jour

### **Update Code**

```bash
# 1. Modify code locally
vim cmd/scalping_live_bybit/app_live.go

# 2. Redeploy
cd devops/scripts
./full-deploy.sh
```

### **Update Config Only**

```bash
# 1. Edit config
vim config/config.yaml

# 2. Upload
scp config/config.yaml root@31.57.224.79:/root/data/scalping-live-bybit/config/

# 3. Restart job
./manage-job.sh restart
```

---

## 🆘 Troubleshooting

### **VPN ne connecte pas**

```bash
# Check WireGuard
sudo wg show

# If no interface:
sudo wg-quick up wg0

# Check firewall
sudo ufw status
```

### **Nomad inaccessible**

```bash
# Check VPN first
ping 10.8.0.1

# Check NOMAD_ADDR
echo $NOMAD_ADDR

# Should be: http://10.8.0.1:4646
export NOMAD_ADDR="http://10.8.0.1:4646"
```

### **Job ne démarre pas**

```bash
# Check job status
nomad job status scalping-live-bybit

# Check allocation events
ALLOC_ID=$(nomad job allocs scalping-live-bybit -json | jq -r '.[0].ID')
nomad alloc status $ALLOC_ID | grep -A 10 "Recent Events"

# Check logs
nomad alloc logs $ALLOC_ID
nomad alloc logs -stderr $ALLOC_ID
```

### **Bybit API bloquée**

```bash
# Test depuis serveur
ssh root@31.57.224.79 'curl -s "https://api.bybit.com/v5/market/kline?category=linear&symbol=SOLUSDT&interval=5&limit=1"'

# Should return JSON with "retCode":0
# If 403: Geographic issue (should not happen from Singapore)
```

---

## 📝 Checklist Production

- [ ] VPN WireGuard configuré et actif
- [ ] Nomad Server opérationnel
- [ ] Caddy installé
- [ ] Connexion VPN testée (ping 10.8.0.1)
- [ ] Bybit API accessible depuis serveur
- [ ] Application déployée via Nomad
- [ ] Logs montrent démarrage OK
- [ ] Notifications fonctionnelles
- [ ] Monitoring actif

---

## 📚 Documentation Complète

- **README Principal:** `devops/README.md`
- **Guide Nomad:** `devops/docs/01-nomad-server-setup.md`
- **Guide VPN:** `devops/docs/02-wireguard-vpn.md`
- **Guide Déploiement:** `devops/docs/03-deployment-workflow.md`
- **Scripts:** `devops/scripts/README.md`

---

## ⚡ Commandes Rapides

```bash
# Status rapide
./manage-job.sh status

# Logs
./manage-job.sh logs

# Redémarrer
./manage-job.sh restart

# Redéployer complet
./full-deploy.sh

# Check infrastructure
ssh root@31.57.224.79 'wg show && systemctl status nomad'
```

---

## 🎯 Architecture Finale

```
┌─────────────────────────────────────────────┐
│  Machine Locale (Dev)                       │
│  └─ VPN: 10.8.0.2                          │
└──────────┬──────────────────────────────────┘
           │ WireGuard VPN
           │ (Encrypted)
┌──────────▼──────────────────────────────────┐
│  Serveur Singapour (31.57.224.79)          │
│  └─ VPN: 10.8.0.1                          │
│  └─ Nomad: http://10.8.0.1:4646           │
│  └─ Caddy: http://10.8.0.1:80             │
│  └─ App: /root/data/scalping-live-bybit   │
└──────────┬──────────────────────────────────┘
           │ HTTPS
┌──────────▼──────────────────────────────────┐
│  Bybit Exchange                             │
│  └─ api.bybit.com                          │
│  └─ USDT Perpetual (linear)               │
└─────────────────────────────────────────────┘
```

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-06  
**Status:** ✅ Production Ready
