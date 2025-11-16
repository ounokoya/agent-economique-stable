# BingX SDK - Spécifications Complètes

## 📋 RÉSUMÉ EXÉCUTIF - SDK BINGX

### **🎯 VUE D'ENSEMBLE :**

**SDK BingX complet en Go** permettant trading automatisé sur **Spot + Futures Perpétuels** avec gestion **multi-comptes** et **scaling multi-serveurs**. **Pas de WebSocket** - API REST uniquement.

---

### **✅ ACTIONS TRADING FONDAMENTALES SUPPORTÉES :**

#### **📊 SPOT TRADING - Actions de Base :**

##### **🟢 ACHETER (BUY) :**
- **Demo** : Simulation achat avec fonds virtuels
- **Live** : Achat réel avec conversion USDT → crypto
- **Types** : Market (immédiat) ou Limit (prix cible)
- **Quantité** : Montant USDT ou quantité crypto
- **Surveillance** : Status ordre jusqu'à exécution complète

##### **🔴 VENDRE (SELL) :**
- **Demo** : Simulation vente avec retour fonds virtuels  
- **Live** : Vente réelle avec conversion crypto → USDT
- **Types** : Market (immédiat) ou Limit (prix cible)
- **Quantité** : Quantité crypto à vendre
- **Surveillance** : Status ordre jusqu'à exécution complète

#### **⚡ FUTURES PERPÉTUELS - Actions de Base :**

##### **📈 OUVRIR POSITION LONG :**
- **Demo** : Simulation position haussière
- **Live** : Position réelle avec effet de levier
- **Configuration** : Levier (1x-125x), marge (Cross/Isolated)
- **Taille** : Quantité en USDT ou nombre de contrats
- **Entry** : Market (immédiat) ou Limit (prix d'entrée)

##### **📉 OUVRIR POSITION SHORT :**
- **Demo** : Simulation position baissière  
- **Live** : Position réelle avec effet de levier
- **Configuration** : Levier (1x-125x), marge (Cross/Isolated)
- **Taille** : Quantité en USDT ou nombre de contrats
- **Entry** : Market (immédiat) ou Limit (prix d'entrée)

##### **✅ FERMER POSITION LONG :**
- **Fermeture totale** : 100% de la position
- **Fermeture partielle** : Pourcentage ou quantité spécifique
- **Types** : Market (immédiat) ou Limit (prix de sortie)
- **PnL réalisé** : Calcul automatique profit/perte

##### **✅ FERMER POSITION SHORT :**
- **Fermeture totale** : 100% de la position
- **Fermeture partielle** : Pourcentage ou quantité spécifique  
- **Types** : Market (immédiat) ou Limit (prix de sortie)
- **PnL réalisé** : Calcul automatique profit/perte

---

### **🔧 FONCTIONNALITÉS AVANCÉES PAR ACTION :**

#### **📊 Spot - Détails Fonctionnels :**

##### **Acheter/Vendre - Capacités :**
- **Vérification soldes** avant ordre
- **Calcul frais** automatique
- **Validation prix** et quantités
- **Historique trades** complet
- **Annulation ordres** en attente
- **Trailing stop** (custom si pas natif)

#### **⚡ Futures - Détails Fonctionnels :**

##### **Ouvrir Positions - Capacités :**
- **Calcul marge** requise automatique
- **Vérification levier** maximum autorisé
- **Mode position** : Hedge (long+short) ou One-way
- **Prix liquidation** calculé en temps réel
- **Funding rate** impact sur position
- **Stop loss** et **Take profit** intégrés

##### **Fermer Positions - Capacités :**
- **PnL temps réel** pendant position ouverte
- **Trailing stop** ajustement automatique
- **Fermeture d'urgence** si marge insuffisante  
- **Réduction only** mode pour diminuer exposition
- **Historique PnL** détaillé par position

---

### **🌐 ENVIRONNEMENTS PAR ACTION :**

#### **📊 SPOT :**
```
Demo VST  : Acheter/Vendre avec fonds virtuels illimités
Live Prod : Acheter/Vendre avec fonds réels + frais réels
```

#### **⚡ FUTURES :**
```
Demo VST  : Positions Long/Short avec marge virtuelle
Live Prod : Positions Long/Short avec marge réelle + liquidation
```

---

### **🚀 SCALING MULTI-SERVEURS (SANS WEBSOCKET) :**

#### **📊 Capacités confirmées :**
```
Rate limit: 10 req/sec par IP (Market Data)
Bot optimal: 1 req/sec par bot (polling prix + ordres)
Capacité: 10 bots par serveur max

3 serveurs = 30 bots = 30 req/sec total ⭐
```

#### **⚡ Optimisations polling :**
- **Cache prix** avec TTL courte (5-10 secondes)
- **Batch requests** quand possible
- **Priorité requêtes** critiques (ordres vs prix)
- **Rate limiter** intelligent par type endpoint

---

### **🏦 MULTI-COMPTES PAR ACTION :**

#### **💰 Transferts automatiques :**
- **Allocation budgets** par bot/stratégie
- **Récupération profits** vers compte principal
- **Isolation risques** par sous-compte
- **Monitoring centralisé** toutes actions

#### **🔐 Sécurité par action :**
- **Permissions granulaires** par API key
- **Limites trading** par sous-compte
- **Audit trail** complet toutes actions

---

### **📋 MATRICE ACTIONS COMPLÈTE :**

| Action | Spot Demo | Spot Live | Futures Demo | Futures Live |
|--------|-----------|-----------|--------------|--------------|
| Acheter | ✅ | ✅ | ❌ | ❌ |
| Vendre | ✅ | ✅ | ❌ | ❌ |
| Ouvrir Long | ❌ | ❌ | ✅ | ✅ |
| Fermer Long | ❌ | ❌ | ✅ | ✅ |
| Ouvrir Short | ❌ | ❌ | ✅ | ✅ |
| Fermer Short | ❌ | ❌ | ✅ | ✅ |
| Prix temps réel | ✅ | ✅ | ✅ | ✅ |
| Candles historiques | ✅ | ✅ | ✅ | ✅ |
| Trailing Stop | ⚠️ | ⚠️ | ✅ | ✅ |
| Multi-comptes | ✅ | ✅ | ✅ | ✅ |

---

### **🎯 CONCLUSION RÉSUMÉ :**

**SDK BingX REST API uniquement** avec **actions trading fondamentales complètes**.

**Spot** : Acheter/Vendre selon environnement avec gestion ordres complète.

**Futures** : Ouvrir/Fermer Long/Short selon environnement avec levier et marge.

**Architecture scalable 30+ bots** avec **isolation risques multi-comptes**.

**Prêt pour validation et implémentation !**

---

## 📋 SPÉCIFICATIONS DÉTAILLÉES

Spécifications complètes pour l'implémentation d'un SDK BingX en Go, basé sur l'analyse approfondie de l'API officielle BingX et l'architecture existante du projet.

## 🎯 OBJECTIFS DU SDK

- **Réutiliser 95%** de l'architecture Binance existante
- **Support complet** des sous-comptes pour isolation des bots
- **Données historiques** pour backtests (klines, trades)
- **Trading live** avec gestion multi-comptes
- **Tests unitaires** complets (couverture 100%)

---

## 🔗 ENDPOINTS API BINGX DISPONIBLES

### 📊 SPOT TRADING API

#### Market Data (Public - Sans authentification)
```
/openApi/spot/v1/common/symbols          # Liste des paires trading
/openApi/spot/v1/market/depth            # Carnet d'ordres (order book)
/openApi/spot/v1/market/trades           # Trades récents
/openApi/spot/v1/market/kline            # Données klines/chandelier ⭐
/openApi/spot/v1/ticker/24hr             # Statistiques 24h
/openApi/spot/v1/ticker/price            # Prix actuel
/openApi/spot/v1/ticker/bookTicker       # Meilleur bid/ask
```

#### Account & Trading (Authentifié)
```
/openApi/spot/v1/account/balance         # Soldes compte
/openApi/spot/v1/account/tradeFee        # Frais de trading
/openApi/spot/v1/trade/order             # Passer ordre (POST)
/openApi/spot/v1/trade/openOrders        # Ordres ouverts
/openApi/spot/v1/trade/historyOrders     # Historique ordres
/openApi/spot/v1/trade/cancel            # Annuler ordre
/openApi/spot/v1/trade/myTrades          # Mes trades exécutés
```

### ⚡ PERPETUAL FUTURES API (Swap V2) - PRIORITAIRE

#### Market Data Futures
```
/openApi/swap/v2/quote/contracts         # Contrats disponibles
/openApi/swap/v2/quote/depth             # Carnet d'ordres futures
/openApi/swap/v2/quote/trades            # Trades récents futures
/openApi/swap/v2/quote/klines            # ⭐ KLINES FUTURES (CIBLE PRINCIPALE)
/openApi/swap/v2/quote/ticker            # Ticker 24h futures
/openApi/swap/v2/quote/price             # Prix mark/index
/openApi/swap/v2/quote/bookTicker        # Meilleur bid/ask futures
/openApi/swap/v2/quote/openInterest     # Open Interest
/openApi/swap/v2/quote/fundingRate      # Funding rate
```

#### Trading Futures
```
/openApi/swap/v2/user/balance            # Solde futures
/openApi/swap/v2/user/positions          # Positions ouvertes
/openApi/swap/v2/user/income             # Historique PnL
/openApi/swap/v2/trade/order             # Passer ordre futures
/openApi/swap/v2/trade/batchOrders       # Ordres en lot
/openApi/swap/v2/trade/closePosition     # Fermer position
/openApi/swap/v2/trade/leverage          # Effet de levier
```

### 🏦 SUB-ACCOUNTS API - FONCTIONNALITÉ CLÉ

#### Gestion Sous-Comptes
```
/openApi/api/v3/sub-account/create       # ✅ Créer sous-compte
/openApi/api/v3/sub-account/list         # ✅ Lister sous-comptes  
/openApi/api/v3/sub-account/uid          # ✅ Query account UID
/openApi/api/v3/sub-account/freeze       # ✅ Freeze/unfreeze sous-comptes
```

#### API Keys Sous-Comptes
```
/openApi/api/v3/sub-account/apikey/create # ✅ Créer API key sous-compte
/openApi/api/v3/sub-account/apikey/query  # ✅ Consulter API keys
/openApi/api/v3/sub-account/apikey/reset  # ✅ Reset API key
/openApi/api/v3/sub-account/apikey/delete # ✅ Supprimer API key
```

#### Transferts et Assets
```
/openApi/api/v3/sub-account/transfer/authorize # ✅ Autoriser transferts
/openApi/api/v3/sub-account/transfer/internal  # ✅ Transfert interne
/openApi/api/v3/sub-account/transfer/history   # ✅ Historique transferts
/openApi/api/v3/sub-account/spot/assets        # ✅ Assets spot sous-compte
/openApi/api/v3/sub-account/balance            # ✅ Soldes sous-comptes
```

### 🔄 WEBSOCKET STREAMING API

#### Endpoints WebSocket
```
wss://open-api-ws.bingx.com/market      # Market data publique
wss://open-api-ws.bingx.com/private     # Données privées

# Streams disponibles
@trade          # Flux trades
@kline_1m       # Flux klines 1 minute  
@kline_5m       # Flux klines 5 minutes ⭐
@depth          # Flux order book
@ticker         # Flux ticker
@account        # Flux soldes compte (privé)
@order          # Flux mises à jour ordres (privé)
@position       # Flux positions (privé)
```

---

## 🔐 AUTHENTIFICATION BINGX

### Signature HMAC SHA256

**Exemple d'authentification :**
```bash
# 1. Paramètres API
quoteOrderQty=20&side=BUY&symbol=ETHUSDT&timestamp=1649404670162&type=MARKET

# 2. Génération signature
echo -n "quoteOrderQty=20&side=BUY&symbol=ETHUSDT&timestamp=1649404670162&type=MARKET" | \
openssl dgst -sha256 -hmac "SECRET_KEY" -hex

# 3. Headers requis
X-BX-APIKEY: [API_KEY]
signature: [HMAC_SIGNATURE]
```

### URLs de Base
```
Production:  https://open-api.bingx.com
Demo/Test:   https://open-api-vst.bingx.com
```

---

## 🏗️ ARCHITECTURE MULTI-COMPTES PROPOSÉE

### Structure Organisationnelle

#### 🏦 Compte Principal (Master Account)
- **Fonction** : Dépôt des fonds principaux
- **Rôle** : Gestion centralisée des assets
- **Opérations** : 
  - Distribution automatique vers bots
  - Monitoring global performance
  - Récupération profits/pertes
  - Allocation budgétaire par stratégie

#### 🤖 Sous-Comptes par Bot
- **Principe** : Un sous-compte = Un bot trading
- **Isolation** : Risques complètement séparés
- **Budget** : Allocation fixe par bot
- **API** : Clés dédiées avec permissions granulaires
- **Monitoring** : Performance individuelle

### Avantages Stratégiques

#### 🛡️ Sécurité
- Bot compromis = impact limité à son sous-compte
- Fonds principaux protégés sur compte master
- API keys avec permissions strictes

#### 📊 Risk Management
- Limite assets par sous-compte
- Contrôle transferts entrants/sortants
- Metrics de risque par bot et globales
- Arrêt d'urgence par bot individuel

#### ⚡ Opérationnel
- Scaling illimité (nouveau bot = nouveau sous-compte)
- Configuration automatique API keys
- Monitoring unifié de la flotte
- Mise à jour stratégies sans impact

---

## 🎯 PARAMÈTRES TRADING POUR BACKTESTS

### Symboles Cibles
```
BTC-USDT    # Bitcoin perpetual
ETH-USDT    # Ethereum perpetual  
SOL-USDT    # Solana perpetual ⭐ (prioritaire mémoire)
SUI-USDT    # Sui perpetual ⭐ (prioritaire mémoire)
```

### Timeframes Supportés
```
"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d", "3d", "1w", "1M"

# Timeframes prioritaires pour backtests (mémoire stratégie)
"5m", "15m", "1h", "4h"
```

### Paramètres Klines
```
# Endpoint principal pour backtests
/openApi/swap/v2/quote/klines

# Paramètres
symbol: "SOL-USDT"      # Paire trading
interval: "5m"          # Timeframe
limit: 1500             # Max klines par requête (optimal)
startTime: timestamp    # Début période
endTime: timestamp      # Fin période
```

---

## 🔧 ADAPTATIONS ARCHITECTURE EXISTANTE

### Réutilisable à 95%

#### ✅ Modules Compatibles
- **Cache système** : Format OHLCV identique
- **Streaming processing** : Même structure de données
- **Aggregation** : Timeframes compatibles  
- **Tests unitaires** : Logique réutilisable
- **Parsers** : Format CSV/JSON adaptable
- **Statistics** : Calculs identiques

#### 🔄 Modifications Spécifiques BingX
- **URLs endpoints** : open-api.bingx.com vs data.binance.vision
- **Headers auth** : X-BX-APIKEY vs X-MBX-APIKEY  
- **Symboles format** : BTC-USDT vs BTCUSDT
- **Rate limits** : 20 req/sec vs 1200 req/min
- **Response format** : Légères différences JSON

### Structure de Fichiers Proposée
```
internal/datasource/bingx/
├── cache.go              # Réutilise binance/cache.go
├── client.go             # Client HTTP avec auth BingX
├── streaming.go          # Réutilise binance/streaming.go
├── downloader.go         # Adaptation endpoints BingX
├── parsers.go            # Réutilise binance/parsers.go
├── aggregator.go         # Réutilise binance/aggregator.go
├── statistics.go         # Réutilise binance/statistics.go
├── subaccounts.go        # ⭐ NOUVEAU - Gestion multi-comptes
├── websocket.go          # ⭐ NOUVEAU - Streaming live
└── types.go              # Types spécifiques BingX
```

---

## 🧪 STRATÉGIE DE TESTS

### Tests Unitaires par Module
```
internal/datasource/bingx/
├── cache_test.go         # Réutilise tests binance
├── client_test.go        # Tests auth HMAC SHA256
├── streaming_test.go     # Tests avec mocks ZIP
├── downloader_test.go    # Tests endpoints BingX
├── parsers_test.go       # Tests parsing klines
├── aggregator_test.go    # Tests agrégation OHLCV
├── statistics_test.go    # Tests calculs statistiques
├── subaccounts_test.go   # ⭐ Tests multi-comptes
└── websocket_test.go     # ⭐ Tests streaming live
```

### Couverture Tests Cible
- **Objectif** : 100% coverage (aligné sur contraintes mémoire)
- **Mock HTTP** : Requêtes API mockées
- **Mock WebSocket** : Streams temps réel mockés
- **Tests intégration** : Workflow complet bout en bout

---

## 📋 PROCHAINES ÉTAPES

### Phase 1 : SDK de Base
1. **Client HTTP** avec authentification HMAC
2. **Endpoints market data** (klines, trades, ticker)
3. **Cache système** adapté de Binance
4. **Tests unitaires** complets

### Phase 2 : Trading Live
1. **Endpoints trading** (ordres, positions, soldes)
2. **WebSocket streaming** temps réel
3. **Risk management** intégré
4. **Tests d'intégration**

### Phase 3 : Multi-Comptes
1. **Gestion sous-comptes** (création, configuration)
2. **API keys automatiques** par sous-compte
3. **Transferts internes** automatisés
4. **Monitoring centralisé** flotte de bots

### Phase 4 : Intégration Complète
1. **Adaptation engine trading** existant
2. **Support stratégies MACD/CCI/DMI** sur BingX
3. **Backtests complets** avec données BingX
4. **Production ready** avec monitoring

---

## 💡 NOTES TECHNIQUES

### Rate Limits BingX - DÉTAILS CRITIQUES

#### 📊 Limites Confirmées
```
Market Data: 100 requêtes/10 secondes par IP  # = 10 req/sec
WebSocket: 10 connections max par IP
Trading: ~15-20 req/sec par IP (estimé conservateur)
Account: ~5-10 req/sec par IP (estimé conservateur)
```

#### 🎯 SCALING MULTI-SERVEURS - CALCULS STRATÉGIQUES

##### Capacité par IP/Serveur
```
Bot usage optimal: 1 requête/seconde par bot
Rate limit: 10 req/sec par IP
Capacité: 10 bots maximum par serveur
```

##### Architecture Multi-Serveurs
```
📊 CALCUL SIMPLE:
Serveur 1 (IP-A): 10 bots × 1 req/sec = 10 req/sec
Serveur 2 (IP-B): 10 bots × 1 req/sec = 10 req/sec  
Serveur 3 (IP-C): 10 bots × 1 req/sec = 10 req/sec
─────────────────────────────────────────────────
TOTAL: 30 bots × 1 req/sec = 30 req/sec ⭐
```

##### 🚀 Stratégies Scaling Avancées
```go
// RÉPARTITION OPTIMALE
// Serveur 1 - Stratégies MACD: 10 bots
// Serveur 2 - Stratégies CCI:  10 bots  
// Serveur 3 - Stratégies DMI:  10 bots

// AVEC SOUS-COMPTES (À TESTER)
// Si rate limits par API key (pas par IP):
// Serveur 1: 5 sous-comptes × 10 req/sec = 50 req/sec
// → 50 bots par serveur au lieu de 10
```

##### ⚡ Optimisations WebSocket
```
ÉCONOMIE MASSIVE:
Au lieu de: 30 bots × 1 req/sec = 30 req/sec API calls
WebSocket: 6 connexions stream → données partagées
Résultat: 30 bots avec <10 req/sec total
```

#### 📈 Capacités Théoriques Maximales

##### Scénario Conservateur
```
3 serveurs × 10 bots = 30 bots simultanés
Rate: 1 req/sec par bot
Total: 30 req/sec
```

##### Scénario Optimisé WebSocket
```
3 serveurs × 10 bots = 30 bots
WebSocket streams: 80% des données
API calls: 20% seulement = 6 req/sec total
```

##### Scénario Multi-Sous-Comptes (À VALIDER)
```
3 serveurs × 5 sous-comptes × 10 bots = 150 bots
WebSocket + sous-comptes
API calls: <20 req/sec total
```

#### 🛡️ Rate Limiter Recommandé
```go
const (
    MarketDataRateLimit   = 8   // req/sec (80% de 10 - marge sécurité)
    TradingRateLimit      = 15  // req/sec (conservateur) 
    AccountRateLimit      = 5   // req/sec (très conservateur)
    WebSocketConnections  = 8   // connexions (80% de 10)
    
    // BURST ALLOWANCE
    BurstLimit           = 20   // requêtes burst courte
    BurstWindow          = 2    // secondes
)

type RateLimiter struct {
    marketLimiter   *rate.Limiter
    tradingLimiter  *rate.Limiter  
    accountLimiter  *rate.Limiter
    globalLimiter   *rate.Limiter
}
```

### Format Réponses
```json
{
  "code": 0,           // 0 = success
  "msg": "success",    // Message statut
  "data": {...}        // Données réponse
}
```

### Gestion Erreurs
```
Code 0: Success
Code 100001: Invalid parameters
Code 100009: Order does not exist
Code 401: Unauthorized
Code 429: Rate limit exceeded
```

---

---

## 🎯 FONCTIONNALITÉS REQUISES - SPÉCIFICATIONS UTILISATEUR

### 📊 SPOT TRADING - FONCTIONNALITÉS COMPLÈTES

#### 📈 Market Data Spot
```go
// PRIX ACTUEL
/openApi/spot/v1/ticker/price            // Prix current
/openApi/spot/v1/ticker/bookTicker       // Meilleur bid/ask

// CANDLES/KLINES
/openApi/spot/v1/market/kline            // Données chandelier historiques
// Paramètres: symbol, interval, limit, startTime, endTime

// CARNET D'ORDRES
/openApi/spot/v1/market/depth            // Order book temps réel
```

#### ⚡ Trading Spot
```go
// PLACER TRADES
/openApi/spot/v1/trade/order             // POST - Ordre market/limit
// Paramètres: symbol, side, type, quantity, price, timeInForce

// TRAILING STOP (Si supporté par BingX Spot)
// ⚠️ À VÉRIFIER: Trailing stop natif ou implémentation custom

// SURVEILLANCE ORDRES
/openApi/spot/v1/trade/openOrders        // Ordres ouverts
/openApi/spot/v1/trade/cancel            // Annuler ordre
/openApi/spot/v1/trade/myTrades          // Trades exécutés
```

#### 💰 Gestion Compte Spot
```go
// SOLDES
/openApi/spot/v1/account/balance         // Soldes disponibles
/openApi/spot/v1/account/tradeFee        // Frais trading
```

---

### ⚡ PERPETUAL FUTURES - FONCTIONNALITÉS COMPLÈTES

#### 📊 Market Data Futures
```go
// PRIX ACTUEL
/openApi/swap/v2/quote/price             // Prix mark/index
/openApi/swap/v2/quote/ticker            // Ticker 24h
/openApi/swap/v2/quote/bookTicker        // Meilleur bid/ask

// CANDLES/KLINES ⭐ PRIORITAIRE
/openApi/swap/v2/quote/klines            // Données chandelier futures
// Paramètres: symbol, interval, limit, startTime, endTime

// INFORMATIONS CONTRATS
/openApi/swap/v2/quote/contracts         // Contrats disponibles
/openApi/swap/v2/quote/fundingRate       // Funding rate
/openApi/swap/v2/quote/openInterest     // Open interest
```

#### 🎛️ Configuration Futures
```go
// EFFET DE LEVIER
/openApi/swap/v2/trade/leverage          // Ajuster levier
// Paramètres: symbol, leverage

// MODE MARGE
/openApi/swap/v2/trade/marginType        // Cross/Isolated
// Paramètres: symbol, marginType

// MODE POSITION
/openApi/swap/v2/trade/positionSide      // Hedge/One-way
// Paramètres: dualSidePosition
```

#### ⚡ Trading Futures
```go
// OUVRIR POSITIONS
/openApi/swap/v2/trade/order             // POST - Ordre futures
// Paramètres: symbol, side, positionSide, type, quantity, price

// TRAILING STOP NATIF BingX
/openApi/swap/v2/trade/order             // Type: TRAILING_STOP_MARKET
// Paramètres: symbol, side, quantity, callbackRate, activationPrice

// FERMER POSITIONS
/openApi/swap/v2/trade/closePosition     // Fermeture totale
/openApi/swap/v2/trade/cancel            // Annuler ordre

// ORDRES EN LOT
/openApi/swap/v2/trade/batchOrders       // Plusieurs ordres simultanés
```

#### 📊 Monitoring Futures
```go
// POSITIONS OUVERTES
/openApi/swap/v2/user/positions          // Positions actuelles avec PnL
// Retour: symbol, size, side, unrealizedPnl, percentage, leverage

// PNL ET HISTORIQUE
/openApi/swap/v2/user/income             // Historique PnL détaillé
// Paramètres: symbol, incomeType, startTime, endTime

// SOLDES FUTURES
/openApi/swap/v2/user/balance            // Solde wallet futures
```

---

### 🏦 GESTION MULTI-COMPTES

#### 💰 Transferts Sous-Comptes
```go
// TRANSFERTS INTERNES
/openApi/api/v3/sub-account/transfer/internal  // Transfert entre sous-comptes
// Paramètres: fromUid, toUid, asset, amount

// SOLDES SOUS-COMPTES
/openApi/api/v3/sub-account/balance            // Solde par sous-compte
/openApi/api/v3/sub-account/spot/assets        // Assets spot sous-compte

// AUTORISATION TRANSFERTS
/openApi/api/v3/sub-account/transfer/authorize // Activer/désactiver transferts
```

---

### 🌐 ENVIRONNEMENTS - DEMO vs LIVE

#### 🧪 Demo/Test (VST)
```go
const DemoBaseURL = "https://open-api-vst.bingx.com"

// SPOT DEMO
DemoSpot := &BingXClient{
    BaseURL: DemoBaseURL,
    APIKey:  "demo_api_key",
    Secret:  "demo_secret_key",
}

// FUTURES DEMO
DemoFutures := &BingXClient{
    BaseURL: DemoBaseURL,
    APIKey:  "demo_api_key", 
    Secret:  "demo_secret_key",
}
```

#### 🚀 Production/Live
```go
const LiveBaseURL = "https://open-api.bingx.com"

// SPOT LIVE
LiveSpot := &BingXClient{
    BaseURL: LiveBaseURL,
    APIKey:  "live_api_key",
    Secret:  "live_secret_key",
}

// FUTURES LIVE  
LiveFutures := &BingXClient{
    BaseURL: LiveBaseURL,
    APIKey:  "live_api_key",
    Secret:  "live_secret_key", 
}
```

---

### 🔧 IMPLÉMENTATION TRAILING STOP INTELLIGENT

#### 🎯 Trailing Stop avec Conditions
```go
type TrailingStopManager struct {
    client          *BingXClient
    positions       map[string]*Position
    conditions      []AdjustmentCondition
    monitoring      bool
}

// CONDITIONS D'AJUSTEMENT
type AdjustmentCondition struct {
    Indicator       string    // "CCI", "MACD", "DMI"
    Trigger         string    // "inverse_zone", "signal_cross" 
    Action          string    // "tighten", "loosen", "close"
    AdjustmentPct   float64   // Pourcentage ajustement
}

// SURVEILLANCE CONTINUE
func (tsm *TrailingStopManager) MonitorStops() {
    // 1. Vérifier trailing stops natifs BingX
    // 2. Appliquer conditions personnalisées
    // 3. Ajuster stops selon indicateurs
    // 4. Détecter fermetures automatiques
}
```

#### ⚡ Workflow Trailing Stop
```go
// 1. OUVRIR POSITION AVEC TRAILING STOP
order := CreatePositionWithTrailingStop(
    symbol:        "SOL-USDT",
    side:          "BUY", 
    quantity:      100,
    leverage:      10,
    callbackRate:  0.5,  // 0.5% trailing
)

// 2. SURVEILLER ET AJUSTER
for position.IsOpen() {
    // Vérifier conditions MACD/CCI/DMI
    if CCIInverseZone() && position.PnL > 0 {
        AdjustTrailingStop(position, 0.3) // Resserrer à 0.3%
    }
    
    if MACDInverseSignal() && position.PnL > 0.02 {
        ClosePosition(position) // Sortie anticipée
    }
    
    // Vérifier si stop touché
    if TrailingStopTriggered(position) {
        LogPositionClosed(position)
        break
    }
}
```

---

### 📋 MATRICE FONCTIONNALITÉS COMPLÈTE

| Fonctionnalité | Spot Demo | Spot Live | Futures Demo | Futures Live |
|---------------|-----------|-----------|--------------|--------------|
| Prix actuel | ✅ | ✅ | ✅ | ✅ |
| Candles/Klines | ✅ | ✅ | ✅ | ✅ |
| Placer trades | ✅ | ✅ | ✅ | ✅ |
| Trailing stop | ⚠️ | ⚠️ | ✅ | ✅ |
| Surveiller stop | ✅ | ✅ | ✅ | ✅ |
| Ajuster levier | ❌ | ❌ | ✅ | ✅ |
| Mode marge | ❌ | ❌ | ✅ | ✅ |
| Ouvrir position | ✅ | ✅ | ✅ | ✅ |
| Accéder PnL | ✅ | ✅ | ✅ | ✅ |
| Transferts sous-comptes | ✅ | ✅ | ✅ | ✅ |
| Soldes compte | ✅ | ✅ | ✅ | ✅ |

**⚠️ Note**: Trailing stop spot peut nécessiter implémentation custom si pas natif BingX.

---

## 🎯 CONCLUSION

**SDK BingX totalement faisable** avec réutilisation massive de l'architecture existante. 

**Fonctionnalités sous-comptes** ouvrent des possibilités énormes pour scaling et risk management.

**Toutes les fonctionnalités requises sont supportées** par l'API BingX avec environnements demo/live complets.

**Prêt pour implémentation** avec spécifications complètes documentées.
