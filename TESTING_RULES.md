# 📋 RÈGLES DE TESTING UNIFIÉES

## 🎯 RÈGLE FONDAMENTALE
**TOUS les tests d'indicateurs doivent utiliser exactement 300 klines**

### Pourquoi 300 klines ?
- ✅ **Précision suffisante** pour calculs fiables
- ✅ **Performance optimale** pour temps de test rapide
- ✅ **Standard unifié** pour comparaison entre indicateurs
- ✅ **Stabilité** des résultats (pas de variations aléatoires)

---

## 📊 CONFIGURATION STANDARD

### Paramètres par défaut pour tous les tests :
```go
// Nombre de klines
KLINE_COUNT = 300

// Timeframe par défaut
TIMEFRAME = "5m"

// Symbole par défaut
SYMBOL = "SOL_USDT"

// Exchange par défaut
EXCHANGE = "Gate.io"
```

### Paramètres indicateurs standards :
```go
// DMI
DMI_PERIOD = 14

// MACD  
MACD_FAST = 12
MACD_SLOW = 26
MACD_SIGNAL = 9

// Stochastic
STOCH_K = 14
STOCH_SMOOTH_K = 3
STOCH_D = 3

// CCI
CCI_PERIOD = 20

// MFI
MFI_PERIOD = 14
```

---

## 🔧 IMPLÉMENTATION MODÈLE

### Structure de test standard :
```go
func main() {
    fmt.Println("🎯 INDICATEUR - TEST STANDARD")
    fmt.Println("=" + strings.Repeat("=", 45))

    // 1. Configuration client
    client := gateio.NewClient()
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // 2. Récupération 300 klines (OBLIGATOIRE)
    fmt.Println("📡 Récupération des 300 dernières klines depuis Gate.io...")
    klines, err := client.GetKlines(ctx, "SOL_USDT", "5m", 300)
    
    // 3. Tri chronologique (OBLIGATOIRE)
    for i := 0; i < len(klines); i++ {
        for j := i + 1; j < len(klines); j++ {
            if klines[j].OpenTime.Before(klines[i].OpenTime) {
                klines[i], klines[j] = klines[j], klines[i]
            }
        }
    }
    
    // 4. Calcul indicateur
    // 5. Affichage 15 dernières valeurs
    // 6. Analyse et statistiques
}
```

---

## 📈 FORMAT D'AFFICHAGE STANDARD

### Tableau de résultats (15 dernières valeurs) :
```
TIME         CLOSE      INDICATEUR1  INDICATEUR2  SIGNAL
---------------------------------------------------------
13:00        175.57     valeur1      valeur2      SIGNAL
13:05        175.50     valeur1      valeur2      SIGNAL
...
```

### Analyse obligatoire :
- ✅ Dernière valeur complète
- ✅ Statistiques sur 15 dernières valeurs
- ✅ Valeurs extrêmes (min/max)
- ✅ Configuration actuelle
- ✅ Recommandations de trading

---

## ⚠️ CONTRAINTES OBLIGATOIRES

### À respecter pour TOUS les tests :
1. **300 klines exactement** - ni plus, ni moins
2. **Tri chronologique** - obligatoire avant calculs
3. **Gate.io comme source** - pour cohérence
4. **Timeframe 5m** - standard pour réactivité
5. **Affichage 15 dernières** - pour lisibilité
6. **Format unifié** - pour comparaison

### Interdictions :
- ❌ Utiliser moins de 300 klines
- ❌ Utiliser plus de 300 klines  
- ❌ Changer de timeframe sans justification
- ❌ Omettre le tri chronologique
- ❌ Utiliser des formats d'affichage différents

---

## 🎯 VALIDATION AUTOMATIQUE

### Checklist de validation :
- [ ] Nombre de klines = 300
- [ ] Tri chronologique effectué
- [ ] Paramètres standards respectés
- [ ] Format d'affichage unifié
- [ ] Analyse complète présente
- [ ] Recommandations incluses

---

## 📁 FICHIERS DE RÉFÉRENCE

### Tests validés respectant les règles :
- ✅ `cci_gateio_application.go`
- ✅ `dmi_gateio_application.go`  
- ✅ `macd_gateio_application.go`
- ✅ `stoch_gateio_application.go`

### Tests à corriger :
- ❌ `dmi_rma_precision_comparison.go` (utilise 100 klines)

---

*Document créé le 03/11/2025 - Règles unifiées pour tous les tests d'indicateurs*
