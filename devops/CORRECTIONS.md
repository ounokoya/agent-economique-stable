# 🔧 Corrections Infrastructure DevOps (Nov 2025)

Ce document liste les corrections critiques apportées à l'infrastructure DevOps suite au déploiement sur le nouveau serveur Singapore.

---

## ✅ Corrections Appliquées

### 1. **Certificats TLS avec SANs**

**Problème :** Nomad rejetait les certificats avec erreur `x509: certificate relies on legacy Common Name field`

**Solution :** Utilisation du flag `-hostname` dans `cfssl` pour générer les SANs

**Fichier modifié :** `scripts/generate-nomad-certs.sh`

```bash
# Avant (INCORRECT)
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=server-config.json server-csr.json

# Après (CORRECT)
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=server-config.json \
  -hostname="server.global.nomad,localhost,127.0.0.1,$SERVER_IP,$SERVER_VPN_IP" \
  server-csr.json
```

**Impact :** Tous les certificats (server, client, cli) incluent maintenant les SANs requis.

---

### 2. **Driver Nomad : raw_exec → exec**

**Problème :** `raw_exec` ne supporte pas les host volumes

```
Error: volumes: task driver "raw_exec" does not support host volumes
```

**Solution :** Utiliser le driver `exec` qui supporte les volumes

**Fichiers modifiés :**
- `configs/scalping-live-bybit.nomad`
- `configs/nomad-server.hcl`

```hcl
# Avant
task "scalping-live-bybit-5m" {
  driver = "raw_exec"
  volume_mount { ... }  # ❌ Échoue
}

# Après
task "scalping-live-bybit-5m" {
  driver = "exec"
  volume_mount { ... }  # ✅ Fonctionne
}
```

**Configuration Nomad :**
```hcl
client {
  options = {
    "driver.exec.enable"     = "1"  # ✅ Ajouté
    "driver.raw_exec.enable" = "1"
  }
}
```

---

### 3. **Datacenter : sg1 → dc1**

**Problème :** Incohérence entre config serveur et job

**Solution :** Alignement sur `dc1` (standard)

**Fichiers modifiés :**
- `configs/nomad-server.hcl`
- `configs/scalping-live-bybit.nomad`

```hcl
# Avant
datacenter = "sg1"  # configs/nomad-server.hcl
datacenters = ["sg1"]  # job

# Après
datacenter = "dc1"  # configs/nomad-server.hcl
datacenters = ["dc1"]  # job
```

---

### 4. **Structure Config : trading → strategy**

**Problème :** Config template Nomad utilisait structure incorrecte

```
Erreur: unsupported interval
Cause: timeframe vide car structure config incorrecte
```

**Solution :** Utiliser structure `strategy:` conforme au code

**Fichier modifié :** `configs/scalping-live-bybit.nomad`

```yaml
# Avant (INCORRECT)
trading:
  mode: "live"
  strategy: "SCALPING"
  symbol: "SOLUSDT"
  timeframe: "5m"

# Après (CORRECT)
strategy:
  name: "SCALPING"
  scalping:
    timeframe: "5m"
    cci_surachat: 100.0
    mfi_surachat: 60.0
    # ...
```

**Référence :** Basé sur `deploy/scalping-live-bybit.nomad` (ancien serveur)

---

### 5. **Firewall UFW : Ports Caddy**

**Problème :** Ports 80/8080 non ouverts → timeout connexion Nomad UI

**Solution :** Ajouter règles UFW

```bash
sudo ufw allow 80/tcp
sudo ufw allow 8080/tcp
```

**Résultat :**
```
80/tcp      ALLOW    Anywhere  # Nomad UI via Caddy
8080/tcp    ALLOW    Anywhere  # Health checks
4646/tcp    ALLOW    Anywhere  # Nomad API
51820/udp   ALLOW    Anywhere  # WireGuard VPN
```

---

### 6. **Caddy mTLS : Certificats Client**

**Problème :** Erreur `502 Bad Gateway` - Nomad exige mTLS

```
Error: remote error: tls: certificate required
```

**Solution :** Configurer Caddy avec certificats client Nomad

**Fichier :** `/etc/caddy/Caddyfile`

```caddyfile
# Avant (INCORRECT)
http://10.8.0.1:80 {
    reverse_proxy localhost:4646  # ❌ HTTP vers HTTPS
}

# Après (CORRECT)
http://10.8.0.1:80 {
    reverse_proxy https://localhost:4646 {
        transport http {
            tls
            tls_client_auth /etc/nomad.d/certs/client.pem /etc/nomad.d/certs/client-key.pem
            tls_trusted_ca_certs /etc/nomad.d/certs/ca.pem
        }
    }
}
```

**Impact :** Caddy peut maintenant se connecter à Nomad HTTPS avec mTLS

---

## 🎯 Workflow Mis à Jour

### **Installation Serveur**

```bash
# 1. Upload scripts
rsync -avz devops/ root@31.57.224.79:/root/agent_economique_stable/devops/

# 2. Installation (manuel recommandé pour interactivité)
ssh root@31.57.224.79
cd /root/agent_economique_stable/devops/scripts

./setup-wireguard.sh server    # VPN
./generate-nomad-certs.sh       # TLS (avec SANs ✅)
./install-nomad.sh              # Nomad (driver exec ✅)
./install-caddy.sh              # Caddy

# 3. Configurer Caddy mTLS
cat > /etc/caddy/Caddyfile << 'EOF'
{
    auto_https off
    admin localhost:2019
}

http://10.8.0.1:80 {
    reverse_proxy https://localhost:4646 {
        transport http {
            tls
            tls_client_auth /etc/nomad.d/certs/client.pem /etc/nomad.d/certs/client-key.pem
            tls_trusted_ca_certs /etc/nomad.d/certs/ca.pem
        }
    }
    log {
        output file /var/log/caddy/nomad.log
    }
}

http://10.8.0.1:8080 {
    respond /health 200
    respond /ready 200
}
EOF

systemctl restart caddy

# 4. Firewall
ufw allow 80/tcp
ufw allow 8080/tcp
ufw allow 4646/tcp
ufw allow 51820/udp
```

### **Client Local**

```bash
# 1. VPN
cd devops/scripts
sudo ./setup-wireguard.sh client

# 2. Certificats
./get-nomad-certs.sh
source ~/.nomad-certs/nomad-env.sh

# 3. Test
ping 10.8.0.1
nomad server members
```

### **Déploiement Application**

```bash
# Les configs sont déjà corrigées ✅
cd devops/scripts
./full-deploy.sh
```

---

## 🌐 Accès Nomad UI

### **Option 1 : Chrome WSL (Graphique)**
```bash
google-chrome --no-sandbox http://10.8.0.1:80
```

### **Option 2 : Tunnel SSH (Windows)**
```bash
# Dans WSL
ssh -L 8080:10.8.0.1:80 root@31.57.224.79 -N

# Dans navigateur Windows
http://localhost:8080
```

### **Option 3 : Navigateur texte**
```bash
lynx http://10.8.0.1:80
```

---

## ✅ Vérification

```bash
# Status complet
ssh root@31.57.224.79 '/root/agent_economique_stable/devops/scripts/check-server.sh'

# Doit afficher :
# ✅ WireGuard actif
# ✅ Certificats TLS (7/7)
# ✅ Nomad HTTPS opérationnel
# ✅ Caddy reverse proxy
# ✅ Ports ouverts
# ✅ Job scalping-live-bybit: healthy
```

---

## 📚 Références

- **Scripts corrigés :**
  - `scripts/generate-nomad-certs.sh` (SANs)
  - `scripts/install-caddy.sh` (mTLS config)
  
- **Configs corrigées :**
  - `configs/nomad-server.hcl` (driver exec, dc1)
  - `configs/scalping-live-bybit.nomad` (strategy, exec driver)

- **Ancienne config de référence :**
  - `deploy/scalping-live-bybit.nomad` (structure correcte)

---

## 🔴 Corrections Applicatives - Génération de Signaux (Nov 2025)

### **CRITIQUE : Contrainte de Synchronisation Manquante**

**Contexte :**
Suite à l'analyse des contraintes de génération de signaux, une **contrainte critique** était absente de l'implémentation : la **synchronisation des mouvements** des 3 indicateurs (CCI, MFI, Stochastic) entre N-2 et N-1.

**Problème :**
- Triple extrême détecté sur **UNE seule bougie**
- Synchronisation des indicateurs **absente**
- Risque de signaux avec **divergences** entre indicateurs
- Exemple problématique : CCI↗ + MFI↘ + Stoch↗ → Signal incohérent

**Solution :**

**1. Nouvelle fonction `getTripleExtremeTypeFlexible()`**
```go
// Remplace isTripleExtreme()
// ✅ Retourne "SURACHAT", "SURVENTE" ou ""
// ✅ Chaque indicateur vérifié sur N-1 OU N-2 (flexibilité)
func (s *ScalpingStrategy) getTripleExtremeTypeFlexible(n2Index, n1Index int) string
```

**2. Nouvelle fonction `checkMovementSynchronization()` 🆕**
```go
// ✅ Vérifie que les 3 indicateurs évoluent dans le MÊME SENS
// LONG : CCI↗ + MFI↗ + Stoch↗ (hausse N-2 → N-1)
// SHORT : CCI↘ + MFI↘ + Stoch↘ (baisse N-2 → N-1)
func (s *ScalpingStrategy) checkMovementSynchronization(n2Index, n1Index int, signalType string) bool
```

**3. Modification `DetectSignals()`**
```go
// Nouvelle logique de validation :
// 1️⃣ Triple extrême flexible
// 2️⃣ Croisement stochastique
// 3️⃣ Synchronisation mouvements (NOUVEAU) ⭐
// 4️⃣ Cohérence directionnelle
// 5️⃣ Validation window
// 6️⃣ Volume conditionné
```

**Fichiers modifiés :**
- ✅ `cmd/scalping_live_bybit/app_live.go`
- ✅ `cmd/scalping_live_gateio/app_live.go`
- ✅ `cmd/scalping_engine/app.go`

**Documentation créée :**
- ✅ `docs/CONTRAINTES_SIGNAUX_SCALPING.md` (454 lignes)
  - 6 contraintes détaillées avec exemples
  - Référence complète pour validation conformité

**Déploiement :**
- ✅ `scalping_live_bybit` redéployé avec corrections
- ✅ Job Nomad : `scalping-live-bybit` (running, healthy)
- ✅ Binaire : 9.8M

**Impact :**
- ✅ Prévient signaux avec divergences d'indicateurs
- ✅ Garantit cohérence (tous en hausse ou tous en baisse)
- ✅ Améliore qualité des signaux
- ✅ Documentation alignée avec implémentation

**Voir :** `CHANGELOG.md` (racine) pour détails complets

---

## 🔒 Sécurité

**Tous les changements maintiennent ou améliorent la sécurité :**
- ✅ TLS avec SANs (meilleure validation)
- ✅ mTLS Caddy ↔ Nomad (authentification mutuelle)
- ✅ VPN WireGuard (isolation réseau)
- ✅ Firewall UFW (ports minimaux)
- ✅ Driver exec (moins permissif que raw_exec)

---

**Date :** 6 novembre 2025  
**Infrastructure :** Production (31.57.224.79 - Singapore)  
**Status :** ✅ Opérationnelle
