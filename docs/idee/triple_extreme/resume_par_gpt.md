# Résumé Complet — Système de Trading « Harmonie Capital »

> Synthèse par GPT — 6 novembre 2025  
> Sources: recherche_part_1.md, recherche_part_2.md, recherche_part_3.md, recherche_part_4.md

---

## 🎯 Vue d'ensemble

- **Philosophie**: Stop/TP = art d'équilibrer protection du capital et liberté du profit.  
- **Rejet des ratios fixes**: préférer des règles adaptatives basées sur volatilité (ATR), structure (DMI/DX), flux (MFI) et forme (CCI/Stoch).  
- **Système vivant**: le stop et le TP **évoluent** avec le marché (contexte ↔ exécution ↔ MM).

---

## 🏗️ Architecture 2 Couches (universelle)

- **Contexte (TF supérieur)**: DMI, ATR, MFI, CCI → décide si on a le droit d’agir (impulsion/respiration/désordre).
- **Exécution (TF inférieur)**: CCI, MFI, Stoch, DMI → décide quand et comment agir (timing précis).

Exemples d’appairage:
- 5m (contexte) ↔ 1m (exécution) — scalping ultra-court.
- 30m (contexte) ↔ 5m (exécution) — mini-swing.
- 4H (contexte) ↔ 1H (exécution) — investissement actif court/moyen terme.

---

## 📊 Indicateurs et lectures clés

- **DMI/DX**  
  - DI+ > DI−: biais haussier. DI− > DI+: biais baissier.  
  - DX↑: impulsion (bloquer contrarien). DX↓: respiration (autoriser contrarien).  
  - Filtre local: DX>ADX = force réelle; DX<ADX = contrarien possible.

- **ATR%**  
  - ATR% = ATR/Close×100.  
  - Régimes (ex crypto 5m): compression <0.6 %, normal 0.6–1.2 %, expansion >1.2 %, spike = range > 2× ATR%.  
  - Sert à calibrer SL initial (k×ATR%) et trailing (m×ATR%).

- **CCI**  
  - Exécution: 14–20 (1m), 20–30 (5m).  
  - Zones: ±200 = extrême local (contrarien); ±100 = excès global.  
  - Perte de pente = fin de cycle/signal de sortie.

- **MFI**  
  - Contexte: 60. Exécution: 14–20.  
  - <30 accumulation, >70 distribution, pente = pression du flux.

- **Stochastique**  
  - 9,3,3 (1m) ou 14,3,3 (5m).  
  - Croisement contrarien en extrême; croisement dans le sens DI pour continuation.

---

## 💰 Money Management (MM) dynamique

### Machine à 5 états
1. **FLAT** → pas de position.  
2. **OPEN_PROTECT** → entrée, SL initial posé.  
3. **SECURED_BE** → lock au break-even.  
4. **TRAIL** → suivi par ATR% (m).  
5. **EXIT** → clôture (retour FLAT).

### Règles clés
- **SL initial**: `SL = k × ATR%`  
  k = 0.8 (compression) / 1.0 (normal) / 1.3 (expansion).  
  Bornes typiques: 1m (≤0.35 %), 5m (≤0.60 %).

- **Lock BE**:  
  1m: +0.12–0.20 % si pente CCI favorable (2 barres)  
  5m: +0.20–0.35 % si pente CCI favorable (2–3 barres)

- **Trailing** (long): `Stop = max(BE, Close − m × ATR%)`  
  m ≈ 0.6–0.8 (nerveux → m plus grand).

- **Time-stop**:  
  1m: 2–3 barres sans progrès → EXIT  
  5m: 3–5 barres sans progrès → EXIT  
  Global: ~10 barres après BE sans nouveau plus-haut/bas → EXIT

### Gestion des grosses bougies (spikes)
- Définition: TR ≥ 2× ATR_prev ou range > seuil.
- Dans ton sens: gèle le stop 1 bougie; reprendre si bougie suivante garde >50 % du corps.  
- Contre toi + DX↑: sortie immédiate (impulsion contraire).  
- Contre toi + DX↓: gèle 1 bougie puis réévaluation.  
- Range neutralisé (wicks longs, petit corps): pas d’action, attendre 2 bougies.

### Ratio de respiration (bonus)
`R = (|CCI| + |MFI|) / DX`  
- R baisse >20 % → marché respire → geler le stop.  
- R + DX en baisse → sortie/réduction.

---

## 🧩 Entrées: Contrarien vs Directionnel

- **Contrarien (Respiration)**  
  Contexte: DX_context↓ ou ATR_context↓; MFI_context non-poussant.  
  Exécution: CCI extrême (±200) + MFI extrême + Stoch croise à contre-sens + DX_exec<ADX_exec.  
  Entrée unique, taille fixe, SL_init immédiat.

- **Directionnel (Impulsion)**  
  Contexte: DX_context↑ & ATR_context↑, DI dominant.
  Exécution: CCI côté de 0 (dans le sens), Stoch croise dans le sens, DX_exec>ADX_exec, MFI confirme (>50).  
  Entrée unique, taille fixe, SL_init immédiat.  
  Pas d’entrée si spike ATR% au contexte.

---

## 🎨 Deux produits complémentaires

### 1) BOT Scalping (revenus complémentaires)
- **TF**: 5m (contexte) / 1m (exécution).  
- **Indicateurs**: DMI(48,6)/MFI(60)/CCI(60) ; exec CCI(14–20), MFI(14), Stoch(9,3,3), DMI(14,3), ATR(24).  
- **Cibles**: 0.2–0.5 % par trade; ~10 % variation captée/mois (sans levier).  
- **MM**: taille fixe, SL/TP dynamiques, BE rapide, trailing par ATR.  
- **Actifs**: SOL, SUI, AVAX, LINK, ARB.

### 2) BOT Investissement Actif (croissance/patrimoine)
- **TF**: 4H (contexte) / 1H (exécution).  
- **Indicateurs**: DMI(24,6)/ATR(24)/CCI(30)/MFI(30) ; exec CCI(20), MFI(14), Stoch(9,3,3), DMI(14,3).  
- **Cibles**: 4–6 % brut/mois sans levier (≈70–100 %/an composé).  
- **Principe**: vendre le montant investi initial; conserver **les bénéfices** en actions → capitalisation.  
- **Actifs**: actions US tokenisées (TSLAon, NVDAon, AAPLon, METAon, MSFTon).

---

## 🧪 Sélection objective des actifs

### Critères de compatibilité
- CCI: oscillations symétriques (pas "collé" à ±200).  
- MFI: corrélé au prix (corr(MFI(14), ΔClose) > 0.5).  
- Stoch: croisements nets (peu de vibration).  
- DMI: alternance DI+/DI−, DX_mean ~25–55 (sur 30 j).  
- ATR%: variance faible, profil lissé/progressif.

### Top 5 Crypto (scalping)
1. SOL/USDT — cycles propres; réglages: CCI(14–18), k=0.9–1.1, m=0.6–0.8.  
2. SUI/USDT — volatilité rythmée; CCI(14), k=1.0–1.3, m=0.8.  
3. AVAX/USDT — directionnalité régulière; CCI(16–20), k=0.9–1.1, m=0.7–0.8.  
4. LINK/USDT — flux MFI propre; CCI(18–22), k=0.8–1.0, m=0.6–0.7.  
5. ARB/USDT — bonne liquidité; CCI(14–18), k=0.9–1.2, m=0.7–0.8.  
Alternatives: OP, SEI, PENDLE.  
À éviter: meme coins bruités (PEPE, FLOKI, …).

### Actions tokenisées (investissement)
- **TSLAon** (priorité #1): ATR% élevé, cycles 4H/1H nets, MFI cohérent, DMI réactif.  
- **NVDAon**: très directionnelle; moins de respirations courtes.  
- **AAPLon/MSFTon**: trop calmes pour rotation rapide.  
- **METAon**: volume irrégulier → MFI moins fiable.

---

## 📈 Métriques pour piloter et optimiser

- **Contexte**: Volatility Regime (ATR%), DX Slope, Flow Pressure (MFI), Confluence Score, Noise Index (ATR%/DX).  
- **Ouverture**: CCI Distance (|CCI|/200), MFI Divergence (ΔCCI/ΔMFI), Stoch Alignment (angle %K/%D), Entry Efficiency.  
- **Dynamique**: Speed Ratio (t_peak/t_total), Return Efficiency (max_gain/max_DD), Volatility Response (ΔStop/ΔATR), Momentum Persistence (durée avant inversion CCI), Spike Sensitivity.  
- **Sortie**: Exit Type, Exit Efficiency (% gain max capté), CCI Exhaustion Accuracy, Lock Timing, Context Reversal Timing.  
- **Agrégées**: Setup Score (pondéré), Win Rate, Profit Factor, Max Drawdown, Sharpe-like.

---

## 🔧 Implémentation (workflow synthétique)

1) Contexte (TF sup) → DMI/DX + ATR% (autoriser/empêcher contrarien; classer régime).  
2) Signal (TF inf) → CCI + MFI + Stoch; filtre DMI (DX vs ADX).  
3) Entrée unique, taille fixe → poser SL_init = k×ATR%.  
4) Gérer time-stop; locker BE si seuil atteint + pente CCI ok.  
5) Activer trailing par `m×ATR%`; geler stop sur spike dans le sens.  
6) EXIT sur: perte de pente CCI/MFI, flip DX de contexte, time-stop, spike contre + DX↑.  
7) Logger métriques (contexte, ouverture, dynamique, sortie) + Setup Score.

---

## ✅ Checklist de mise en production

- Indicateurs (CCI, MFI, Stoch, DMI, ATR) validés sur historiques.  
- Machine à états (FLAT→OPEN→BE→TRAIL→EXIT).  
- SL initial dynamique (k), BE auto, trailing (m), time-stop, spikes.  
- Détection contexte (DX/ATR), gating contrarien/directionnel.  
- Journal complet + calcul des métriques + dashboard.  
- Backtests multi-TF, optimisation par paire, walk-forward, tolérance au slippage/latence.

---

## 📚 Glossaire
ATR% (ATR/Close×100) · BE (Break-Even) · CCI · DI+/DI− · DMI · DX · ADX · MFI · Stoch (%K/%D) · SL (Stop Loss) · TP (Take Profit) · MM (Money Management) · TF (TimeFrame).

---

Fin du résumé — Document prêt à l’implémentation et au backtest.
