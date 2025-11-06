# User Stories - Application CLI Agent Économique

**Version :** 1.0.0  
**Priorité :** MVP (Minimum Viable Product)

## Epic 1: Configuration et Initialisation

### US-001: Configuration via fichier YAML
**En tant qu'** analyste quantitatif  
**Je veux** pouvoir configurer l'application via un fichier YAML  
**Afin de** définir facilement les symboles, timeframes et paramètres de téléchargement

**Critères d'acceptation :**
- ✅ Le fichier config.yaml est validé au démarrage
- ✅ Erreur explicite si configuration invalide
- ✅ Support configuration par défaut
- ✅ Validation des paramètres obligatoires

**Définition de fini :**
- Configuration chargée sans erreur
- Validation complète des champs
- Messages d'erreur informatifs
- Tests unitaires passent

---

### US-002: Validation paramètres d'entrée
**En tant qu'** utilisateur CLI  
**Je veux** recevoir des messages d'erreur clairs en cas de paramètres invalides  
**Afin de** corriger rapidement ma configuration

**Critères d'acceptation :**
- ✅ Validation symboles (format correct)
- ✅ Validation dates (range valide)
- ✅ Validation timeframes supportés
- ✅ Validation chemins d'accès

---

## Epic 2: Téléchargement de Données

### US-003: Téléchargement automatique données Binance
**En tant qu'** trader algorithmique  
**Je veux** télécharger automatiquement les données historiques Binance  
**Afin d'** alimenter mes algorithmes de trading

**Critères d'acceptation :**
- ✅ Téléchargement klines et trades
- ✅ Support multiple symboles en parallèle
- ✅ Retry automatique en cas d'échec
- ✅ Validation checksums des fichiers

**Définition de fini :**
- Téléchargement réussi sans intervention
- Fichiers intègres validés
- Performance > 10 MB/s
- Logs détaillés du processus

---

### US-004: Gestion cache intelligent
**En tant qu'** utilisateur récurrent  
**Je veux** que l'application évite de re-télécharger les données existantes  
**Afin d'** économiser bande passante et temps

**Critères d'acceptation :**
- ✅ Vérification existence fichiers locaux
- ✅ Validation intégrité cache
- ✅ Nettoyage fichiers corrompus
- ✅ Option force re-download

---

### US-005: Monitoring téléchargement en temps réel
**En tant qu'** opérateur système  
**Je veux** voir la progression du téléchargement en temps réel  
**Afin de** monitorer l'état de l'opération

**Critères d'acceptation :**
- ✅ Barre de progression par fichier
- ✅ Vitesse téléchargement (MB/s)
- ✅ ETA (temps estimé)
- ✅ Statistiques globales

---

## Epic 3: Traitement Multi-Timeframes

### US-006: Parsing automatique des données
**En tant qu'** analyste de données  
**Je veux** que l'application parse automatiquement les fichiers ZIP de tous timeframes  
**Afin d'** obtenir des données structurées prêtes à l'analyse

**Critères d'acceptation :**
- ✅ Parsing klines CSV → structures Go (5m, 15m, 1h, 4h, 1d)
- ✅ Parsing trades CSV → structures Go
- ✅ Validation format et cohérence par timeframe
- ✅ Gestion erreurs parsing gracieuse

---

### US-007: Téléchargement direct multi-timeframes
**En tant qu'** stratégiste trading  
**Je veux** télécharger directement les données dans tous les timeframes disponibles  
**Afin d'** analyser le marché sur plusieurs échelles temporelles sans agrégation

**Critères d'acceptation :**
- ✅ Téléchargement direct 5m, 15m, 1h, 4h, 1d depuis Binance Vision
- ✅ Support parallèle de multiples timeframes
- ✅ Validation disponibilité timeframes par symbole
- ✅ Cache intelligent par timeframe

---

### US-008: Calcul statistiques avancées
**En tant qu'** quantitative analyst  
**Je veux** obtenir des statistiques détaillées sur les données  
**Afin d'** évaluer la qualité et les caractéristiques du marché

**Critères d'acceptation :**
- ✅ Statistiques prix (min, max, moyenne)
- ✅ Statistiques volume et trades
- ✅ Ratios buyer/seller
- ✅ Métriques de qualité données

---

## Epic 4: Performance et Robustesse

### US-009: Optimisation mémoire
**En tant qu'** utilisateur avec ressources limitées  
**Je veux** que l'application respecte les contraintes mémoire  
**Afin de** pouvoir l'exécuter sur des serveurs modestes

**Critères d'acceptation :**
- ✅ Utilisation mémoire < limite configurée
- ✅ Streaming des gros fichiers
- ✅ Garbage collection optimisé
- ✅ Monitoring mémoire temps réel

---

### US-010: Traitement parallèle
**En tant qu'** utilisateur avec besoins performance  
**Je veux** que l'application utilise tous les cœurs CPU disponibles  
**Afin de** minimiser le temps de traitement

**Critères d'acceptation :**
- ✅ Téléchargements parallèles
- ✅ Parsing concurrent
- ✅ Agrégation multi-threaded
- ✅ Configuration nombre workers

---

### US-011: Gestion interruptions gracieuses
**En tant qu'** opérateur système  
**Je veux** pouvoir interrompre l'application proprement  
**Afin de** redémarrer sans perte de données

**Critères d'acceptation :**
- ✅ Signal SIGTERM/SIGINT gérés
- ✅ Sauvegarde état avant arrêt
- ✅ Reprise au dernier checkpoint
- ✅ Nettoyage ressources

---

## Epic 5: Monitoring et Rapports

### US-012: Logs structurés
**En tant qu'** administrateur système  
**Je veux** des logs structurés et searchables  
**Afin de** debugger et analyser le comportement

**Critères d'acceptation :**
- ✅ Logs format JSON
- ✅ Niveaux DEBUG, INFO, WARN, ERROR
- ✅ Correlation IDs pour traçabilité
- ✅ Rotation automatique logs

---

### US-013: Rapport final d'exécution
**En tant qu'** utilisateur final  
**Je veux** un rapport détaillé à la fin de l'exécution  
**Afin de** valider le succès de l'opération

**Critères d'acceptation :**
- ✅ Résumé symboles traités
- ✅ Statistiques téléchargement
- ✅ Métriques performance
- ✅ Liste erreurs/warnings

---

### US-014: Export résultats
**En tant qu'** data scientist  
**Je veux** exporter les données agrégées dans différents formats  
**Afin de** les utiliser dans mes outils d'analyse

**Critères d'acceptation :**
- ✅ Export CSV pour Excel/Python
- ✅ Export JSON pour APIs
- ✅ Export Parquet pour big data
- ✅ Compression optionnelle

---

## Epic 6: Interface Utilisateur

### US-015: Interface ligne de commande intuitive
**En tant qu'** utilisateur CLI  
**Je veux** une interface simple et cohérente  
**Afin de** utiliser l'outil efficacement

**Critères d'acceptation :**
- ✅ Help contextuelle complète
- ✅ Autocomplétion paramètres
- ✅ Validation arguments temps réel
- ✅ Messages d'erreur utiles

---

### US-016: Modes d'exécution flexibles
**En tant qu'** power user  
**Je veux** différents modes d'exécution  
**Afin d'** adapter l'outil à mes besoins spécifiques

**Critères d'acceptation :**
- ✅ Mode téléchargement seul
- ✅ Mode traitement seul  
- ✅ Mode streaming temps réel
- ✅ Mode batch automatisé

---

## Critères de Priorisation

### Priorité P0 (Bloquant MVP)
- US-001, US-003, US-006, US-007, US-012

### Priorité P1 (Important)
- US-002, US-004, US-008, US-013, US-015

### Priorité P2 (Nice to have)
- US-005, US-009, US-010, US-014, US-016

### Priorité P3 (Future)
- US-011

---

**Définition globale de "Fini" :**
- Toutes les user stories P0 implémentées
- Tests unitaires > 90% couverture
- Documentation technique complète
- Performance validée sur environnement cible
