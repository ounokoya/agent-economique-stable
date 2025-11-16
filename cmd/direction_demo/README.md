# Direction Demo - Grille Adaptative VWMA6

## 📋 Objectif

Cette démo implémente une **grille de trading adaptative** basée sur l'indicateur VWMA6. Au lieu d'utiliser des niveaux de prix fixes, elle détecte automatiquement les **vagues naturelles du marché** pour identifier les phases d'achat (LONG) et de vente (SHORT).

**Concept** : Grid Trading Intelligent basé sur le momentum prix-volume.

---

## 🎯 Philosophie

### Grille Classique (Prix Fixe)
```
Vendre à : 165, 170, 175, 180...
Acheter à : 160, 155, 150, 145...
```
❌ Rigide
❌ Ne s'adapte pas
❌ Peut rater des mouvements

### Grille VWMA6 (Dynamique)
```
Intervalle ↗ : ACHETER (LONG)
Intervalle ↘ : VENDRE (SHORT)
```
✅ S'adapte au marché
✅ Suit les vagues naturelles
✅ Filtre le bruit automatiquement

---

## 🧮 Indicateur : VWMA6

### **VWMA (Volume Weighted Moving Average)**
Moyenne mobile pondérée par le volume sur **3 périodes** (configurable).

**Pourquoi VWMA et pas SMA ?**
- Intègre le **volume** : Les mouvements avec fort volume ont plus de poids
- Plus réactif aux vraies tendances
- Moins sensible aux faux mouvements à faible volume

### **Calcul de la variation**
```
Variation% = (VWMA6[i] - VWMA6[i-PERIODE_PENTE]) / VWMA6[i-PERIODE_PENTE] × 100
```

### **Détection du sens**
```
Si Variation% > +0.10% → ↗ CROISSANT (LONG)
Si Variation% < -0.10% → ↘ DÉCROISSANT (SHORT)
Sinon                   → → STABLE (ignoré)
```

**Unité** : Pourcentage (relatif au prix) pour s'adapter à tous les niveaux de prix.

---

## 🧩 Regroupement en Intervalles

### **Règle 1 : STABLE n'interrompt PAS**
```
↗ → ↗ → ↗ = UN seul intervalle CROISSANT
```
Les périodes **STABLE** (→) sont absorbées dans l'intervalle en cours.

### **Règle 2 : K-Confirmation (Anti-Bruit)**
Un changement de direction (↗→↘ ou ↘→↗) doit se **confirmer pendant K bougies** :

```
Avec K = 2 :
↗↗↗ ↘ ↗ → Intervalle ↗ continue (↘ rejeté, durée 1 < K)
↗↗↗ ↘↘ ↗ → Intervalle ↗ SE FERME (↘ confirmé sur 2 bougies)
              Nouvel intervalle ↘ commence
```

**But** : Éviter les faux changements de direction causés par le bruit du marché.

### **Résultat**
```
Intervalle #15 ↘ : 86 bougies (17.4%) - VENDRE
Intervalle #6 ↗  : 47 bougies (9.5%)  - ACHETER
Intervalle #18 ↗ : 47 bougies (9.5%)  - ACHETER
```

Chaque intervalle = **une vague complète** du marché.

---

## ⚙️ Paramètres Configurables

```go
// Symbole et données
SYMBOL     = "SOL_USDT"  // Format Gate.io
TIMEFRAME  = "1m"        // 1m, 5m, 15m, 30m, 1h...
NB_CANDLES = 500

// Période VWMA
VWMA_RAPIDE = 3          // Nombre de bougies pour VWMA

// Calibrage pentes
PERIODE_PENTE    = 2     // Nombre de bougies pour calculer la variation
SEUIL_PENTE_VWMA = 0.10  // Variation minimale en % (0.10% = 0.001)
K_CONFIRMATION   = 2     // Nombre de bougies pour confirmer changement
```

### **Calibrage selon timeframe**

| Timeframe | VWMA_RAPIDE | PERIODE_PENTE | SEUIL_PENTE_VWMA | Commentaire |
|-----------|-------------|---------------|------------------|-------------|
| **1m**    | 3           | 2             | 0.10%            | ✅ Scalping rapide, très réactif |
| **5m**    | 6           | 3             | 0.15%            | Moyen terme, filtre plus de bruit |
| **15m**   | 6           | 3             | 0.20%            | Tendances plus longues |
| **1h**    | 6           | 4             | 0.25%            | Position trading |

---

## 🚀 Exécution

```bash
cd /root/projects/trading_space/windsurf_space/harmonie_60_space/agent_economique_stable
go run cmd/direction_demo/main.go
```

---

## 📊 Sorties

### **1. Tableau de calibrage (30 dernières bougies)**
```
Date/Heure          | VWMA6      | Var VWMA%    | Sens VWMA6
------------------------------------------------------------
2025-11-08 08:24:00 |     160.72 |       +0.11% | ↗
2025-11-08 08:25:00 |     160.88 |       +0.14% | ↗
2025-11-08 08:26:00 |     161.01 |       +0.18% | ↗
```

**Usage** : Vérifier si les seuils sont bien calibrés. Si trop de → STABLE, baisser `SEUIL_PENTE_VWMA`.

### **2. Intervalles VWMA6**
```
#    | %      | Sens           | Date Début          | Date Fin            | Bougies  | VWMA6 Moy
-----|--------|----------------|---------------------|---------------------|----------|----------
6    |  9.5%  | ↗ CROISSANT    | 2025-11-08 01:59:00 | 2025-11-08 02:45:00 |       47 |    162.97
15   | 17.4%  | ↘ DÉCROISSANT  | 2025-11-08 04:50:00 | 2025-11-08 06:15:00 |       86 |    161.73
18   |  9.5%  | ↗ CROISSANT    | 2025-11-08 07:29:00 | 2025-11-08 08:15:00 |       47 |    160.67
```

**Colonne %** : Pourcentage du temps capté par cet intervalle.

### **3. Statistiques**
```
INTERVALLES VWMA6:
  Total intervalles    : 20
  - Croissant (↗)      : 10 intervalles (246 bougies, 49.9%)
  - Décroissant (↘)    : 10 intervalles (247 bougies, 50.1%)
```

**Équilibre** : Un marché équilibré aura ~50/50. Un marché tendanciel sera déséquilibré (ex: 70/30).

---

## 📈 Interprétation des Résultats

### **Exemple : SOL/USDT 1m (500 bougies)**

```
Total intervalles    : 20
  - Croissant (↗)    : 10 intervalles (246 bougies, 49.9%)
  - Décroissant (↘)  : 10 intervalles (247 bougies, 50.1%)

Plus long intervalle : #15 ↘ (86 bougies, 17.4%)
```

**Analyse** :
- ✅ Marché **équilibré** (50/50)
- ✅ **20 intervalles** = vagues bien définies (pas trop fragmenté)
- ✅ **Intervalle max = 86 bougies** = vraies tendances détectées
- 💡 Idéal pour du **scalping bidirectionnel** (LONG et SHORT)

### **Marché Tendanciel (exemple)**
```
Total intervalles    : 8
  - Croissant (↗)    : 2 intervalles (120 bougies, 25%)
  - Décroissant (↘)  : 6 intervalles (360 bougies, 75%)
```
👉 **Marché baissier fort** : Privilégier les positions SHORT !

---

## 💡 Stratégies d'Utilisation

### **1. Grid Trading Adaptatif**
```
Entrée LONG  : Début d'intervalle ↗
Sortie LONG  : Fin d'intervalle ↗ (↘ confirmé sur K bougies)

Entrée SHORT : Début d'intervalle ↘
Sortie SHORT : Fin d'intervalle ↘ (↗ confirmé sur K bougies)
```

### **2. Filtre de Direction**
Combiner avec d'autres signaux :
```
Signal achat + Intervalle ↗ = ✅ ENTRER
Signal achat + Intervalle ↘ = ❌ ÉVITER
```

### **3. Position Sizing Adaptatif**
```
Intervalle court (< 20 bougies) = Position petite (risque élevé)
Intervalle long (> 50 bougies)  = Position grande (tendance forte)
```

### **4. Stop Loss Dynamique**
```
LONG : Placer SL sous le début de l'intervalle ↗
SHORT : Placer SL au-dessus du début de l'intervalle ↘
```

---

## 🔍 Notes Techniques

### **Pourquoi la variation est en % ?**
```go
// ❌ Absolu (dépend du prix)
variation = VWMA6[i] - VWMA6[i-2]  // 0.5 à 50$ ≠ 0.5 à 150$

// ✅ Relatif (indépendant du prix)
variation = (VWMA6[i] - VWMA6[i-2]) / VWMA6[i-2] * 100  // 0.3% partout
```

Le **pourcentage** permet d'utiliser les mêmes seuils quel que soit le niveau de prix !

### **Pourquoi K-Confirmation ?**
Sans K-confirmation :
```
↗↗↗ ↘ ↗↗↗ = 3 intervalles (bruit)
```

Avec K=2 :
```
↗↗↗ ↘ ↗↗↗ = 1 intervalle ↗ (↘ rejeté)
```

Réduit le **sur-trading** et améliore le **ratio signal/bruit** !

---

## 🎯 Avantages vs Approches Classiques

| Aspect | Grille Fixe | MA Crossover | **Grille VWMA6** |
|--------|-------------|--------------|------------------|
| **Adaptation** | ❌ Statique | ⚠️ Lente | ✅ Temps réel |
| **Volume** | ❌ Ignoré | ❌ Ignoré | ✅ Intégré |
| **Bruit** | ❌ Aucun filtre | ⚠️ Lag | ✅ K-Confirmation |
| **Simplicité** | ✅ Simple | ✅ Simple | ✅ Simple |
| **Backtest** | ✅ Facile | ✅ Facile | ✅ Facile |

---

## 🚦 Prochaines Étapes

1. **Backtesting** : Tester sur historique complet
2. **Optimisation** : Trouver meilleurs paramètres par paire/timeframe
3. **Entry/Exit** : Ajouter signaux précis d'entrée/sortie dans les intervalles
4. **Risk Management** : Calculer taille de position selon durée intervalle
5. **Multi-Timeframe** : Combiner 1m + 5m + 15m pour confirmation

---

## 📚 Différence avec `trend_demo`

| Aspect | `trend_demo` | `direction_demo` |
|--------|--------------|------------------|
| **Méthode** | Croisements VWMA + DMI | Intervalles directionnels VWMA |
| **Indicateurs** | VWMA6 + VWMA24 + DMI | VWMA3/6 uniquement |
| **Output** | Signaux ponctuels | Phases continues |
| **Usage** | Entrées/sorties précises | Contexte de marché |
| **Complexité** | Élevée | Faible |

**Complémentarité** :
- `direction_demo` → Identifier les vagues (contexte)
- `trend_demo` → Entrer/sortir dans les vagues (timing)

---

## 📊 Exemple de Résultat (1m SOL/USDT)

```
=== DEMO ANALYSE DIRECTIONNELLE (VWMA6 uniquement) ===

Configuration:
Symbole            : SOL_USDT
Timeframe          : 1m
VWMA rapide        : 3
Seuil pente VWMA6  : 0.10%
K confirmation     : 2 bougies

INTERVALLES VWMA6:
  Total intervalles    : 20
  - Croissant (↗)      : 10 intervalles (246 bougies, 49.9%)
  - Décroissant (↘)    : 10 intervalles (247 bougies, 50.1%)

Top 3 intervalles:
  #15 ↘ : 86 bougies (17.4%) - 04:50→06:15
  #6  ↗ : 47 bougies (9.5%)  - 01:59→02:45
  #18 ↗ : 47 bougies (9.5%)  - 07:29→08:15
```

**Interprétation** :
- Marché équilibré LONG/SHORT
- 20 vagues bien définies sur 8h20
- Plus longue vague : 86 minutes (SHORT)
- ✅ **Parfait pour scalping bidirectionnel** !

---

## 🏁 Conclusion

`direction_demo` implémente une **grille de trading adaptative** qui s'ajuste automatiquement au marché en détectant les vagues naturelles via VWMA6.

**Simple, robuste, et exploitable !** 🚀
