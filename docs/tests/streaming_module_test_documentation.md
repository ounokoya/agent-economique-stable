# Documentation Tests - Module Streaming

**Module:** `internal/data/binance/streaming.go`  
**Version:** 0.1.0  
**Objet:** Description de la logique à tester pour chaque fonction

## Vue d'ensemble des tests requis

Le module streaming nécessite des tests rigoureux pour valider :
- **Performance mémoire** : Contrainte stricte < 100MB
- **Décompression** : Streaming ZIP sans extraction complète
- **Robustesse** : Gestion erreurs et fichiers corrompus
- **Compatibilité** : Support formats Binance Vision

---

## Fonction: `NewZipStreamReader(filePath string) (*ZipStreamReader, error)`

### Logique à tester

#### Test 1: Ouverture fichier ZIP valide
**Objectif:** Initialisation correcte du lecteur streaming
**Logique testée:**
- Ouverture fichier ZIP sans extraction en mémoire
- Validation headers ZIP et structure
- Initialisation scanner bufferisé
- Préparation première entrée pour lecture

**Conditions d'entrée:**
- Fichier ZIP valide de données Binance (klines ou trades)
- Permissions lecture sur le fichier
- Espace suffisant pour buffers internes

**Résultats attendus:**
- ZipStreamReader retourné non-nil et initialisé
- Première entrée ZIP accessible via NextEntry()
- Consommation mémoire < 10MB pour initialisation
- Aucune erreur retournée

#### Test 2: Fichier ZIP corrompu
**Objectif:** Détection précoce corruption ZIP
**Logique testée:**
- Validation headers ZIP lors ouverture
- Détection signature ZIP manquante/invalide
- Gestion fichiers tronqués
- Erreur explicite sans crash

**Conditions d'entrée:**
- Fichier avec extension .zip mais contenu corrompu
- Headers ZIP invalides ou manquants
- Fichier tronqué au milieu des headers

**Résultats attendus:**
- Erreur explicite décrivant le problème
- Retour nil pour ZipStreamReader
- Pas de memory leak sur échec initialisation

#### Test 3: Fichier inexistant ou inaccessible
**Objectif:** Gestion robuste erreurs accès fichier
**Logique testée:**
- Détection fichier inexistant
- Gestion permissions insuffisantes
- Fichier verrouillé par autre processus

**Conditions d'entrée:**
- Chemin vers fichier inexistant
- Fichier sans permissions lecture
- Fichier en cours d'écriture par autre processus

**Résultats attendus:**
- Erreurs système appropriées propagées
- Messages d'erreur informatifs
- Pas de handles fichier laissés ouverts

---

## Fonction: `NextEntry() (*StreamEntry, error)`

### Logique à tester

#### Test 1: Navigation séquentielle entrées ZIP
**Objectif:** Parcours complet des entrées du ZIP
**Logique testée:**
- Lecture headers d'entrée ZIP successives
- Positionnement correct dans le stream
- Métadonnées d'entrée extraites (nom, taille)
- Préparation lecteur pour contenu de l'entrée

**Conditions d'entrée:**
- ZIP contenant plusieurs fichiers CSV (cas typique Binance)
- Lecteur initialisé et positionné au début

**Résultats attendus:**
- Première entrée retournée avec métadonnées correctes
- Appels successifs retournent entrées suivantes
- Nom fichier et taille cohérents avec contenu ZIP

#### Test 2: Fin de ZIP atteinte
**Objectif:** Gestion propre de fin de stream
**Logique testée:**
- Détection fin du ZIP (plus d'entrées)
- Retour approprié (nil + EOF ou équivalent)
- État interne cohérent après fin de stream

**Conditions d'entrée:**
- ZIP avec N entrées, après lecture de la N-ième
- Tentative lecture entrée suivante

**Résultats attendus:**
- Retour nil + io.EOF ou équivalent
- Pas d'erreur inattendue
- Ressources nettoyées appropriément

#### Test 3: Entrée ZIP corrompue
**Objectif:** Robustesse face à corruption interne
**Logique testée:**
- Détection headers d'entrée corrompus
- Gestion CRC invalides
- Tailles incohérentes entre header et contenu

**Conditions d'entrée:**
- ZIP avec une entrée corrompue au milieu
- Headers partiellement valides

**Résultats attendus:**
- Erreur explicite sur entrée corrompue
- Capacité continuer avec entrées suivantes si possible
- Logs détaillés pour diagnostic

---

## Fonction: `ReadLine() ([]byte, error)`

### Logique à tester

#### Test 1: Lecture séquentielle lignes CSV
**Objectif:** Parsing correct format CSV Binance
**Logique testée:**
- Décompression streaming du contenu d'entrée
- Détection délimiteurs de ligne (\n, \r\n)
- Buffer circulaire pour optimiser mémoire
- Retour ligne complète sans délimiteurs

**Conditions d'entrée:**
- Entrée ZIP contenant données klines CSV format Binance
- Lignes de longueur variable (< 1KB typique)

**Résultats attendus:**
- Chaque ligne CSV retournée complète
- Délimiteurs supprimés automatiquement
- Performance > 50MB/s décompression
- Mémoire buffer < 100KB per appel

#### Test 2: Gestion lignes très longues
**Objectif:** Robustesse face à lignes exceptionnelles
**Logique testée:**
- Lignes dépassant taille buffer initial
- Réallocation dynamique si nécessaire
- Limite maximale pour éviter DoS
- Performance maintenue sur lignes normales

**Conditions d'entrée:**
- Entrée avec lignes > 4KB (anormal mais possible)
- Mix lignes normales et très longues

**Résultats attendus:**
- Lignes longues traitées correctement
- Limite raisonnable appliquée (ex: 1MB max)
- Pas de dégradation sur lignes normales

#### Test 3: Fin de contenu d'entrée
**Objectif:** Transition propre entre entrées
**Logique testée:**
- Détection fin du contenu de l'entrée courante
- Retour EOF approprié
- Préparation pour entrée suivante via NextEntry()

**Conditions d'entrée:**
- Lecture complète d'une entrée ZIP
- Tentative lecture ligne suivante

**Résultats attendus:**
- Retour io.EOF pour indiquer fin d'entrée
- État interne prêt pour NextEntry()
- Ressources de décompression libérées

---

## Fonction: `HasNext() bool`

### Logique à tester

#### Test 1: Détection contenu disponible
**Objectif:** Optimisation des boucles de lecture
**Logique testée:**
- Vérification buffer interne non vide
- Look-ahead dans stream sans consommer
- Performance optimisée (pas d'I/O si possible)

**Conditions d'entrée:**
- Stream avec contenu restant à lire
- Buffer interne avec/sans données

**Résultats attendus:**
- Retour `true` si données disponibles
- Pas de consommation données lors vérification
- Performance constante O(1)

#### Test 2: Fin de stream
**Objectif:** Détection propre épuisement données
**Logique testée:**
- Toutes entrées ZIP traitées
- Buffer interne vide
- Stream sous-jacent fermé ou EOF

**Conditions d'entrée:**
- Lecture complète de toutes entrées du ZIP
- Aucune donnée restante

**Résultats attendus:**
- Retour `false` indiquant fin
- État cohérent avec ReadLine() retournant EOF

---

## Fonction: `Close() error`

### Logique à tester

#### Test 1: Fermeture propre ressources
**Objectif:** Nettoyage complet sans leaks
**Logique testée:**
- Fermeture fichier ZIP sous-jacent
- Libération buffers mémoire
- Nettoyage décompresseurs internes
- Invalidation état pour empêcher réutilisation

**Conditions d'entrée:**
- ZipStreamReader en cours d'utilisation
- Buffers alloués et ressources ouvertes

**Résultats attendus:**
- Toutes ressources libérées
- File handles fermés
- Mémoire disponible pour GC
- Retour nil (succès)

#### Test 2: Double fermeture
**Objectif:** Idempotence de l'opération Close
**Logique testée:**
- Détection état déjà fermé
- Pas d'erreur sur fermeture répétée
- Pas de corruption ou double-free

**Conditions d'entrée:**
- ZipStreamReader déjà fermé via Close()
- Appel Close() supplémentaire

**Résultats attendus:**
- Pas d'erreur sur double Close()
- État reste stable
- Pas d'effets de bord

#### Test 3: Fermeture avec erreur
**Objectif:** Gestion erreurs lors fermeture
**Logique testée:**
- Erreurs I/O lors fermeture fichier
- Corruption détectée en fin de stream
- Nettoyage partiel si certaines opérations échouent

**Conditions d'entrée:**
- Fichier avec erreurs I/O
- Système avec ressources limitées

**Résultats attendus:**
- Erreur retournée mais ressources nettoyées
- Pas de blocage ou hang
- Logs appropriés pour diagnostic

---

## Tests de performance critiques

### Test contrainte mémoire
**Objectif:** Validation stricte < 100MB RAM
**Logique testée:**
- Traitement fichier ZIP > 500MB
- Mesure consommation mémoire continue
- Pas de croissance mémoire avec taille fichier
- GC efficace des buffers temporaires

**Configuration test:**
- Fichiers test de différentes tailles (10MB à 1GB)
- Monitoring mémoire RSS et heap
- Détection memory leaks

**Seuils acceptables:**
- Mémoire max < 100MB quelle que soit taille fichier
- Croissance linéaire uniquement avec profondeur stream
- GC libère > 95% mémoire temporaire

### Test performance décompression
**Objectif:** Validation débit > 50MB/s
**Logique testée:**
- Décompression streaming optimisée
- Balance CPU vs I/O
- Pas de goulots d'étranglement buffer

**Benchmarks:**
- Fichiers ZIP avec différents ratios compression
- Mesure débit net (données décompressées/seconde)
- Profiling CPU pour identifier hotspots

### Test robustesse long terme
**Objectif:** Stabilité sur usage prolongé
**Logique testée:**
- Traitement séquentiel de 100+ fichiers ZIP
- Pas de dégradation performance
- Stabilité mémoire sur durée
- Gestion propre erreurs intermittentes

**Durée test:** 1h+ traitement continu
**Métriques:** Performance, mémoire, taux erreur

---

## Tests de compatibilité formats

### Test formats Binance Klines
**Objectif:** Support exact format CSV klines
**Logique testée:**
- Parsing lignes format: timestamp,open,high,low,close,volume,...
- Gestion variations colonnes selon timeframes
- Encoding UTF-8 et caractères spéciaux

### Test formats Binance Trades  
**Objectif:** Support exact format CSV trades
**Logique testée:**
- Parsing lignes format: trade_id,price,quantity,timestamp,...
- Gestion volumes élevés (millions de trades/jour)
- Précision données numériques

### Test compression ZIP
**Objectif:** Support variantes compression Binance
**Logique testée:**
- Différents niveaux compression ZIP
- Algorithmes compression variés
- Headers ZIP avec/sans métadonnées étendues

---
*Documentation tests v0.1.0 - Module Streaming*
