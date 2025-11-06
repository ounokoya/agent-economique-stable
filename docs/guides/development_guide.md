# Guide de Développement

**Version:** 0.1  
**Statut:** Guide pratique  
**Scope:** Processus de développement et contribution

## 🚀 **Démarrage rapide**

### **Prérequis**
- Go 1.21+
- Connexion internet (téléchargement données Binance)
- ~10GB espace disque (cache données historiques)

### **Installation**
```bash
# Cloner le repository
git clone <repository-url>
cd agent_economique_stable

# Initialiser le module Go
go mod init agent-economique

# Installer les dépendances
go mod tidy
```

### **Configuration de base**
```yaml
# config/config.yaml
binance_data:
  cache_root: "data/binance"
  symbols: ["SOLUSDT", "SUIUSDT", "ETHUSDT"]
  timeframes: ["5m", "15m", "1h", "4h"]
  start_date: "2023-06-01"
  end_date: "2025-06-29"
```

## 🔧 **Standards de développement**

### **Méthodologie obligatoire**
**📋 Processus complet**: [`workflow_methodology_development.md`](../workflow_methodology_development.md)

**Étapes obligatoires :**
1. **Consulter** la méthodologie avant tout développement
2. **Respecter** les contraintes Go ([`constraints_development_go.md`](../constraints_development_go.md))
3. **Suivre** l'organisation des fichiers ([`FILE_ORGANIZATION_RULES.md`](../FILE_ORGANIZATION_RULES.md))
4. **Valider** avec l'utilisateur avant implémentation
5. **Tests** unitaires obligatoires (>90% couverture)
6. **Documentation** mise à jour (changelogs + user stories)

### **Contraintes techniques**
- **Stack**: Go uniquement (pas Python)
- **Taille fichiers**: Max 500 lignes par fichier
- **Architecture**: Éviter pointeurs, fonctions pures privilégiées
- **Tests**: Unitaires obligatoires pour chaque fonction
- **Organisation**: Structure Go standard (internal/, tests/)

### **Workflow Git**
```bash
# 1. Fork du repository
# 2. Créer une branche feature
git checkout -b feature/nouvelle-fonctionnalite

# 3. Développer en suivant les contraintes
# 4. Tests obligatoires
go test ./...

# 5. Commit avec messages descriptifs
git commit -m "feat: ajouter validation checksums Binance"

# 6. Pull Request avec description détaillée
```

## 🧪 **Tests et Qualité**

### **Types de tests**
- **Unitaires**: Chaque fonction testée individuellement
- **Intégration**: Flux complets end-to-end  
- **Performance**: Benchmarks mémoire et CPU
- **Robustesse**: Simulation pannes réseau, corruptions

### **Commandes de test**
```bash
# Tests unitaires complets
go test ./...

# Tests avec couverture
go test -cover ./internal/data/binance/...

# Tests d'intégration
go test -tags=integration ./tests/integration/...

# Benchmarks
go test -bench=. ./internal/data/binance/...
```

### **Documentation des tests**
Chaque module a sa documentation de tests détaillée dans [`../tests/`](../tests/)

## 📋 **Code Review Checklist**

### **Avant soumission :**
- [ ] Méthodologie suivie ([`workflow_methodology_development.md`](../workflow_methodology_development.md))
- [ ] Contraintes architecturales respectées
- [ ] Tests unitaires ajoutés/mis à jour (>90% couverture)
- [ ] Documentation technique mise à jour
- [ ] Performance validée (benchmarks)
- [ ] Logs structurés appropriés
- [ ] Pas de régression sur fonctionnalités existantes

### **Validation finale :**
- [ ] Build réussit sans warnings
- [ ] Tous les tests passent
- [ ] Code formaté (`go fmt`)
- [ ] Pas de vulnérabilités (`go mod audit`)

## 🆘 **Support et Aide**

### **En cas de problème :**
1. **Consulter** [`NAVIGATION.md`](../NAVIGATION.md) pour orientation
2. **Vérifier** [`workflow_methodology_development.md`](../workflow_methodology_development.md) pour processus
3. **Examiner** les exemples dans les user stories
4. **Référencer** la documentation des tests appropriée

### **Resources externes :**
- [Documentation Go](https://golang.org/doc/) - Référence langage
- [Binance Data Vision](https://data.binance.vision) - Source des données
- [Keep a Changelog](https://keepachangelog.com/fr/) - Format changelog
