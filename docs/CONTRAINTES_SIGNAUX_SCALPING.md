# 📊 Contraintes de Génération de Signaux - Scalping Live Bybit

**Version** : 1.0  
**Application** : `scalping_live_bybit`  
**Stratégie** : SCALPING (Triple Extrême Synchronisé)  
**Timeframe** : 5 minutes  
**Symbole** : SOLUSDT (Bybit Linear Perpetual)

---

## 🎯 Vue d'ensemble

Un signal de trading est généré UNIQUEMENT si **TOUTES les 6 contraintes** suivantes sont validées dans l'ordre strict :

1. ✅ **Triple Extrême Flexible** (détection zone)
2. ✅ **Synchronisation des Mouvements** (même sens N-2 → N-1)
3. ✅ **Croisement Stochastique** (confirmation technique)
4. ✅ **Cohérence Directionnelle** (logique trading)
5. ✅ **Bougie de Validation** (confirmation visuelle)
6. ✅ **Volume Conditionné** (force du mouvement)

**❌ Si UNE SEULE contrainte échoue → Aucun signal généré**

---

## 1️⃣ TRIPLE EXTRÊME FLEXIBLE

### Principe
Les **3 indicateurs** (CCI, MFI, Stochastique) doivent **TOUS être en zone extrême**, mais **chacun peut l'être sur N-1 OU N-2** (bougie différente).

### Conditions SURACHAT (pour signal SHORT potentiel)

Chaque indicateur vérifié **indépendamment** sur N-1 OU N-2 :

| Indicateur | Condition Extrême | Bougie |
|------------|-------------------|--------|
| **CCI** | > 100 | N-1 OU N-2 |
| **MFI** | > 60 | N-1 OU N-2 |
| **Stochastique** | K ≥ 70 OU D ≥ 70 | N-1 OU N-2 |

**Validation** : Au moins une des 2 bougies (N-1 ou N-2) doit montrer l'indicateur en zone extrême.

### Conditions SURVENTE (pour signal LONG potentiel)

| Indicateur | Condition Extrême | Bougie |
|------------|-------------------|--------|
| **CCI** | < -100 | N-1 OU N-2 |
| **MFI** | < 40 | N-1 OU N-2 |
| **Stochastique** | K ≤ 30 OU D ≤ 30 | N-1 OU N-2 |

### Exemples

#### ✅ Exemple VALIDE (extrêmes flexibles)
```
Indicateur    N-2      N-1      Extrême détecté
CCI          -110     -105      N-2 (< -100) ✅
MFI            42       38      N-1 (< 40) ✅
Stoch K        32       28      N-1 (< 30) ✅
```
→ **Les 3 en SURVENTE** (même si sur bougies différentes)

#### ❌ Exemple INVALIDE (un indicateur manquant)
```
Indicateur    N-2      N-1      Extrême détecté
CCI          -110     -105      N-2 (< -100) ✅
MFI            45       42      Aucun (> 40) ❌
Stoch K        28       25      N-1 (< 30) ✅
```
→ **REJETÉ** : MFI pas en survente

### ❌ STOP si :
- Un des 3 indicateurs n'est en extrême ni sur N-1 ni sur N-2
- Les 3 ne pointent pas dans la même direction (SURACHAT vs SURVENTE)

---

## 2️⃣ SYNCHRONISATION DES MOUVEMENTS ⭐ **CRITIQUE**

### Principe
**TOUS les 3 indicateurs** doivent évoluer **dans le MÊME SENS** entre N-2 et N-1, selon le type de signal.

### Pour signal LONG (sortie de SURVENTE)

Les 3 indicateurs doivent **TOUS être en HAUSSE** :

```
CCI(N-1) > CCI(N-2)   ↗
MFI(N-1) > MFI(N-2)   ↗
Stoch(N-1) > Stoch(N-2)   ↗
```

**Logique** : Sortie progressive de la zone de survente (retournement haussier)

### Pour signal SHORT (sortie de SURACHAT)

Les 3 indicateurs doivent **TOUS être en BAISSE** :

```
CCI(N-1) < CCI(N-2)   ↘
MFI(N-1) < MFI(N-2)   ↘
Stoch(N-1) < Stoch(N-2)   ↘
```

**Logique** : Sortie progressive de la zone de surachat (retournement baissier)

### Exemples

#### ✅ Exemple VALIDE LONG (synchronisation parfaite)
```
Indicateur    N-2      N-1      Mouvement
CCI          -110     -105      Hausse ↗ ✅
MFI            38       42      Hausse ↗ ✅
Stoch K        28       32      Hausse ↗ ✅
```
→ **Les 3 montent ensemble** : sortie de survente confirmée

#### ✅ Exemple VALIDE SHORT (synchronisation parfaite)
```
Indicateur    N-2      N-1      Mouvement
CCI           115      110      Baisse ↘ ✅
MFI            64       61      Baisse ↘ ✅
Stoch K        73       71      Baisse ↘ ✅
```
→ **Les 3 descendent ensemble** : sortie de surachat confirmée

#### ❌ Exemple INVALIDE (divergence de mouvement)
```
Indicateur    N-2      N-1      Mouvement
CCI          -110     -105      Hausse ↗ ✅
MFI            42       38      Baisse ↘ ❌ DIVERGENCE !
Stoch K        28       32      Hausse ↗ ✅
```
→ **REJETÉ** : MFI ne synchronise pas avec CCI et Stoch

#### ❌ Exemple INVALIDE (mouvement plat)
```
Indicateur    N-2      N-1      Mouvement
CCI           115      110      Baisse ↘ ✅
MFI            62       62      Plat ➡ ❌ PAS DE MOUVEMENT !
Stoch K        73       71      Baisse ↘ ✅
```
→ **REJETÉ** : MFI stagnant (pas de mouvement clair)

### ❌ STOP si :
- **Un seul indicateur** évolue dans le sens inverse
- **Un seul indicateur** stagne (pas de mouvement)
- Mouvements non cohérents avec le type de signal attendu

---

## 3️⃣ CROISEMENT STOCHASTIQUE

### Principe
Détection d'un **croisement** entre %K et %D sur la transition **N-2 → N-1**.

### Croisement HAUSSIER (signal SHORT potentiel)

**K passe SOUS D** (croisement bearish) :

```
Sur N-2 : K > D
Sur N-1 : K < D
→ Signal SHORT
```

### Croisement BAISSIER (signal LONG potentiel)

**K passe AU-DESSUS de D** (croisement bullish) :

```
Sur N-2 : K < D
Sur N-1 : K > D
→ Signal LONG
```

### Exemples

#### ✅ Croisement LONG détecté
```
        N-2      N-1
K :      25       32
D :      28       30
```
→ K passe AU-DESSUS de D (25 < 28 puis 32 > 30) ✅

#### ✅ Croisement SHORT détecté
```
        N-2      N-1
K :      75       71
D :      72       73
```
→ K passe SOUS D (75 > 72 puis 71 < 73) ✅

#### ❌ Pas de croisement
```
        N-2      N-1
K :      75       78
D :      72       74
```
→ K reste AU-DESSUS de D (pas de croisement) ❌

### ❌ STOP si :
- Aucun croisement détecté entre N-2 et N-1
- K et D évoluent en parallèle sans se croiser

---

## 4️⃣ COHÉRENCE DIRECTIONNELLE

### Principe
Le **type d'extrême** et le **type de croisement** doivent être **cohérents** avec la logique de trading.

### Règle de cohérence STRICTE

| Zone Extrême | Croisement Requis | Signal Généré | Logique |
|--------------|-------------------|---------------|---------|
| **SURACHAT** | **SHORT (K sous D)** | SHORT | Vendre après pic |
| **SURVENTE** | **LONG (K sur D)** | LONG | Acheter après creux |

### Combinaisons VALIDES ✅

```
SURACHAT + Croisement SHORT = ✅ Signal SHORT
SURVENTE + Croisement LONG  = ✅ Signal LONG
```

### Combinaisons INVALIDES ❌

```
SURACHAT + Croisement LONG  = ❌ REJETÉ (acheter au sommet)
SURVENTE + Croisement SHORT = ❌ REJETÉ (vendre au creux)
```

### Exemple REJETÉ (incohérence)

```
Étapes validées :
1. Triple extrême SURACHAT ✅
2. Synchronisation baisse ✅
3. Croisement LONG détecté ✅
4. Cohérence : SURACHAT + LONG = ❌ INCOHÉRENT !

→ Signal REJETÉ (pas de trade contre-productif)
```

### ❌ STOP si :
- SURACHAT avec croisement LONG
- SURVENTE avec croisement SHORT

---

## 5️⃣ BOUGIE DE VALIDATION

### Principe
Rechercher une **bougie conforme** au signal dans une **fenêtre de 6 bougies** suivant le croisement.

### Fenêtre de recherche

**Début** : Bougie N-2 (où le croisement est validé)  
**Fin** : N-2 + 5 (6 bougies au total)

### Bougie conforme pour signal SHORT

**Bougie ROUGE** (bearish) :
```
Close < Open
```

### Bougie conforme pour signal LONG

**Bougie VERTE** (bullish) :
```
Close > Open
```

### Logique

Dès qu'**une seule bougie conforme** est trouvée dans la fenêtre :
→ Passer à la validation volume (contrainte 6)

Si **aucune bougie conforme** dans les 6 bougies :
→ ❌ Signal rejeté

### Exemples

#### ✅ Bougie trouvée (SHORT)
```
Signal : SHORT
Fenêtre : N-2 à N-2+5

Bougie N-2 : Open=100, Close=102 (verte) ❌
Bougie N-1 : Open=102, Close=101 (rouge) ✅ TROUVÉE !
→ Passer à validation volume
```

#### ❌ Aucune bougie (LONG)
```
Signal : LONG
Fenêtre : N-2 à N-2+5

Toutes les 6 bougies sont rouges (bearish) ❌
→ Signal rejeté
```

### ❌ STOP si :
- Aucune bougie conforme dans les 6 périodes
- Toutes les bougies vont dans le sens inverse du signal

---

## 6️⃣ VALIDATION VOLUME

### Principe
Le **volume de la bougie candidate** doit être **significativement supérieur** à la moyenne des volumes des **bougies INVERSES** précédentes.

### Type de bougies INVERSES recherchées

| Signal | Bougie Candidate | Bougies Inverses Recherchées |
|--------|------------------|------------------------------|
| **LONG** | VERTE (bullish) | ROUGES (bearish) précédentes |
| **SHORT** | ROUGE (bearish) | VERTES (bullish) précédentes |

**Logique** : Le mouvement du signal doit "écraser" les mouvements inverses précédents.

### Fenêtre de recherche DYNAMIQUE

**Période initiale** : 3 bougies  
**Extensions automatiques** : 3 → 6 → 12 → 24 → 48 → 100 (max)

Le système **étend automatiquement** la recherche jusqu'à trouver **au moins 2 bougies inverses**.

### Calcul du seuil

1. **Collecter** volumes des bougies inverses dans la fenêtre
2. **Calculer** moyenne des volumes inverses
3. **Seuil** = Moyenne × 0.25 (25%)
4. **Valider** : Volume candidate > Seuil

### Formule

```
Moyenne_Inverse = Σ(volumes_inverses) / nombre_inverses

Seuil = Moyenne_Inverse × 0.25

Validation : Volume_Candidate > Seuil
```

### Exemples

#### ✅ Volume validé (LONG)
```
Signal : LONG
Bougie candidate (verte) : Volume = 1200

Fenêtre 3 périodes (bougies rouges précédentes) :
- Bougie -3 : Volume = 800
- Bougie -2 : Volume = 900
- Bougie -1 : Volume = 850

Moyenne inverse = (800 + 900 + 850) / 3 = 850
Seuil = 850 × 0.25 = 212.5
Validation : 1200 > 212.5 ✅ OK

→ Signal LONG validé !
```

#### ✅ Volume validé après extension (SHORT)
```
Signal : SHORT
Bougie candidate (rouge) : Volume = 500

Fenêtre 3 périodes : 0 bougie verte ❌
Extension 6 périodes : 1 bougie verte ❌
Extension 12 périodes : 3 bougies vertes ✅

Bougies vertes trouvées :
- Volume = 1200
- Volume = 1100
- Volume = 1300

Moyenne inverse = (1200 + 1100 + 1300) / 3 = 1200
Seuil = 1200 × 0.25 = 300
Validation : 500 > 300 ✅ OK

→ Signal SHORT validé !
```

#### ❌ Volume insuffisant
```
Signal : LONG
Bougie candidate (verte) : Volume = 200

Bougies rouges précédentes (6 périodes) :
- Volumes : 1000, 1100, 1200, 1050

Moyenne inverse = 1087.5
Seuil = 1087.5 × 0.25 = 271.875
Validation : 200 < 271.875 ❌ REJETÉ

→ Volume trop faible, signal rejeté
```

#### ❌ Pas assez de bougies inverses
```
Signal : SHORT
Bougie candidate (rouge) : Volume = 600

Extensions jusqu'à 100 périodes :
- Seulement 1 bougie verte trouvée ❌

→ Pas assez de bougies inverses (< 2)
→ Signal rejeté
```

### ❌ STOP si :
- Volume candidate < Seuil (25% moyenne inverses)
- Moins de 2 bougies inverses trouvées (même après extension max)

---

## 📊 Paramètres de Configuration

### Indicateurs

```yaml
# Périodes de calcul
indicators:
  cci:
    period: 20
  mfi:
    period: 14
  stochastic:
    period_k: 14
    smooth_k: 3
    period_d: 3
```

### Seuils d'extrêmes

```yaml
scalping:
  # SURACHAT
  cci_surachat: 100.0
  mfi_surachat: 60.0
  stoch_surachat: 70.0
  
  # SURVENTE
  cci_survente: -100.0
  mfi_survente: 40.0
  stoch_survente: 30.0
```

### Validation

```yaml
scalping:
  # Fenêtre bougie
  validation_window: 6
  
  # Volume
  volume_threshold: 0.25      # 25%
  volume_period: 3            # Période initiale
  volume_max_ext: 4           # Max extensions (3→6→12→24)
```

---

## 🎯 Résumé : Signal VALIDÉ

Un signal est **émis** uniquement si les **6 contraintes** sont **TOUTES validées** :

```
1. ✅ Triple extrême flexible (3 indicateurs en zone, N-1 ou N-2)
2. ✅ Synchronisation (3 indicateurs même sens N-2 → N-1)
3. ✅ Croisement stochastique (K croise D)
4. ✅ Cohérence directionnelle (zone ↔ croisement)
5. ✅ Bougie conforme (rouge/verte dans 6 périodes)
6. ✅ Volume suffisant (> 25% moyenne inverses)

→ SIGNAL ÉMIS : LONG ou SHORT
```

### Contenu du signal

```go
Signal {
    Type:      "LONG" ou "SHORT"
    Timestamp: Timestamp bougie validée
    Price:     Close bougie validée
    CCI:       Valeur CCI au croisement
    MFI:       Valeur MFI au croisement
    StochK:    Valeur K au croisement
    StochD:    Valeur D au croisement
    Volume:    Volume bougie validée
}
```

---

## 🚫 Causes de Rejet

| Contrainte | Cause Rejet | Conséquence |
|------------|-------------|-------------|
| 1. Triple Extrême | Un indicateur pas en zone | ❌ Pas de signal |
| 2. Synchronisation | Mouvements divergents N-2→N-1 | ❌ Pas de signal |
| 3. Croisement | K et D ne se croisent pas | ❌ Pas de signal |
| 4. Cohérence | SURACHAT+LONG ou SURVENTE+SHORT | ❌ Pas de signal |
| 5. Bougie | Aucune bougie conforme en 6 périodes | ❌ Pas de signal |
| 6. Volume | Volume < 25% moyenne inverses | ❌ Pas de signal |

**Une seule contrainte échouée = Aucun signal émis**

---

## 📝 Notes Importantes

### Flexibilité vs Rigidité

- **FLEXIBLE** : Détection extrêmes (N-1 ou N-2 par indicateur)
- **STRICT** : Synchronisation mouvements (tous ensemble N-2→N-1)
- **STRICT** : Cohérence zone ↔ croisement
- **FLEXIBLE** : Recherche bougie (6 périodes)
- **DYNAMIQUE** : Extension fenêtre volume (jusqu'à 100)

### Fréquence de vérification

- **Tick** : Toutes les 10 secondes (monitoring)
- **Traitement** : Uniquement sur **clôture bougie 5m**
- **Données** : 300 klines récupérées à chaque clôture

### Mode de fonctionnement

- **Stateless** : Pas de stockage klines entre marqueurs
- **Fresh data** : Récupération API à chaque clôture
- **Async** : Traitement en goroutine (pas de blocage)

---

**Version** : 1.0  
**Date** : 6 novembre 2025  
**Fichier implémentation** : `cmd/scalping_live_bybit/app_live.go`  
**Configuration** : `devops/configs/scalping-live-bybit.nomad`
