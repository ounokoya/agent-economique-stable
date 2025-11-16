# Setups par indicateur (version de travail pour validation)

## Objectif
- Définir des setups “edge‑triggered” par indicateur (DMI, VWMA, Stoch) avec seuils configurables par setup.
- Sortie attendue par setup: orientation LONG/SHORT, et possibilité de verbosité (détaillé vs heure seule).

## Terminologie et conventions
- DI_sup / DI_inf
  - Long: DI_sup = DI+, DI_inf = DI−.
  - Short: DI_sup = DI−, DI_inf = DI+.
- Pentes
  - “DX croissant” signifie DX[i] > DX[i−1]. “DX décroissant” signifie DX[i] < DX[i−1]. Idem pour ADX.
- Edge‑triggered (i−1 → i)
  - “était sous/sur/au‑dessus” s’évalue sur i−1, “passe sous/sur/au‑dessus” s’évalue au pas i.
  - “sur” → égalité approchée avec marge eps (petite tolérance).
- Seuils par setup (tous optionnels)
  - min_dx, min_adx: valeurs minimales exigées pour DX/ADX à i.
  - seuil_resp: seuil de respiration utilisé dans DMI respiration.
  - eps: tolérance d’égalité pour “sur”.


# DMI

- Remarque commune
  - Dans toutes les variantes où l’on exige “DX reste sous DI_sup” et/ou “ADX reste sous DI_sup”, la comparaison est numérique au pas i (on peut appliquer eps si désiré).
  - Exception ouverture: pour le setup « ouverture », appliquer uniquement le plafond ADX < DI_sup (pas de contrainte DX < DI_sup).
  - Règle globale DI de fond: pour tout croisement de DX, le contexte DI de fond (DI_sup > DI_inf côté LONG, inverse côté SHORT) est vérifié uniquement sur la bougie de départ i; il n’est pas re‑vérifié sur une éventuelle bougie de confirmation i+1.

- Setup: croisement ouverture tendance DMI LONG
  - Conditions (i−1 → i)
    - Contexte DI à i: DI+ > DI−. Aucun croisement DI exigé.
    - Pentes: DX croissant et ADX croissant.
    - Croisement DX/ADX: DX était ≤ ADX à i−1 et DX passe au‑dessus d’ADX à i.
    - Seuil directionnel: DX > DI_inf (ici DI_inf = DI−) au moment i.
    - Distance minimale (optionnelle): |DX − ADX| ≥ `DMI_OPEN_MIN_DX_ADX_GAP` si `DMI_OPEN_USE_MIN_DX_ADX_GAP` est activé.
    - Plafonds:
      - Relatif: ADX < DI_sup (ici DI_sup = DI+).
      - Absolu (optionnel): ADX ≤ DMI_OPEN_MAX_ADX (si DMI_OPEN_MAX_ADX > 0; 0 = désactivé).
    - Seuils optionnels: DX[i] ≥ min_dx, ADX[i] ≥ min_adx.
    - Périodes: ADX peut utiliser une période de lissage distincte (periodADX ≠ periodDI).
  - Orientation: LONG.

- Setup: croisement ouverture tendance DMI SHORT
  - Conditions (i−1 → i)
    - Contexte DI à i: DI− > DI+ (aucun croisement DI exigé).
    - Pentes: DX croissant et ADX croissant.
    - Croisement DX/ADX: DX était ≤ ADX à i−1 et DX passe au‑dessus d’ADX à i.
    - Seuil directionnel: DX > DI_inf (ici DI_inf = DI+) au moment i.
    - Distance minimale (optionnelle): |DX − ADX| ≥ `DMI_OPEN_MIN_DX_ADX_GAP` si `DMI_OPEN_USE_MIN_DX_ADX_GAP` est activé.
    - Plafonds:
      - Relatif: ADX < DI_sup (ici DI_sup = DI−).
      - Absolu (optionnel): ADX ≤ DMI_OPEN_MAX_ADX (si DMI_OPEN_MAX_ADX > 0; 0 = désactivé).
    - Seuils optionnels: DX[i] ≥ min_dx, ADX[i] ≥ min_adx.
    - Périodes: ADX peut utiliser une période de lissage distincte (periodADX ≠ periodDI).
  - Orientation: SHORT.

- Setup: croisement respiration tendance DMI LONG
  - Contexte: DI+ > DI− (tendance haussière en cours).
  - Contexte évalué sur la bougie i uniquement (non re‑vérifié sur i+1).
  - Pentes: DX décroissant ET ADX décroissant.
  - Transition: DX était “sur” ADX (≈ égalité avec eps) et DX passe sous ADX.
  - Gestion indépendante selon la position de DX vis‑à‑vis de DI_inf (= DI−):
    - Cas A (immédiat 1): si DX ≤ DI_inf → validation immédiate à i.
    - Cas B (immédiat 2): si DX > DI_inf ET DX − DI_inf ≤ seuil_resp → validation immédiate à i.
    - Cas C (attente 1 bougie): si DX > DI_inf ET DX − DI_inf > seuil_resp → NE PAS valider à i. Placer un état “en attente”.
      - Bougie de confirmation (i+1): valider seulement si, à i+1, DX > DI_inf et DX − DI_inf ≤ seuil_resp. Ne pas re‑vérifier le contexte DI (évalué à i uniquement).
      - Si ces conditions échouent à i+1, annuler l’attente (pas de setup émis).
  - Seuils optionnels: min_dx, min_adx peuvent s’appliquer à i (et/ou i+1, à préciser si désiré).
  - Orientation: LONG.

- Setup: croisement respiration tendance DMI SHORT
  - Miroir LONG (DI− > DI+, pentes décroissantes, DX passe sous ADX, mêmes trois cas A/B/C avec DI_inf = DI+).
  - Contexte évalué sur la bougie i uniquement (non re‑vérifié sur i+1).
  - Bougie de confirmation (i+1): ne pas re‑vérifier DI; valider seulement si, à i+1, DX > DI_inf et DX − DI_inf ≤ seuil_resp.
  - Orientation: SHORT.

- Setup: croisement reprise tendance DMI LONG
  - Contexte: DI+ > DI−.
  - Pente: DX croissant.
  - Transition: DX était sous ADX, et DX passe au‑dessus d’ADX.
  - Plafond: DX < DI_sup et ADX < DI_sup (DI_sup = DI+).
  - Orientation: LONG.

- Setup: croisement reprise tendance DMI SHORT
  - Miroir LONG.
  - Orientation: SHORT.


# VWMA

- Paramètres usuels: 20, 60, 240 (à confirmer). “Sous/sur” = strict, possibilité d’utiliser eps.

- Cassure VWMA20 haussière
  - Cas 1 (même bougie): open[i] < vwma20[i] ET close[i] > vwma20[i].
  - Cas 2 (n−1 → n): open[i−1] < vwma20[i−1] ET close[i] > vwma20[i].
  - Orientation: LONG.

- Cassure VWMA20 baissière
  - Cas 1: open[i] > vwma20[i] ET close[i] < vwma20[i].
  - Cas 2: open[i−1] > vwma20[i−1] ET close[i] < vwma20[i].
  - Orientation: SHORT.

- Cassure VWMA60 haussière/baissière: identique à VWMA20 en remplaçant par vwma60.
- Cassure VWMA240 haussière/baissière: identique à VWMA20 en remplaçant par vwma240.

- Croisement tendance haussière VWMA 60_240
  - vwma60[i−1] ≤ vwma240[i−1] ET vwma60[i] > vwma240[i].
  - Orientation: LONG.

- Croisement tendance baissière VWMA 60_240
  - vwma60[i−1] ≥ vwma240[i−1] ET vwma60[i] < vwma240[i].
  - Orientation: SHORT.

- Croisement haussier VWMA 20_60
  - vwma20[i−1] ≤ vwma60[i−1] ET vwma20[i] > vwma60[i].
  - Orientation: LONG.

- Croisement baissier VWMA 20_60
  - vwma20[i−1] ≥ vwma60[i−1] ET vwma20[i] < vwma60[i].
  - Orientation: SHORT.

- Filtre indépendant: croisement VWMA(P1, P2)
  - Paramètres: P1, P2 configurables; type ∈ {ANY, GOLDEN, DEAD}.
  - Détection:
    - GOLDEN: vwma(P1) était ≤ vwma(P2) à i−1 et vwma(P1) passe au‑dessus de vwma(P2) à i.
    - DEAD: vwma(P1) était ≥ vwma(P2) à i−1 et vwma(P1) passe sous vwma(P2) à i.
    - ANY: accepter GOLDEN ou DEAD.
  - Orientation: GOLDEN → LONG, DEAD → SHORT.
  - Toggles (démo): FILTER_VWMA_ENABLED, FILTER_VWMA_LOG_VERBOSE, FILTER_VWMA_P1, FILTER_VWMA_P2, FILTER_VWMA_CROSS_TYPE.

- États de tendance VWMA
  - Haussier: vwma60 > vwma240.
  - Baissier: vwma60 < vwma240.


# Stochastique

- Paramètres à confirmer: K_length, D_length, smoothing; zones: surachat (≥ 80), survente (≤ 20) par défaut.

- Croisement haussier Stoch
  - %K était sous %D, et %K passe au‑dessus de %D.
  - Contrainte: %K et %D non en surachat (par ex. < 80).
  - Orientation: LONG.

- Croisement baissier Stoch
  - %K était sur/au‑dessus de %D, et %K passe sous %D.
  - Contrainte: %K et %D non en survente (par ex. > 20).
  - Orientation: SHORT.

# Filtre 2: ratio moyen signé 3 bougies (body/ATR)

- Fenêtre glissante de 3 bougies sur les analyses de corps/non‑corps rapportés à l’ATR.
- Règle de validation (bougie i):
  - |avg_signed_body/atr (sur 3 bougies)| ≥ seuil FILTER2_MIN_BODY_ATR.
  - BodyATRRatio(i) > 0.
  - Cohérence de signe: signe(BodyATRRatio(i)) = signe(avg_signed_body/atr sur 3 bougies).
- Toggles (démo): FILTER2_ENABLED, FILTER2_MIN_BODY_ATR, FILTER2_LOG_VERBOSE.
- Logs: si verbose → détails; sinon → heure d’ouverture seulement.


## Notes d’implémentation (pour la démo)
- Chaque setup aura des toggles: *_ENABLED, *_LOG_VERBOSE, et ses seuils (min_dx, min_adx, seuil_resp, eps…).
- Détection edge‑triggered, séries non filtrées.
- Logs:
  - Verbose: détails complets.
  - Sinon: “open‑time‑only + orientation”.
- Le cas DMI respiration “attente 1 bougie” est indépendant et ne s’applique qu’au scénario DX > DI_inf avec dépassement du seuil_resp. Les autres cas valident immédiatement.
