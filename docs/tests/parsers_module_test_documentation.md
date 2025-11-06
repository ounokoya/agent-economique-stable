# Documentation Tests - Modules Parsers (Klines & Trades)

**Modules:** `internal/data/binance/klines.go` & `internal/data/binance/trades.go`  
**Version:** 0.1.0  
**Objet:** Description de la logique à tester pour chaque fonction des parsers

## Vue d'ensemble des tests requis

Les modules parsers nécessitent des tests exhaustifs pour valider :
- **Parsing précis** : Formats CSV Binance exacts
- **Performance** : Traitement streaming haute vitesse
- **Validation** : Cohérence temporelle et données
- **Robustesse** : Gestion erreurs et données malformées

---

# Module Parser Klines

## Fonction: `NewKlineParser(timeframe string) *KlineParser`

### Logique à tester

#### Test 1: Timeframes supportés
**Objectif:** Validation timeframes Binance standards
**Logique testée:**
- Timeframes valides: "5m", "15m", "1h", "4h"
- Configuration parser selon timeframe
- Validation format timeframe en entrée
- Initialisation structures internes appropriées

**Conditions d'entrée:**
- Timeframes standards: "5m", "15m", "1h", "4h"
- Timeframes invalides: "3m", "2h", "1d", "", "invalid"

**Résultats attendus:**
- Retour KlineParser valide pour timeframes supportés
- Erreur explicite pour timeframes non supportés
- Configuration interne cohérente avec timeframe

#### Test 2: Initialisation structures internes
**Objectif:** Vérifier préparation parser pour parsing
**Logique testée:**
- Allocation buffers appropriés pour timeframe
- Configuration validation temporelle
- Initialisation compteurs et métriques
- Préparation détecteurs d'anomalies

**Résultats attendus:**
- Structures internes allouées et initialisées
- Buffers dimensionnés selon timeframe (5m = plus petit buffer)
- Validation rules configurées par timeframe

---

## Fonction: `Parse(zipReader *ZipStreamReader) (*KlineIterator, error)`

### Logique à tester

#### Test 1: Parsing fichier klines valide
**Objectif:** Extraction complète données OHLCV
**Logique testée:**
- Lecture streaming depuis ZipStreamReader
- Parsing chaque ligne CSV au format Binance
- Validation nombre colonnes (typiquement 12)
- Conversion types string → float64/int64
- Construction structures Kline complètes

**Conditions d'entrée:**
- ZIP contenant fichier CSV klines format Binance standard
- Lignes format: `timestamp,open,high,low,close,volume,close_time,quote_asset_volume,count,taker_buy_base_asset_volume,taker_buy_quote_asset_volume,ignore`

**Résultats attendus:**
- KlineIterator retourné prêt pour iteration
- Toutes lignes CSV parsées sans erreur
- Structures Kline avec tous champs populés
- Validation temporelle initiale passée

#### Test 2: Fichier avec données malformées
**Objectif:** Robustesse face à corruption CSV
**Logique testée:**
- Lignes avec nombre colonnes incorrect
- Valeurs numériques non parsables
- Timestamps invalides ou incohérents
- Lignes vides ou avec caractères spéciaux

**Conditions d'entrée:**
- CSV avec lignes corrompues mélangées à données valides
- Formats de nombres invalides (letters dans float)
- Timestamps hors plage valide

**Résultats attendus:**
- Lignes valides parsées correctement
- Lignes corrompues ignorées avec warning
- Erreur globale si trop de corruption (>5%)
- Logs détaillés des problèmes détectés

#### Test 3: Fichier vide ou headers seulement
**Objectif:** Gestion cas limites
**Logique testée:**
- Fichier CSV vide
- Seulement headers sans données
- Une seule ligne de données

**Résultats attendus:**
- Gestion gracieuse sans crash
- KlineIterator vide mais valide
- Warnings appropriés dans logs

---

## Fonction: `Next() (*Kline, error)` (KlineIterator)

### Logique à tester

#### Test 1: Iteration séquentielle normale
**Objectif:** Parcours complet des klines parsées
**Logique testée:**
- Retour klines dans l'ordre temporel
- Structures Kline complètes et cohérentes
- Performance constante par appel
- Gestion state internal de l'iterator

**Conditions d'entrée:**
- KlineIterator avec 1000+ klines chargées
- Appels Next() séquentiels

**Résultats attendus:**
- Chaque kline retournée complète et valide
- Ordre chronologique respecté
- Performance < 0.1ms par kline
- Pas de memory leak sur iteration longue

#### Test 2: Validation données kline
**Objectif:** Cohérence données financières
**Logique testée:**
- High >= max(Open, Close) et Low <= min(Open, Close)
- Volume >= 0 et Trades >= 0
- CloseTime = OpenTime + timeframe duration
- Prix et volumes dans plages réalistes

**Conditions d'entrée:**
- Klines avec données edge cases mais valides
- Klines avec micro-variations prix

**Résultats attendus:**
- Validation passed pour données cohérentes
- Détection anomalies financières flagrantes
- Warnings sur données suspectes mais non-bloquantes

---

## Fonction: `ValidateKlineSequence(klines []Kline) error`

### Logique à tester

#### Test 1: Séquence temporelle cohérente
**Objectif:** Validation continuité temporelle
**Logique testée:**
- Timestamps croissants strictement
- Gaps détectés selon timeframe (5m = 300s intervals)
- Overlaps ou duplicates détectés
- Validation timeframe consistency

**Conditions d'entrée:**
- Séquence klines 5m avec gaps de 10-15 minutes
- Séquence avec duplicates timestamps
- Séquence avec ordre non chronologique

**Résultats attendus:**
- Gaps détectés et reportés avec précision
- Duplicates identifiés avec position
- Ordre temporel validé
- Errors détaillées pour chaque anomalie

#### Test 2: Validation cross-timeframe
**Objectif:** Cohérence entre timeframes différents
**Logique testée:**
- Klines 5m agrégées cohérentes avec 15m
- Volumes totaux cohérents entre timeframes
- Prix OHLC cohérents sur agrégation

**Conditions d'entrée:**
- Klines 5m couvrant exactement 3 périodes 15m
- Données permettant agrégation complète

**Résultats attendus:**
- Validation agrégation arithmétique correcte
- Détection incohérences entre timeframes
- Métriques qualité données générées

---

# Module Parser Trades

## Fonction: `NewTradeParser() *TradeParser`

### Logique à tester

#### Test 1: Initialisation parser trades
**Objectif:** Configuration appropriée pour trades
**Logique testée:**
- Buffers optimisés pour volume élevé trades
- Configuration aggregation par fenêtres temps
- Initialisation détecteurs patterns trading
- Préparation métriques microstructure

**Résultats attendus:**
- TradeParser configuré pour haute fréquence
- Buffers dimensionnés pour millions trades/jour
- Aggregators initialisés avec fenêtres configurables

---

## Fonction: `Parse(zipReader *ZipStreamReader) (*TradeIterator, error)`

### Logique à tester

#### Test 1: Parsing trades haute fréquence
**Objectif:** Traitement efficace millions de trades
**Logique testée:**
- Parsing CSV format: `trade_id,price,quantity,timestamp,is_buyer_maker,best_match`
- Conversion types optimisée pour vitesse
- Gestion précision prix et quantités
- Validation cohérence trade_id séquentiel

**Conditions d'entrée:**
- Fichier trades journalier (1M+ trades typique)
- Format CSV Binance standard
- Trades avec microsecondes precision

**Résultats attendus:**
- Parsing > 100k trades/seconde
- Précision préservée pour prix et quantités
- Trade_IDs séquentiels validés
- Mémoire usage contrôlé malgré volume

#### Test 2: Détection anomalies trades
**Objectif:** Identification trades suspects
**Logique testée:**
- Prix outliers vs prix récents
- Quantités anormalement élevées
- Timestamps non-chronologiques
- IsBuyerMaker patterns suspects

**Conditions d'entrée:**
- Trades avec outliers price/quantity injectés
- Trades avec ordre temporel perturbé

**Résultats attendus:**
- Outliers détectés avec seuils configurables
- Trades suspects flaggés mais non supprimés
- Métriques qualité mises à jour

---

## Fonction: `AggregateTradesByTimeWindow(trades []Trade, windowMs int64) []TradeWindow`

### Logique à tester

#### Test 1: Agrégation fenêtres temps
**Objectif:** Agrégation OHLCV depuis trades
**Logique testée:**
- Groupement trades par fenêtres windowMs
- Calcul OHLC depuis premier/dernier/min/max prix
- Sommation volumes et comptage trades
- Calcul VWAP et métriques order flow

**Conditions d'entrée:**
- 10k trades sur période 1 heure
- WindowMs = 60000 (1 minute windows)
- Mix buy/sell trades avec variété prix

**Résultats attendus:**
- 60 TradeWindows d'1 minute chacune
- OHLC cohérents avec trades source
- Volumes totaux conservés
- VWAP calculé correctement

#### Test 2: Gestion fenêtres partielles
**Objectif:** Robustesse bordures temporelles
**Logique testée:**
- Première fenêtre possiblement incomplète
- Dernière fenêtre possiblement incomplète
- Fenêtres vides (pas de trades)
- Transitions entre fenêtres

**Conditions d'entrée:**
- Trades débutant milieu d'une fenêtre
- Gaps sans trades sur certaines fenêtres

**Résultats attendus:**
- Fenêtres partielles gérées correctement
- Fenêtres vides reportées avec métadonnées
- Pas de double-comptage aux bordures

---

## Tests de performance intégrés

### Test performance parsing klines
**Objectif:** Validation vitesse traitement
**Logique testée:**
- Parsing 100k+ klines en < 1 seconde
- Mémoire constante indépendamment volume
- CPU usage raisonnable

**Benchmarks:**
- Fichiers différentes tailles (1k à 1M klines)
- Timeframes différents
- Profiling mémoire et CPU

### Test robustesse formats
**Objectif:** Compatibilité variations Binance
**Logique testée:**
- Variations colonnes entre versions Binance
- Encodings différents (UTF-8, ASCII)
- Séparateurs alternatifs (, vs ;)
- Précision décimales variables

### Test intégration streaming
**Objectif:** Compatibilité avec ZipStreamReader
**Logique testée:**
- Parsing depuis stream sans full-load
- Gestion backpressure si parsing plus lent
- Coordination avec contraintes mémoire streaming

---
*Documentation tests v0.1.0 - Modules Parsers*
