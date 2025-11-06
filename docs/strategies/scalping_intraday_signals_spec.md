# 📋 SPÉCIFICATIONS - APPLICATION DÉMO SIGNAUX SCALPING & INTRADAY

**Date:** 2025-11-03  
**Version:** 1.0.0  
**Type:** Spécifications application de trading  

---

## 🎯 **OBJECTIF**

Créer une application de démonstration qui détecte et affiche des signaux de trading pour trois stratégies distinctes sur données Binance Futures (SOLUSDT, timeframe 5m) :

1. **Scalping** : Signaux basés sur double extrême (CCI + MFI) + volume conditionné
2. **Intraday** : Signaux basés sur croisements MACD + tendance DMI
3. **Multi-Timeframe** : Signaux basés sur triple extrême contraire (Stoch + MFI + CCI) avec validation multi-TF

---

## 📊 **DONNÉES SOURCE**

- **Exchange** : Binance Futures perpétuels
- **Paire** : SOLUSDT  
- **Timeframe** : 5m
- **Période analyse** : **300 dernières bougies uniquement** ⭐ **SPÉCIFIÉ**
- **Précision** : 100% (client Binance Futures existant)
- **Indicateurs** : **Obligatoirement tv_standard** ⭐ **SPÉCIFIÉ**

**Indicateurs par stratégie :**
- **Scalping** : CCI TV Standard (période 20) + MFI TV Standard (période 14)
- **Intraday** : MACD TV Standard (12,26,9) + DMI TV Standard (14)
- **Multi-Timeframe** : Stochastique TV Standard (14,3,3) + MFI TV Standard (14) + CCI TV Standard (20)

**Note importante** : 
- L'application analyse exclusivement les 300 dernières bougies pour une détection de signaux en temps réel sur données récentes.
- **TOUS** les indicateurs techniques doivent utiliser les implémentations `tv_standard` pour garantir la précision 100% TradingView.

---

## 🎯 **STRATÉGIE TRIPLE EXTRÊME (Multi-Timeframe)**

### **Vue d'ensemble**
Stratégie universelle basée sur le **triple extrême simultané** (CCI, MFI, Stochastique) avec croisement Stochastique et validation par bougie inverse + volume conditionné.

**⚡ DÉCOUVERTE IMPORTANTE** : Cette stratégie est **TRÈS PUISSANTE** sur les timeframes supérieurs (1h, 4h, 1D) avec une qualité de signaux significativement améliorée.

### **Timeframes recommandés**
- **5m-15m** : Scalping (nombreux signaux, réactivité rapide)
- **1h-4h** : Swing intraday (signaux de qualité, moins de bruit) ⭐ **RECOMMANDÉ**
- **1D** : Swing trading (tendances solides, meilleur R:R)

### **Conditions d'ouverture**

#### 1️⃣ **TRIPLE extrême simultané (CCI + MFI + Stochastique)**
```go
// Pour signal SHORT (SURACHAT)
CCI > 100
MFI > 80
Stoch K ≥ 80 OU Stoch D ≥ 80

// Pour signal LONG (SURVENTE)
CCI < -100
MFI < 20
Stoch K ≤ 20 OU Stoch D ≤ 20
```

#### 2️⃣ **Croisement Stochastique dans l'extrême (sur 2 barres fermées)**
```go
// RÈGLE : Utiliser les 2 DERNIÈRES BARRES FERMÉES (N-2 et N-1)
// JAMAIS la barre actuelle en cours

// Pour signal SHORT (après SURACHAT)
Bougie N-2 : K ≥ D
Bougie N-1 : K < D (K passe SOUS D) → Croisement BAISSIER

// Pour signal LONG (après SURVENTE)
Bougie N-2 : K ≤ D
Bougie N-1 : K > D (K passe AU-DESSUS de D) → Croisement HAUSSIER
```

#### 3️⃣ **Fenêtre de validation bougie + volume (N=3 bougies)**
```go
// OUVERTURE FENÊTRE dès croisement détecté
windowStart = indexCroisement
windowEnd = indexCroisement + N  // N=3 par défaut

// Chercher PREMIÈRE bougie validante dans fenêtre
for i in [windowStart, windowEnd):
    // Bougie inverse requise
    if (SURACHAT && bougie ROUGE) OU (SURVENTE && bougie VERTE):
        // Volume conditionné
        if VolumeActuel > 25% × VolumeMoyenBougiesInverses:
            SIGNAL_GÉNÉRÉ at i ✅
            break
            
// Si fenêtre expirée sans validation → Signal perdu
```

#### 4️⃣ **Calcul volume moyen bougies inverses**
```go
// Algorithme extension automatique
fonction CalculerVolumeMoyenInverses(signal, periodes=5):
    extension = 1
    TANT QUE vrai:
        periodesAnalyse = periodes × extension
        // IMPORTANT: cherche bougies INVERSES à la bougie actuelle
        // Objectif: vérifier si volume actuel > 25% moyenne tendance précédente
        bougiesInverses = identifierBougiesInverses(periodesAnalyse, signal)
        
        SI bougiesInverses.nonVide():
            volumeMoyen = moyenne(bougiesInverses.volume)
            RETOURNER volumeMoyen, periodesAnalyse
        
        extension *= 2  // N×2, N×4, N×8...
        
        SI periodesAnalyse > 100:  // Limite sécurité
            RETOURNER 0, 0
```

### **Logique de signal**

**SIGNAL SHORT :**
```
Bougie N-1 (dernière fermée) :
1. Triple extrême SURACHAT (CCI>100 + MFI>80 + Stoch≥80) ✅

Comparaison N-2 vs N-1 :
2. Croisement Stoch baissier (K[N-2]≥D[N-2] → K[N-1]<D[N-1]) ✅

Fenêtre validation [N-1, N, N+1] :
3. Chercher bougie ROUGE + volume>25% moyenne inverses ✅
4. Dès validation → SIGNAL SHORT généré
```

**SIGNAL LONG :**
```
Bougie N-1 (dernière fermée) :
1. Triple extrême SURVENTE (CCI<-100 + MFI<20 + Stoch≤20) ✅

Comparaison N-2 vs N-1 :
2. Croisement Stoch haussier (K[N-2]≤D[N-2] → K[N-1]>D[N-1]) ✅

Fenêtre validation [N-1, N, N+1] :
3. Chercher bougie VERTE + volume>25% moyenne inverses ✅
4. Dès validation → SIGNAL LONG généré
```

**Paramètres :**
- N = 3 (taille fenêtre validation, configurable)
- Barres utilisées : N-2 et N-1 (JAMAIS la barre actuelle)
- Volume : 25% moyenne bougies inverses
- Extension volume : 5 → 10 → 20 → 40 → max 100 périodes

### **Seuils recommandés par timeframe**

**5m-15m (Scalping)** :
```
CCI : ±100
MFI : 20/80
Stoch : 20/80
```

**1h (Intraday)** :
```
CCI : ±100
MFI : 35/65
Stoch : 30/70
```

**4h (Swing)** ⭐ **OPTIMAL** :
```
CCI : ±100
MFI : 40/60
Stoch : 30/70
Sélectivité : ~1 signal/27 bougies
```

**1D (Position)** :
```
CCI : ±100
MFI : 45/55
Stoch : 30/70
```

### **Résultats testés (SOLUSDT, 300 bougies)**

| Timeframe | Signaux | LONG/SHORT | Sélectivité | Qualité |
|-----------|---------|------------|-------------|---------|
| 5m        | ~15-20  | Variable   | 1/15-20     | Moyenne |
| 15m       | ~7-9    | Équilibré  | 1/33        | Bonne   |
| 1h        | ~10     | 50/50      | 1/30        | Très bonne |
| 4h        | ~11     | 64/36      | 1/27        | Excellente ⭐ |

---

## 📈 **STRATÉGIE INTRADAY**

### **Logique de contre-tendance avec fenêtre de croisement**

#### **Conditions de marché requises**

**Pour signal SHORT (vente) :**
```go
DI+ > DI-          // Forte tendance haussière
MACD > 0           // Position MACD positive  
Signal > 0         // Signal positif
ADX > DI+          // Force supérieure à la tendance haussière
```

**Pour signal LONG (achat) :**
```go
DI- > DI+          // Forte tendance baissière
MACD < 0           // Position MACD négative
Signal < 0         // Signal négatif  
ADX > DI-          // Force supérieure à la tendance baissière
```

#### **Déclenchement sur croisements simultanés**

**Fenêtre de recherche : M = 6 bougies (configurable)**

**Pour SHORT :**
- **Croisement MACD** : MACD passe **sous** Signal (baissier)
- **Croisement DX/ADX** : DX passe **sous** ADX (baissier)
- **Simultanéité** : Les deux croisements dans fenêtre M=6

**Pour LONG :**
- **Croisement MACD** : MACD passe **au-dessus** de Signal (haussier)  
- **Croisement DX/ADX** : DX passe **au-dessus** de ADX (haussier)
- **Simultanéité** : Les deux croisements dans fenêtre M=6

#### **🔄 Algorithme de détection par fenêtre glissante**

```go
// 1. Attendre premier croisement (MACD ou DX/ADX)
firstCross = waitForFirstCrossing()

if (firstCross.detected) {
    // 2. OUVRIR FENÊTRE de M=6 périodes à partir du premier croisement
    windowStart = firstCross.index
    windowEnd = firstCross.index + 6
    
    // 3. Chercher deuxième croisement dans la fenêtre
    secondCross = findSecondCrossing(windowStart, windowEnd)
    
    if (secondCross.detected) {
        // 4. Valider conditions indépendantes dans la fenêtre restante
        validationStart = secondCross.index
        validationEnd = windowEnd
        
        // 5. Vérifier chaque condition séparément
        dxValidated = validateDX(validationStart, validationEnd)
        adxValidated = validateADX(validationStart, validationEnd) 
        diValidated = validateDI(validationStart, validationEnd)
        
        // 6. Si TOUTES conditions validées → SIGNAL
        if (dxValidated && adxValidated && diValidated) {
            SIGNAL_GENERATED(validationMoment, signalType)
        }
    }
    
    // 7. FERMER FENÊTRE (validée ou non)
    // 8. Attendre nouveau premier croisement
}
```

#### **📋 Logique de fenêtre par étapes**

**Étape 1 - Premier croisement (déclencheur) :**
- **MACD** ou **DX/ADX** se produit à bougie X
- **OUVERTURE FENÊTRE** : [X, X+6] (6 périodes après le croisement)
- **Type signal déterminé** par le type du premier croisement

**Étape 2 - Deuxième croisement requis :**
- **Recherche active** de l'autre croisement dans la fenêtre [X, X+6]
- **MACD + DX/ADX** doivent être présents dans la fenêtre
- **Types compatibles** requis pour le signal visé

**Étape 3 - Validation des conditions indépendantes :**
- **Période de validation** : [deuxièmeCroisement, X+6]
- **DX** : validé indépendamment à n'importe quelle bougie
- **ADX** : validé indépendamment à n'importe quelle bougie  
- **DI** : validé indépendamment à n'importe quelle bougie
- **Chaque condition** peut être validée à des moments différents

**Étape 4 - Génération du signal :**
- **Signal généré** quand TOUTES conditions sont validées
- **Moment du signal** = instant où la dernière condition est validée
- **Fenêtre fermée** après validation ou échec

**🎯 Avantages :**
- **Flexibilité temporelle** : chaque condition validée à son propre moment
- **Précision** : fenêtre définie par le premier croisement réel
- **Robustesse** : validation indépendante évite les rejets prématurés

### **Logique de signal**
- **Contre-tendance haussière** : DI+>DI-, MACD>0, Signal>0, ADX>DI+ → croisements simultanés MACD↓ + DX↓ → SIGNAL SHORT
- **Contre-tendance baissière** : DI->DI+, MACD<0, Signal<0, ADX>DI- → croisements simultanés MACD↑ + DX↑ → SIGNAL LONG

### **🔄 Logique de détection des croisements**
**Un croisement nécessite DEUX bougies et génère UN SEUL signal :**

#### **Croisement MACD haussier (LONG) :**
```
Bougie N-1 : MACD < Signal
Bougie N   : MACD > Signal
→ SIGNAL LONG détecté à la bougie N
```

#### **Croisement MACD baissier (SHORT) :**
```
Bougie N-1 : MACD ≥ Signal  
Bougie N   : MACD < Signal
→ SIGNAL SHORT détecté à la bougie N
```

#### **Croisement DX/ADX haussier (LONG) :**
```
Bougie N-1 : DX < ADX
Bougie N   : DX > ADX  
→ SIGNAL LONG détecté à la bougie N
```

#### **Croisement DX/ADX baissier (SHORT) :**
```
Bougie N-1 : DX ≥ ADX
Bougie N   : DX < ADX
→ SIGNAL SHORT détecté à la bougie N
```

---

## **📊 Exemple concret de logique par fenêtre**

### **Scénario réel :**
- **6h15** : Premier croisement MACD SHORT (bougie 66)
- **6h20** : Deuxième croisement DX/ADX SHORT (bougie 67)
- **6h25** : Condition ADX validée
- **6h30** : Condition DI validée
- **6h35** : Condition DX validée

### **✅ Logique par fenêtre glissante :**

```go
// 1. Premier croisement à 6h15 (bougie 66)
firstCross = MACD_SHORT@66
windowStart = 66
windowEnd = 66 + 6 = 72 (6h45)

// 2. Recherche deuxième croisement dans fenêtre [66, 72]
secondCross = DX_SHORT@67 (trouvé dans fenêtre ✅)

// 3. Validation indépendante dans [67, 72]
validationStart = 67
validationEnd = 72

// 4. Vérification séparée des conditions :
- ADX validé à bougie 68 (6h25) ✅
- DI validé à bougie 70 (6h30) ✅  
- DX validé à bougie 71 (6h35) ✅

// 5. TOUTES conditions validées → SIGNAL
SIGNAL_GENERATED@71 (6h35, SHORT)
```

### **🎯 Avantages de cette approche :**

1. **Flexibilité temporelle** : Chaque condition validée à son propre rythme
2. **Fenêtre précise** : Définie par le premier croisement réel (6h15→6h45)
3. **Validation indépendante** : ADX à 6h25, DI à 6h30, DX à 6h35
4. **Signal au moment optimal** : Généré quand dernière condition validée (6h35)

### **❌ Cas d'échec :**
```go
// Si condition DI non validée dans [67, 72] :
- ADX validé ✅
- DI validé ❌ (jamais dans fenêtre)
- DX validé ✅
→ FENÊTRE FERMÉE sans signal
→ Attendre nouveau premier croisement
```

**⚠️ Important :** Un signal à 14:05 et un signal à 14:10 sont deux signaux distincts, chacun validé par sa propre paire de bougies !

---

## ⚙️ **PARAMÈTRES CONFIGURATION**

### **Indicateurs techniques**
```go
// Périodes par défaut
CCI_Period     = 20
MFI_Period     = 14
Stoch_KPeriod  = 14
Stoch_DPeriod  = 3
MACD_Fast      = 12
MACD_Slow      = 26
MACD_Signal    = 9
DMI_Period     = 14

// Paramètres stratégie intraday
Intraday_WindowCroisement = 6  // Fenêtre M pour croisements simultanés
```

### **Volume conditionné**
```go
// Paramètres volume
Volume_AnalysePeriode   = 5      // Périodes initiales
Volume_SeuilPourcentage = 25.0   // 25% du volume moyen
Volume_MaxExtension     = 100    // Limite sécurité
```

### **Seuils zones extrêmes**
```go
// Zones extrêmes scalping
CCI_Surachat   = 100
CCI_Survente   = -100
MFI_Surachat   = 80
MFI_Survente   = 20
Stoch_Surachat = 80
Stoch_Survente = 20
```

---

## 🏗️ **STRUCTURE APPLICATION**

### **Structure application**
```
cmd/signals_demo/
├── main.go              # Application principale avec stratégies
├── types.go             # Types Signal et StrategyConfig  
├── scalping_strategy.go # Implémentation stratégie scalping
├── intraday_strategy.go # Implémentation stratégie intraday
└── README.md            # Documentation utilisation
```

### **Implémentation indicateurs - OBLIGATOIRE tv_standard** ⭐
```go
// CCI TradingView Standard
cciTV := indicators.NewCCITVStandard(20)
cciValues := cciTV.Calculate(high, low, close)

// MFI TradingView Standard  
mfiTV := indicators.NewMFITVStandard(14)
mfiValues := mfiTV.Calculate(high, low, close, volume)

// Stochastic TradingView Standard
stochTV := indicators.NewStochTVStandard(14, 3, 3)
stochK, stochD := stochTV.Calculate(high, low, close)

// MACD TradingView Standard
macdTV := indicators.NewMACDTVStandard(12, 26, 9)
macd, signal, hist := macdTV.Calculate(close)

// DMI TradingView Standard (avec DX pour croisements)
dmiTV := indicators.NewDMITVStandard(14)
diPlus, diMinus, adx := dmiTV.Calculate(high, low, close)
dx := dmiTV.CalculateDX(high, low, close)  // DX nécessaire pour croisements
```

### **Architecture**
```go
// Structures principales
type Signal struct {
    Timestamp     time.Time
    Strategy      string  // "SCALPING" ou "INTRADAY"
    Direction     string  // "LONG" ou "SHORT"
    Price         float64
    Conditions    []string
    Confidence    float64
}

type StrategyConfig struct {
    CCIPeriod     int
    MFIPeriod     int
    VolumePeriod  int
    VolumeSeuil   float64
}

type SignalResult struct {
    ScalpingSignals []Signal
    IntradaySignals []Signal
    Summary         StrategySummary
}
```

---

## 📊 **FORMAT AFFICHAGE**

### **Tableau résumé**
```
🔍 DÉTECTION SIGNAUX - SOLUSDT 5m
=====================================

📊 RÉSUMÉ SCALPING:
- Signaux détectés: 3 (2 LONG, 1 SHORT)
- Taux réussite: 67% (2/3)
- Profit moyen: +1.2%

📊 RÉSUMÉ INTRADAY:  
- Signaux détectés: 2 (1 LONG, 1 SHORT)
- Taux réussite: 100% (2/2)
- Profit moyen: +2.1%
```

### **Détail signaux**
```
🎯 SIGNAUX SCALPING DÉTECTÉS:
┌─────────────────────┬────────┬─────────┬──────────┬─────────────────┐
│ Heure               │ Signal │ Prix    │ Confidence│ Conditions      │
├─────────────────────┼────────┼─────────┼──────────┼─────────────────┤
│ 15:45               │ LONG   │ 185.23  │ 85%      │ CCI:-120,MFI:15  │
│ 16:20               │ SHORT  │ 186.45  │ 92%      │ CCI:+145,MFI:88  │
│ 17:05               │ LONG   │ 184.78  │ 78%      │ CCI:-105,MFI:18  │
└─────────────────────┴────────┴─────────┴──────────┴─────────────────┘

🎯 SIGNAUX INTRADAY DÉTECTÉS:
┌─────────────────────┬────────┬─────────┬──────────┬─────────────────┐
│ Heure               │ Signal │ Prix    │ Confidence│ Conditions      │
├─────────────────────┼────────┼─────────┼──────────┼─────────────────┤
│ 15:30               │ LONG   │ 185.12  │ 88%      │ MACD↑,DI+>DI-    │
│ 16:55               │ SHORT  │ 186.89  │ 91%      │ MACD↓,DI->DI+    │
└─────────────────────┴────────┴─────────┴──────────┴─────────────────┘
```

---

## 🧪 **VALIDATION**

### **Tests unitaires**
```go
// Test condition volume inversé
func TestVolumeInverseCondition(t *testing.T)

// Test triple extrême simultané  
func TestTripleExtremeSimultane(t *testing.T)

// Test croisement MACD
func TestMACDCrossOver(t *testing.T)

// Test tendance DMI
func TestDMITrendStrength(t *testing.T)
```

### **Tests intégration**
```bash
# Lancer application démo
go run cmd/signals_demo/main.go

# Résultats attendus
- Signaux scalping: 0-5 par session
- Signaux intraday: 1-3 par session  
- Temps exécution: <2 secondes
- Précision calculs: 100%
```

---

## 🚀 **UTILISATION**

### **Lancement**
```bash
cd cmd/signals_demo/
go run main.go

# Options disponibles
--exchange=binance      # Exchange (défaut: binance)
--symbol=SOLUSDT       # Paire (défaut: SOLUSDT)
--timeframe=5m         # Timeframe (défaut: 5m)
--periods=300          # **FIXÉ à 300 bougies** ⭐ **SPÉCIFIÉ**
--verbose              # Mode debug
```

### **Sortie attendue**
```
🔍 DÉTECTION SIGNAUX SCALPING & INTRADAY
=========================================
📡 Connexion Binance Futures...
✅ 300 klines récupérées (2025-11-03 12:20 → 17:20) ⭐ **300 dernières bougies**

🎯 ANALYSE SCALPING:
🔍 EXTREME SURACHAT détecté à 14:25 - CCI:145.2 MFI:89.1 STOCH:87.3/88.9
   ✅ Bougie inverse: 185.45 < 185.67 (bearish)
   ✅ Volume: 45678 > 12345 (25% moyenne sur 5 bougies inverses)
🎯 SIGNAL SCALPING SHORT à 14:25 - Prix: 185.45 - Confiance: 85.0%

🔍 EXTREME SURVENTE détecté à 15:10 - CCI:-156.3 MFI:12.4 STOCH:8.2/9.1
   ✅ Bougie inverse: 177.23 > 176.89 (bullish)
   ✅ Volume: 234567 > 45678 (25% moyenne sur 10 bougies inverses)
🎯 SIGNAL SCALPING LONG à 15:10 - Prix: 177.23 - Confiance: 90.0%

   Conditions extrêmes trouvées: 28
   Signaux volume validés: 8
   Signaux scalping générés: 8

📈 ANALYSE INTRADAY:
🔍 COND MARCHÉ SHORT: DI+>DI- (25.3>18.1), MACD>0 (0.234), Signal>0 (0.198), ADX>DI+ (28.7>25.3)
   ✅ Croisement MACD baissier dans fenêtre M=6: 0.234 < 0.198
   ✅ Croisement DX/ADX baissier dans fenêtre M=6: 26.1 < 28.7
🎯 SIGNAL INTRADAY SHORT à 16:15 - Prix: 182.45 - Confiance: 95.0%

   Conditions marché trouvées: 5
   Croisements simultanés validés: 1
   Signaux intraday générés: 1

🎯 SIGNAUX TOTAUX: 9 (8 SCALPING + 1 INTRADAY)
```

---

## 📈 **MÉTRIQUES PERFORMANCE**

### **Indicateurs à suivre**
- **Nombre signaux** : Scalping vs Intraday
- **Taux réussite** : Signaux profitables
- **Profit moyen** : PIPS ou pourcentage
- **Durée moyenne** : Temps en position
- **Confiance moyenne** : Score 0-100%

### **Optimisations futures**
- **Adaptation automatique** : Ajustement périodes selon volatilité
- **Filtre temporel** : Éviter signaux faible volume
- **Multi-timeframe** : Confirmation sur 15m avant signal 5m
- **Risk management** : Calcul stop-loss automatique

---

## 🎯 **STRATÉGIE MULTI-TIMEFRAME - EXTREMES CONTRAIRES**

### **Concept stratégique**
Détection des meilleurs points d'entrée contraires en utilisant 3 timeframes et 3 indicateurs en zone extrême pour une confiance maximale.

### **Timeframes utilisés**
- **TF Principal** : 5m (décision finale)
- **TF Confirmation** : 15m (tendance globale)
- **TF Contexte** : 1h (marché global)

### **Conditions d'exécution - Triple Extrême Contraire**

#### 1️⃣ **Stochastique Extrême + Croisement Logique**
```go
// Zone SURVENTE (≤20) : Uniquement croisement LONG
if (stochK <= 20 || stochD <= 20) && croisementType == "LONG"

// Zone SURACHAT (≥80) : Uniquement croisement SHORT  
if (stochK >= 80 || stochD >= 80) && croisementType == "SHORT"
```

#### 2️⃣ **MFI Extrême Confirmé**
```go
// Zones extrêmes obligatoires
MFI <= 20  // SURVENTE pour signal LONG
MFI >= 80  // SURACHAT pour signal SHORT
```

#### 3️⃣ **CCI Extrême Confirmé**
```go
// Zones extrêmes obligatoires  
CCI <= -100  // SURVENTE pour signal LONG
CCI >= 100   // SURACHAT pour signal SHORT
```

### **Algorithme Multi-Timeframe**

#### **Étape 1 - TF Principal (5m)**
```go
// Détection triple extrême
if stochExtremeLogique && mfiExtreme && cciExtreme {
    confiance = 100%
    // Passer à confirmation TF supérieur
}
```

#### **Étape 2 - TF Context (1h) - Détermination du SENS**
```go
// RÈGLE STRICTE : Utiliser les 2 DERNIÈRES bougies 1h 100% FERMÉES
// JAMAIS la bougie en cours (données incomplètes et non fiables)

// Exemple : à 06:45
// - Bougie en cours [06:00-07:00] → INTERDITE ❌
// - Bougie CURRENT [05:00-06:00] → Dernière fermée ✅
// - Bougie PREV [04:00-05:00] → Avant-dernière fermée ✅

indexCurrent = dernière bougie 1h 100% fermée
indexPrev = avant-dernière bougie 1h 100% fermée

// SENS DU CONTEXT : Vérifier la VARIATION de CHAQUE indicateur

if signalType == "LONG" {
    // Pour LONG, context doit être HAUSSIER (tous croissants)
    stochKBullish := stochK1h[indexCurrent] > stochK1h[indexPrev]  // K CROISSANT ↗️
    mfiBullish := mfi1h[indexCurrent] > mfi1h[indexPrev]          // MFI CROISSANT ↗️
    cciBullish := cci1h[indexCurrent] > cci1h[indexPrev]          // CCI CROISSANT ↗️
    
    if stochKBullish && mfiBullish && cciBullish {
        // Context HAUSSIER validé → Chercher LONG en survente 5m ✅
    }
}

if signalType == "SHORT" {
    // Pour SHORT, context doit être BAISSIER (tous décroissants)
    stochKBearish := stochK1h[indexCurrent] < stochK1h[indexPrev]  // K DÉCROISSANT ↘️
    mfiBearish := mfi1h[indexCurrent] < mfi1h[indexPrev]          // MFI DÉCROISSANT ↘️
    cciBearish := cci1h[indexCurrent] < cci1h[indexPrev]          // CCI DÉCROISSANT ↘️
    
    if stochKBearish && mfiBearish && cciBearish {
        // Context BAISSIER validé → Chercher SHORT en surachat 5m ✅
    }
}

// LOGIQUE STRATÉGIQUE :
// - Signal LONG  : Context 1h haussier + Exécution 5m survente = Acheter le pullback
// - Signal SHORT : Context 1h baissier + Exécution 5m surachat = Vendre le rebond
```

#### **Étape 3 - Tableau de Validation Multi-TF**
```go
// Affichage avec double validation
┌──────┬────────┬────────┬──────────┬──────────┬─────────────────┬──────────┬──────────┐
│ Index│ Heure  │ Type   │    %K    │    %D    │   Conditions    │ Validé 5m│ Validé 1h│
├──────┼────────┼────────┼──────────┼──────────┼─────────────────┼──────────┼──────────┤
│   36 │ 05:05  │ LONG   │   9.440  │   8.967  │ 7.284≤8.967→... │ ✅        │ ✅        │
│   47 │ 06:00  │ LONG   │   8.369  │   7.168  │ 5.178≤7.168→... │ ✅        │ ✅        │
└──────┴────────┴────────┴──────────┴──────────┴─────────────────┴──────────┴──────────┘

// Seuls les signaux ✅✅ sont exécutés (double validation)
```

### **Tableau de signaux Multi-Timeframe**
```
🎯 SIGNAUX MULTI-TIMEFRAME EXTREMES:
┌──────┬────────┬────────┬──────────┬──────────┬──────────┬──────────┬───────────┬──────────┐
│ Index│ Heure  │ Type   │    %K    │    %D    │   MFI    │   CCI    │  Confiance│ TF Conf  │
├──────┼────────┼────────┼──────────┼──────────┼──────────┼──────────┼───────────┼──────────┤
│   35 │ 04:40  │ LONG   │   6.981  │   6.073  │   12.4   │  -191.0  │    100%   │   15m ✓  │
│   51 │ 06:00  │ LONG   │   8.369  │   7.168  │   14.7   │  -102.1  │    100%   │   15m ✓  │
└──────┴────────┴────────┴──────────┴──────────┴──────────┴──────────┴───────────┴──────────┘
```

### **Avantages stratégiques**
- **Psychologique** : Acheter quand tout le monde vend
- **Probabilité** : 3 indicateurs + 3 timeframes = confiance maximale
- **Risk/Reward** : Entrées contraires aux extrêmes émotionnels
- **Filtrage** : Élimine les faux signaux par validation multi-TF

### **Paramètres configurables**
```go
type MultiTimeframeConfig struct {
    // Timeframes
    PrincipalTF    string  // "5m"
    ConfirmationTF string  // "15m" 
    ContexteTF     string  // "1h"
    
    // Seuils extrêmes
    StochSurachat  float64 // 80
    StochSurvente  float64 // 20
    MFISurachat    float64 // 80
    MFISurvente    float64 // 20
    CCISurachat    float64 // 100
    CCISurvente    float64 // -100
}
```

### **Exemple de signal parfait**
```
🎯 SIGNAL MULTI-TIMEFRAME LONG à 04:40 - Prix: 157.23 - Confiance: 100%
┌── TF Principal (5m) ────────────────────────┐
│ Stoch: 6.98→6.07 (LONG en SURVENTE) ✅      │
│ MFI: 12.4 (SURVENTE) ✅                     │  
│ CCI: -191.0 (SURVENTE) ✅                   │
└─────────────────────────────────────────────┘
┌── TF Confirmation (15m) ─────────────────────┐
│ Tendance: Baissière modérée ✅               │
│ Volume: Confirmé ✅                         │
└─────────────────────────────────────────────┘
┌── TF Contexte (1h) ──────────────────────────┐
│ Tendance: Neutre ✅                         │
│ Volatilité: normale ✅                      │
└─────────────────────────────────────────────┘
```

---

## 💡 **CONCLUSION**

Cette application démo fournira :

✅ **Détection précise** des signaux scalping et intraday  
✅ **Analyse volume conditionné** avec extension automatique  
✅ **Interface claire** avec tableaux de signaux détaillés  
✅ **Configuration flexible** des paramètres indicateurs  
✅ **Base extensible** pour futures stratégies  

**Prêt pour implémentation et tests sur données Binance Futures !**
