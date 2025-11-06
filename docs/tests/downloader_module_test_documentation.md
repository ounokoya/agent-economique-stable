# Documentation Tests - Module Téléchargeur

**Module:** `internal/data/binance/downloader.go`  
**Version:** 0.1.0  
**Objet:** Description de la logique à tester pour chaque fonction

## Vue d'ensemble des tests requis

Le module téléchargeur nécessite des tests approfondis pour valider :
- **Robustesse réseau** : Gestion interruptions et reprises
- **Performance** : Vitesse et optimisation bande passante  
- **Intégrité** : Validation checksums et détection corruption
- **Concurrence** : Téléchargements parallèles et thread-safety

---

## Fonction: `NewDownloader(config DownloaderConfig) *Downloader`

### Logique à tester

#### Test 1: Configuration valide
**Objectif:** Vérifier initialisation correcte du client HTTP
**Logique testée:**
- Validation paramètres configuration (timeouts, retry, etc.)
- Création client HTTP avec bonnes options
- Initialisation structures internes (activeDownloads map)
- Setup User-Agent et headers par défaut

**Conditions d'entrée:**
```go
config := DownloaderConfig{
    BaseURL: "https://data.binance.vision",
    Timeout: 30 * time.Second,
    MaxRetries: 3,
    ChunkSize: 8192,
}
```

**Résultats attendus:**
- Downloader retourné non-nil avec client HTTP configuré
- Timeout appliqué correctement au client
- Map activeDownloads initialisée vide
- Mutex correctement initialisé

#### Test 2: Configuration invalide
**Objectif:** Valider gestion erreurs configuration
**Logique testée:**
- URL de base malformée ou vide
- Timeouts négatifs ou zéro
- Nombre de retries incohérent
- Taille de chunk invalide

**Conditions d'entrée:**
- BaseURL vide ou malformée
- Timeout négatif
- MaxRetries < 0

**Résultats attendus:**
- Panic ou erreur explicite sur configuration invalide
- Pas de création Downloader avec config incohérente
- Messages d'erreur clairs pour debugging

---

## Fonction: `DownloadFile(url, localPath string) error`

### Logique à tester

#### Test 1: Téléchargement complet réussi
**Objectif:** Valider téléchargement normal sans interruption
**Logique testée:**
- Requête HTTP GET vers URL Binance valide
- Création répertoires parents du fichier local
- Écriture streaming des données reçues
- Validation taille finale vs Content-Length
- Mise à jour état interne pendant téléchargement

**Conditions d'entrée:**
- URL valide: `https://data.binance.vision/.../SOLUSDT-5m-2023-06-01.zip`
- Chemin local accessible en écriture
- Réseau stable et serveur disponible

**Résultats attendus:**
- Fichier téléchargé intégralement
- Taille correspond à Content-Length
- Retour nil (pas d'erreur)
- DownloadState mis à jour correctement

#### Test 2: Erreur réseau temporaire
**Objectif:** Tester mécanisme de retry automatique
**Logique testée:**
- Détection erreur réseau (timeout, connexion refusée)
- Déclenchement retry avec backoff exponentiel
- Tentatives multiples selon configuration
- Abandon après nombre max de retries

**Conditions d'entrée:**
- Serveur simulé qui échoue N fois puis réussit
- Configuration MaxRetries = 3
- InitialRetryDelay = 1s

**Résultats attendus:**
- 3 tentatives avec délais croissants (1s, 2s, 4s)
- Succès final après échecs temporaires
- Logs des tentatives et erreurs
- Temps total cohérent avec backoff

#### Test 3: Fichier inexistant (404)
**Objectif:** Valider gestion erreurs client
**Logique testée:**
- Détection status HTTP 404
- Pas de retry sur erreur client (4xx)
- Abandon immédiat avec erreur explicite
- Nettoyage fichier partiel si créé

**Conditions d'entrée:**
- URL pointant vers fichier inexistant
- Serveur retourne 404 Not Found

**Résultats attendus:**
- Retour erreur immédiate (pas de retry)
- Message d'erreur explicite avec status HTTP
- Pas de fichier partiel créé localement

---

## Fonction: `ResumeDownload(url, localPath string, offset int64) error`

### Logique à tester

#### Test 1: Reprise téléchargement interrompu
**Objectif:** Valider reprise avec HTTP Range
**Logique testée:**
- Requête HTTP avec header `Range: bytes=offset-`
- Validation que serveur supporte ranges (206 Partial Content)
- Append données à fichier existant à partir de l'offset
- Vérification intégrité finale du fichier complet

**Conditions d'entrée:**
- Fichier partiellement téléchargé (50% complet)
- Offset correspond à la taille actuelle du fichier
- Serveur supporte HTTP ranges

**Résultats attendus:**
- Requête avec header Range correct
- Réception 206 Partial Content
- Fichier complété sans corruption
- Checksum final valide

#### Test 2: Serveur ne supporte pas ranges
**Objectif:** Fallback sur téléchargement complet
**Logique testée:**
- Détection que serveur retourne 200 au lieu de 206
- Suppression fichier partiel existant
- Redémarrage téléchargement depuis le début
- Avertissement dans les logs

**Conditions d'entrée:**
- Serveur qui ignore header Range
- Fichier partiel existant

**Résultats attendus:**
- Détection absence support ranges
- Suppression fichier partiel
- Téléchargement complet depuis début
- Log warning sur absence support ranges

#### Test 3: Offset invalide
**Objectif:** Gestion robuste paramètres offset
**Logique testée:**
- Offset négatif ou supérieur à taille fichier
- Offset non aligné sur taille fichier existant
- Fichier local corrompu/modifié depuis interruption

**Conditions d'entrée:**
- Offset > taille réelle fichier distant
- Fichier local modifié après interruption

**Résultats attendus:**
- Détection incohérence offset
- Erreur explicite ou fallback téléchargement complet
- Pas de corruption données

---

## Fonction: `ValidateChecksum(filePath, expectedChecksum string) bool`

### Logique à tester

#### Test 1: Checksum valide
**Objectif:** Validation intégrité fichier correct
**Logique testée:**
- Lecture complète fichier par chunks
- Calcul SHA256 progressif
- Comparaison avec checksum attendu
- Performance optimisée pour gros fichiers

**Conditions d'entrée:**
- Fichier ZIP téléchargé intégralement
- Checksum SHA256 attendu fourni par Binance

**Résultats attendus:**
- Retour `true` pour fichier intègre
- Calcul performant même pour fichiers > 100MB
- Pas de modification fichier pendant validation

#### Test 2: Checksum invalide
**Objectif:** Détection corruption fichier
**Logique testée:**
- Calcul checksum différent de l'attendu
- Identification type de corruption (troncature, modification)
- Logging détaillé pour diagnostic

**Conditions d'entrée:**
- Fichier avec corruption simulée
- Checksum attendu différent du calculé

**Résultats attendus:**
- Retour `false` pour fichier corrompu
- Log détaillant checksums attendu vs calculé
- Pas de faux positifs

#### Test 3: Erreurs validation
**Objectif:** Robustesse accès fichier
**Logique testée:**
- Fichier inexistant ou inaccessible
- Permissions lecture insuffisantes
- Erreurs I/O pendant lecture
- Format checksum attendu invalide

**Résultats attendus:**
- Gestion gracieuse erreurs accès
- Distinction entre corruption et erreur technique
- Messages d'erreur informatifs

---

## Fonction: `GetFileSize(url string) (int64, error)`

### Logique à tester

#### Test 1: Récupération taille fichier
**Objectif:** Obtenir Content-Length via HEAD request
**Logique testée:**
- Requête HTTP HEAD vers URL
- Extraction header Content-Length
- Conversion string vers int64
- Gestion serveurs sans Content-Length

**Conditions d'entrée:**
- URL Binance valide avec fichier existant
- Serveur retourne Content-Length correct

**Résultats attendus:**
- Taille retournée correspond au fichier réel
- Pas de téléchargement du contenu (HEAD seulement)
- Performance rapide (< 1s)

#### Test 2: Serveur sans Content-Length
**Objectif:** Gestion serveurs ne fournissant pas la taille
**Logique testée:**
- Détection absence header Content-Length
- Fallback possible ou erreur explicite
- Gestion Transfer-Encoding: chunked

**Conditions d'entrée:**
- Serveur retournant 200 OK sans Content-Length

**Résultats attendus:**
- Retour erreur explicite ou taille -1
- Indication claire que taille est inconnue

---

## Fonction: `GetProgress(localPath string) (*DownloadProgress, error)`

### Logique à tester

#### Test 1: Progression téléchargement actif
**Objectif:** Suivi temps réel du téléchargement
**Logique testée:**
- Consultation DownloadState pour le fichier
- Calcul pourcentage basé sur taille totale/téléchargée
- Estimation ETA basée sur vitesse moyenne
- Métriques de performance (MB/s)

**Conditions d'entrée:**
- Téléchargement en cours (50% complété)
- DownloadState avec métriques à jour

**Résultats attendus:**
- Pourcentage correct (50%)
- ETA réaliste basé sur vitesse actuelle
- Vitesse calculée cohérente
- Timestamps à jour

#### Test 2: Fichier non en cours de téléchargement
**Objectif:** Gestion fichiers inactifs
**Logique testée:**
- Recherche dans activeDownloads map
- Détection absence de téléchargement actif
- Erreur appropriée pour fichier inexistant

**Conditions d'entrée:**
- Chemin fichier non présent dans activeDownloads

**Résultats attendus:**
- Erreur explicite "téléchargement non trouvé"
- Pas de confusion avec fichiers terminés

---

## Tests d'intégration requis

### Test téléchargements concurrents
**Objectif:** Valider parallélisme sécurisé
**Logique testée:**
- Téléchargements simultanés de fichiers différents
- Thread-safety des structures partagées
- Gestion bande passante entre téléchargements
- Pas de corruption croisée

**Configuration test:**
- 3-5 téléchargements simultanés
- Fichiers de tailles différentes
- Simulation interruptions aléatoires

### Test robustesse réseau
**Objectif:** Simulation conditions réseau dégradées
**Logique testée:**
- Timeouts réseau aléatoires
- Déconnexions temporaires
- Bande passante limitée
- Reprises multiples en cascade

**Simulation:**
- Proxy avec latence variable
- Interruptions réseau programmées
- Limitation bande passante

### Test performance
**Objectif:** Validation métriques de performance
**Logique testée:**
- Vitesse téléchargement > 10 MB/s sur connexion normale
- Overhead mémoire < 50MB per téléchargement
- CPU usage raisonnable pendant téléchargement
- Pas de memory leaks sur téléchargements répétés

**Benchmarks:**
- Fichiers de différentes tailles (1MB à 1GB)
- Mesure vitesse, mémoire, CPU
- Tests de charge prolongés

---
*Documentation tests v0.1.0 - Module Téléchargeur*
