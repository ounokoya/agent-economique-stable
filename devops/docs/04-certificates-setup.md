# 🔒 Setup Certificats TLS Nomad

Guide complet pour générer et utiliser les certificats TLS Nomad.

---

## 📍 Où sont les Certificats ?

### **Sur le Serveur (31.57.224.79)**

```
/etc/nomad.d/certs/           ← Certificats de production
├── ca.pem                    ← Certificate Authority
├── ca-key.pem                ← CA private key
├── server.pem                ← Server certificate
├── server-key.pem            ← Server private key
├── client.pem                ← Client certificate
├── client-key.pem            ← Client private key
├── cli.pem                   ← CLI certificate
└── cli-key.pem               ← CLI private key

/tmp/nomad-client-certs/      ← Package pour machines locales
├── ca.pem
├── cli.pem
└── cli-key.pem
```

### **Sur Machine Locale (Après récupération)**

```
~/.nomad-certs/               ← Certificats client
├── ca.pem
├── cli.pem
├── cli-key.pem
└── nomad-env.sh              ← Script environment variables
```

---

## 🔧 Génération Certificats (Serveur)

### **Automatique via setup-server.sh**

Le script principal génère automatiquement les certificats :

```bash
ssh root@31.57.224.79
cd /root/agent_economique_stable/devops/scripts
./setup-server.sh
```

### **Manuel (si besoin)**

```bash
ssh root@31.57.224.79
cd /root/agent_economique_stable/devops/scripts
./generate-nomad-certs.sh
```

---

## 📥 Récupération Certificats (Local)

### **Option 1 : Script Automatique** ✅ RECOMMANDÉ

```bash
# Sur machine locale
cd /root/projects/.../agent_economique_stable/devops/scripts

# Récupérer certificats
./get-nomad-certs.sh
```

**Ce script fait :**
1. Télécharge certificats depuis serveur
2. Les place dans `~/.nomad-certs/`
3. Configure permissions (600)
4. Crée script `nomad-env.sh`

### **Option 2 : Copie Manuelle**

```bash
# Télécharger certificats
scp -r root@31.57.224.79:/tmp/nomad-client-certs ~/.nomad-certs

# Configurer permissions
chmod 600 ~/.nomad-certs/*.pem
```

---

## ⚙️ Configuration Nomad CLI

### **Méthode 1 : Source Script** ✅ RECOMMANDÉ

```bash
# Charger environnement (temporaire)
source ~/.nomad-certs/nomad-env.sh

# Permanent (ajouter à ~/.bashrc)
echo 'source ~/.nomad-certs/nomad-env.sh' >> ~/.bashrc
source ~/.bashrc
```

### **Méthode 2 : Export Manuel**

```bash
export NOMAD_ADDR="https://10.8.0.1:4646"
export NOMAD_CACERT="$HOME/.nomad-certs/ca.pem"
export NOMAD_CLIENT_CERT="$HOME/.nomad-certs/cli.pem"
export NOMAD_CLIENT_KEY="$HOME/.nomad-certs/cli-key.pem"
```

### **Méthode 3 : Flags Nomad CLI**

```bash
nomad server members \
  -address=https://10.8.0.1:4646 \
  -ca-cert=$HOME/.nomad-certs/ca.pem \
  -client-cert=$HOME/.nomad-certs/cli.pem \
  -client-key=$HOME/.nomad-certs/cli-key.pem
```

---

## ✅ Vérification

### **1. Vérifier VPN Actif**

```bash
# Ping serveur via VPN
ping 10.8.0.1

# Voir interface WireGuard
wg show
```

### **2. Vérifier Variables d'Environnement**

```bash
# Charger environnement si pas déjà fait
source ~/.nomad-certs/nomad-env.sh

# Vérifier variables
echo $NOMAD_ADDR
echo $NOMAD_CACERT
echo $NOMAD_CLIENT_CERT
echo $NOMAD_CLIENT_KEY
```

### **3. Tester Connexion Nomad**

```bash
# Status serveurs
nomad server members

# Expected:
# Name                   Address     Port  Status  Leader  Raft Version
# server.global.nomad    10.8.0.1    4648  alive   true    3

# Status nodes
nomad node status

# Expected: 1 node ready
```

### **4. Test Complet**

```bash
# Job status
nomad job status

# Si aucun job:
# No running jobs

# Si jobs existent:
# ID                      Type     Status
# scalping-live-bybit     service  running
```

---

## 🔄 Renouvellement Certificats

Les certificats sont valides **10 ans** (généré avec expiry: 87600h).

### **Vérifier Date Expiration**

```bash
# Sur serveur
openssl x509 -in /etc/nomad.d/certs/ca.pem -noout -dates
openssl x509 -in /etc/nomad.d/certs/server.pem -noout -dates
openssl x509 -in /etc/nomad.d/certs/cli.pem -noout -dates
```

### **Regénérer si Nécessaire**

```bash
# Sur serveur
cd /root/agent_economique_stable/devops/scripts

# Backup old certs
mv /etc/nomad.d/certs /etc/nomad.d/certs.backup.$(date +%Y%m%d)

# Generate new
./generate-nomad-certs.sh

# Restart Nomad
systemctl restart nomad

# Sur machine locale - récupérer nouveaux certs
./get-nomad-certs.sh
```

---

## 🆘 Troubleshooting

### **Erreur : "certificate verify failed"**

```bash
# Vérifier NOMAD_CACERT pointe vers bon fichier
ls -l $NOMAD_CACERT

# Si erreur, recharger environnement
source ~/.nomad-certs/nomad-env.sh
```

### **Erreur : "connection refused"**

```bash
# Vérifier VPN actif
ping 10.8.0.1

# Si pas de réponse, activer VPN
sudo wg-quick up wg0
```

### **Erreur : "TLS handshake timeout"**

```bash
# Vérifier Nomad écoute bien en HTTPS
ssh root@31.57.224.79 'netstat -tlnp | grep 4646'

# Expected: 0.0.0.0:4646 (listening)
```

### **Permissions Incorrectes**

```bash
# Fixer permissions certificats
chmod 600 ~/.nomad-certs/*.pem
chmod 700 ~/.nomad-certs
```

---

## 🔐 Sécurité

### **Bonnes Pratiques**

1. ✅ **Certificats privés** : Ne jamais commit dans git
2. ✅ **Permissions 600** : Seulement propriétaire peut lire
3. ✅ **Backup sécurisé** : Stocker copie chiffrée
4. ✅ **Rotation régulière** : Même si validité 10 ans

### **Fichiers à Protéger**

```bash
# Ajouter au .gitignore
echo "**/*.pem" >> .gitignore
echo "**/*-key.pem" >> .gitignore
echo ".nomad-certs/" >> .gitignore
```

### **Backup Certificats**

```bash
# Sur serveur
tar -czf nomad-certs-backup-$(date +%Y%m%d).tar.gz \
  /etc/nomad.d/certs/

# Télécharger backup
scp root@31.57.224.79:~/nomad-certs-backup-*.tar.gz \
  ~/backups/
```

---

## 📊 Résumé Workflow

```
1. Serveur: ./setup-server.sh
   └─> Génère certificats dans /etc/nomad.d/certs/
   └─> Crée package dans /tmp/nomad-client-certs/

2. Local: ./get-nomad-certs.sh
   └─> Télécharge package
   └─> Place dans ~/.nomad-certs/
   └─> Crée nomad-env.sh

3. Local: source ~/.nomad-certs/nomad-env.sh
   └─> Configure NOMAD_ADDR, NOMAD_CACERT, etc.

4. Local: nomad server members
   └─> Test connexion TLS
```

---

## 🎯 Commandes Rapides

```bash
# Récupérer certificats
./get-nomad-certs.sh

# Configurer environnement
source ~/.nomad-certs/nomad-env.sh

# Tester connexion
nomad server members

# Voir variables
env | grep NOMAD

# Redémarrer Nomad (serveur)
ssh root@31.57.224.79 'systemctl restart nomad'
```

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-06  
**Serveur:** 31.57.224.79 (Singapore)
