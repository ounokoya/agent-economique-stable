# ban_eco — Stratégie fenêtre actuelle

## Résumé
- Mode fenêtre: activé.
- Ancre: croisement VWMA4/VWMA50 (`cross_4_50`).
- Seuil d'écart (gap) exigé dans la fenêtre: `0.5 × ATR(3)`.
- Périodes: VWMA4 = 2, VWMA50 = 30, ATR = 3.
- Logique d'accumulation: chaque condition activée est validée une fois dans la fenêtre; aucune persistance exigée.
- Signal: dès que toutes les conditions activées sont validées, on log un `[WIN]` et l'index avance de `WINDOW_SIZE/2`.
- Fenêtre: pré-fenêtre de `i - WINDOW_SIZE/2` jusqu'à `i-1` (bornée à 0), post-fenêtre de `i+1` jusqu'à `i + WINDOW_SIZE`.
- Spécificité croisement: si l'ancre est un croisement (`cross_*`), la validation d'écart ATR se fait uniquement en post-fenêtre (après la bougie d'ancre).
- Horodatage du signal: le champ `at=` correspond à la bougie de validation finale `j*` (post-ancre), pas à la bougie d'ancre `i`.

## Paramètres actifs (extraits de cmd/ban_eco/app.go)
- `WINDOW_MODE_ENABLED = true`
- `WINDOW_SIZE = 5`
- `WINDOW_ANCHOR = "cross_4_50"`
- `CROSS_VWMA4_50_ENABLED = true`
- `CROSS_VWMA4_50_ATR_MULT = 0.5`
- `POST_CROSS_SCORE_ENABLED = false`
- `POST_CROSS_SCORE_MIN = 0.5`
- Désactivés: `BASE_ENABLED`, `VWMA4_ENABLED`, `VWMA10_ENABLED`, `VWMA50_ENABLED`, `VWMA200_ENABLED`, `COMBO_VWMA4_10_ENABLED`, `COMBO_VWMA10_50_ENABLED`, `CROSS_VWMA4_10_ENABLED`, `CROSS_VWMA10_50_ENABLED`.

## Logique de détection
1) Calculs préalables
- Séries ATR(3), VWMA4, VWMA10, VWMA50, VWMA200 (selon besoin/ancre/conditions).

2) Ancre `cross_4_50`
- Détection du croisement entre VWMA4 (rapide) et VWMA50 (lente) à la bougie `i`:
  - LONG si `pf <= ps` et `cf > cs`.
  - SHORT si `pf >= ps` et `cf < cs`.
  - Notation: `pf, ps` = valeurs à `i-1`; `cf, cs` = valeurs à `i`.
- Aucune validation d'écart à l'ancre; `sig` (LONG/SHORT) est fixé par le sens du croisement.

3) Fenêtre d'accumulation des conditions
- Une fois l'ancre validée à l'index `i`, on scanne:
  - Pré-fenêtre: `j ∈ [max(0, i - WINDOW_SIZE/2), i)`.
  - Post-fenêtre: `j ∈ (i, i + WINDOW_SIZE]`.
- Condition d'écart VWMA4↔VWMA50 (si `CROSS_VWMA4_50_ENABLED = true`):
  - LONG: `gap = VWMA4[j] - VWMA50[j]`.
  - SHORT: `gap = VWMA50[j] - VWMA4[j]`.
  - Critère: `gap + EPS >= CROSS_VWMA4_50_ATR_MULT × ATR[j]`.
  - Validation dans la fenêtre:
    - Si l'ancre est un croisement (`cross_*`): post-fenêtre uniquement (j > i).
    - Sinon: pré ou post.
  - Dès validation, la condition est cochée (pas de persistance requise).
- Les autres conditions activées (base, cassures prix↔VWMA, combos) suivent la même logique d'accumulation.

### Filtre optionnel: Score post-croisement (aligné vs contraire)
- Active si `POST_CROSS_SCORE_ENABLED = true` et uniquement quand l'ancre est un croisement (`cross_*`).
- Cumul en post-fenêtre de `i+1` jusqu'à la bougie `j*` qui valide l'écart VWMA.
- Contributions séparées:
  - `body = |Close - Open|`, `atr = ATR[j]`, `bodyATR = body / atr`.
  - Alignée au signal (`sig`): `posScore += bodyATR`.
  - Contraire au signal: `negScore += bodyATR`.
  - Score net: `score = posScore - negScore`.
- Critère au moment de la validation d'écart (à `j*`): `score + EPS >= POST_CROSS_SCORE_MIN`.
- Objectif: s'assurer que la dynamique des bougies post-croisement va dans le sens du signal avant d'accepter le croisement.

### Spécification DI/DX/ADX (nouvelle)
- Ancrage `cross_di` (spécification):
  - LONG si `+DI` croise au-dessus de `-DI` à la bougie d'ancre `i`.
  - SHORT si `+DI` croise sous `-DI` à `i`.
  - Période de calcul (proposition): `DI_ADX_PERIOD = 14` (Wilder).
- Condition fenêtre DX/ADX (utilisable avec ancre VWMA ou ancre `cross_di`):
  - Croisement `DX`↔`ADX` exigé dans la fenêtre, avec portée configurable:
    - `DX_ADX_SCOPE ∈ {pre, post, both}` (avant/après ancre, ou les deux).
  - Direction du croisement:
    - `DX_ADX_DIRECTION ∈ {up, down, any}` où `up` signifie `DX` croise au-dessus de `ADX`.
  - Seuil d'écart après croisement:
    - `DX_ADX_GAP_MIN`: exiger qu'à un moment (dans la même portée) `|DX - ADX| ≥ DX_ADX_GAP_MIN`.
  - Lissage ADX:
    - `ADX_SMOOTH_PERIOD`: période dissage pour ADX.
    - Formule: `ADX = RMA(DX, ADX_SMOOTH_PERIOD)` (Wilder).
  - Garde finale DX (optionnelle):
    - `DX_REJECT_IF_ABOVE_BOTH_DI` (bool): au moment de la validation finale `j*`, si `DX(j*) > DI+(j*)` ET `DX(j*) > DI-(j*)`, le signal est rejeté (on continue à chercher un `j*` ultérieur dans la fenêtre).
  - Contraintes relatives aux DI au moment du croisement DX/ADX (activables):
    - `DX_ADX_REQUIRE_UNDER_DI_INFERIOR`: à la bougie de croisement, imposer `DX ≤ min(+DI, -DI)` ET `ADX ≤ min(+DI, -DI)`.
    - `DX_ADX_REQUIRE_UNDER_DI_SUPERIOR`: à la bougie de croisement, imposer `DX ≤ max(+DI, -DI)` ET `ADX ≤ max(+DI, -DI)`.
    - Si les deux sont activées, appliquer la contrainte la plus stricte (sous `DI` inférieur).
- Compatibilité ancre VWMA:
  - Avec `WINDOW_ANCHOR = cross_*` VWMA, cette condition DX/ADX agit comme un filtre de fenêtre additionnel; elle ne change pas la direction `sig` (définie par l'ancre).
- Horodatage:
  - Le champ `at=` reste la bougie de validation finale `j*` (post-ancre), une fois toutes les conditions (VWMA/DI/DX/ADX/score/combos) cochées.

### Principe ancre vs filtre (spécification)
- Complémentarité: VWMA et DI/DX/ADX peuvent échanger leurs rôles.
- Si une famille A est l'ancre, l'autre B agit comme filtre fenêtre complet:
  - B doit d'abord avoir SON propre croisement B détecté dans la fenêtre par rapport à `i` (portée selon `B_SCOPE`: `pre`, `post` ou `both`).
  - Les validations de B sont ensuite vérifiées uniquement APRÈS ce croisement B (post-croisement B).
    - Exemple VWMA en filtre: après le croisement VWMA trouvé, on vérifie l'écart ATR VWMA en post-only.
    - Exemple DX/ADX en filtre: après le croisement DX↔ADX trouvé, on vérifie `|DX−ADX| ≥ DX_ADX_GAP_MIN` (et, si activées, les contraintes vs DI s'appliquent sur la bougie du croisement DX/ADX).
  - La direction `sig` reste celle de l'ancre A; B n'altère pas `sig`.
  - Le signal est daté à `j*` (dernière validation dans la fenêtre).

4) Emission du signal
- Si toutes les conditions activées sont cochées, on émet `[WIN]` immédiatement et on saute `i += WINDOW_SIZE/2`. L'horodatage `at=` est celui de `j*`.

## Conditions activables
- Base (body% et body/ATR avec couleur de bougie).
- Cassures prix↔VWMA: 4, 10, 50, 200.
- Combos prix↔VWMA: (4 & 10), (10 & 50).
- Croisements VWMA↔VWMA: (4/10), (4/50), (10/50) avec seuils d'ATR indépendants.
- Score post-croisement (aligné vs contraire) avec seuil `POST_CROSS_SCORE_MIN`.
- DX/ADX (croisement et écart) avec options de portée/direction/contraintes vs DI (spécification).
- Ancre configurable: `base`, `vwma4`, `vwma10`, `vwma50`, `vwma200`, `cross_4_10`, `cross_4_50`, `cross_10_50`, `cross_di`.

## Paramètres DI/DX/ADX (spécification)
- `DI_ADX_PERIOD` (par défaut proposé: 14)
- `ADX_SMOOTH_PERIOD` (par défaut proposé: 14)
- `DX_ADX_ENABLED`
- `DX_ADX_SCOPE ∈ {pre, post, both}`
- `DX_ADX_DIRECTION ∈ {up, down, any}`
- `DX_ADX_GAP_MIN`
- `DX_ADX_REQUIRE_UNDER_DI_INFERIOR` (bool)
- `DX_ADX_REQUIRE_UNDER_DI_SUPERIOR` (bool)
- `DX_REJECT_IF_ABOVE_BOTH_DI` (bool)

## Utilisation
- Exemple d'exécution (lecture klines Gate.io):
```
go run ./cmd/ban_eco -symbol SOLUSDT -n 1000
```
- Sorties typiques:
  - `[WIN] anchor=cross_4_50 | dir=LONG|SHORT | at=... | window=5 | ok=ALL`
  - `[BAN] ...` pour les bougies répondant aux filtres de base (quand le mode fenêtre est désactivé).

## Notes
- L'écart du croisement est décorrélé du moment du croisement et se valide dans la fenêtre selon la direction (`sig`).
- Les constantes sont en tête de `cmd/ban_eco/app.go`.
- TF par défaut: `1m`.
