# 📝 Changelog DevOps

## [1.1.0] - 2025-11-06

### ✅ Ajouts
- **CORRECTIONS.md** : Documentation complète des correctifs critiques
- **check-server.sh** : Script de vérification infrastructure
- **manage-job.sh** : Utilitaire de gestion des jobs Nomad
- **get-nomad-certs.sh** : Récupération automatique certificats client

### 🔧 Corrections Critiques

#### Certificats TLS
- **Ajout SANs** : Génération certificats avec Subject Alternative Names
- **Flag `-hostname`** : Utilisation correcte de cfssl pour SANs
- **mTLS Caddy** : Configuration Caddy avec certificats client Nomad

#### Configuration Nomad
- **Driver** : `raw_exec` → `exec` (support volumes)
- **Datacenter** : `sg1` → `dc1` (standardisation)
- **Driver exec** : Activation dans `nomad-server.hcl`

#### Configuration Application
- **Structure config** : `trading:` → `strategy:` (conformité code)
- **Template job** : Alignement sur `deploy/scalping-live-bybit.nomad`
- **Timeframe** : Correction détection intervalle Bybit

#### Infrastructure Réseau
- **Firewall UFW** : Ajout ports 80/8080
- **Caddy mTLS** : Configuration reverse proxy avec authentification
- **VPN Routing** : Peer WireGuard correctement configuré

### 📚 Documentation
- **README.md** : Ajout référence CORRECTIONS.md
- **Structure** : Mise à jour arborescence complète
- **Accès UI** : Documentation Chrome WSL + tunnel SSH

### 🎯 Impact
- ✅ Infrastructure 100% fonctionnelle
- ✅ Nomad UI accessible (HTTP + HTTPS)
- ✅ Job scalping-live-bybit : healthy
- ✅ Sécurité maintenue (mTLS, VPN, TLS)

---

## [1.0.0] - 2025-11-05

### 🎉 Version Initiale
- Infrastructure DevOps complète
- Scripts d'installation automatisés
- Documentation guides (4 docs)
- Configuration Nomad + WireGuard + Caddy
- Support Singapore server (31.57.224.79)

### ⚠️ Problèmes Connus (Corrigés en 1.1.0)
- Certificats sans SANs
- Driver raw_exec incompatible volumes
- Config job structure incorrecte
- Ports firewall manquants
- Caddy sans mTLS

---

**Légende :**
- ✅ Ajout
- 🔧 Correction
- 📚 Documentation
- ⚠️ Problème connu
- 🎯 Impact
