# 🚀 Guide de Déploiement - Scalping Live

Guide complet pour déployer le système de trading Scalping Live sur serveur distant avec Nomad.

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
- ✅ Connectivité internet (pour Binance API)

### **Certificats TLS (Optionnel):**
- ✅ Dossier `certs/` avec certificats Nomad
- ✅ Fichiers: `ca.pem`, `client.pem`, `client-key.pem`

---

## 🔧 Configuration Serveur

### **Serveur Cible:**
```
Host: 193.29.62.96
User: root
Base Dir: /root/data/scalping-live/
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
# Depuis la racine du projet

# Vérifier structure
ls -la cmd/scalping_live/main.go
ls -la cmd/scalping_live/app_live.go
ls -la deploy/deploy_scalping_live.sh
ls -la deploy/deploy_scalping_live_nomad.sh
ls -la deploy/scalping-live.nomad
ls -la config/config.yaml

# Vérifier connectivité SSH
ssh root@193.29.62.96 "echo 'SSH OK'"
```

### **Étape 2: Compilation et Upload du Binaire**

```bash
# Rendre le script exécutable
chmod +x deploy/deploy_scalping_live.sh

# Lancer compilation et déploiement
./deploy/deploy_scalping_live.sh
```

**Ce que fait le script:**
1. ✅ Vérifie Go installé
2. ✅ Compile `cmd/scalping_live/main.go`
3. ✅ Test connectivité SSH
4. ✅ Crée arborescence distante (`config/`, `logs/`, `state/`, `data/`)
5. ✅ Upload binaire via SCP
6. ✅ Upload config.yaml
7. ✅ Configure permissions (chmod +x)
8. ✅ Test binaire distant
9. ✅ Nettoyage binaire local

**Sortie attendue:**
```
🔨 Compilation et Déploiement Scalping Live
================================================
• Binaire: scalping_live
• Serveur: root@193.29.62.96
• Destination: /root/data/scalping-live/
================================================
✅ Prérequis validés
✅ Binaire compilé
✅ Connectivité SSH validée
✅ Arborescence distante préparée
✅ Binaire uploadé
✅ Configuration uploadée
✅ Permissions configurées
✅ Binaire fonctionnel sur le serveur distant
✅ Déploiement du binaire terminé!
```

### **Étape 3: Déploiement Job Nomad**

```bash
# Rendre le script exécutable
chmod +x deploy/deploy_scalping_live_nomad.sh

# Lancer déploiement Nomad
./deploy/deploy_scalping_live_nomad.sh
```

**Ce que fait le script:**
1. ✅ Vérifie Nomad CLI installé
2. ✅ Vérifie fichier job `scalping-live.nomad`
3. ✅ Vérifie certificats TLS (optionnel)
4. ✅ Arrête job existant si présent
5. ✅ Déploie nouveau job Nomad
6. ✅ Vérifie statut et allocations
7. ✅ Affiche logs récents

**Sortie attendue:**
```
🚀 Déploiement Job Nomad Scalping Live
=============================================
• Job: scalping-live
• Fichier: deploy/scalping-live.nomad
• Cluster: http://193.29.62.96:4646/
=============================================
✅ Prérequis validés
✅ Job déployé avec succès
✅ Déploiement Nomad terminé!
```

---

## 📊 Monitoring et Gestion

### **Vérifier Statut Job**

```bash
# Via script local
nomad job status -address http://193.29.62.96:4646/ -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 scalping-live

# Sur serveur distant
ssh root@193.29.62.96
nomad job status scalping-live
```

### **Voir Logs en Temps Réel**

```bash
# 1. Récupérer l'allocation ID
ALLOC_ID=$(nomad job allocs -address http://193.29.62.96:4646/ -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 scalping-live -json | jq -r '.[0].ID')

# 2. Suivre les logs
nomad alloc logs -address http://193.29.62.96:4646/ -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 -f $ALLOC_ID
```

### **Arrêter le Job**

```bash
nomad job stop -address http://193.29.62.96:4646/ -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 scalping-live
```

### **Redémarrer le Job**

```bash
nomad job restart -address http://193.29.62.96:4646/ -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 scalping-live
```

---

## 📱 Notifications

### **Configuration Ntfy**

Les notifications sont envoyées sur le topic **`scalping-live`**.

**Pour s'abonner:**
1. Installer l'app ntfy (iOS/Android)
2. Ajouter un topic
3. **Serveur:** `https://notifications.koyad.com`
4. **Topic:** `scalping-live`

### **Types de Notifications**

| Type | Quand | Exemple |
|------|-------|---------|
| 🚀 **Démarrage** | Au lancement | "🚀 Démarrage Scalping live<br>📊 Symbole: SOLUSDT<br>⏱️ Timeframe: 5m" |
| 🎯 **Signal LONG** | Triple extrême + croisement haussier | "🎯 Signal LONG détecté<br>💰 Prix: 185.43 SOLUSDT<br>📈 CCI: -105.2, MFI: 18.3" |
| 🎯 **Signal SHORT** | Triple extrême + croisement baissier | "🎯 Signal SHORT détecté<br>💰 Prix: 187.12 SOLUSDT<br>📉 CCI: 112.5, MFI: 72.8" |
| ⚠️ **Erreur** | En cas de problème | "⚠️ Erreur Scalping Engine<br>Binance API timeout" |
| 🛑 **Arrêt** | À la fermeture | "🛑 Arrêt Scalping live<br>📊 Signaux détectés: 3" |

---

## 🔧 Modification de la Configuration

### **Changer Symbole ou Timeframe**

```bash
# Éditer le job Nomad
vim deploy/scalping-live.nomad

# Modifier les lignes du template (ligne 47-48)
  symbols: 
    - "ETHUSDT"  # Au lieu de SOLUSDT
  timeframes:
    - "15m"      # Au lieu de 5m

# Redéployer
./deploy/deploy_scalping_live_nomad.sh
```

### **Ajuster Seuils Indicateurs**

```bash
# Dans deploy/scalping-live.nomad, section strategy.scalping (lignes 80-85)
    cci_surachat: 150.0     # Au lieu de 100.0 (moins de signaux)
    cci_survente: -150.0    # Au lieu de -100.0
    mfi_surachat: 70.0      # Au lieu de 60.0
    mfi_survente: 30.0      # Au lieu de 40.0
```

---

## 🐛 Dépannage

### **Le binaire ne démarre pas**

```bash
# Se connecter au serveur
ssh root@193.29.62.96

# Tester manuellement
cd /root/data/scalping-live
./scalping_live -config config/config.yaml

# Vérifier les logs
cat logs/*.log
```

### **Pas de notifications reçues**

1. ✅ Vérifier abonnement au topic `scalping-live`
2. ✅ Vérifier logs du job : `nomad alloc logs ...`
3. ✅ Tester manuellement : `curl -d "Test" https://notifications.koyad.com/scalping-live`

### **Job Nomad en échec**

```bash
# Voir raison de l'échec
nomad job status -address http://193.29.62.96:4646/ -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 scalping-live

# Voir logs allocation
nomad alloc logs -address http://193.29.62.96:4646/ -token 1fc424de-5992-f4a5-c90e-cccabd7ef5d9 <ALLOC_ID>
```

---

## 📚 Fichiers Importants

| Fichier | Description |
|---------|-------------|
| `deploy/deploy_scalping_live.sh` | Script compilation + upload binaire |
| `deploy/deploy_scalping_live_nomad.sh` | Script déploiement job Nomad |
| `deploy/scalping-live.nomad` | Définition job Nomad (config incluse) |
| `cmd/scalping_live/main.go` | Code source principal |
| `cmd/scalping_live/app_live.go` | Logique application |
| `config/config.yaml` | Configuration par défaut |

---

## ✅ Checklist Déploiement

- [ ] Code compilé sans erreur
- [ ] SSH configuré vers serveur
- [ ] Binaire uploadé sur serveur
- [ ] Configuration uploadée
- [ ] Job Nomad déployé
- [ ] Job en statut "running"
- [ ] Abonné au topic ntfy `scalping-live`
- [ ] Notification de démarrage reçue
- [ ] Logs consultés et normaux

---

## 🚀 Prochaines Étapes

1. **Surveiller premiers signaux** : Observer les détections sur quelques heures
2. **Ajuster seuils** : Si trop/pas assez de signaux, modifier les seuils
3. **Backtester** : Valider les paramètres sur historique avant production
4. **Money Management** : Implémenter la gestion de positions réelles
5. **Multi-symboles** : Déployer sur d'autres paires (ETH, BTC, etc.)

---

## 📞 Support

En cas de problème, vérifier :
- Logs Nomad
- Logs serveur `/root/data/scalping-live/logs/`
- Connectivité réseau
- API Binance disponible
