# 🚀 Guide de Déploiement - Agent Economique Live Notifications

Guide complet pour déployer le système de notifications live Agent Economique sur serveur distant avec Nomad.

---

## 📋 Prérequis

### **Sur Machine Locale:**
- ✅ Go 1.22+ installé
- ✅ SSH configuré vers serveur distant
- ✅ Clés SSH sans mot de passe (ou agent SSH actif)
- ✅ Code source complet du projet

### **Sur Serveur Distant:**
- ✅ Nomad cluster actif (port 4646)
- ✅ Accès root ou sudo
- ✅ Connectivité internet (pour fetch BingX)

### **Certificats TLS:**
- ✅ Dossier `certs/` avec certificats Nomad
- ✅ Fichiers requis: `ca.pem`, `client.pem`, `client-key.pem`

---

## 🔧 Configuration Serveur

### **Serveur Cible:**
```
Host: 193.29.62.96
User: root
Base Dir: /root/data/backtest-optimizer/
```

### **Nomad Cluster:**
```
URL: http://193.29.62.96:4646/
Token: 1fc424de-5992-f4a5-c90e-cccabd7ef5d9
Datacenter: dc1
```

---

## 🚀 Procédure de Déploiement

### **Étape 1: Vérification Préalable**

```bash
# Depuis la racine du projet agent_economique_standalone/

# Vérifier structure
ls -la backend/cmd/agent_economique_live_notifications/main.go
ls -la deploy/deploy_binary.sh
ls -la deploy/deploy_nomad.sh
ls -la deploy/agent-economique-live.nomad
ls -la certs/ca.pem certs/client.pem certs/client-key.pem

# Vérifier connectivité SSH
ssh root@193.29.62.96 "echo 'SSH OK'"
```

### **Étape 2: Compilation et Upload du Binaire**

```bash
# Rendre le script exécutable
chmod +x deploy/deploy_binary.sh

# Lancer compilation et déploiement
./deploy/deploy_binary.sh
```

**Ce que fait le script:**
1. ✅ Vérifie Go installé
2. ✅ Compile `backend/cmd/agent_economique_live_notifications/main.go`
3. ✅ Test connectivité SSH
4. ✅ Crée arborescence distante (`out/`, `state/`, `configs/`, etc.)
5. ✅ Upload binaire via SCP
6. ✅ Configure permissions (chmod +x)
7. ✅ Test binaire distant
8. ✅ Nettoyage binaire local

**Sortie attendue:**
```
🔨 Compilation et Déploiement du Binaire Agent Economique
================================================
✅ Prérequis validés
✅ Binaire compilé: backend/agent_economique_live_notifications
✅ Connectivité SSH validée
✅ Arborescence distante préparée
✅ Binaire uploadé
✅ Permissions configurées
✅ Binaire fonctionnel sur le serveur distant
✅ Déploiement du binaire terminé!
```

### **Étape 3: Déploiement Job Nomad**

```bash
# Rendre le script exécutable
chmod +x deploy/deploy_nomad.sh

# Lancer déploiement Nomad
./deploy/deploy_nomad.sh
```

**Ce que fait le script:**
1. ✅ Vérifie fichier job Nomad existe
2. ✅ Vérifie certificats TLS présents
3. ✅ Installe Nomad CLI si absent
4. ✅ Stop job existant si présent
5. ✅ Deploy nouveau job avec TLS
6. ✅ Affiche statut + allocations
7. ✅ Affiche logs récents

**Sortie attendue:**
```
🚀 Déploiement Job Nomad Agent Economique
=============================================
✅ Certificats TLS trouvés
✅ Prérequis validés
✅ Job déployé avec succès
✅ Déploiement Nomad terminé!

📱 Pour recevoir les notifications:
   1. Installer l'app ntfy
   2. S'abonner à: notifications.koyad.com/notification-agent-eco
```

---

## 📊 Vérification du Déploiement

### **Vérifier Job Actif:**
```bash
nomad job status \
  -address http://193.29.62.96:4646/ \
  -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 \
  agent-economique-live
```

### **Voir les Logs en Temps Réel:**
```bash
# Récupérer l'allocation ID
ALLOC_ID=$(nomad job allocs \
  -address http://193.29.62.96:4646/ \
  -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 \
  agent-economique-live -json | jq -r '.[0].ID')

# Suivre les logs
nomad alloc logs \
  -address http://193.29.62.96:4646/ \
  -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 \
  -f $ALLOC_ID
```

### **État du Système:**
```bash
# Sur le serveur distant
ssh root@193.29.62.96

# Vérifier binaire
ls -lh /root/data/backtest-optimizer/agent_economique_live_notifications

# Vérifier état sauvegardé
cat /root/data/backtest-optimizer/state/agent_eco_state.json

# Vérifier logs système
tail -f /root/data/backtest-optimizer/logs/*.log
```

---

## 📱 Configuration Notifications

### **Installer ntfy (Mobile/Desktop):**

**Android:**
1. Google Play Store → Installer "ntfy"
2. Ouvrir app
3. Cliquer "+" → "Subscribe to topic"
4. Serveur: `notifications.koyad.com`
5. Topic: `notification-agent-eco`

**iOS:**
1. App Store → Installer "ntfy"
2. Même procédure qu'Android

**Web:**
1. Navigateur: `https://notifications.koyad.com`
2. S'abonner à `notification-agent-eco`

### **Test Notification:**
```bash
# Envoyer notification test
curl -d "Test Agent Economique" https://notifications.koyad.com/notification-agent-eco
```

---

## 🔧 Modification de la Configuration

### **Changer les Paramètres:**

Éditer `deploy/agent-economique-live.nomad`:
```hcl
template {
  data = <<EOH
{
  "symbol": "BTCUSDT",           # Changer paire
  "exec_tf": "5m",               # Changer timeframe
  "params": {
    "cci_period": 120,           # Modifier params
    "trailing_stop_pct": 3.0,    # Modifier stop
    ...
  }
}
EOH
  ...
}
```

### **Redéployer:**
```bash
# Pas besoin de recompiler le binaire
./deploy/deploy_nomad.sh
```

---

## 🛑 Arrêter le Bot

```bash
nomad job stop \
  -address http://193.29.62.96:4646/ \
  -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 \
  agent-economique-live
```

---

## 🐛 Dépannage

### **Problème: Compilation échoue**
```bash
# Vérifier Go installé
go version

# Vérifier dépendances
cd backend
go mod tidy
go mod download
```

### **Problème: SSH échoue**
```bash
# Tester connexion manuelle
ssh root@193.29.62.96

# Vérifier clés SSH
ssh-add -l

# Ajouter clé si nécessaire
ssh-add ~/.ssh/id_rsa
```

### **Problème: Certificats manquants**
```bash
# Copier depuis backtest_optimizer
cp -r ../backtest_optimizer/certs ./

# Vérifier présence
ls -la certs/ca.pem certs/client.pem certs/client-key.pem
```

### **Problème: Job ne démarre pas**
```bash
# Voir raison exacte
nomad job status -address http://193.29.62.96:4646/ \
  -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 \
  agent-economique-live

# Voir logs allocation
nomad alloc logs -address http://193.29.62.96:4646/ \
  -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 \
  $ALLOC_ID
```

### **Problème: Pas de données BingX**
```bash
# Vérifier connectivité internet sur serveur
ssh root@193.29.62.96 "curl -I https://open-api.bingx.com"

# Tester fetch manuel
ssh root@193.29.62.96
/root/data/backtest-optimizer/agent_economique_live_notifications \
  -config /tmp/test_config.json
```

### **Problème: Notifications non reçues**
```bash
# Vérifier serveur ntfy accessible
curl https://notifications.koyad.com

# Test notification directe
curl -d "Test" https://notifications.koyad.com/notification-agent-eco

# Vérifier topic dans l'app
```

---

## 📈 Monitoring Production

### **Dashboard Nomad:**
```
URL: http://193.29.62.96:4646/ui/jobs/agent-economique-live
```

### **Métriques à Surveiller:**
- ✅ Job status: running
- ✅ Allocations: 1/1 healthy
- ✅ CPU usage: < 80%
- ✅ Memory usage: < 200MB
- ✅ Restarts: < 3/jour

### **Logs Importants:**
```bash
# Signaux détectés
grep "POSITION OPENED" logs.txt

# Paliers atteints
grep "PALIER ATTEINT" logs.txt

# Stops touchés
grep "STOP HIT" logs.txt

# Erreurs
grep "ERROR\|Failed" logs.txt
```

---

## 🔄 Workflow Complet

```
1. Développement Local
   └─> Test backtest avec même config
   
2. Compilation
   └─> ./deploy/deploy_binary.sh
   
3. Déploiement Nomad
   └─> ./deploy/deploy_nomad.sh
   
4. Vérification
   └─> Logs + Statut + Notifications test
   
5. Monitoring
   └─> App ntfy + Dashboard Nomad
   
6. Ajustements (si nécessaire)
   └─> Modifier job Nomad → Redeploy
```

---

## ⚙️ Configuration Avancée

### **Multi-Paires:**
Créer plusieurs jobs Nomad (un par paire):
```bash
cp deploy/agent-economique-live.nomad deploy/agent-economique-btc.nomad
# Modifier symbol: "BTCUSDT"
# Modifier job name: "agent-economique-btc"
# Modifier topic: "notification-agent-eco-btc"
```

### **Timeframes Multiples:**
Un job par timeframe:
```bash
# Job 5m
job "agent-economique-5m" { exec_tf: "5m" }

# Job 15m  
job "agent-economique-15m" { exec_tf: "15m" }

# Job 1h
job "agent-economique-1h" { exec_tf: "1h" }
```

---

## 📚 Ressources

- **Code Source:** `backend/cmd/agent_economique_live_notifications/`
- **Documentation:** `backend/cmd/agent_economique_live_notifications/README.md`
- **Job Nomad:** `deploy/agent-economique-live.nomad`
- **Certificats:** `certs/`

---

**💡 Bon Trading ! 🚀**
