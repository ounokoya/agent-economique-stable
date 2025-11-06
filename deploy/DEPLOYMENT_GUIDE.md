# 🚀 Guide de Déploiement - Agent Economique

**Date:** 21 Octobre 2025  
**Serveur:** 193.29.62.96  
**Stack:** Nomad + ArangoDB + Agent Live

---

## 📋 Prérequis

### **Sur votre machine locale**
- Accès SSH au serveur 193.29.62.96
- Clés SSH configurées
- Certificats TLS Nomad dans `certs/`

### **Sur le serveur distant**
- Nomad installé et configuré
- Docker installé
- Ports ouverts: 8529 (ArangoDB), 4646 (Nomad)

---

## 🗂️ Architecture Déployée

```
Serveur 193.29.62.96
├── ArangoDB (Port 8529)
│   ├── Database: agent_economique
│   └── Collections: notification_*, paper_*, live_*
│
└── Agent Economique Live (Nomad Job)
    ├── Mode: Notification
    ├── Timeframe: 15m
    ├── Symbol: SUIUSDT
    └── Persistence: ArangoDB
```

---

## 🚀 Déploiement Complet (Ordre Correct)

### **Étape 1: Déployer ArangoDB** ⚡ OBLIGATOIRE EN PREMIER

```bash
# Depuis la racine du projet
cd /root/projects/trading_space/windsurf_space/harmonie_60_space/agent_economique_standalone

# Déployer ArangoDB
./deploy/deploy_arangodb.sh
```

**Ce que fait ce script:**
1. ✅ Vérifie connectivité SSH
2. ✅ Crée `/opt/arangodb_data` sur le serveur
3. ✅ Upload du job Nomad ArangoDB
4. ✅ Déploie ArangoDB sur Nomad
5. ✅ Vérifie que la DB est accessible

**Temps estimé:** 2-3 minutes

**Vérification:**
```bash
# Tester l'accès à ArangoDB
curl http://10.0.0.1:8529/_api/version

# Accéder à l'interface web
# URL: http://10.0.0.1:8529
# User: root
# Pass: agent_economique_2025
```

⚠️ **IMPORTANT:** Ne passez pas à l'étape suivante tant qu'ArangoDB n'est pas accessible.

---

### **Étape 2: Compiler le Binaire**

```bash
# Depuis le dossier backend
cd backend

# Compiler avec support ArangoDB
go mod tidy
go build -o agent_economique_live_notifications cmd/agent_economique_live_notifications/main.go

# Vérifier que le binaire existe
ls -lh agent_economique_live_notifications
```

---

### **Étape 3: Déployer le Binaire**

```bash
# Depuis la racine du projet
./deploy/deploy_binary.sh
```

**Ce que fait ce script:**
1. ✅ Compile le binaire (si pas fait)
2. ✅ Upload vers le serveur
3. ✅ Configure les permissions
4. ✅ Prépare l'arborescence

---

### **Étape 4: Déployer sur Nomad**

```bash
# Déployer le job Nomad
./deploy/deploy_nomad.sh
```

**Ce que fait ce script:**
1. ✅ Vérifie les certificats TLS
2. ✅ Arrête le job existant si présent
3. ✅ Déploie le nouveau job
4. ✅ Affiche les logs

**Configuration automatique:**
- Symbol: SUIUSDT
- Timeframe: 15m
- Database: http://10.0.0.1:8529
- Collections: notification_*
- Notifications: notifications.koyad.com

---

## 🔍 Vérification du Déploiement

### **1. Vérifier ArangoDB**

```bash
# SSH sur le serveur
ssh root@193.29.62.96

# Status du job
nomad job status arangodb-agent-economique

# Logs
nomad alloc logs $(nomad job allocs arangodb-agent-economique -json | jq -r '.[0].ID')

# Test connexion
curl http://localhost:8529/_api/version
```

### **2. Vérifier l'Agent Live**

```bash
# Status
nomad job status agent-economique-live

# Logs en temps réel
nomad alloc logs -f $(nomad job allocs agent-economique-live -json | jq -r '.[0].ID')

# Vérifier les notifications
# → Installer ntfy et s'abonner à notifications.koyad.com/notification-agent-eco
```

### **3. Vérifier les Données en DB**

**Via Interface Web:**
1. Aller sur http://10.0.0.1:8529
2. Login: root / agent_economique_2025
3. Database: agent_economique
4. Voir les collections: notification_trades, etc.

**Via AQL:**
```aql
FOR trade IN notification_trades
  SORT trade.entry_time DESC
  LIMIT 10
  RETURN trade
```

---

## 🔄 Commandes Utiles

### **Gestion ArangoDB**

```bash
# Restart
nomad job restart arangodb-agent-economique

# Stop
nomad job stop arangodb-agent-economique

# Logs
nomad alloc logs <alloc-id>

# Status
nomad job status arangodb-agent-economique
```

### **Gestion Agent Live**

```bash
# Restart
nomad job restart agent-economique-live

# Stop
nomad job stop agent-economique-live

# Logs temps réel
nomad alloc logs -f <alloc-id>

# Modifier config et redéployer
vim deploy/agent-economique-live.nomad
./deploy/deploy_nomad.sh
```

### **Backup Base de Données**

```bash
# Sur le serveur
ssh root@193.29.62.96

# Backup
docker exec <container-id> arangodump \
  --server.endpoint tcp://127.0.0.1:8529 \
  --server.username root \
  --server.password agent_economique_2025 \
  --output-directory /tmp/backup

# Récupérer le backup
docker cp <container-id>:/tmp/backup ./backup_$(date +%Y%m%d)
```

---

## 🐛 Troubleshooting

### **ArangoDB ne démarre pas**

```bash
# Vérifier les logs
nomad alloc logs <alloc-id>

# Vérifier le volume
ssh root@193.29.62.96 "ls -la /opt/arangodb_data"

# Vérifier la config Nomad
ssh root@193.29.62.96 "cat /etc/nomad.d/nomad.hcl | grep -A 5 arangodb_data"
```

**Solution:** Configurer le host volume dans `/etc/nomad.d/nomad.hcl`:
```hcl
client {
  host_volume "arangodb_data" {
    path      = "/opt/arangodb_data"
    read_only = false
  }
}
```

Puis: `sudo systemctl restart nomad`

---

### **Agent ne se connecte pas à la DB**

```bash
# Tester la connexion depuis le serveur
ssh root@193.29.62.96
curl http://localhost:8529/_api/version

# Vérifier les logs de l'agent
nomad alloc logs <alloc-id> | grep -i "database\|arango\|connection"
```

**Solutions:**
- Vérifier que ArangoDB est bien démarré
- Vérifier l'URL dans le job Nomad (doit être 10.0.0.1:8529)
- Vérifier le mot de passe

---

### **Pas de notifications**

```bash
# Vérifier que l'agent tourne
nomad job status agent-economique-live

# Vérifier les logs
nomad alloc logs -f <alloc-id>

# Tester ntfy manuellement
curl -d "Test" https://notifications.koyad.com/notification-agent-eco
```

---

## 📊 Monitoring

### **Dashboard Nomad**

URL: http://193.29.62.96:4646

Vérifier:
- Jobs running
- Allocations healthy
- Resources usage

### **Interface ArangoDB**

URL: http://10.0.0.1:8529

Vérifier:
- Collections créées
- Nombre de documents
- Taille de la DB

### **Logs en Temps Réel**

```bash
# Agent
nomad alloc logs -f $(nomad job allocs agent-economique-live -json | jq -r '.[0].ID')

# ArangoDB
nomad alloc logs -f $(nomad job allocs arangodb-agent-economique -json | jq -r '.[0].ID')
```

---

## 🎯 Prochaines Étapes

### **1. Valider Mode Notification (en cours)**
- ✅ ArangoDB déployé
- ✅ Agent live déployé
- 🔄 Attendre signaux et vérifier données en DB
- 🔄 Valider que les trades sont correctement enregistrés

### **2. Développer Mode Paper Trading**
- [ ] Créer `cmd/agent_economique_paper_trading/main.go`
- [ ] Simuler exécution avec slippage/fees
- [ ] Enregistrer dans collections paper_*
- [ ] Déployer sur Nomad

### **3. Développer Mode Live Trading**
- [ ] Créer `cmd/agent_economique_live_trading/main.go`
- [ ] Intégrer API BingX
- [ ] Gestion orders réels
- [ ] Enregistrer dans collections live_*
- [ ] Déployer sur Nomad

### **4. Dashboard Analytics**
- [ ] Backend API Go
- [ ] Frontend React
- [ ] Graphiques comparatifs
- [ ] Export données

---

## 📝 Checklist de Déploiement

**Avant de déployer:**
- [ ] Certificats TLS présents dans `certs/`
- [ ] Accès SSH au serveur fonctionnel
- [ ] Binaire compilé

**Déploiement:**
- [ ] ArangoDB déployé (`deploy_arangodb.sh`)
- [ ] DB accessible sur port 8529
- [ ] Binaire uploadé (`deploy_binary.sh`)
- [ ] Job Nomad déployé (`deploy_nomad.sh`)

**Validation:**
- [ ] Job Nomad running
- [ ] Logs de l'agent sans erreurs
- [ ] Première notification reçue
- [ ] Premier trade enregistré en DB

---

**Support:** Voir les logs en cas de problème  
**Documentation:** `/docs/` pour architecture complète
