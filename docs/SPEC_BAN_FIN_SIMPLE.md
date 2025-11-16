# Spécification BAN_FIN_SIMPLE (VWMA + DMI/DX-ADX + ATR + MFI)

Objectif: version simplifiée de BAN_FIN qui supprime toute dépendance au CCI et utilise MFI comme confirmation positive.

- Long si MFI en extrême haussier (MFI ≥ seuil_haussier)
- Short si MFI en extrême baissier (MFI ≤ seuil_baissier)
- Edge-trigger: n’émettre qu’à la première bougie valide après le dernier croisement VWMA, dans une fenêtre glissante W

## Modules et données

- Données d’entrée: séquence de bougies OHLCV non triées
  - Trier par timestamp croissant
  - Dédupliquer par timestamp
  - Exclure bougies incomplètes

- VWMA
  - `VWMAShort` (période courte)
  - `VWMALong` (période longue)
  - Détection des croisements haussier/baisser court vs long

- DMI (TV standard)
  - `DI+`, `DI−`, `ADX`
  - `DX` dérivé de `DI+/DI−`
  - Détection des croisements `DX` vs `ADX`

- ATR(3)
  - Volatilité de base pour le gating (gap ≥ n × ATR3)

- MFI (TV standard)
  - Oscillateur 0–100
  - Détection d’extrêmes: `MFI ≥ overbought` (extrême haussier), `MFI ≤ oversold` (extrême baissier)

## Paramètres (Config)

- VWMAShortPeriod: int (ex: 10)
- VWMALongPeriod: int (ex: 20)
- DMIPeriod: int (ex: 14)
- WindowMatching (W): int (ex: 5)
- ATRPeriod: int (fixé à 3 dans les règles de base)
- GapATRMultiplier (n): float (ex: 1.0)
- GapBasis: string
  - "vwma_spread" (par défaut) = |VWMAShort − VWMALong|
  - "price_vs_vwma_short" = |Close − VWMAShort|
  - "price_vs_vwma_long" = |Close − VWMALong|
- EnableGapGating: bool (ex: true)
- Slopes (optionnels):
  - EnableSlopeVWMAShort: bool, SlopeVWMAShortMin: float
  - EnableSlopeVWMALong: bool, SlopeVWMALongMin: float
  - EnableSlopeDX: bool, SlopeDXMin: float
  - EnableSlopeADX: bool, SlopeADXMin: float
- MFI
  - MFIPeriod: int (ex: 14)
  - MFIOverboughtThreshold: float (ex: 80)
  - MFIOversoldThreshold: float (ex: 20)
  - MFIMode: string = "positive_extremes" (LONG si ≥ overbought, SHORT si ≤ oversold)
- DX/ADX
  - DXADXRequiredDirectionalCross: bool (ex: true)
- EdgeTrigger: bool (ex: true)

## Conditions de base (fenêtre W)

Pour la bougie candidate i* (dernière bougie fermée):

1) Existence d’un dernier croisement VWMA dans W
- Chercher dans [i* − W + 1, i*] le dernier indice t_cross où VWMAShort croise VWMALong
- Déterminer la cible: LONG si VWMAShort > VWMALong à i*, sinon SHORT

2) Alignement DMI à i*
- LONG: DI+ > DI−
- SHORT: DI− > DI+

3) Direction DX/ADX dans W
- Chercher le dernier croisement DX vs ADX dans W
- Si `DXADXRequiredDirectionalCross = true`:
  - LONG: dernier croisement doit être UP (DX passe au-dessus d’ADX)
  - SHORT: dernier croisement doit être DOWN (DX passe en dessous d’ADX)

Si l’une des 3 bases échoue, aucun signal.

## Gating par bougie (à i*)

- Pentes (si activées):
  - VWMAShort: signe cohérent avec la cible, |Δ| ≥ min si spécifié
  - VWMALong: idem
  - DX/ADX: idem
- Gap vs ATR(3) (si `EnableGapGating`):
  - gap(i*) ≥ n × ATR3(i*), base selon `GapBasis`
- MFI (confirmation POSITIVE):
  - LONG: MFI(i*) ≥ MFIOverboughtThreshold
  - SHORT: MFI(i*) ≤ MFIOversoldThreshold

Toutes les clauses gating doivent être vraies pour valider i*.

## Edge-trigger (événement)

- Définir j0: première bougie dans [t_cross .. i*] qui satisfait Bases + Gating (chaque j évalué avec sa propre fenêtre W glissante)
- N’émettre un signal que si i* = j0
- Sinon, aucun nouveau signal (les bougies ultérieures alignées restent un ÉTAT, pas un nouvel événement). Réarmement sur un nouveau croisement VWMA.

## Sortie (événement de signal)

- Type: LONG | SHORT (toujours mode TENDANCE car le DMI aligné et DX/ADX directionnel imposent la direction)
- Timestamp: ouverture de la bougie i*
- Prix: Close(i*)
- Métadonnées minimales:
  - `vwma_short`, `vwma_long`, `di_plus`, `di_minus`, `dx`, `adx`, `atr3`, `mfi`
  - `vwma_cross_index` (t_cross), `window_matching` (W), `gap_basis`, `gap_n`
  - `dxadx_required_direction`: bool

## Algorithme (pseudo-code)

1) Normaliser bougies: trier, dédupliquer, exclure incomplètes
2) Calculer séries: VWMAShort/Long, DI+/DI−, ADX, DX, ATR3, MFI
3) Pour i* = dernière bougie fermée:
   - W := WindowMatching; winStart := max(1, i* − W + 1)
   - Bases:
     - t_cross := dernier croisement VWMA dans [winStart .. i*]; si none → stop
     - targetLong := (VWMAShort[i*] > VWMALong[i*])
     - DI aligné à i*; sinon → stop
     - last DX/ADX cross dans [winStart .. i*]; si none → stop
       - si `DXADXRequiredDirectionalCross` et (targetLong && !UP) ou (!targetLong && !DOWN) → stop
   - Gating à i* (pentes optionnelles, gap vs ATR3 si activé, MFI extrême positif selon la cible)
     - LONG: MFI ≥ overbought; SHORT: MFI ≤ oversold; sinon → stop
   - Edge-trigger:
     - Chercher j0 ∈ [t_cross .. i*] première bougie validant Bases+Gating
     - Si i* ≠ j0 → no signal; sinon émettre signal

## Valeurs par défaut recommandées

- VWMAShortPeriod: 10
- VWMALongPeriod: 20
- DMIPeriod: 14
- WindowMatching: 5
- ATRPeriod: 3
- EnableGapGating: true, GapATRMultiplier: 1.0, GapBasis: "vwma_spread"
- MFIPeriod: 14, Overbought: 80, Oversold: 20, MFIMode: "positive_extremes"
- Slopes: désactivées par défaut
- DXADXRequiredDirectionalCross: true
- EdgeTrigger: true

## Notes d’implémentation

- Utiliser les implémentations TV standard existantes: VWMA, DMI/ADX, ATR, MFI
- Les tableaux doivent propager NaN pour warm-up; les validations doivent traiter NaN comme invalidants
- Les pentes s’évaluent simple (Δ1 bougie) par défaut; éviter des bases dépendantes de lookback supplémentaires dans cette version simple

---

Ce document définit BAN_FIN_SIMPLE en retirant le CCI et en remplaçant la contrainte CCI par une confirmation positive basée sur MFI: LONG si MFI en zone d’extrême haussier, SHORT si MFI en zone d’extrême baissier, le tout dans le même cadre Edge-trigger + fenêtre W que BAN_FIN. 
