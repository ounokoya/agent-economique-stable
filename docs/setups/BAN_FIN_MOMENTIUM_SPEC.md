# BAN_FIN_MOMENTIUM – Spécification du setup (structure ban_fin, datasource Bybit)

## Objectif
- Détecter un setup « momentum » scalping/ban_fin pour l’ouverture et la fermeture de positions, basé sur le couple tendance (CCI long) / timing (CCI court + Stoch cross), avec information de force (Body/ATR, Volume/SMA) utilisée pour la confiance.
- Respecter la structure et le style de `internal/signals/ban_fin` (évaluation de la dernière bougie fermée, métadonnées, snapshot de gating), et utiliser Bybit pour la démo/paper/live ; le backtest utilise les données téléchargées en local.

## Hypothèses
- Le module consomme des klines OHLCV et calcule en interne ATR(10), SMA(volume), Stoch (K/D), MFI, CCI.
- Les valeurs NaN invalides rejettent la détection (gating échoue si une mesure requise est NaN).
- Le setup peut fonctionner en mode « bougie simple » ou « agrégation 3 bougies ». Par défaut: « bougie simple ».
 - Les indicateurs de tendance (CCI long et, si utilisé, VWMA) sont conçus pour être calculés sur un timeframe supérieur à celui des signaux. Exemples typiques de couples: signaux en 1m avec tendance en 15m, signaux en 5m avec tendance en 1h, signaux en 15m avec tendance en 4h.

## Terminologie
- Body: |Close − Open| de la bougie (ou agrégée 3 bougies, cf. Agrégation).
- Body/ATR: Body / ATR(10).
- Volume normalisé: Volume courant; en agrégation: moyenne sur les 3 bougies.
- Volume/SMA: Volume normalisé / SMA(volume, `VolumeSMAPeriod`).
- Bougie verte: Close > Open. Bougie rouge: Close < Open. Doji (Close == Open) ignoré.
- Agrégation 3 bougies (optionnelle):
  - Open agrégé = Open(i−2), Close agrégé = Close(i), Volume agrégé = moyenne(Volume[i−2..i]).
  - Couleur agrégée = signe(Close(i) − Open(i−2)).

## Règles de détection
Filtre de tendance (appliqué aux signaux directionnels, optionnel mais activé par défaut):
- Tendance principale par CCI long (`CCITrendPeriod`), typiquement calculé sur un timeframe supérieur à celui des signaux.
- Si `EnableCCITrendGate = true`:
  - Pour un signal LONG: `CCI_trend > 0`.
  - Pour un signal SHORT: `CCI_trend < 0`.

### Signaux LONG / SHORT

Les signaux sont construits en quatre couches, dont trois filtres activables indépendamment.

1. **Croisement Stoch (obligatoire)**
   - Un signal n’est envisagé que s’il y a croisement de Stoch:
     - Signal LONG potentiel si K croise **à la hausse** D (K passe de <D à >D).
     - Signal SHORT potentiel si K croise **à la baisse** D (K passe de >D à <D).

2. **Filtre de tendance (CCI_trend)**
   - On ne retient le signal que si la tendance CCI long (calculé sur le timeframe de tendance) est cohérente:
   - LONG: `CCI_trend > 0`.
   - SHORT: `CCI_trend < 0`.
   - VWMA fast/slow peut s’ajouter en filtre de tendance si `EnableVWMATrendGate=true`.

3. **Filtre "analyse de barre" (optionnel, gate activable)**
   - Ce filtre valide le croisement Stoch par la structure de la bougie. Deux modes sont possibles:
     - Analyse **bougie simple**.
     - Analyse **agrégation 3 bougies** (si `Aggregate3=true`).
   - Si le filtre est **activé**, on exige par exemple:
     - Body/ATR > `BodyATRMultiplier`.
     - Volume normalisé / SMA(volume) > `VolumeCoeff`.
     - Couleur cohérente avec le sens du signal (verte pour LONG, rouge pour SHORT, en mode simple ou agrégé).
   - Si le filtre est **désactivé**, Body/ATR et Volume/SMA ne rejettent pas la bougie (restent utilisés pour la confiance et le diagnostic).

4. **Filtre CCI court (optionnel)**
   - CCI court sert à détecter un **extrême inverse de la tendance**, pour entrer en pullback:
     - En tendance LONG: on valide le signal uniquement si le CCI court est en **survente LONG** (extrême inférieur),
       typiquement `CCI_short <= CCIOversoldLong`.
     - En tendance SHORT: on valide le signal uniquement si le CCI court est en **surachat SHORT** (extrême supérieur),
       typiquement `CCI_short >= CCIOverboughtShort`.
   - Si le filtre CCI est désactivé, le CCI court n’intervient pas dans le gating (il reste disponible en métadonnée).

5. **Filtre MFI (optionnel)**
   - MFI suit la même logique que le CCI court, en extrême inverse de tendance:
     - En tendance LONG: validation sur **MFI oversold LONG**.
     - En tendance SHORT: validation sur **MFI overbought SHORT**.
   - Le filtre MFI peut être activé ou désactivé indépendamment du filtre CCI.

En résumé, chaque signal LONG/SHORT est:
- obligatoirement déclenché par un croisement Stoch,
- filtré par la tendance CCI_trend,
- éventuellement filtré par l’analyse de barre (Step1),
- éventuellement filtré par CCI court extrême inverse de la tendance,
- éventuellement filtré par MFI extrême inverse de la tendance.

Ces signaux sont **directionnels** (LONG ou SHORT) et servent autant pour **ouvrir** que pour **fermer** une position :
- en absence de position, un signal LONG/SHORT est interprété comme une **entrée** dans ce sens,
- si une position existe dans le sens opposé, un signal inverse est interprété comme une **sortie** de cette position (et éventuellement une inversion de position, selon la logique de la couche d’exécution).

## Trailing stop (exécution, non géré dans le générateur)
- Le stop est un **trailing stop standard en pourcentage du prix d'entrée**, implémenté dans `internal/execution.PercentTrailing`.
- Paramètre principal de la démo: `TRAIL_PCT` (ex: `0.003` = `0,3%`) défini dans `cmd/ban_fin_momentium/main.go`.
- À l'ouverture d'une position:
  - LONG: `Trail = entry_price × (1 − TRAIL_PCT)`, et `maxPrice = entry_price`.
  - SHORT: `Trail = entry_price × (1 + TRAIL_PCT)`, et `minPrice = entry_price`.
- À chaque nouvelle bougie (dans la démo: Close 5m):
  - LONG:
    - Si `price > maxPrice`, on met à jour `maxPrice = price` et `Trail = maxPrice × (1 − TRAIL_PCT)`.
    - Le stop ne peut que **monter** dans le sens de la position (il ne redescend jamais).
  - SHORT:
    - Si `price < minPrice`, on met à jour `minPrice = price` et `Trail = minPrice × (1 + TRAIL_PCT)`.
    - Le stop ne peut que **descendre** dans le sens de la position (il ne remonte jamais).
- Condition de déclenchement (hit) dans la démo:
  - LONG: fermeture si `Close <= Trail`.
  - SHORT: fermeture si `Close >= Trail`.
- Le trailing stop est appliqué dans la couche d'exécution de la démo (backtest/paper), après les signaux du générateur, sans feed‑back vers le générateur.

## API proposée (structure « ban_fin »)
- Package: `internal/signals/ban_fin_momentium`

Types/Config:
- `Config`:
  - Momentum/volume:
    - `ATRPeriod` (int, défaut 10)
    - `BodyATRMultiplier` (float64)
    - `VolumeSMAPeriod` (int)
    - `VolumeCoeff` (float64)
    - `Aggregate3` (bool)
  - VWMA (filtre tendance):
    - `VWMAFastPeriod`, `VWMASlowPeriod` (int)
    - `EnableVWMATrendGate` (bool)
  - Drapeaux de setup (indépendants):
    - `EnableOpenLong`, `EnableOpenShort`, `EnableCloseLong`, `EnableCloseShort` (bool)
  - Stoch:
    - `StochKPeriod`, `StochKSmooth`, `StochDPeriod` (int)
    - Filtre Stoch optionnel: `EnableStochCross` (croisement K/D) utilisé pour le **timing directionnel** (LONG: cross haussier, SHORT: cross baissier)
    - Extrêmes Stoch K (optionnels, liés à la tendance et à la direction):
      - Bornes LONG: `StochKOversoldLong`, `StochKOverboughtLong` avec toggles `EnableStochKOversoldLongGate`, `EnableStochKOverboughtLongGate`.
        - Si un de ces toggles est activé, la condition s’applique uniquement aux signaux LONG.
        - Le générateur n’accepte le signal LONG que si `K` est **inférieur ou égal** à chacune des bornes actives (elles jouent le rôle de **borne supérieure** de la zone autorisée).
        - Exemple d’usage typique: tendance haussière, on n’accepte que les croisements haussiers avec `K` en dessous d’un seuil « trop haut » (par ex. `K <= 80`).
      - Bornes SHORT: `StochKOversoldShort`, `StochKOverboughtShort` avec toggles `EnableStochKOversoldShortGate`, `EnableStochKOverboughtShortGate`.
        - Si un de ces toggles est activé, la condition s’applique uniquement aux signaux SHORT.
        - Le générateur n’accepte le signal SHORT que si `K` est **supérieur ou égal** à chacune des bornes actives (elles jouent le rôle de **borne inférieure** de la zone autorisée).
        - Exemple d’usage typique: tendance baissière, on n’accepte que les croisements baissiers avec `K` au-dessus d’un seuil « trop bas » (par ex. `K >= 20`).
      - Il est recommandé de ne pas activer simultanément toutes les bornes pour une même direction, mais de choisir une configuration cohérente avec la tendance:
        - en tendance LONG: typiquement, croisement haussier + `CCI_trend > 0` + `K` sous une borne haute (par ex. 80);
        - en tendance SHORT: croisement baissier + `CCI_trend < 0` + `K` au-dessus d’une borne basse (par ex. 20).
  - MFI/CCI (4 extrêmes par indicateur et par tendance, pour CCI court):
    - `MFIPeriod`
    - LONG: `MFIOversoldLong`, `MFIOverboughtLong`
    - SHORT: `MFIOversoldShort`, `MFIOverboughtShort`
    - `CCIPeriod`
    - LONG: `CCIOversoldLong`, `CCIOverboughtLong`
    - SHORT: `CCIOversoldShort`, `CCIOverboughtShort`
    - Filtres optionnels: `EnableMFIGate`, `EnableCCIGate`
  - CCI tendance (CCI long):
    - `CCITrendPeriod`
    - Filtre optionnel: `EnableCCITrendGate` (tendance par le signe de `CCI_trend`, sans notion de range)

- `Generator`:
  - `Name() string`
  - `Initialize(config signals.GeneratorConfig) error`
  - `CalculateIndicators(klines []signals.Kline) error`
  - `EvaluateLast(klines []signals.Kline) (*signals.Signal, error)`
    - Évalue uniquement la dernière bougie fermée (style `ban_fin`).
    - Retourne au plus un **signal directionnel** par bougie (Action=ENTRY, Type=LONG ou SHORT) avec métadonnées.
  - `GetMetrics() signals.GeneratorMetrics`

- Métadonnées (exemples):
  - `generator: "ban_fin_momentium"`, `mode: "MOMENTUM"`, `agg3: bool`
  - `body`, `atr10`, `body_to_atr`, `volume`, `vol_sma`, `vol_to_sma`
  - `vwma_fast`, `vwma_slow`, `stoch_k`, `stoch_d`, `mfi`, `cci`, `cci_trend`
  - `gating_snapshot` (JSON) optionnel pour traçabilité

### Invariant de configuration (aucune logique de gating par défaut)

- Le générateur BAN_FIN_MOMENTIUM **ne définit aucun filtre implicite** en dehors de ce qui est piloté par la Config.
- Les seules valeurs « par défaut » sont celles des champs de `Config` et des constantes dans la démo; il n’existe pas de logique interne qui activerait un gate lorsque son toggle est à `false`.
- En particulier:
  - si `EnableBarGate=false`, Body/ATR et Volume/SMA **n’interviennent pas** dans le gating (ils restent disponibles en métadonnée et pour la confiance);
  - si `EnableStochCross=false` et que tous les `EnableStochK*Gate` sont `false`, Stoch K/D n’intervient pas dans le gating;
  - si `EnableMFIGate=false` et tous les toggles extrêmes MFI sont `false`, MFI n’intervient pas dans le gating;
  - si `EnableCCIGate=false` et tous les toggles extrêmes CCI sont `false`, le CCI court n’intervient pas dans le gating;
  - si `EnableVWMATrendGate=false` et `EnableVWMACross=false`, VWMA ne filtre pas les signaux (hors calcul éventuel pour métadonnées);
  - si `EnableMACD*Gate=false` pour tous les gates MACD, MACD ne filtre pas les signaux.
- Si **aucun gate** n’est actif (c’est‑à‑dire: `EnableBarGate=false`, `EnableStochCross=false`, tous les toggles Stoch K/MFI/CCI à `false`, `EnableVWMATrendGate=false`, `EnableVWMACross=false`, tous les gates MACD à `false`), alors:
  - `EvaluateLast` **ne produit aucun signal** (retourne toujours `nil` tant qu’aucun gate n’est activé);
  - `DetectSignals` ne renvoie aucun signal sur le flux;
  - le `WindowFinder` (mode fenêtre) ne produit aucun signal non plus.
- Autrement dit, **toute la logique de gating est entièrement pilotée par la Config**: si vous désactivez tous les filtres dans la Config/démo, le setup BAN_FIN_MOMENTIUM devient silencieux (0 signal) et ne rajoute aucune règle implicite.

## Comportement de EvaluateLast
1) Vérifier que les indicateurs requis sont valides à la dernière bougie fermée, uniquement pour les filtres activés (StochK/StochD si `EnableStochCross`, MFI si `EnableMFIGate`, CCI court si le filtre CCI est activé, CCI long si `EnableCCITrendGate`). ATR et SMA(volume) sont requis pour le calcul de Body/ATR et Volume/SMA.
2) Calculer Body/ATR et Volume/SMA en mode simple ou agrégé (selon `Aggregate3`). Si le filtre d’analyse de barre est activé, ces mesures (Body/ATR, Volume/SMA, couleur) participent au gating; sinon, elles sont utilisées uniquement pour la confiance et le diagnostic.
3) Évaluer la tendance via `CCI_trend` (signe, et éventuellement VWMA si `EnableVWMATrendGate`) pour déterminer si la bougie est éligible côté LONG ou SHORT.
4) Vérifier la présence d’un croisement Stoch dans le sens de la tendance si `EnableStochCross`.
5) Appliquer ensuite les filtres optionnels activés: analyse de barre, CCI court extrême inverse de la tendance, MFI extrême inverse de la tendance.
6) Si toutes les conditions sont réunies, émettre au plus un signal directionnel par bougie (Action=ENTRY, Type=LONG ou SHORT).
7) Renseigner `Confidence` via une heuristique simple (par exemple pondération de Body/ATR et Volume/SMA), et enrichir `Metadata` avec les indicateurs disponibles, ainsi qu’un éventuel `gating_snapshot` pour traçabilité.

## Résumé logique (vue d’ensemble)
- La bougie est considérée **éligible LONG** si:
  - `CCI_trend` (calculé sur le timeframe de tendance) indique une tendance haussière (signe > 0),
  - le cross Stoch est haussier (si le cross est activé),
  - et tous les filtres optionnels activés (analyse de barre, CCI court/MFI extrêmes inverses) valident le signal.
- La bougie est considérée **éligible SHORT** si:
  - `CCI_trend` indique une tendance baissière (signe < 0),
  - le cross Stoch est baissier (si le cross est activé),
  - et tous les filtres optionnels activés valident le signal.
- Les signaux sont **directionnels** (LONG ou SHORT) et servent à la fois à l’ouverture et à la fermeture via signal inverse dans la couche d’exécution; les sorties explicites (EXIT) ne font pas partie de cette spec.
- Body/ATR et Volume/SMA sont des mesures de qualité/force du signal; lorsque le filtre d’analyse de barre est activé, elles participent au gating, sinon elles servent uniquement à la confiance.

## Mode fenêtre optionnel (WindowMode)

En plus du mode « bougie simple » décrit ci‑dessus (évaluation de la dernière bougie fermée uniquement), le setup BAN_FIN_MOMENTIUM peut fonctionner en **mode fenêtre** optionnel. Ce mode permet de laisser plusieurs bougies à un setup pour se compléter avant d’émettre un signal.

- Paramètres conceptuels:
  - `EnableWindowMode` (bool):
    - si `false`, le comportement reste celui de `EvaluateLast` pur (toutes les conditions sont évaluées sur la même bougie);
    - si `true`, un Finder par fenêtre est utilisé pour orchestrer les signaux.
  - `WindowBars` (int): taille de la fenêtre en nombre de bougies (par défaut 10).
  - `WindowOpenMode` (string): stratégie d’ouverture de fenêtre, avec par exemple les valeurs suivantes:
    - `"auto"` (défaut): si au moins un croisement est activé, la fenêtre s’ouvre sur le premier croisement dans le sens de la direction; sinon, elle s’ouvre sur l’analyse de barre si `EnableBarGate` est actif et valide.
    - `"cross"`: la fenêtre ne peut s’ouvrir que sur un croisement d’un indicateur dont le gate de cross est activé (Stoch, VWMA, MACD); si aucun croisement n’est actif, aucun setup par fenêtre ne sera déclenché.
    - `"bar"`: la fenêtre ne s’ouvre que via l’analyse de barre (bar gate), même si des gates de croisement sont activés.

- Ouverture de fenêtre (par direction LONG/SHORT):
  - Si au moins un gate de croisement est activé (`EnableStochCross`, `EnableVWMACross`, `EnableMACDCrossGate`):
    - une fenêtre LONG s’ouvre lorsqu’un premier croisement haussier se produit (Stoch, VWMA ou MACD) dans le sens LONG;
    - une fenêtre SHORT s’ouvre lorsqu’un premier croisement baissier se produit dans le sens SHORT;
    - pour chaque indicateur à croisement, la direction de son **dernier cross** observé dans cette fenêtre est mémorisée.
  - Si aucun croisement n’est activé:
    - la fenêtre s’ouvre via l’analyse de barre si `EnableBarGate` est actif et si la bougie est valide dans la direction (par exemple: bougie verte avec Body/ATR et Volume/SMA au‑dessus des seuils pour LONG, rouge pour SHORT).
  - À l’ouverture, on mémorise un indice de départ (bougie d’ouverture) et une échéance `WindowBars` bougies plus loin; la fenêtre est donc définie sur un intervalle de 10 bougies par défaut.

- Vie de la fenêtre:
  - Pour chaque bougie incluse dans la fenêtre, le Finder évalue, pour la direction considérée (LONG ou SHORT):
    - la cohérence de la tendance CCI_trend (et éventuellement VWMA trend si `EnableVWMATrendGate` est actif);
    - les filtres d’analyse de barre (body/ATR, volume/SMA) si `EnableBarGate` est actif;
    - les extrêmes Stoch K, CCI, MFI selon leurs bornes et toggles directionnels, lorsque leurs filtres sont activés;
    - les gates MACD (signe de la ligne et de la ligne signal, histogramme) lorsque `EnableMACDSignGate` ou `EnableMACDHistGate` sont activés;
    - les nouveaux croisements Stoch/VWMA/MACD, en mettant à jour la direction du **dernier cross** pour chaque indicateur à croisement utilisé dans la fenêtre.
  - Les filtres dont les toggles sont à `false` ne participent pas au rejet du signal (ils restent disponibles en métadonnée et pour le diagnostic).

- Validation d’un signal dans la fenêtre:
  - À une bougie donnée de la fenêtre, un signal directionnel (LONG ou SHORT) peut être validé si, pour la direction de la fenêtre:
    - toutes les conditions de tendance activées (CCI_trend, VWMA trend) sont satisfaites sur cette bougie;
    - tous les filtres activés (analyse de barre, Stoch K extrêmes, CCI, MFI, MACD sign/histogramme) sont satisfaits sur cette bougie;
    - pour chaque indicateur à croisement activé dans la configuration, le **dernier croisement observé dans la fenêtre** est dans le même sens que celui qui a ouvert le setup (par exemple, dernier cross haussier pour une fenêtre LONG).
  - La bougie sur laquelle ces conditions sont pour la première fois réunies devient la **barre de validation**: le Finder émet alors un signal ENTRY (LONG ou SHORT) sur cette bougie et la fenêtre est fermée.

- Expiration de la fenêtre:
  - Si aucune bougie de l’intervalle de travail de la fenêtre n’atteint l’ensemble des conditions requises, aucun signal n’est émis;
  - à l’échéance (après `WindowBars` bougies), la fenêtre expire silencieusement et le setup est réinitialisé en attente d’un nouveau déclencheur (nouveau croisement ou nouvelle validation d’analyse de barre, selon la configuration).

Ce mode fenêtre est **optionnel** et ne remplace pas la logique de base de BAN_FIN_MOMENTIUM; il offre une façon supplémentaire de valider les setups en autorisant un délai de complétion sur plusieurs bougies, tout en garantissant la cohérence des derniers croisements et filtres au moment de la validation.

## Notes
- La définition « Volume > coefficient volume » est implémentée comme `Volume > VolumeCoeff × SMA(volume)` par défaut; peut être remplacée par une autre base (EMA/median) selon vos préférences.
- L’évaluation est strictement sur la dernière bougie fermée (style `ban_fin`).
- Le trailing est séparé (moteur d’exécution de la démo) pour préserver la réutilisabilité du générateur.

## Démo / Backtest (Bybit)
- Binaire: `cmd/ban_fin_momentium/main.go`
- Datasource: `internal/datasource/bybit` (timeframe 1m/5m). Paramètres: symbole, nombre de bougies, périodes indicateurs, coefficients.
- Flux backtest:
  1. Récupérer klines Bybit, trier chrono, convertir en `signals.Kline`.
  2. `Initialize` → `CalculateIndicators` → `EvaluateLast` en boucle (bougie par bougie).
  3. Exécuter une couche de simulation avec trailing stop (règle ci‑dessus), une position à la fois.
  4. Logs: tableau des signaux, tableau des intervalles ENTRY→EXIT, statistiques (captures brutes et orientées).
  5. Export JSON (klines, intervalles, signaux) dans `out/`.

## Valeurs par défaut (recommandées pour démo)
- `ATRPeriod=10`, `BodyATRMultiplier=1.0`
- `VolumeSMAPeriod=20`, `VolumeCoeff=1.2`
- `Aggregate3=false` (analyse bougie unique)
- Setups: `EnableOpenLong=true`, `EnableOpenShort=true`, `EnableCloseLong=true`, `EnableCloseShort=true`
- VWMA: `VWMAFastPeriod=60`, `VWMASlowPeriod=240`, `EnableVWMATrendGate=false` (VWMA disponible mais non utilisé par défaut)
- Stoch: `14,2,3`, `EnableStochCross=true` (cross utilisé comme trigger d’entrée)
- MFI: `Period=30`, LONG: `MFIOversoldLong=20`, `MFIOverboughtLong=80`; SHORT: `MFIOversoldShort=20`, `MFIOverboughtShort=80`; `EnableMFIGate=false`
- CCI court: `Period=20`, LONG: `CCIOversoldLong=-100`, `CCIOverboughtLong=100`; SHORT: `CCIOversoldShort=-100`, `CCIOverboughtShort=100`; `EnableCCIGate=false`
- CCI tendance: `CCITrendPeriod=100`, `EnableCCITrendGate=true` (tendance par le signe de `CCI_trend`, sans range de gating).
- Trailing standard en %: `TRAIL_PCT=0.003` (0,3% de trailing autour du prix d'entrée) dans la démo BAN_FIN_MOMENTIUM.

## Résumé logique (vue d’ensemble)
- La bougie est considérée **éligible LONG** si:
  - CCI_trend indique une tendance haussière (signe > 0),
  - le cross Stoch est haussier (si le cross est activé),
  - et tous les filtres optionnels activés (analyse de barre, CCI court/MFI extrêmes inverses, autres gates) valident le signal.
- La bougie est considérée **éligible SHORT** si:
  - CCI_trend indique une tendance baissière (signe < 0),
  - le cross Stoch est baissier (si le cross est activé),
  - et tous les filtres optionnels activés valident le signal.
- Les signaux sont **directionnels** (LONG ou SHORT) et servent à la fois à l’ouverture et à la fermeture via signal inverse dans la couche d’exécution; les sorties explicites (EXIT) ne font pas partie de cette spec.
- Body/ATR et Volume/SMA sont des mesures de qualité/force du signal; lorsque le filtre d’analyse de barre est activé, elles participent au gating, sinon elles servent uniquement à la confiance.
