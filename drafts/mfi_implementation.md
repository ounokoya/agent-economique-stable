# MFI (Money Flow Index) - Implémentation Précision Chirurgicale

## 📊 FORMULE OFFICIELLE (TradingView/MetaTrader)

### Étapes de calcul exactes :

**1. Typical Price (TP)**
```
TP = (High + Low + Close) / 3
```

**2. Raw Money Flow**
```
Raw MF = TP × Volume
```

**3. Classification du Money Flow**
- Si `TP(actuel) > TP(précédent)` → **Positive Money Flow**
- Si `TP(actuel) < TP(précédent)` → **Negative Money Flow** 
- Si `TP(actuel) = TP(précédent)` → **Ni positive ni negative** (ignoré)

**4. Sommes sur période N**
```
PMF = Sum(Positive Money Flow sur N périodes)
NMF = Sum(Negative Money Flow sur N périodes)
```

**5. Money Ratio**
```
MR = PMF / NMF
```
*Note : Si NMF = 0, alors MFI = 100*

**6. Money Flow Index final**
```
MFI = 100 - (100 / (1 + MR))
```

## 🎯 SPÉCIFICATIONS TECHNIQUES

### Sources de référence :
- ✅ **MetaTrader 4** - Documentation officielle
- ✅ **TradingView** - Implémentation standard
- ✅ **IFC Markets** - Formule détaillée

### Période standard :
- **14 périodes** (configurable)

### Plage de valeurs :
- **0 à 100**
- **> 80** : Surachat (overbought)  
- **< 20** : Survente (oversold)

## 🔧 PLAN D'IMPLÉMENTATION

### Phase 1 : Fonction de base
```go
func CalculateMFI(klines []Kline, period int) []float64
```

### Phase 2 : Tests unitaires
- Test avec données manuelles simples (5-10 bougies)
- Validation étape par étape
- Comparaison avec TradingView/MT4

### Phase 3 : Intégration
- Remplacement dans `MFIFromKlines()`
- Tests avec données BingX réelles
- Validation sur 500 klines

## 📝 POINTS CRITIQUES

### Gestion des cas limites :
1. **Première bougie** : Pas de MF (pas de TP précédent)
2. **NMF = 0** : MFI = 100 
3. **PMF = 0** : MFI = 0
4. **TP égaux** : Money Flow neutre (ignoré)

### Précision requise :
- **float64** pour tous les calculs
- **Pas d'arrondi intermédiaire**
- **Ordre des opérations respecté**

## 📚 RÉFÉRENCES TECHNIQUES ANALYSÉES

### Bibliothèques de référence :
1. **TA-Lib** (référence industrielle)
2. **pandas-ta** (twopirllc/pandas-ta sur GitHub)
3. **ta** (bukosabino/ta - Technical Analysis Library)
4. **MetaTrader 4/5** (documentation officielle)
5. **TradingView** (Pine Script)

### Formule consensus (toutes sources) :
```python
# Étape 1: Typical Price
tp = (high + low + close) / 3

# Étape 2: Money Flow brut
raw_mf = tp * volume

# Étape 3: Classification
# Comparer TP actuel vs TP précédent
if tp[i] > tp[i-1]:
    positive_mf += raw_mf[i]
elif tp[i] < tp[i-1]:
    negative_mf += raw_mf[i]
# Si tp[i] == tp[i-1] → neutre (ignoré)

# Étape 4: Sommes glissantes sur N périodes
pmf = sum(positive_money_flows[-period:])
nmf = sum(negative_money_flows[-period:])

# Étape 5: Calcul final MFI
if nmf == 0:
    mfi = 100.0
else:
    money_ratio = pmf / nmf
    mfi = 100.0 - (100.0 / (1.0 + money_ratio))
```

## 🧪 DONNÉES DE TEST

### Test Case 1 : Validation manuelle
```
Period = 3
Bougie 0: H=10, L=8,  C=9,  V=100 → TP=9.00
Bougie 1: H=11, L=9,  C=10, V=200 → TP=10.00 (>9.00) → +MF = 2000
Bougie 2: H=10, L=8,  C=9,  V=150 → TP=9.00  (<10.00) → -MF = 1350  
Bougie 3: H=12, L=10, C=11, V=300 → TP=11.00 (>9.00) → +MF = 3300

Sur période 3 (bougies 1-3):
PMF = 2000 + 0 + 3300 = 5300
NMF = 0 + 1350 + 0 = 1350
MR = 5300/1350 = 3.925925...
MFI = 100 - (100/(1+3.925925)) = 79.73
```

### Test Case 2 : Cas limites
```
- PMF = 0, NMF > 0 → MFI = 0
- PMF > 0, NMF = 0 → MFI = 100  
- TP identiques → Neutre (pas de MF)
```

### Validation attendue :
- **TA-Lib MFI(14)** : Référence absolue
- **pandas-ta mfi()** : Doit matcher TA-Lib
- **Notre implémentation Go** : Doit matcher pandas-ta

---
**Objectif** : Précision **EXACTE** vs TA-Lib (référence industrie)
