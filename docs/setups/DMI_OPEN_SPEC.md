# DMI Ouverture – Spécification du module réutilisable

## Objectif
- Détecter un setup DMI « ouverture » basé sur le croisement DX>ADX (edge‑trigger i−1→i), avec règles de gating configurables et re‑vérifications indépendantes jusqu’à +6 barres.
- Être réutilisable et améliorable indépendamment (pas d’I/O ni de logs directs, seulement des diagnostics).

## Hypothèses
- Le module consomme des séries déjà calculées: DI+, DI−, DX, ADX.
- Périodes DI et ADX sont gérées en amont (ex: DI sur `DMI_PERIOD`, ADX sur `ADX_PERIOD`).
- Les valeurs NaN invalident les validations (gating échoue si une mesure requise est NaN).

## Terminologie
- DI_sup / DI_inf
  - LONG: DI_sup = DI+, DI_inf = DI−.
  - SHORT: DI_sup = DI−, DI_inf = DI+.
- Croisement DX/ADX (edge‑trigger)
  - Croisement haussier de DX au‑dessus d’ADX sur i si `DX[i−1] <= ADX[i−1]` et `DX[i] > ADX[i]`.
  - Epsilon `eps` optionnel pour tolérance d’égalité (par défaut 0): `DX[i−1] <= ADX[i−1] + eps`.
- Contexte DI
  - LONG si `DI+[i] > DI−[i]`. SHORT si `DI−[i] > DI+[i]`. Pas d’exigence de croisement DI.
- Pentes (par défaut requises)
  - `DX[i] > DX[i−1]` et `ADX[i] > ADX[i−1]`.

## Règles de gating
- Relatives (activables):
  - ADX sous DI_sup: `ADX[i] < DI_sup[i]`.
  - DX au‑dessus de DI_inf: `DX[i] > DI_inf[i]`.
- Numériques (activables):
  - `DX[i] >= MinDX`.
  - `ADX[i] >= MinADX`.
  - `ADX[i] <= MaxADX`.
  - Distance minimale: `|DX[i] − ADX[i]| >= MinDXADXGap`.
- Echec de n’importe quelle règle active ⇒ gating échoue.

## Re‑vérifications indépendantes (lags)
- Lorsqu’un croisement est détecté à i mais que le gating échoue, on planifie des rechecks indépendants à i+L, pour chaque L activé dans {1,2,3,4,5,6}.
- À i+L, on ne re‑demande PAS un nouveau croisement ni les pentes; on ré‑évalue uniquement le gating sur les valeurs courantes.
- Nouveauté (intégrité pendant recheck):
  - On peut exiger que l’orientation DI reste conforme au côté recherché (LONG: DI+>DI−, SHORT: DI−>DI+).
  - On peut exiger que la relation DX>ADX reste vraie.
  - Si l’une de ces conditions s’inverse pendant le checking et que l’option d’annulation est activée, on annule (purge) tous les rechecks restants liés au même croisement (le signal s’est inversé).
- Si un recheck valide le gating (et satisfait les contraintes ci‑dessus si activées), on émet un unique évènement pour ce croisement avec `Lag=L`, et on abandonne les autres rechecks associés au même croisement.
- Un recheck tenté est consommé (pas de re‑planification), qu’il réussisse ou non.

## API proposée
- Package: `internal/setups/dmiopen`

Types:
- `Side` = {`LONG`, `SHORT`}.
- `Config`:
  - Gating relatifs:
    - `UseRelAdxUnderDiSup bool`
    - `UseRelDxOverDiInf bool`
  - Gating numériques:
    - `UseMinDX bool`, `MinDX float64`
    - `UseMinADX bool`, `MinADX float64`
    - `UseMaxADX bool`, `MaxADX float64`
    - `UseMinGapDXADX bool`, `MinGapDXADX float64`
  - Rechecks (indépendants):
    - `UseRecheck1 bool`, ..., `UseRecheck6 bool`
  - Intégrité pendant rechecks:
    - `RecheckRequireContextSide bool` (exiger DI côté recherché pendant recheck)
    - `RecheckRequireDXAboveADX bool` (exiger DX>ADX pendant recheck)
    - `RecheckCancelSiblingsOnFlip bool` (annuler tous les rechecks restants si inversion DI ou DX<=ADX)
  - Détection:
    - `RequireDXUp bool` (par défaut true)
    - `RequireADXUp bool` (par défaut true)
    - `Eps float64` (tolérance croisement, par défaut 0)
  - Logs (niveaux, fournis en Diag, pas de print dans le module):
    - `LogLevel int` (0=Off, 1=Summary, 2=Detailed, 3=Debug)
- `Inputs`:
  - `Index int` (i)
  - `DIpPrev, DImPrev, DXPrev, ADXPrev float64`
  - `DIp, DIm, DX, ADX float64`
- `Diagnostics`:
  - `DIp, DIm, DX, ADX float64`
  - `GapDXADX float64`
  - `Applied struct { RelAdxUnderDiSup, RelDxOverDiInf, MinDX, MinADX, MaxADX, MinGapDXADX bool }`
  - `Checks struct { DXUp, ADXUp, CrossEdge, ContextLong, ContextShort bool }`
- `Event`:
  - `Triggered bool`
  - `Side Side`
  - `Lag int` // 0 (immédiat) ou 1..6
  - `Index int` // i de validation
  - `CrossAt int` // index du croisement initial
  - `Diag Diagnostics`

Méthodes:
- `func New(cfg Config) *DmiOpen`
- `func (d *DmiOpen) Reset()`
- `func (d *DmiOpen) Step(in Inputs) (Event, bool)`

## Comportement de Step
1) Rechecks planifiés pour `in.Index`:
   - Pour chaque pending dont `recheckAt == in.Index`:
     - Si `RecheckRequireContextSide` ⇒ vérifier orientation DI conforme au `Side`.
     - Si `RecheckRequireDXAboveADX` ⇒ vérifier `DX>ADX`.
     - Si l’une échoue et `RecheckCancelSiblingsOnFlip` ⇒ purger tous les pendings du même `crossIndex` (annulation), consommer le courant.
     - Sinon, évaluer le gating uniquement.
   - Si validé et aucun évènement émis encore à cette bougie, émettre `Event{Triggered:true, Side:pending.side, Lag:pending.lag, Index:i, CrossAt:pending.crossIndex}` et purger les autres pendings du même `crossIndex`.
   - Sinon, consommer le pending (ne pas re‑planifier).
2) Détection croisement sur i (si i>0):
   - Vérifier `prevDX <= prevADX + eps` et `curDX > curADX`.
   - Vérifier pentes si `RequireDXUp/RequireADXUp`.
   - Déterminer `Side` par contexte DI à i (DI_sup > DI_inf).
   - Évaluer le gating:
     - Si OK → `Lag=0`, émettre Event immédiat.
     - Sinon → planifier rechecks i+L pour chaque `UseRecheckL` activé (L∈{1..6}).

## Logs (exploités par l’appelant selon LogLevel)
- Level 1 (Summary): `time/index, side, lag`.
- Level 2 (Detailed): Summary + valeurs `DIp, DIm, DX, ADX, GapDXADX`, seuils actifs et valeurs.
- Level 3 (Debug): Detailed + statut des checks (DXUp, ADXUp, CrossEdge, ContextLong/Short, chaque règle de gating true/false).

## Valeurs par défaut (recommandées démo)
- Relatifs: `UseRelAdxUnderDiSup=true`, `UseRelDxOverDiInf=false`.
- Numériques: `UseMinDX=false`, `UseMinADX=false`, `UseMaxADX=true (MaxADX=25)`, `UseMinGapDXADX=false (ou true si besoin, ex 10.0)`.
- Rechecks: `UseRecheck1=true`, `UseRecheck2=true`, `UseRecheck3=true`, `UseRecheck4=false`, `UseRecheck5=false`, `UseRecheck6=false`.
- Détection: `RequireDXUp=true`, `RequireADXUp=true`, `Eps=0`.
- Intégrité rechecks: `RecheckRequireContextSide=true`, `RecheckRequireDXAboveADX=true`, `RecheckCancelSiblingsOnFlip=true`.
- Logs: `LogLevel=1` (summary) par défaut côté démo.

## Intégration démo (main.go)
- Mapping des constantes existantes vers `Config`.
- Remplacer la logique inline par une instance `dmiopen.DmiOpen` et des appels `Step`.
- Le main garde la responsabilité d’imprimer les logs en fonction du `LogLevel` sélectionné, à partir d’`Event.Diag`.

## Pseudocode
```
mod := dmiopen.New(cfg)
for i := 1; i < n; i++ {
  evt, ok := mod.Step(Inputs{
    Index:i,
    DIpPrev:diPlus[i-1], DImPrev:diMinus[i-1], DXPrev:dx[i-1], ADXPrev:adx[i-1],
    DIp:diPlus[i], DIm:diMinus[i], DX:dx[i], ADX:adx[i],
  })
  if ok {
    // Logs selon LogLevel + émission ORIENTATION (LONG/SHORT) dans la sortie
  }
}
```

## Notes
- Le module n’a pas de dépendance temporelle réelle: `Index` sert d’identifiant séquentiel.
- On peut empêcher l’explosion des pendings en limitant à `{i+1..i+6}` et en purgeant à l’émission.
