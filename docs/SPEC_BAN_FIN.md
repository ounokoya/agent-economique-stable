# 📋 SPÉCIFICATION STRATÉGIE BAN_FIN

**Date**: 2025-11-08  
**Version**: 1.0  
**Auteur**: Agent Économique Stable

---

## 🎯 PRINCIPE GÉNÉRAL

La stratégie **BAN_FIN** combine :
1. **Croisement VWMA court/long** : Détection de changement de direction
2. **Position relative DI** : Force directionnelle (DI+ vs DI-)
3. **Croisement DX/ADX** : Évolution de la force (augmente ou diminue)

**RÈGLE ABSOLUE** : 
- Croisement VWMA dicte TOUJOURS la direction du signal (LONG ou SHORT)
- DI position et DX/ADX croisement ajoutent qualification (Tendance ou Contre-Tendance)
- **Pas de signals ENTRY/EXIT distincts** : seulement des signaux LONG et SHORT
- **Le signal est contextuel** : ENTRY si aucune position, EXIT si position inverse ouverte

---

## 📊 LES 4 COMBINAISONS DE BASE

| # | VWMA Cross | DI Position | DX/ADX Cross | Signal Généré |
|---|------------|-------------|--------------|---------------|
| **1** | Court > Long ↗ | DI+ > DI- | DX croise ADX ↑ | **LONG** |
| **2** | Court > Long ↗ | DI- > DI+ | DX croise ADX ↓ | **LONG** |
| **3** | Court < Long ↘ | DI- > DI+ | DX croise ADX ↑ | **SHORT** |
| **4** | Court < Long ↘ | DI+ > DI- | DX croise ADX ↓ | **SHORT** |

### Interprétation

**Tendance** (DX croise ADX ↑) :
- VWMA cross et DI position vont dans le même sens
- La force dans cette direction augmente
- Signal fort, confirmé

**Contre-Tendance** (DX croise ADX ↓) :
- VWMA cross commence une direction, mais DI encore dans l'ancienne
- La force de l'ancienne direction diminue
- Signal d'anticipation de retournement

### Comportement contextuel des signaux

**Logique automatique** :
- **Si aucune position ouverte** → Signal = **ENTRY**
- **Si position LONG ouverte** → Signal SHORT = **EXIT LONG**
- **Si position SHORT ouverte** → Signal LONG = **EXIT SHORT**

**Exemple** :
- Position = NONE → Signal généré = **LONG** → Action = **ENTRY LONG**
- Position = LONG → Signal généré = **SHORT** → Action = **EXIT LONG**
- Position = SHORT → Signal généré = **LONG** → Action = **EXIT SHORT**

---

## 🟢 SIGNAUX LONG

### Signal 1 : LONG (Tendance)
**Conditions** :
- ✅ Croisement VWMA court > long (détection du croisement)
- ✅ DI+ > DI- (position relative, force haussière domine)
- ✅ DX croise ADX vers le haut (force haussière augmente)

**Comportement** :
- Si aucune position → **ENTRY LONG**
- Si position SHORT → **EXIT SHORT**

### Signal 2 : LONG (Contre-Tendance)
**Conditions** :
- ✅ Croisement VWMA court > long (détection du croisement)
- ✅ DI- > DI+ (position relative, force baissière domine ENCORE)
- ✅ DX croise ADX vers le bas (force baissière diminue)

**Comportement** :
- Si aucune position → **ENTRY LONG**
- Si position SHORT → **EXIT SHORT**

---

## 🔴 SIGNAUX SHORT

### Signal 3 : SHORT (Tendance)
**Conditions** :
- ✅ Croisement VWMA court < long (détection du croisement)
- ✅ DI- > DI+ (position relative, force baissière domine)
- ✅ DX croise ADX vers le haut (force baissière augmente)

**Comportement** :
- Si aucune position → **ENTRY SHORT**
- Si position LONG → **EXIT LONG**

### Signal 4 : SHORT (Contre-Tendance)
**Conditions** :
- ✅ Croisement VWMA court < long (détection du croisement)
- ✅ DI+ > DI- (position relative, force haussière domine ENCORE)
- ✅ DX croise ADX vers le bas (force haussière diminue)

**Comportement** :
- Si aucune position → **ENTRY SHORT**
- Si position LONG → **EXIT LONG**


---

## 🔄 FENÊTRE DE MATCHING

### Règle de validation obligatoire

**Les 3 conditions doivent être validées dans une fenêtre de W bougies consécutives, peu importe l'ordre** :

1. ✅ **VWMA Cross** : Croisement court/long détecté
2. ✅ **DI Position** : Position relative DI+ vs DI- (simple comparaison)
3. ✅ **DX/ADX Cross** : Croisement DX vs ADX détecté

### Paramètre de fenêtre
 
 - Fenêtre W (window_matching) paramétrable; valeur par défaut: 5 bougies.

### Logique de détection
 
 Le Finder détecte les 3 conditions de base dans la fenêtre W (ordre indifférent), définit t_cross comme la première bougie du croisement VWMA, et t_signal comme la bougie où la 3e condition est validée. À chaque bougie candidate i, si les 3 conditions de base sont présentes dans la fenêtre W glissante, il applique le gating à i: (1) pentes configurées avec le signe attendu; (2) gap ≥ n × ATR(3) selon la base de gap choisie. Si le gating est valide à i, le signal est émis à i; sinon, on réévalue à la bougie suivante tant que la fenêtre reste valide.
 
 - Gap suffisant: gap ≥ n × ATR(3) à la bougie candidate (base de gap configurable).

### Exemples

#### Signal LONG Tendance dans fenêtre W=5

**Bougie T0** : DX croise ADX ↑ ✅  
**Bougie T1** : Rien  
**Bougie T2** : VWMA court croise long vers haut ✅  
**Bougie T3** : Rien  
**Bougie T4** : DI+ > DI- (position) ✅  

**Résultat** : Signal LONG Tendance généré à **T4** (timestamp=T4, prix=Close[T4])

#### Conditions simultanées

**Bougie T0** : VWMA cross + DI position + DX/ADX cross tous validés ✅✅✅  
**Résultat** : Signal généré **immédiatement** à T0 (fenêtre = 1 bougie)

### Timestamp et prix du signal

- **Timestamp** : Toujours la dernière bougie de la fenêtre (quand 3ème condition validée)
- **Prix d'exécution** : Close de cette dernière bougie
- **Fenêtre référence** : Sauvegardée dans métadonnées pour debug

---

## 📐 PARAMÈTRES TECHNIQUES
 
 ### Paramètres VWMA
 
 - vwma_short_period: 10
 - vwma_long_period: 20
 - enable_slope_vwma_short: true
 - enable_slope_vwma_long: true
 - slope_vwma_short_min: 0.0
 - slope_vwma_long_min: 0.0
 - slope_basis_vwma: "delta_1_bougie" (alternatives: "spread_vwma")
 
 ### Paramètres DMI
 
 - dmi_period: 14
 - dmi_smooth: 14
 - enable_dmi_position_aligned: true
 
 ### Paramètres fenêtre
 
 - window_matching: 5
 
 ### Paramètres ATR / Gap
 
 - atr_period: 3
 - gap_atr_multiplier: 1.0
 - gap_basis: "vwma_spread" (|VWMAShort − VWMALong|)
 - enable_gap_gating: true
 
 ### Paramètres CCI
 
 - cci_period: 20
 - enable_cci_extremes: true
 - cci_overbought: +100
 - cci_oversold: -100
 - enable_slope_cci: false
 - slope_cci_min: 0.0
 
 ### Paramètres Pentes (toggles et seuils)
 
 - enable_slope_dx: false
 - enable_slope_adx: false
 - slope_dx_min: 0.0
 - slope_adx_min: 0.0
 
 ### Paramètres DX/ADX
 
 - enable_dx_adx_spread: false
 - dx_adx_spread_min: 0.0
 - dx_adx_required_directional_cross: true
 
 ### Paramètres Trailing Stop
 
 - enable_trailing_stop: true
 - ts_init_coef: 2.0  (trail_pct_init = clamp(ts_init_coef × ATR(3)_entry / entry_price, ts_min_pct, ts_max_pct))
 - ts_min_pct: 0.003  (0.3%)
 - ts_max_pct: 0.03   (3%)
 - ts_profit_threshold_pct: 0.03  (3%)
 - ts_profit_trail_factor: 0.3333  (≈ 1/3 du % de bénéfice)

---

## 📊 ALIGNEMENT SIGNAL/POSITION ET RÈGLES DE TRAILING STOP
  
### Alignement Signal/Position
  
- **Signal du même côté que la position** :
  - Si plusieurs signaux valides successifs du même type apparaissent, la position existante reste ouverte (pas de ré‑entrée, pas de pyramiding par défaut).
  - La position ne se ferme pas sauf déclenchement d'un stop (trailing ou autre règle de sortie explicite).
- **Signal opposé au côté de la position** :
  - Sortie immédiate de la position en cours.
  - Pas d'ouverture opposée automatique sur la même bougie; réévaluation ultérieure par le générateur.
- **Priorité** : EXIT par signal opposé ou par stop > tout autre signal.
  
### Règles de Trailing Stop
  
- **Activation** : optionnelle (enable_trailing_stop).
- **Initialisation à l'ouverture** :
  - trail_pct_init = clamp(ts_init_coef × ATR(3)_entry / entry_price, ts_min_pct, ts_max_pct).
  - trail_price_init = entry_price × (1 − trail_pct_init) pour LONG; entry_price × (1 + trail_pct_init) pour SHORT.
- **Avant le seuil de profit** : tant que profit_pct < ts_profit_threshold_pct, trailing_pct = trail_pct_init (inchangé).
- **Au‑delà du seuil de profit** : si profit_pct ≥ ts_profit_threshold_pct,
  - trailing_pct_candidate = ts_profit_trail_factor × profit_pct.
  - trailing_pct = clamp(max(trailing_pct, trailing_pct_candidate), ts_min_pct, ts_max_pct) si ts_monotonic_increase_only = true.
  - Monotonicité : trailing_pct ne diminue jamais; il n'est relevé que s'il dépasse la valeur courante.
- **Événements d'update autorisés** : uniquement sur croisement DMI contre‑tendance (croisement DI et croisement DX/ADX dans le sens opposé à la direction VWMA), si ts_update_on_dmi_countertrend_only = true.
- **Sortie par stop** : si le prix touche le niveau trailing calculé, fermeture immédiate de la position (même en présence d'un signal du même côté).
- **Désactivation** : si enable_trailing_stop = false, aucun trailing n'est appliqué.

---

## 🔍 DÉTECTION ET VALIDATION

### Étapes de détection

1. **Calculer indicateurs** :
   - VWMA court, VWMA long
   - DMI (DI+, DI-)
   - DX, ADX

2. **Détecter croisements/états sur chaque bougie** :
   - Croisement VWMA court vs long
   - Position relative DI+ vs DI-
   - Croisement DX vs ADX

3. **Appliquer fenêtre de matching** :
   - Vérifier si les 3 conditions sont présentes dans fenêtre W
   - Classifier le signal combiné

4. **Générer signal final** :
   - Toujours générer si conditions validées (pas de filtres)

---

## 🎨 MÉTADONNÉES DES SIGNAUX

---

## ⚠️ RÈGLES IMPORTANTES

### 1. VWMA Cross = Direction absolue
- VWMA court > long → LONG uniquement
- VWMA court < long → SHORT uniquement
- **JAMAIS** de LONG si VWMA court < long
- **JAMAIS** de SHORT si VWMA court > long

### 2. Logique contextuelle unique
- **Signal LONG** : ENTRY si aucune position, EXIT si position SHORT
- **Signal SHORT** : ENTRY si aucune position, EXIT si position LONG
- **Un seul signal actif** à la fois selon contexte de position

### 3. Priorité automatique
- **EXIT > ENTRY** implicite par contexte (un signal inverse ferme toujours)
- Si plusieurs signaux valides même direction : prendre priorité 1 (Tendance) puis 2 (Contre-Tendance)

### 4. Gestion positions
- **Une seule position à la fois** (par défaut)
- Fermer position existante avant ouvrir nouvelle

---

## 📊 EXEMPLE COMPLET

### Scénario : Retournement haussier

**T0 : Baisse en cours**
- VWMA court < long
- DI- > DI+ (position)
- DX croise ADX ↑ (force baisse augmente)
- → **Signal SHORT** → Si aucune position = **ENTRY SHORT**

**T1 : VWMA se retourne**
- VWMA **court > long** (croisement)
- DI- > DI+ **encore** (position)
- DX croise ADX ↓ (force baisse diminue)
- → **Signal LONG** → Si position SHORT = **EXIT SHORT**

**T2 : DMI confirme**
- VWMA court > long
- DI+ > DI- (position bascule)
- DX croise ADX ↑ (force hausse augmente)
- → **Signal LONG** → Si aucune position = **ENTRY LONG**

---

## 🧩 ARCHITECTURE MODULAIRE

### Modules d'indicateurs — Filtres et Croisements

- VWMA
  - Entrées: Close, Volume, vwma_short_period, vwma_long_period
  - Filtres (états):
    - Direction: VWMADirection = "UP" si VWMAShort >= VWMALong, sinon "DOWN"
    - Pentes: calculées pour VWMAShort et VWMALong; signe attendu configurable par direction/mode. Gating par bougie: vérifiées à la bougie candidate; si invalides, pas d'émission et réévaluation aux bougies suivantes; aucune exigence de validité continue. Base de pente: différence 1-bougie (par défaut) ou spread (VWMAShort − VWMALong). Seuil minimal optionnel par série.
  - Croisements (événements):
    - CrossUp: VWMAShort passe de <= VWMALong à > VWMALong
    - CrossDown: VWMAShort passe de >= VWMALong à < VWMALong
  - Sorties: VWMAShort, VWMALong, VWMACrossDetected, VWMADirection

- DMI
  - Entrées: High, Low, Close, dmi_period, dmi_smooth
  - Filtres (états):
    - DIDominant = "DI_PLUS" si DIPlus > DIMinus, sinon "DI_MINUS"
  - Croisements (optionnels, non requis par les règles):
    - DIPlusCrossDIMinus: DI+ croise DI- (peut servir à l'analyse, pas utilisé par le Finder)
  - Sorties: DIPlus, DIMinus, DIDominant

- DX/ADX
  - Entrées: High, Low, Close, dmi_period, dmi_smooth
  - Filtres (états):
    - Intensité: DX et ADX bruts (aucun seuil imposé par la spec)
   - Pentes (états):
     - slopeDX, slopeADX avec signe attendu configurable par direction/mode. Gating par bougie: vérifiées à la bougie candidate; pas d'exigence de validité continue.
   - Écart (état):
     - spreadDXADX = |DX − ADX|, avec signe/dominance conforme au sens du croisement (après CrossUp: DX > ADX; après CrossDown: ADX > DX). Seuil minimal optionnel (dx_adx_spread_min). Gating par bougie: vérifié à la bougie candidate; pas d'exigence de validité continue.
    - Croisements (événements):
      - CrossUp: DX passe de <= ADX à > ADX
      - CrossDown: DX passe de >= ADX à < ADX
    - Sorties: DX, ADX, DXCrossADX, DXCrossDirection

- CCI
  - Entrées: Close
  - Paramètre: cci_period
  - Filtres (états):
    - Extrêmes: surachat/survente via seuils cci_overbought et cci_oversold. Gating par bougie: pas de LONG en surachat, pas de SHORT en survente à la bougie candidate; réévaluation aux bougies suivantes tant que la fenêtre reste valide.
    - CCI brut (si utilisé pour d'autres filtres)
  - Pentes (états):
    - slopeCCI avec signe attendu configurable par direction/mode. Gating par bougie: vérifiée à la bougie candidate; pas d'exigence de validité continue.
  - Sorties: CCI

- ATR (3)
  - Entrées: High, Low, Close
  - Paramètre: période = 3
  - Sorties: ATR3

### Module de recherche de signaux (BAN_FIN Finder)

- Entrées: états/croisements des modules, ATR3, window_matching, gap_atr_multiplier (n), position courante (optionnelle)
 - Règles:
  - Les 3 conditions de base (VWMA cross, DI position, DX/ADX cross) doivent apparaître dans une fenêtre W, ordre indifférent.
  - Fenêtre post-croisement: définir t_cross comme la première bougie où le croisement VWMA est détecté. Le signal est daté t_signal (bougie où la 3e condition de base est validée). Le segment d'évaluation est [t_cross, t_signal].
  - Gating par bougie: à chaque bougie i, si les 3 conditions de base sont présentes dans la fenêtre W glissante, alors vérifier pentes/gap/CCI à i. Émettre un signal daté i si tout est valide; sinon, attendre i+1 tant que la fenêtre continue de contenir les 3 conditions. Si la fenêtre expire sans bougie valide, aucun signal.
   - Pentes: pour chaque série activée (VWMAShort, VWMALong, DX, ADX, CCI), la pente respecte le signe attendu (seuil minimal optionnel) à la bougie candidate.
   - Gap: gap ≥ n × ATR(3) à la bougie candidate. Base par défaut: |VWMAShort − VWMALong|; alternatives configurables: distance prix vs VWMAShort ou prix vs VWMALong.
   - Ancrage DX/ADX: t_dx_cross = bougie du croisement DX/ADX. Si pentes DX/ADX ou écarts/dominance DX−ADX sont invalides à i, ne pas émettre; réévaluer aux bougies suivantes. Pas d'exigence de validité continue depuis t_dx_cross.
   - Filtre extrêmes CCI (si activé): LONG rejeté si CCI ≥ cci_overbought; SHORT rejeté si CCI ≤ cci_oversold, à la bougie candidate.
  - Classification:
    - Court>Long + DI+>DI- + DX↑ → LONG TREND
    - Court>Long + DI->DI+ + DX↓ → LONG COUNTER_TREND
    - Court<Long + DI->DI+ + DX↑ → SHORT TREND
    - Court<Long + DI+>DI- + DX↓ → SHORT COUNTER_TREND
  - Priorités: EXIT > ENTRY; TREND > COUNTER_TREND (même direction)
- Sorties: événements BanFinSignal

Note: Aucune référence à des chemins de fichiers; implémentation libre respectant ces interfaces.

## ❌ CAS INVALIDES
 
 - **DX/ADX contraire à la direction VWMA**
   - VWMA LONG et dernier croisement DX/ADX = Down (ADX > DX) → base DX/ADX non satisfaite → pas de signal à la bougie candidate; attendre un croisement Up (DX > ADX) tant que W le permet.
   - VWMA SHORT et dernier croisement DX/ADX = Up (DX > ADX) → base non satisfaite → attendre un croisement Down.
 
 - **Inversion DX/ADX pendant le gating**
   - Un nouveau croisement dans le sens inverse avant émission devient le « dernier croisement » de la fenêtre W et invalide la base si contraire à VWMA. On attend un croisement cohérent dans W; à défaut → aucun signal.
 
 - **DMI opposé à VWMA (dominance DI)**
   - VWMA LONG et DI− > DI+ → rejet de la bougie candidate.
   - VWMA SHORT et DI+ > DI− → rejet de la bougie candidate.
 
 - **Pentes invalides à la bougie candidate (gating)**
   - Pour toute série activée (VWMAShort, VWMALong, DX, ADX, CCI), si la pente ne respecte pas le signe attendu (et seuil si défini) → rejet à la bougie; réévaluation aux bougies suivantes tant que W est valide.
 
 - **Gap insuffisant à la bougie candidate**
   - gap < n × ATR(3) (base de gap configurée) → rejet; réévaluer aux bougies suivantes tant que W est valide.
 
 - **CCI extrême contraire (gating)**
   - LONG rejeté si CCI ≥ cci_overbought.
   - SHORT rejeté si CCI ≤ cci_oversold.
 
 - **Expiration de fenêtre (W)**
    - Les 3 conditions de base ne sont jamais toutes présentes dans W.
    - Les 3 bases sont présentes mais aucune bougie n’a un gating valide avant que l’une des bases sorte de W → aucun signal.

## ⚙️ GÉNÉRATEUR BAN_FIN

- **Objectif**
  - Déterminer s’il existe un signal sur la dernière bougie fermée (i*), conformément à BAN_FIN.

- **Entrées**
  - Config stratégie (périodes VWMA/DMI/ATR/CCI, W, n, bases de gap, activations et seuils de pentes, seuils CCI, dx_adx_spread_min éventuel).
  - Klines OHLCV suffisantes pour tous les indicateurs.

- **Sorties**
  - Signal présent ou non à i*.
  - Si présent: Type (LONG/SHORT), Mode (TREND), principales métadonnées (indices d’ancre, valeurs clés, raisons de validation).

- **Workflow (edge‑triggered, “première bougie valide après dernier VWMA cross”)**
  0. Ordonner chronologiquement les klines par timestamp croissant; dédupliquer si nécessaire; ignorer les bougies incomplètes.
  1. Pré‑calculer VWMA (short/long), DMI (DI+/DI−), DX/ADX (et croisements), ATR(3), CCI, pentes activées.
  2. Définir la fenêtre glissante W = [i*−W+1, i*].
  3. Trouver le dernier croisement VWMA dans W.
     - S’il n’y en a pas → pas de signal.
     - Sinon fixer la direction cible: CrossUp → LONG, CrossDown → SHORT.
  4. Vérifier les 3 bases à la bougie candidate i*:
     - DI aligné au sens VWMA (LONG: DI+>DI−; SHORT: DI−>DI+).
     - DX/ADX directionnel cohérent dans W via le DERNIER croisement (LONG: Up, SHORT: Down). Un croisement inverse récent invalide la base.
     - VWMA cross présent (déjà acquis à l’étape 3).
  5. Gating par bougie à i*:
     - Pentes activées conformes (signe/seuil) pour VWMAShort, VWMALong, DX, ADX, CCI.
     - Gap ≥ n×ATR(3) (base de gap selon config).
     - CCI non extrême pour le côté (pas de LONG en surachat, pas de SHORT en survente).
     - Dominance/écart DX/ADX (si activé) conforme au dernier croisement et au seuil éventuel.
  6. Émission edge‑triggered:
     - Chercher dans [t_cross, i*] la première bougie j0 qui satisfait bases + gating.
     - Émettre uniquement si i* = j0.
     - Sinon, aucun nouveau signal (les bougies ultérieures alignées restent un ÉTAT, pas un nouvel événement). Réarmement sur un nouveau croisement VWMA.
  7. Cas dynamique DX/ADX:
     - Si un croisement inverse survient avant l’émission, la base DX/ADX devient contraire → pas de signal tant qu’un croisement cohérent ne réapparaît pas dans W.

### Tests à effectuer

1. Vérifier que les 4 signaux sont correctement détectés
2. Tester avec différentes tailles de fenêtre (3, 5, 7 bougies)
3. Valider que Entry/Exit sont indépendants
4. Comparer performances vs Direction simple et Direction+DMI
5. Optimiser paramètres VWMA (short/long periods)

---

 

---

**Version finale validée le 2025-11-08**
**Cette spécification fait référence - Ne pas modifier sans discussion**
