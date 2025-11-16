# Vérificateur de Calculs - Direction Generator Demo

## 🎯 Objectif

Script de vérification qui recalcule toutes les variations de prix à partir des fichiers JSON exportés par `direction_generator_demo` et compare avec les résultats affichés.

## 📊 Vérifications effectuées

1. **Variation par intervalle** : Recalcule `(PrixFin - PrixDébut) / PrixDébut × 100`
2. **Totaux LONG/SHORT** : Somme des variations par type
3. **Total capté** : Vérifie la formule `variationLong - variationShort`

## 🚀 Utilisation

### 1. Générer les données

```bash
cd cmd/direction_generator_demo
go run main.go
```

Cela créera un dossier dans `out/direction_demo_<timeframe>_<params>/` avec :
- `klines.json` : Données klines brutes
- `intervalles.json` : Intervalles détectés avec variations

### 2. Vérifier les calculs

```bash
cd cmd/verify_direction_calculs
go run main.go ../../out/direction_demo_5m_vwma6_slope4_k3_atr4_coef0.50
```

## 📋 Output

```
═══════════════════════════════════════════════════════════════════
  VÉRIFICATION DES CALCULS - Direction Generator Demo
═══════════════════════════════════════════════════════════════════

📂 Dossier: out/direction_demo_5m_vwma6_slope4_k3_atr4_coef0.50
✅ Klines chargées: 501
✅ Intervalles chargés: 30

──────────────────────────────────────────────────────────────────
VÉRIFICATION INTERVALLE PAR INTERVALLE
──────────────────────────────────────────────────────────────────
✅ #1  | SHORT | 156.31 → 157.03 | Variation: +0.46% (recalc: +0.46%, diff: 0.0000%)
✅ #2  | LONG  | 157.03 → 155.92 | Variation: -0.71% (recalc: -0.71%, diff: 0.0000%)
...

──────────────────────────────────────────────────────────────────
VÉRIFICATION DES TOTAUX
──────────────────────────────────────────────────────────────────

📈 LONG:
   • Intervalles: 15
   • Variation démo:      -1.33%
   • Variation recalc:    -1.33%
   • Différence:          0.0000%

📉 SHORT:
   • Intervalles: 15
   • Variation démo:      +4.59%
   • Variation recalc:    +4.59%
   • Différence:          0.0000%

💰 TOTAL CAPTÉ (bidirectionnel):
   • Démo:       -5.93%
   • Recalculé:  -5.93%
   • Différence: 0.0000%

═══════════════════════════════════════════════════════════════════
VERDICT
═══════════════════════════════════════════════════════════════════
✅ TOUS LES CALCULS SONT CORRECTS
   • Intervalles vérifiés: 30/30
   • Erreurs détectées: 0
   • Précision: < 0.01%
═══════════════════════════════════════════════════════════════════
```

## 🔍 Détection d'erreurs

Le script détecte :
- ❌ Variations individuelles incorrectes (tolérance 0.01%)
- ❌ Totaux LONG/SHORT incorrects
- ❌ Formule "Total capté" incorrecte

## 💡 Pourquoi cette vérification ?

Garantit que :
1. Les prix sont correctement extraits des klines
2. Les calculs de variations sont exacts
3. La formule bidirectionnelle `LONG - SHORT` est bien appliquée
4. Aucune erreur d'arrondi significative
