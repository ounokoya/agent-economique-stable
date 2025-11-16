# Résumé Complet - Système de Trading "Harmonie Capital"

> **Synthèse par Claude - 6 novembre 2025**  
> Sources: recherche_part_1 à 4

---

## 📋 Table des Matières

1. [Vue d'Ensemble](#vue-densemble)
2. [Architecture 2 Couches](#architecture-2-couches)
3. [Indicateurs Techniques](#indicateurs-techniques)
4. [Money Management Dynamique](#money-management-dynamique)
5. [Les Deux Produits](#les-deux-produits)
6. [Sélection des Actifs](#sélection-des-actifs)
7. [Métriques de Performance](#métriques-de-performance)
8. [Implémentation](#implémentation)

---

## 🎯 Vue d'Ensemble

### Philosophie Centrale

> **"Le Stop et le TP ne sont pas des paramètres — c'est un art."**

Le système rejette les ratios fixes (1:2, stop 0.5%/TP 1%) au profit d'une **lecture vivante du marché**:

- **Volatilité** (ATR) → dimension de souffle
- **Cycle** (DMI/DX) → force ou respiration  
- **Volume** (MFI) → flux réel
- **Prix** (CCI/Stoch) → forme du mouvement

### Principe du "Système Vivant"

Le stop et TP **évoluent naturellement** avec le marché plutôt que d'être "placés" de façon arbitraire.

---

## 🏗️ Architecture 2 Couches

### Structure Fondamentale

**Toujours 2 timeframes** (jamais plus pour éviter la complexité):

| Couche | Rôle | Indicateurs |
|--------|------|-------------|
| **Contexte** | Phase du marché (impulsion/respiration) | DMI, ATR, MFI, CCI |
| **Exécution** | Timing précis entrée/sortie | CCI, MFI, Stoch, DMI local |

**Relation:**
- Contexte → *si tu as le droit d'agir*
- Exécution → *quand et comment agir*

### Adaptabilité Multi-TF

Structure identique, seuls les paramètres changent:

| Usage | Contexte | Exécution | Cible | Durée |
|-------|----------|-----------|-------|-------|
| **Scalping ultra** | 5m | 1m | 0.2-0.5% | 2-3 min |
| **Scalping court** | 30m | 5m | 0.3-0.8% | 15-25 min |
| **Invest actif** | 4H | 1H | 1-2×ATR | 1-5 jours |

---

## 📊 Indicateurs Techniques

### DMI (Directional Movement Index)

**Rôle:** Structure de tendance et force

**Paramètres:**
- Contexte: DMI(48,6) ou (24,6)
- Exécution: DMI(14,3)

**Lecture:**
- DX ↑ → impulsion (bloquer contrarien)
- DX ↓ → respiration (autoriser contrarien)
- DX > ADX → force réelle
- DX < ADX → impulsion faible

### ATR% (Average True Range en %)

**Calcul:** `ATR(période) / Close × 100`

**Régimes (crypto 5m):**
- Compression: < 0.6%
- Normal: 0.6-1.2%
- Expansion: > 1.2%
- Spike: bougie > 2× ATR%

**Usage:** Calibrer SL/TP, détecter anomalies

### CCI (Commodity Channel Index)

**Périodes:**
- Contexte: CCI(60) ou (30)
- Exécution: CCI(14-20) pour 1m, (20-30) pour 5m

**Zones:**
- ±200 → extrême local (contrarien)
- ±100 → excès global
- Perte de pente → fin de cycle

### MFI (Money Flow Index)

**Périodes:** MFI(60) contexte, MFI(14-20) exécution

**Zones:**
- < 30 → accumulation
- > 70 → distribution
- Pente = direction du flux

### Stochastique

**Paramètre:** Stoch(9,3,3) ou (14,3,3)

**Rôle:** Déclencheur précis
- Croisement contrarien dans extrêmes
- Croisement dans sens DI pour continuation

---

## 💰 Money Management Dynamique

### Machine à États

1. **FLAT** → pas de position
2. **OPEN_PROTECT** → SL initial posé
3. **SECURED_BE** → stop au break-even
4. **TRAIL** → trailing actif
5. **EXIT** → clôture

### Stop Loss Hiérarchisé

**SL Initial:**
```
SL = k × ATR%
k = 0.8 (compression) | 1.0 (normal) | 1.3 (expansion)
```

**Lock Break-Even:**
- Déclenché à +0.12-0.20% (1m) ou +0.20-0.35% (5m)
- Si pente CCI reste favorable 2-3 bougies

**Trailing:**
```
Stop = max(BE, Close − m × ATR%)  [Long]
m = 0.6-0.8 selon nervosité
```

**Stop Temporel:**
- 2-3 bougies (1m) sans progrès → EXIT
- 3-5 bougies (5m) sans progrès → EXIT

### Gestion Grosses Bougies

**Spike:** TR ≥ 2× ATR_prev

| Cas | Action |
|-----|--------|
| Dans sens | Gèle stop 1 bougie |
| Contre + DX↑ | EXIT immédiat |
| Contre + DX↓ | Gèle 1 bougie, réévalue |

### Ratio de Respiration

```
R = (|CCI| + |MFI|) / DX
```
Si R baisse > 20% → marché respire → gèle stop

---

## 🎨 Les Deux Produits

### 1️⃣ BOT SCALPING "Revenus Complémentaires"

| Élément | Détail |
|---------|--------|
| **TF** | 5m / 1m |
| **Actifs** | SOL, SUI, AVAX, LINK, ARB |
| **Cible** | ~10% var/mois (~100% avec lev 10) |
| **Durée** | 1-3 minutes |
| **SL init** | 0.25-0.35% |
| **Lock BE** | +0.12-0.20% |
| **Client** | Traders actifs, risque modéré |

**Indicateurs:**
- Contexte 5m: DMI(48,6), ATR(48), MFI(60), CCI(60)
- Exécution 1m: CCI(14-20), MFI(14), Stoch(9,3,3), DMI(14,3), ATR(24)

### 2️⃣ BOT INVESTISSEMENT ACTIF "Croissance Richesse"

| Élément | Détail |
|---------|--------|
| **TF** | 4H / 1H |
| **Actifs** | TSLAon, NVDAon, AAPLon, METAon, MSFTon |
| **Cible** | 4-6%/mois (~70-100%/an composé) |
| **Durée** | 1-5 jours |
| **SL init** | 1× ATR(1H) |
| **Client** | Investisseurs prudents, long terme |

**Principe:** Vendre montant initial, **conserver bénéfices en actions** → effet boule de neige

**Indicateurs:**
- Contexte 4H: DMI(24,6), ATR(24), CCI(30), MFI(30)
- Exécution 1H: CCI(20), MFI(14), Stoch(9,3,3), DMI(14,3), ATR(14)

---

## 🎯 Sélection des Actifs

### Critères de Compatibilité

| Indicateur | Condition | Test |
|------------|-----------|------|
| **CCI** | Oscillations symétriques | Pas saturé ±200 |
| **MFI** | Suit flux réel | Corr(MFI,ΔClose) > 0.5 |
| **Stoch** | Croisements nets | Pas vibration permanente |
| **DMI** | Alternance DI+/DI− | DX_mean 25-55 |
| **ATR** | Lissé progressif | Variance < 0.0005 |

### Top 5 Crypto (Scalping)

1. **SOL/USDT** - Cycles propres, 8-12%/mois - k=0.9-1.1, m=0.6-0.8
2. **SUI/USDT** - Volatilité rythmée, 12-15%/mois - k=1.0-1.3, m=0.8
3. **AVAX/USDT** - Directionnel régulier, 8-12%/mois - k=0.9-1.1, m=0.7-0.8
4. **LINK/USDT** - MFI propre, 7-11%/mois - k=0.8-1.0, m=0.6-0.7
5. **ARB/USDT** - Bonne liquidité, 7-10%/mois - k=0.9-1.2, m=0.7-0.8

**À éviter:** Meme coins (PEPE, FLOKI...) - spikes imprévisibles

### Actions Tokenisées (Investissement)

1. **TSLAon (Tesla)** - PRIORITÉ #1
   - ATR% élevé (2-3× autres)
   - Cycles 4H/1H nets
   - MFI cohérent, DMI réactif

2. **NVDAon** - Directionnelle, moins de respirations
3. **AAPLon** - Trop calme (CCI/MFI peu expressifs)
4. **METAon** - Volume irrégulier
5. **MSFTon** - Bon long terme, pas rotation rapide

---

## 📈 Métriques de Performance

### Métriques Contexte (avant trade)

- **Volatility Regime**: ATR% classification
- **DX Slope**: ΔDX sur 3 bougies
- **Flow Pressure**: Zone MFI
- **Confluence Score**: 0-5 éléments favorables
- **Noise Index**: ATR% / DX

### Métriques Ouverture (signal)

- **CCI Distance**: |CCI| / 200
- **MFI Divergence**: ΔCCI/ΔMFI
- **Stoch Alignment**: Angle croisement
- **Entry Efficiency**: Distance entry/extrême

### Métriques Dynamique (pendant)

- **Speed Ratio**: Bougies peak / total
- **Return Efficiency**: Max gain / max DD
- **Momentum Persistence**: Durée avant inversion CCI
- **Spike Sensitivity**: Réaction aux spikes

### Métriques Sortie

- **Exit Type**: TP/Trail/BE/Spike/Time/ContextFlip
- **Exit Efficiency**: (Gain final / Gain max) × 100
- **CCI Exhaustion**: Timing vs inflexion CCI

### Métriques Agrégées

- **Setup Score**: Somme pondérée toutes métriques
- **Win Rate**: % trades positifs
- **Profit Factor**: Gains / Pertes
- **Sharpe-like**: Mean(PnL) / Std(PnL)
- **Max Drawdown**: Plus grosse perte

---

## 🔧 Implémentation

### Structure Signaux

**A. Contrarien (Respiration)**

Contexte: DX↓ ou ATR↓, MFI stable

Exécution:
1. CCI ≥ ±200
2. MFI < 30 ou > 70 + inflexion
3. Stoch croise à contre-sens DI
4. DX < ADX

**B. Directionnel (Impulsion)**

Contexte: DX↑ & ATR↑, DI identifié

Exécution:
1. CCI même côté que 0
2. Stoch croise dans sens DI
3. DX > ADX
4. MFI > 50 dans sens

### Workflow Trade

```
OUVERTURE:
→ Évaluer contexte (DMI + ATR supérieur)
→ Attendre signal (CCI + MFI + Stoch)
→ Valider filtre (DX vs ADX)
→ Entrer + SL_init = k × ATR%

GESTION:
→ Surveiller gain vs seuil BE
→ Si OK + pente CCI → Lock BE
→ Si continue → Trailing
→ Spike dans sens → Geler 1 bougie
→ Spike contre + DX↑ → EXIT

FERMETURE:
→ Perte pente CCI/MFI
→ Flip contexte DX
→ Time-stop
→ EXIT + log métriques
```

### Calculs Clés

**ATR%:**
```python
ATR_pct = (ATR(period) / Close) * 100
```

**SL Initial:**
```python
k = 0.8 if compression else (1.0 if normal else 1.3)
SL_init = k * ATR_pct
```

**Trailing (Long):**
```python
new_stop = max(BE_price, close - m * ATR_pct)
```

**Spike:**
```python
is_spike = (high - low) / close_prev > 2 * ATR_prev
```

---

## 🎯 Objectifs Rendement

### BOT Scalping
- **Cible:** ~10% var/mois
- **Avec levier 10:** 80-120% brut/mois
- **Fréquence:** 10-30 trades/jour
- **Risque:** Moyen à élevé

### BOT Investissement
- **Cible:** 4-6% brut/mois
- **Sur 12 mois composé:** +70-100% annuel
- **Fréquence:** 4-10 trades/mois
- **Risque:** Faible à moyen

---

## 💡 Concepts Clés

### Les 3 Phases du Marché

1. **Impulsion** (DX↑ + ATR↑): mouvement fort, éviter contrarien
2. **Respiration** (DX↓ ou ATR↓): pause, contrarien OK
3. **Désordre** (DX↓ + ATR↑): chaos, s'abstenir

### Le Stop Intelligent

> "Le stop parfait n'évite pas la perte, il *choisit* quand elle est nécessaire."

- Protège quand logique invalidée
- Tolère respiration naturelle
- Attend calme de volatilité

### Le TP Dynamique

> "Le TP parfait ne vole jamais ton profit, mais ne ment pas sur la fin."

- Pas de cible fixe
- Mouvement d'accompagnement
- Respect fin de cycle

### L'Adaptabilité

Structure identique tous TF, il suffit d':
1. Ajuster périodes indicateurs (2-3× durée TF)
2. Recaler seuils (SL/TP prop. ATR%)
3. Garder logique états (entrée → BE → trail → sortie)

---

## ✅ Checklist Implémentation

### Phase 1: Fondations
- [ ] Calcul indicateurs (CCI, MFI, Stoch, DMI, ATR)
- [ ] Validation données historiques
- [ ] Machine à états 5 niveaux
- [ ] SL initial dynamique (k × ATR%)

### Phase 2: MM Avancé
- [ ] Lock break-even auto
- [ ] Trailing adaptatif (m × ATR%)
- [ ] Détection/gestion spikes
- [ ] Time-stop
- [ ] Ratio respiration

### Phase 3: Signaux
- [ ] Détection contexte (DX, ATR trends)
- [ ] Signal contrarien complet
- [ ] Signal directionnel complet
- [ ] Filtrage multi-TF

### Phase 4: Métriques
- [ ] Logging exhaustif
- [ ] Calcul métriques complètes
- [ ] Setup score pondéré
- [ ] Dashboard visualisation

### Phase 5: Production
- [ ] Backtest multi-TF
- [ ] Optimisation par paire
- [ ] Gestion ordres réels
- [ ] Monitoring temps réel
- [ ] Alertes + rapports

---

## 📚 Glossaire

- **ATR%**: Average True Range en % du prix
- **BE**: Break-Even (prix d'entrée)
- **CCI**: Commodity Channel Index
- **DI+/DI−**: Directional Indicators
- **DMI**: Directional Movement Index
- **DX**: Directional Index (écart DI+/DI−)
- **ADX**: Average Directional Index
- **MFI**: Money Flow Index
- **Stoch**: Stochastique (%K, %D)
- **TF**: TimeFrame
- **SL**: Stop Loss
- **TP**: Take Profit
- **MM**: Money Management
- **PnL**: Profit and Loss

---

**FIN DU RÉSUMÉ**

*Document exploitable pour implémenter les systèmes de trading "Harmonie Capital"*
