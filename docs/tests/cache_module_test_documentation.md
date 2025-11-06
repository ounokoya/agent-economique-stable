# Documentation Tests - Module Cache Local

**Module:** `internal/data/binance/cache.go`  
**Version:** 0.1.0  
**Objet:** Description de la logique à tester pour chaque fonction

## Vue d'ensemble des tests requis

Le module cache nécessite des tests robustes pour valider :
- **Intégrité** : Gestion correcte des métadonnées et index
- **Performance** : Temps de réponse et consommation mémoire
- **Concurrence** : Thread-safety des opérations
- **Robustesse** : Gestion d'erreurs et corruption

---

## Fonction: `InitializeCache(rootPath string) (*CacheManager, error)`

### Logique à tester

#### Test 1: Initialisation cache vide
**Objectif:** Vérifier création structure répertoires et index
**Logique testée:**
- Création automatique du répertoire racine si inexistant
- Génération de la structure hiérarchique de base
- Création fichier index JSON initial vide
- Retour d'un CacheManager valide et fonctionnel

**Conditions d'entrée:**
- Répertoire racine n'existe pas
- Aucun fichier de cache préexistant

**Résultats attendus:**
- Structure `data/binance/futures_um/` créée
- Fichier `cache_index.json` créé avec structure valide
- CacheManager retourné non-nil avec index chargé
- Aucune erreur retournée

#### Test 2: Initialisation cache existant
**Objectif:** Vérifier chargement index existant
**Logique testée:**
- Lecture et parsing du fichier index JSON existant
- Validation structure et métadonnées chargées
- Reconstruction état cache en mémoire
- Gestion index corrompu avec fallback

**Conditions d'entrée:**
- Répertoire cache existe avec fichiers
- Index JSON valide avec métadonnées

**Résultats attendus:**
- Index chargé en mémoire avec toutes les entrées
- Métadonnées cohérentes (taille, checksum, dates)
- CacheManager fonctionnel avec état restauré

#### Test 3: Gestion erreurs initialisation
**Objectif:** Valider robustesse face aux erreurs
**Logique testée:**
- Permissions insuffisantes sur répertoire
- Index JSON malformé ou corrompu
- Espace disque insuffisant
- Répertoire racine est un fichier (conflit)

**Résultats attendus:**
- Erreurs appropriées retournées avec messages clairs
- Pas de corruption de l'état système
- Nettoyage ressources partiellement allouées

---

## Fonction: `FileExists(symbol, dataType, date string, timeframe ...string) bool`

### Logique à tester

#### Test 1: Fichier présent dans index
**Objectif:** Vérifier consultation index mémoire
**Logique testée:**
- Génération clé de recherche à partir des paramètres
- Consultation rapide index en mémoire (< 1ms)
- Validation cohérence entrée trouvée
- Retour true pour fichier existant et valide

**Conditions d'entrée:**
- Index contient entrée pour "SOLUSDT", "klines", "2023-06-01", "5m"
- Fichier physique existe et n'est pas corrompu

**Résultats attendus:**
- Retour `true` en < 1ms
- Pas d'accès disque (uniquement index mémoire)

#### Test 2: Fichier absent de l'index
**Objectif:** Vérifier gestion absence
**Logique testée:**
- Recherche exhaustive dans index
- Validation que le fichier n'existe pas physiquement
- Retour false sans erreur

**Conditions d'entrée:**
- Index ne contient pas l'entrée recherchée
- Fichier physique n'existe pas

**Résultats attendus:**
- Retour `false` immédiat
- Performance maintenue même avec gros index

#### Test 3: Validation paramètres d'entrée
**Objectif:** Tester robustesse validation inputs
**Logique testée:**
- Paramètres vides ou null
- Formats de date invalides
- Symboles non supportés
- Types de données inconnus

**Résultats attendus:**
- Gestion gracieuse des paramètres invalides
- Retour false pour inputs malformés
- Pas de panic ou corruption

---

## Fonction: `GetFilePath(symbol, dataType, date string) string`

### Logique à tester

#### Test 1: Génération chemin klines
**Objectif:** Vérifier construction chemin correct
**Logique testée:**
- Parsing date au format "YYYY-MM-DD"
- Construction chemin hiérarchique correct
- Gestion des séparateurs OS (Unix/Windows)
- Format nom fichier conforme Binance

**Conditions d'entrée:**
- symbol="SOLUSDT", dataType="klines", date="2023-06-01", timeframe="5m"

**Résultats attendus:**
- Chemin: `data/binance/futures_um/klines/SOLUSDT/5m/SOLUSDT-5m-2023-06-01.zip`
- Séparateurs corrects pour l'OS
- Répertoires parents inclus dans le chemin

#### Test 2: Génération chemin trades
**Objectif:** Valider format spécifique trades
**Logique testée:**
- Format nom fichier différent pour trades
- Structure répertoire cohérente
- Gestion timeframes absents pour trades

**Conditions d'entrée:**
- symbol="ETHUSDT", dataType="trades", date="2023-06-15"

**Résultats attendus:**
- Chemin: `data/binance/futures_um/trades/ETHUSDT/ETHUSDT-trades-2023-06-15.zip`

#### Test 3: Gestion formats de date
**Objectif:** Tester robustesse parsing dates
**Logique testée:**
- Dates limites (début/fin mois/année)
- Formats alternatifs (avec/sans séparateurs)
- Dates invalides (31 février, etc.)
- Années bissextiles

**Résultats attendus:**
- Parsing correct pour dates valides
- Gestion d'erreur pour dates invalides
- Cohérence avec calendrier Binance

---

## Fonction: `UpdateIndex(fileInfo FileMetadata) error`

### Logique à tester

#### Test 1: Ajout nouvelle entrée
**Objectif:** Vérifier ajout métadonnées
**Logique testée:**
- Insertion nouvelle entrée dans index mémoire
- Calcul et validation checksum fichier
- Mise à jour timestamps (création, modification)
- Sauvegarde atomique index sur disque

**Conditions d'entrée:**
- FileMetadata valide avec toutes les propriétés
- Fichier physique existe et est accessible

**Résultats attendus:**
- Entrée ajoutée à l'index mémoire
- Fichier index JSON mis à jour
- Opération thread-safe
- Retour nil (pas d'erreur)

#### Test 2: Mise à jour entrée existante
**Objectif:** Valider modification métadonnées
**Logique testée:**
- Remplacement entrée existante
- Comparaison checksums ancien/nouveau
- Mise à jour sélective des champs modifiés
- Préservation historique si nécessaire

**Conditions d'entrée:**
- Entrée existe déjà dans l'index
- Nouvelles métadonnées avec changements

**Résultats attendus:**
- Entrée mise à jour avec nouvelles valeurs
- Index cohérent après modification
- Performance maintenue même avec gros index

#### Test 3: Gestion erreurs mise à jour
**Objectif:** Tester robustesse face aux échecs
**Logique testée:**
- Fichier index en lecture seule
- Espace disque insuffisant
- Corruption pendant écriture
- Interruption système

**Résultats attendus:**
- Erreurs appropriées retournées
- Index mémoire cohérent même en cas d'échec
- Pas de corruption données existantes

---

## Fonction: `IsFileCorrupted(filePath string) (bool, error)`

### Logique à tester

#### Test 1: Fichier intègre
**Objectif:** Valider détection fichier sain
**Logique testée:**
- Calcul checksum SHA256 fichier complet
- Comparaison avec checksum attendu de l'index
- Validation taille fichier cohérente
- Vérification accessibilité lecture

**Conditions d'entrée:**
- Fichier ZIP valide téléchargé correctement
- Checksum dans index correspond au fichier

**Résultats attendus:**
- Retour `(false, nil)` - fichier non corrompu
- Performance acceptable même pour gros fichiers
- Pas de modification du fichier pendant vérification

#### Test 2: Fichier corrompu
**Objectif:** Détecter corruption données
**Logique testée:**
- Différence checksum calculé vs attendu
- Taille fichier incohérente
- Fichier tronqué ou partiellement téléchargé
- Headers ZIP corrompus

**Conditions d'entrée:**
- Fichier avec corruption simulée
- Checksum index différent du fichier actuel

**Résultats attendus:**
- Retour `(true, nil)` - corruption détectée
- Log détaillé du type de corruption
- Recommandation re-téléchargement

#### Test 3: Gestion erreurs accès fichier
**Objectif:** Valider robustesse accès fichier
**Logique testée:**
- Fichier inexistant
- Permissions lecture insuffisantes
- Fichier verrouillé par autre processus
- Erreurs I/O disque

**Résultats attendus:**
- Retour `(false, error)` avec erreur explicite
- Distinction claire entre corruption et erreur accès
- Logs appropriés pour debugging

---

## Tests de performance requis

### Test concurrence
**Objectif:** Valider thread-safety
**Logique testée:**
- Accès simultanés lecture/écriture index
- Opérations parallèles sur cache
- Cohérence données sous charge
- Pas de race conditions

### Test mémoire
**Objectif:** Valider consommation RAM
**Logique testée:**
- Index de 10k+ fichiers < 50MB mémoire
- Pas de memory leaks sur opérations répétées
- Croissance linéaire avec taille cache
- GC efficace des ressources

### Test performance
**Objectif:** Valider temps de réponse
**Logique testée:**
- FileExists < 1ms même avec gros index
- UpdateIndex < 100ms avec sauvegarde disque
- Initialisation < 100ms pour cache existant
- Pas de dégradation avec usage prolongé

---
*Documentation tests v0.1.0 - Module Cache*
