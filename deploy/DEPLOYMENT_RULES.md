# 📋 Règles de Déploiement - Agent Economique

**Date:** 22 Octobre 2025  
**Version:** v0.3.0

---

## 🎯 Principes Fondamentaux

### **1. Séparation des Responsabilités**

Chaque script a **UNE SEULE** responsabilité:

```
❌ MAUVAIS: Un script qui fait tout (SSH, config Nomad, deploy, test, etc.)
✅ BON: Scripts séparés avec responsabilités claires
```

### **2. Ne JAMAIS toucher à la config Nomad existante**

- Nomad est installé et configuré depuis des années
- Il fonctionne bien, on ne le modifie PAS
- Les scripts ne doivent PAS modifier `/etc/nomad.d/nomad.hcl`
- Si config nécessaire → Instructions manuelles SEULEMENT

### **3. Structure de Déploiement Standard**

```
Pour un service Docker (ex: ArangoDB):
  1. deploy_XXX_volume.sh    → Crée les volumes/dossiers à distance
  2. deploy_XXX.sh           → Upload et déploie le job Nomad
  3. XXX.nomad               → Job Nomad (description du service)

Pour un binaire compilé (ex: Agent):
  1. deploy_binary.sh        → Compile, crée dossiers, upload le binaire
  2. deploy_nomad.sh         → Déploie le job Nomad
  3. XXX.nomad               → Job Nomad (lance le binaire)
```

---

## 📁 Structure des Scripts

### **Type 1: Service Docker Simple**

**Exemple: ArangoDB**

```bash
# 1️⃣ deploy_arango_volume.sh
# Responsabilité: Créer le volume sur le serveur
ssh user@host "mkdir -p /opt/volume && chmod 755 /opt/volume"

# 2️⃣ deploy_arangodb.sh  
# Responsabilité: Deploy sur Nomad avec CLI local
nomad job run \
    -token "$TOKEN" \
    -address "$ADDR" \
    -ca-cert="certs/ca.pem" \
    job.nomad

# 3️⃣ arangodb.nomad
# Responsabilité: Définition du service Docker
job "arangodb" {
  # Configuration Docker
}
```

**Ce que le script NE FAIT PAS:**
- ❌ Modifier la config Nomad
- ❌ Installer des dépendances système
- ❌ Tests complexes de connectivité
- ❌ Configuration réseau
- ❌ SSH + upload du job (utilise CLI Nomad local)

**Ce qu'il FAIT:**
- ✅ Utilise CLI Nomad local avec certificats TLS
- ✅ Deploy directement via l'API Nomad
- ✅ Afficher commandes utiles

---

### **Type 2: Binaire Compilé**

**Exemple: Agent Live**

```bash
# 1️⃣ deploy_binary.sh
# Responsabilité: Compiler + Upload binaire
cd backend
go build -o binary
scp binary user@host:/path/
ssh user@host "mkdir -p /path/folders"

# 2️⃣ deploy_nomad.sh
# Responsabilité: Deploy job Nomad avec CLI local
nomad job run \
    -token "$TOKEN" \
    -address "$ADDR" \
    -ca-cert="certs/ca.pem" \
    job.nomad

# 3️⃣ agent.nomad
# Responsabilité: Définition job raw_exec
job "agent" {
  task {
    driver = "raw_exec"
    config { command = "/path/binary" }
  }
}
```

---

## 🚫 Ce qu'on NE FAIT JAMAIS

### **1. Modifier Nomad à distance**

```bash
❌ INTERDIT:
ssh user@host "echo 'config' >> /etc/nomad.d/nomad.hcl"
ssh user@host "systemctl restart nomad"

✅ À LA PLACE:
echo "⚠️  Configuration manuelle requise:"
echo "Ajouter dans /etc/nomad.d/nomad.hcl:"
echo "client { ... }"
```

### **2. Scripts monolithiques**

```bash
❌ MAUVAIS:
deploy_everything.sh
  - Teste SSH
  - Configure Nomad
  - Crée volumes
  - Compile code
  - Upload binaire
  - Deploy Nomad
  - Teste service
  - Configure monitoring

✅ BON:
deploy_volume.sh      → Volumes seulement
deploy_binary.sh      → Compile + Upload
deploy_nomad.sh       → Deploy Nomad
```

### **3. Dépendances cachées**

```bash
❌ MAUVAIS:
# Script qui suppose que X est installé sans le dire

✅ BON:
# Vérifier dépendances explicitement
if ! command -v nomad &> /dev/null; then
    echo "❌ Nomad CLI requis"
    exit 1
fi
```

---

## ✅ Checklist pour Nouveau Service

### **Avant d'écrire un script de déploiement:**

1. [ ] Le service est-il un Docker ou un binaire ?
2. [ ] Quels volumes/dossiers sont nécessaires ?
3. [ ] Quelles sont les dépendances ?
4. [ ] Faut-il compiler quelque chose ?
5. [ ] Quelles credentials sont nécessaires ?

### **Structure minimale:**

```
deploy/
├── deploy_XXX_volume.sh    (si volumes nécessaires)
├── deploy_XXX.sh           (upload + deploy)
├── XXX.nomad               (job Nomad)
└── DEPLOYMENT_GUIDE.md     (documentation)
```

### **Chaque script doit:**

- [ ] Avoir UN objectif clair
- [ ] Être exécutable depuis machine locale
- [ ] Ne PAS modifier config système
- [ ] Afficher ce qu'il fait
- [ ] Gérer les erreurs proprement
- [ ] Donner commandes utiles en fin

---

## 📚 Exemples Réels

### **Déploiement ArangoDB (Service Docker)**

```bash
# Étape 1: Créer volume
./deploy/deploy_arango_volume.sh
# → Crée /opt/arangodb_data sur serveur

# Étape 2: Déployer service
./deploy/deploy_arangodb.sh
# → Upload arangodb.nomad + nomad job run

# Étape 3: Vérifier
ssh root@server "nomad job status arangodb-agent-economique"
```

### **Déploiement Agent (Binaire Go)**

```bash
# Étape 1: Compiler et upload binaire
./deploy/deploy_binary.sh
# → Compile Go + Upload vers /root/data/

# Étape 2: Déployer job Nomad
./deploy/deploy_nomad.sh
# → nomad job run avec TLS

# Étape 3: Vérifier
ssh root@server "nomad alloc logs -f <id>"
```

---

## 🔧 Configuration Manuelle Requise

### **Host Volumes Nomad**

**Si un job nécessite un host volume:**

```bash
# ⚠️  Configuration MANUELLE sur le serveur

# 1. Éditer config Nomad
sudo nano /etc/nomad.d/nomad.hcl

# 2. Ajouter dans section client
client {
  host_volume "volume_name" {
    path      = "/path/to/volume"
    read_only = false
  }
}

# 3. Redémarrer Nomad
sudo systemctl restart nomad
```

**Les scripts ne font QUE afficher ces instructions, ils ne les exécutent PAS.**

---

## 📝 Documentation Requise

### **Pour chaque nouveau service déployé:**

1. **README.md** mis à jour avec:
   - Commandes de déploiement
   - Prérequis
   - Ordre des étapes

2. **DEPLOYMENT_GUIDE.md** avec:
   - Guide détaillé étape par étape
   - Troubleshooting
   - Commandes utiles

3. **Changelog** dans `/changelogs/`:
   - Date
   - Changements apportés
   - Scripts créés/modifiés

---

## 🎯 En Résumé

### **DO (À FAIRE):**

✅ Scripts simples avec responsabilités claires  
✅ Séparer volumes / binaires / déploiement  
✅ Documenter chaque étape  
✅ Gérer les erreurs explicitement  
✅ Afficher commandes utiles  

### **DON'T (À NE PAS FAIRE):**

❌ Scripts monolithiques qui font tout  
❌ Modifier config Nomad à distance  
❌ Supposer que X est installé  
❌ Cacher des dépendances  
❌ Oublier la documentation  

---

## 📖 Référence

**Ce document est la référence officielle pour tous les déploiements.**

Tout script qui ne respecte pas ces règles doit être refactoré.

**Dernière mise à jour:** 22 Octobre 2025
