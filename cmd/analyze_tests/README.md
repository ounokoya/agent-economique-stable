# 📊 Analyze Tests - Outil d'Analyse Comparative

## Objectif

Analyser et comparer les performances de multiples configurations de la stratégie Direction en lisant les exports JSON générés par `direction_generator_demo`.

## Utilisation

```bash
# Analyser tous les tests dans le dossier out/
go run cmd/analyze_tests/main.go out

# Analyser un dossier spécifique
go run cmd/analyze_tests/main.go path/to/test/results
```

## Format d'entrée

L'outil lit les fichiers `intervalles.json` dans chaque sous-dossier de `out/`. 

**Structure attendue**:
```
out/
├── direction_demo_5m_vwma20_slope6_k2_atr8_coef0.25/
│   ├── klines.json
│   └── intervalles.json
├── direction_demo_5m_vwma12_slope2_k2_atr4_coef0.50/
│   ├── klines.json
│   └── intervalles.json
└── ...
```

**Format `intervalles.json`**:
```json
[
  {
    "Numero": 1,
    "Type": "LONG",
    "DateDebut": "2025-11-06T19:30:00Z",
    "DateFin": "2025-11-06T20:00:00Z",
    "PrixDebut": 156.89,
    "PrixFin": 157.32,
    "NbBougies": 7,
    "VariationCaptee": 0.274
  }
]
```

## Output

### 1. Tableau de classement

Toutes les configurations triées par **TOTAL CAPTÉ** décroissant:

```
Rank     | TF    | VWMA | Slp | ATR | Coef       | #Int | Long%    | Short%   | CAPTÉ%     | AvgBougie_L | AvgBougie_S
─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
🥇#1     | 5m    |   20 |   6 |   8 |       0.25 |   12 |    +4.60 |    -1.43 |      +6.03 |         35.5 |        40.0
🥈#2     | 5m    |   12 |   2 |   4 |       0.50 |   10 |    +4.89 |    -1.08 |      +5.98 |         43.8 |        48.4
```

**Colonnes**:
- **Rank**: Position avec médailles pour top 3 et étoiles pour top 10
- **TF**: Timeframe
- **VWMA**: Période VWMA
- **Slp**: Période de calcul de pente
- **ATR**: Période ATR
- **Coef**: Coefficient ATR
- **#Int**: Nombre d'intervalles (trades)
- **Long%**: Variation captée LONG cumulée
- **Short%**: Variation captée SHORT cumulée
- **CAPTÉ%**: **TOTAL CAPTÉ** = Long% - Short%
- **AvgBougie_L/S**: Nombre moyen de bougies par intervalle

### 2. Analyse par catégories

#### 🎯 Par VWMA Period
Moyenne, meilleure et pire performance pour chaque valeur de VWMA:
```
VWMA       | Tests  | Avg Capté    | Best         | Worst       
──────────────────────────────────────────────────────────────────────
20         |      5 |        +3.44% |        +6.03% |        -1.59%
12         |     11 |        +1.40% |        +5.98% |        -4.72%
```

#### ⚡ Par ATR Coefficient
Même analyse pour les coefficients ATR:
```
Coeff      | Tests  | Avg Capté    | Best         | Worst       
──────────────────────────────────────────────────────────────────────
0.40       |      1 |        +4.10% |        +4.10% |        +4.10%
0.80       |      7 |        +1.45% |        +4.90% |        -4.43%
```

#### ⏱️ Par durée d'intervalle
Catégorisation automatique:
- **Court terme**: <20 bougies
- **Moyen terme**: 20-50 bougies
- **Long terme**: >50 bougies

```
📍 COURT TERME (<20 bougies) (9 tests):
   • Moyenne capté: -4.92%
   • Meilleur: +0.29% (VWMA=6, ATR_coef=0.25, avg_bougie=12.9)
   • Pire: -15.67% (VWMA=3, ATR_coef=0.25, avg_bougie=6.8)
```

### 3. Recommandations stratégiques

Recommandations automatiques par horizon de trading:

```
🎯 COURT TERME (Scalping, <20 bougies = <2h en 5m):
   • Meilleure config: VWMA=6, Slope=3, ATR=6, Coef=0.25
   • Performance: +0.29% capté
   • Intervalles: 39 (avg 12.9 bougies)
   • Interprétation: VWMA court = réactivité élevée, ATR_coef faible = moins de bruit
```

## Métriques calculées

### TOTAL CAPTÉ
Formule: `LONG - SHORT`

**Pourquoi cette formule ?**
- Les variations LONG profitables sont positives
- Les variations SHORT profitables sont négatives
- Pour obtenir le total bidirectionnel, on soustrait SHORT (ce qui équivaut à l'additionner en valeur absolue)

**Exemple**:
- LONG: +4.60%
- SHORT: -1.43%
- TOTAL CAPTÉ: 4.60 - (-1.43) = **+6.03%**

### Durée moyenne d'intervalle
Moyenne pondérée du nombre de bougies par intervalle LONG et SHORT:
```
avg_duration = (avg_bougies_long + avg_bougies_short) / 2
```

Utilisée pour la catégorisation court/moyen/long terme.

## Parsing du nom de dossier

L'outil extrait automatiquement les paramètres du nom:
```
direction_demo_5m_vwma20_slope6_k2_atr8_coef0.25
               │   │      │      │  │    │
               │   │      │      │  │    └─ ATR Coefficient
               │   │      │      │  └────── ATR Period
               │   │      │      └───────── (K ignoré pour l'instant)
               │   │      └──────────────── Slope Period
               │   └─────────────────────── VWMA Period
               └─────────────────────────── Timeframe
```

**Regex**: `direction_demo_(\w+)_vwma(\d+)_slope(\d+)_k(\d+)_atr(\d+)_coef([\d.]+)`

## Cas d'usage

### 1. Comparer rapidement toutes les configs
```bash
go run cmd/analyze_tests/main.go out | head -50
```
→ Voir le top 10 directement

### 2. Identifier les patterns
```bash
go run cmd/analyze_tests/main.go out > analysis_results.txt
```
→ Chercher "MEILLEURS PAR VWMA" pour comprendre l'impact de chaque paramètre

### 3. Trouver la config optimale pour un horizon
```bash
go run cmd/analyze_tests/main.go out | grep -A 5 "MOYEN TERME"
```
→ Recommandations spécifiques court/moyen/long terme

### 4. Valider une hypothèse
```bash
# "Est-ce que VWMA=20 performe mieux que VWMA=3 ?"
go run cmd/analyze_tests/main.go out | grep "MEILLEURS PAR VWMA" -A 10
```

## Structure du code

```
main.go
├── Types
│   ├── Intervalle        # Structure d'un intervalle (trade)
│   └── TestResult        # Résultat d'un test de config
├── main()                # Orchestration
├── parseDirectoryName()  # Extraction paramètres depuis nom
├── Analysis Functions
│   ├── analyzeByVWMA()
│   ├── analyzeByATRCoeff()
│   └── analyzeByCandleDuration()
└── recommandations()     # Suggestions par horizon
```

## Évolutions possibles

- [ ] Export CSV pour analyse dans Excel/Python
- [ ] Graphiques avec gnuplot ou plotly
- [ ] Filtrage par timeframe (quand 1m disponible)
- [ ] Analyse de corrélation entre paramètres
- [ ] Calcul du ratio Sharpe/drawdown
- [ ] Détection des outliers statistiques
- [ ] Comparaison avant/après période (train/test split)

## Exemples de questions répondues

**Q: Quelle est la meilleure config pour le moyen terme en 5m ?**
```bash
go run cmd/analyze_tests/main.go out | grep -A 4 "MOYEN TERME (Intraday"
```
→ VWMA=20, Slope=6, ATR=8, Coef=0.25 (+6.03%)

**Q: VWMA=3 est-il viable ?**
```bash
go run cmd/analyze_tests/main.go out | grep "VWMA       | Tests" -A 10
```
→ Non, -11.35% en moyenne

**Q: Quel ATR_coef choisir ?**
```bash
go run cmd/analyze_tests/main.go out | grep "Coeff      | Tests" -A 10
```
→ 0.40-0.80 est optimal (+1.45% à +4.10%)

**Q: Combien de tests sont positifs ?**
```bash
go run cmd/analyze_tests/main.go out | grep "^⭐\|^🥇" | wc -l
```
→ Top 10 = configs à considérer

## Limitations

- Nécessite que `direction_generator_demo` ait déjà généré les exports JSON
- Ne supporte que le format de nommage spécifique `direction_demo_*`
- Assume que tous les tests utilisent la même période (500 bougies)
- N'analyse pas les klines individuelles (seulement les intervalles agrégés)
- Calculs basés sur prix de clôture uniquement (pas de simulation spread/fees)

## Voir aussi

- `cmd/direction_generator_demo/` : Générateur de tests
- `cmd/verify_direction_calculs/` : Vérificateur de calculs
- `docs/ANALYSE_PARAMETRES_DIRECTION.md` : Rapport d'analyse détaillé
- `docs/RESUME_ANALYSE_DIRECTION.md` : Résumé exécutif

---

**Auteur**: Agent Économique Stable  
**Date**: 2025-11-08  
**Version**: 1.0
