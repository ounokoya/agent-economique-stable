# SPÉCIFICATION STRATÉGIE SCALPING MOMENTIUM

Date: 2025-11-09
Version: 1.0
Auteur: Agent Économique Stable

---

## PRINCIPE GÉNÉRAL

La stratégie Scalping Momentium exploite des bougies à fort corps et un filtre stochastique pour générer des signaux directionnels. L’exécution ouvre uniquement sur `ENTRY` quand aucune position n’est ouverte, ferme sur `EXIT` du même type ou sur `ENTRY` opposé, et intègre un trailing stop basé sur l’ATR mesuré à l’entrée. Les performances sont évaluées via la variation brute captée (différence de prix).

---

## GÉNÉRATION DES SIGNAUX

Lieu d’implémentation de référence: `cmd/scalping_momentium/main.go` → `detectMomentumSignals(...)`

### 1) Filtres bougie
- Corps absolu: `body = |Close - Open|`
- Pourcentage du corps dans le range: `bodyPct = body / (High - Low)`
- Conditions requises:
  - `bodyPct >= BODY_PCT_MIN`
  - `body >= BODY_ATR_MIN * ATR(i)`

### 2) Direction
- `Type = LONG` si `Close > Open`
- `Type = SHORT` si `Close < Open`
- Bougies doji (`Close == Open`) ignorées

### 3) Filtre Stochastique (TV standard)
- Calcul: `StochTVStandard(STOCH_K_PERIOD, STOCH_K_SMOOTH, STOCH_D_PERIOD)`
- Validation:
  - LONG si `K(i) < STOCH_K_LONG_MAX`
  - SHORT si `K(i) > STOCH_K_SHORT_MIN`

### 4) Étiquetage `ENTRY` vs `EXIT`
- Références locales sur les 2 bougies précédentes, dépendantes de la couleur de chaque bougie:
  - Pour LONG: `ref(j) = max(Open[j], Close[j])` (sommet)
  - Pour SHORT: `ref(j) = min(Open[j], Close[j])` (creux)
- Règle:
  - `ENTRY LONG` si `Close(i) >= max(ref(i-1), ref(i-2))`, sinon `EXIT LONG`
  - `ENTRY SHORT` si `Close(i) <= min(ref(i-1), ref(i-2))`, sinon `EXIT SHORT`

### Métadonnées du signal
- Index
- Timestamp
- Type (LONG | SHORT)
- Label (ENTRY | EXIT)
- Open
- High
- Low
- Close
- Body (|Close−Open|)
- Range (High−Low)
- BodyPctBar (Body/Range)
- ATR (période configurée)
- BodyToATR (Body/ATR)
- StochK

---

## MOTEUR D’EXÉCUTION

Lieu d’implémentation de référence: `displaySignals(...)` (exécution inline au fil du tableau).

### État et transitions
- Une seule position à la fois: `openType ∈ {"", LONG, SHORT}`
- `ENTRY` alors que `openType == ""` → ouvrir position (`entryPrice = Close`, `atrEntry = ATR`, init trailing)
- `EXIT` du même type que `openType` → fermer position
- `ENTRY` de type opposé → fermer position puis ouvrir immédiatement la nouvelle au `Close` courant
- `ENTRY` du même type → ignoré (pas de pyramiding, pas de reset du prix d’entrée)
- `EXIT` du type opposé → ignoré

### Trailing stop (ATR d’entrée)
- À l’entrée: mémoriser `atrEntry = ATR(i_entrée)`
- Stop initial:
  - LONG: `trail = entryPrice - atrEntry`
  - SHORT: `trail = entryPrice + atrEntry`
- Mise à jour à chaque bougie (offset fixe basé sur `atrEntry`):
  - LONG: `trail = max(trail, Close - atrEntry)` (ne baisse jamais)
  - SHORT: `trail = min(trail, Close + atrEntry)` (ne monte jamais)
- Déclenchement stop sur `Close`:
  - LONG: fermer si `Close <= trail`
  - SHORT: fermer si `Close >= trail`
  - Prix d’exécution stop: `stopPrice = trail`
- Comportement combiné sur la même bougie: si stop ferme la position et que le signal courant est `ENTRY`, ré-ouvrir immédiatement (nouveau `entryPrice`, `atrEntry`, `trail`)

### Priorités sur une bougie
1. Mise à jour du trailing stop
2. Vérification du stop et fermeture éventuelle
3. Sinon, si `EXIT` même type → fermeture
4. Sinon, si `ENTRY` opposé → fermeture + ouverture immédiate
5. Sinon, si `ENTRY` même type → ignorer

Pseudocode simplifié:
```
if flat and ENTRY -> open(entryPrice=Close, atrEntry=ATR, trail=entry±ATR)
else:
  updateTrail(atrEntry)
  if stopHit(Close, trail) -> close at stopPrice; if Label==ENTRY -> open new at Close
  else if Label==EXIT and Type==openType -> close at Close
  else if Label==ENTRY and Type!=openType -> close at Close; open new at Close
  else -> hold
```

---

## CALCUL DES VARIATIONS CAPTÉES

### Valeur captée par trade (brute)
- Fermeture sur signal: `capté = Close(sortie) - Open(entrée)`
- Fermeture par stop: `capté = stopPrice - Open(entrée)`
- Affichage: imprimé sur la ligne où la fermeture a lieu

### Agrégations
- `sumLong`: somme des `capté` des fermetures LONG
- `sumShort`: somme des `capté` des fermetures SHORT
- `total_directionnel = sumLong + (-1 * sumShort)`

Remarque: la différence est brute (pas normalisée en % et pas orientée). Le total directionnel convertit les variations SHORT en contribution positive via le facteur `-1`.

---

## PARAMÈTRES TECHNIQUES

Paramètres principaux exposés dans la démo actuelle:

```yaml
symbol: SYMBOL                 # ex: SOL_USDT
interval: TIMEFRAME            # ex: 1m, 5m, 15m
nb_candles: NB_CANDLES         # taille fenêtre de données
atr_period: ATR_PERIOD         # ex: 3
body_pct_min: BODY_PCT_MIN     # ex: 0.60 (60% du range)
body_atr_min: BODY_ATR_MIN     # ex: 0.60 (>= 0.60 * ATR)
stoch_k_period: STOCH_K_PERIOD # ex: 14
stoch_k_smooth: STOCH_K_SMOOTH # ex: 3
stoch_d_period: STOCH_D_PERIOD # ex: 3
stoch_k_long_max: STOCH_K_LONG_MAX   # ex: 50.0
stoch_k_short_min: STOCH_K_SHORT_MIN # ex: 50.0
```

---

## RÈGLES IMPORTANTES

1) Une seule position à la fois (flat/long/short)
2) Ouverture uniquement sur `ENTRY` si flat
3) Fermeture sur `EXIT` du même type, ou `ENTRY` opposé, ou stop
4) Priorité: mise à jour et déclenchement du stop > `EXIT` même type > `ENTRY` opposé > `ENTRY` même type
5) Aucune action sur `ENTRY` même type tant qu’une position est ouverte
6) La valeur captée est une différence de prix brute; l’agrégation directionnelle convertit les SHORT en contribution positive

---

## TESTS À EFFECTUER

1) Vérifier que les signaux respectent les filtres `bodyPct`, `body vs ATR`, et `StochK`
2) Valider l’étiquetage `ENTRY/EXIT` via les références `n-1`/`n-2`
3) Tester l’exécution: ouverture/fermeture, règles d’ignoration (`ENTRY` même type)
4) Tester le trailing stop: initialisation, traînage, déclenchement sur Close, interaction avec `ENTRY` sur la même bougie
5) Valider les calculs `capté` sur sorties par signal et par stop
6) Vérifier `sumLong`, `sumShort`, `total_directionnel`

---

## IMPLÉMENTATION (POINTEURS)

- Démo et logique actuelle: `cmd/scalping_momentium/main.go`
  - Détection signaux: `detectMomentumSignals(...)`
  - Exécution + trailing + rendu: `displaySignals(...)`
- Indicateurs utilisés: `internal/indicators` (ATR TV standard, Stoch TV standard)

---

Version finale 1.0 – Cette spécification décrit la logique fonctionnelle de la stratégie Scalping Momentium sans éléments décoratifs. Ne pas modifier sans discussion préalable.
