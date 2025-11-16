# 📊 Générateurs de Signaux Unifiés

Interface commune pour générer des signaux de trading avec différentes stratégies.

## 🎯 Architecture

```
internal/signals/
├── generator.go          # Interface commune + types Signal/Kline
├── direction/
│   └── generator.go     # Générateur Direction (VWMA6 + ATR)
└── trend/
    └── generator.go     # Générateur Trend (VWMA6↔24 + DMI)
```

---

## 🔌 Interface `Generator`

Tous les générateurs implémentent :

```go
type Generator interface {
    Name() string
    Initialize(config GeneratorConfig) error
    CalculateIndicators(klines []Kline) error
    DetectSignals(klines []Kline) ([]Signal, error)
    GetMetrics() GeneratorMetrics
}
```

---

## 📡 Type `Signal` Unifié

```go
type Signal struct {
    Timestamp  time.Time     // Timestamp du signal
    Action     SignalAction  // ENTRY ou EXIT
    Type       SignalType    // LONG ou SHORT
    Price      float64       // Prix au moment du signal
    Confidence float64       // 0.0 à 1.0
    Metadata   map[string]interface{} // Données spécifiques
    
    // Si Action == EXIT
    EntryPrice *float64
    EntryTime  *time.Time
}
```

**Actions :**
- `SignalActionEntry` : Signal d'ouverture de position
- `SignalActionExit` : Signal de fermeture de position

**Types :**
- `SignalTypeLong` : Position longue (achat)
- `SignalTypeShort` : Position courte (vente)

---

## 🎛️ Générateur DIRECTION

**Stratégie :** Intervalles directionnels basés sur VWMA6

### Configuration

```go
import "agent-economique/internal/signals/direction"

config := direction.Config{
    VWMAPeriod:          3,    // Période VWMA
    SlopePeriod:         2,    // Période calcul pente
    KConfirmation:       2,    // Bougies de confirmation
    UseDynamicThreshold: true, // Seuil dynamique (ATR)
    ATRPeriod:           14,   // Période ATR
    ATRCoefficient:      1.0,  // Coefficient ATR
    FixedThreshold:      0.5,  // Seuil fixe si mode statique
}

generator := direction.NewDirectionGenerator(config)
```

### Signaux Générés

**ENTRY (ouverture intervalle) :**
- Détecté quand VWMA6 change de direction (croissant ↔ décroissant)
- Type : `LONG` (croissant) ou `SHORT` (décroissant)
- Confiance : 0.7 (initiale)

**EXIT (fermeture intervalle) :**
- Détecté quand direction s'inverse
- Inclut `EntryPrice` et `EntryTime`
- Confiance : basée sur durée + variation captée
- Métadonnées : `duration_bars`, `variation_pct`

### Exemple de sortie

```
ENTRY  | LONG  | 160.50 | Conf: 0.70 | VWMA6=160.45
  → Position ouverte
EXIT   | LONG  | 163.20 | Conf: 0.85 | Duration: 47 bars, Variation: +1.68%
  → Position fermée
```

---

## 📈 Générateur TREND

**Stratégie :** Croisements VWMA6↔24 validés par DMI

### Configuration

```go
import "agent-economique/internal/signals/trend"

config := trend.Config{
    VwmaRapide:          6,    // VWMA rapide
    VwmaLent:            24,   // VWMA lent
    DmiPeriode:          14,   // Période DMI
    DmiSmooth:           3,    // Lissage DMI
    AtrPeriode:          30,   // Période ATR
    GammaGapVWMA:        0.5,  // 50% ATR pour gap VWMA
    GammaGapDI:          5.0,  // Gap minimal DI
    GammaGapDX:          5.0,  // Gap minimal DX/ADX
    VolatiliteMin:       0.3,  // ATR% minimal
    WindowGammaValidate: 5,    // Fenêtre validation gamma
    WindowW:             10,   // Fenêtre matching VWMA+DMI
}

generator := trend.NewTrendGenerator(config)
```

### Signaux Générés

**ENTRY (croisement VWMA+DMI) :**
- Détecté quand VWMA6↔24 + DI+↔DI- matchent (±10 barres)
- Type : `LONG` (croisement haussier) ou `SHORT` (baissier)
- Confiance : 0.75-0.95 (selon distance matching)
- Métadonnées : `motif`, `distance_bars`, indicateurs

**EXIT (croisement inverse VWMA) :**
- Détecté sur croisement inverse VWMA6↔24 **sans validation DMI**
- Pour position `LONG` : sortie sur croisement baissier
- Pour position `SHORT` : sortie sur croisement haussier
- Confiance : 0.8
- Métadonnées : `exit_reason`, `duration_bars`, `variation_pct`

### Exemple de sortie

```
ENTRY  | LONG  | 155.20 | Conf: 0.85 | VWMA→DMI (+2 bars)
  → Position ouverte
EXIT   | LONG  | 158.50 | Conf: 0.80 | VWMA inverse cross, Duration: 34 bars, +2.13%
  → Position fermée
```

---

## 🚀 Utilisation dans scalping_live_bybit

### Exemple d'intégration

```go
package main

import (
    "agent-economique/internal/signals"
    "agent-economique/internal/signals/direction"
    "agent-economique/internal/signals/trend"
)

// Choisir le générateur
func createGenerator(generatorType string) signals.Generator {
    switch generatorType {
    case "direction":
        config := direction.Config{
            VWMAPeriod:          3,
            SlopePeriod:         2,
            KConfirmation:       2,
            UseDynamicThreshold: true,
            ATRPeriod:           14,
            ATRCoefficient:      1.0,
        }
        return direction.NewDirectionGenerator(config)
        
    case "trend":
        config := trend.Config{
            VwmaRapide:          6,
            VwmaLent:            24,
            DmiPeriode:          14,
            DmiSmooth:           3,
            AtrPeriode:          30,
            GammaGapVWMA:        0.5,
            GammaGapDI:          5.0,
            GammaGapDX:          5.0,
            VolatiliteMin:       0.3,
            WindowGammaValidate: 5,
            WindowW:             10,
        }
        return trend.NewTrendGenerator(config)
        
    default:
        return nil
    }
}

// Dans processMarker
func (app *ScalpingLiveBybitApp) processMarker(klines []Kline) {
    // 1. Convertir vers format unifié
    unifiedKlines := make([]signals.Kline, len(klines))
    for i, k := range klines {
        unifiedKlines[i] = signals.Kline{
            OpenTime: k.OpenTime,
            Open:     k.Open,
            High:     k.High,
            Low:      k.Low,
            Close:    k.Close,
            Volume:   k.Volume,
        }
    }
    
    // 2. Calculer indicateurs
    if err := app.generator.CalculateIndicators(unifiedKlines); err != nil {
        log.Printf("Erreur calcul indicateurs: %v", err)
        return
    }
    
    // 3. Détecter signaux
    newSignals, err := app.generator.DetectSignals(unifiedKlines)
    if err != nil {
        log.Printf("Erreur détection signaux: %v", err)
        return
    }
    
    // 4. Traiter les signaux
    for _, sig := range newSignals {
        if sig.Action == signals.SignalActionEntry {
            fmt.Printf("🟢 ENTRY %s @ %.2f (conf: %.0f%%)\n", 
                sig.Type, sig.Price, sig.Confidence*100)
            // Ouvrir position
            
        } else if sig.Action == signals.SignalActionExit {
            variation := 0.0
            if sig.EntryPrice != nil {
                variation = (sig.Price - *sig.EntryPrice) / *sig.EntryPrice * 100
            }
            fmt.Printf("🔴 EXIT %s @ %.2f (conf: %.0f%%, var: %+.2f%%)\n",
                sig.Type, sig.Price, sig.Confidence*100, variation)
            // Fermer position
        }
    }
    
    // 5. Afficher métriques
    metrics := app.generator.GetMetrics()
    fmt.Printf("📊 Métriques: %d signaux (%d ENTRY, %d EXIT)\n",
        metrics.TotalSignals, metrics.EntrySignals, metrics.ExitSignals)
}
```

---

## 📊 Comparaison des Générateurs

| Aspect | DIRECTION | TREND |
|--------|-----------|-------|
| **Fréquence signaux** | Haute (26 intervalles/3j @ 5m) | Moyenne (32 signaux/3j @ 5m) |
| **Filtrage** | ATR uniquement | Multi-validation (VWMA+DMI) |
| **Signaux ENTRY** | ✅ Oui | ✅ Oui |
| **Signaux EXIT** | ✅ Oui (fin intervalle) | ✅ Oui (croisement inverse) |
| **Capture variation** | ✅ Complète (89% @ 5m) | ⚠️ Partielle (~70% @ 5m) |
| **Qualité signaux** | Moyenne (confiance 0.5-0.9) | Haute (confiance 0.75-0.95) |
| **Timeframe optimal** | 1m-5m | 5m-15m-30m |
| **Complexité** | Simple | Complexe |

---

## 🔧 Lancement avec CLI

```bash
# Direction sur 1m
go run cmd/scalping_live_bybit/*.go --generator=direction --timeframe=1m

# Trend sur 5m
go run cmd/scalping_live_bybit/*.go --generator=trend --timeframe=5m

# Scalping classique (défaut)
go run cmd/scalping_live_bybit/*.go
```

---

## ✅ Tests

Créer tests unitaires pour chaque générateur :

```bash
go test ./internal/signals/direction -v
go test ./internal/signals/trend -v
```

---

## 🎯 Avantages Architecture

1. **Interchangeabilité** : Changer de stratégie = 1 ligne de code
2. **Réutilisabilité** : Générateurs utilisables partout (paper/live/backtest)
3. **Testabilité** : Chaque générateur testable indépendamment
4. **Extensibilité** : Ajouter nouveau générateur = implémenter interface
5. **Uniformité** : Même signature pour tous les signaux
6. **Simplicité** : API claire et documentée

---

## 📝 TODO

- [ ] Tests unitaires direction/generator.go
- [ ] Tests unitaires trend/generator.go
- [ ] Intégration complète dans scalping_live_bybit
- [ ] Backtest avec les deux générateurs
- [ ] Documentation exemples avancés
- [ ] Générateur hybride (direction + trend)
