# SWE-1.5 — Résumé Technique Complet — Système "Harmonie Capital"

> **Analyse structurée pour implémentation directe**  
> Sources: recherche_part_1.md, recherche_part_2.md, recherche_part_3.md, recherche_part_4.md  
> Date: 2025-11-06

---

## 📋 Executive Summary

Le système "Harmonie Capital" est une architecture de trading algorithmique **multi-timeframe** basée sur une **lecture adaptative du marché**. Il combine 5 indicateurs (CCI, MFI, Stoch, DMI, ATR) dans une structure **Contexte ↔ Exécution ↔ Money Management** pour générer deux produits distincts: un bot de scalping crypto et un bot d'investissement actif sur actions tokenisées.

**Innovation clé:** Remplacement des ratios fixes par un **système vivant** où stop/TP évoluent dynamiquement avec la volatilité, la structure de marché et le flux de volume.

---

## 🏗️ Architecture Fondamentale

### Structure à 2 Couches (Universal Pattern)

```
┌─────────────────┐    ┌─────────────────┐
│   CONTEXTE      │    │   EXÉCUTION     │
│  (TF supérieur) │◄──►│  (TF inférieur) │
│                 │    │                 │
│ • DMI(48,6)     │    │ • CCI(14-20)    │
│ • ATR%(48)      │    │ • MFI(14)       │
│ • MFI(60)       │    │ • Stoch(9,3,3)  │
│ • CCI(60)       │    │ • DMI(14,3)     │
│                 │    │ • ATR%(24)      │
└─────────────────┘    └─────────────────┘
         │                       │
         ▼                       ▼
   Autoriser/Interdire      Timing précis
   les signaux               entrée/sortie
```

### Mapping Timeframes

| Produit | Contexte | Exécution | Cible TP | Durée typique |
|---------|----------|-----------|----------|---------------|
| Scalping ultra | 5m | 1m | 0.2-0.5% | 2-3 min |
| Scalping court | 30m | 5m | 0.3-0.8% | 15-25 min |
| Investissement | 4H | 1H | 1-2×ATR | 1-5 jours |

---

## 📊 Spécifications Indicateurs

### DMI (Directional Movement Index)

**Calcul et paramètres:**
```python
# Contexte
DI_plus, DI_minus, DX = DMI(period=48, smooth=6)
ADX = ADX(period=14)  # pour filtre local

# Exécution  
DI_plus, DI_minus, DX = DMI(period=14, smooth=3)
```

**Règles de trading:**
- `DX > ADX`: Force réelle → bloquer contrarien
- `DX < ADX`: Impulsion faible → autoriser contrarien
- `DX ↑`: Impulsion active
- `DX ↓`: Respiration en cours
- `DI_plus > DI_minus`: Biais haussier
- `DI_minus > DI_plus`: Biais baissier

### ATR% (Average True Range Percentage)

**Calcul:**
```python
ATR_pct = (ATR(period) / Close) * 100
```

**Classification des régimes:**
```python
if ATR_pct < 0.6:     regime = "compression"
elif ATR_pct <= 1.2:  regime = "normal"  
else:                 regime = "expansion"

# Spike detection
is_spike = (high - low) / close_prev > 2 * ATR_pct_prev
```

### CCI (Commodity Channel Index)

**Paramètres adaptatifs:**
- Contexte: CCI(60) ou CCI(30)
- Exécution 1m: CCI(14-20)
- Exécution 5m: CCI(20-30)

**Zones de trading:**
- `|CCI| >= 200`: Extrême local (setup contrarien)
- `|CCI| >= 100`: Excès global
- `pente(CCI, 3) < 0`: Inversion (signal de sortie)

### MFI (Money Flow Index)

**Configuration:**
- Contexte: MFI(60)
- Exécution: MFI(14) ou MFI(14-20)

**Interprétation:**
- `MFI < 30`: Accumulation (acheteurs épuisés)
- `MFI > 70`: Distribution (vendeurs épuisés)
- `pente(MFI, 3)`: Direction du flux

### Stochastique

**Paramètres:**
- 1m: Stoch(9, 3, 3)
- 5m: Stoch(14, 3, 3)

**Signaux:**
- Croisement %K/%D dans extrêmes → contrarien
- Croisement dans sens DI → continuation

---

## 💰 Money Management Dynamique

### Machine à États (State Machine)

```python
class TradingState(Enum):
    FLAT = 1           # Pas de position
    OPEN_PROTECT = 2   # Position ouverte, SL initial
    SECURED_BE = 3     # Stop au break-even
    TRAIL = 4          # Trailing actif
    EXIT = 5           # Clôture en cours
```

### Stop Loss Hiérarchisé

**1. SL Initial (à l'entrée):**
- Formule: SL = k × ATR%
- Valeurs k: 0.8 (compression), 1.0 (normal), 1.3 (expansion)
- Bornes maximales: 1m ≤ 0.35%, 5m ≤ 0.60%

**2. Lock Break-Even:**
- Seuil: +0.15% (1m) ou +0.25% (5m)
- Condition: gain atteint ET pente CCI favorable (2 barres)

**3. Trailing Stop:**
- Formule Long: Stop = max(BE, Close - m × ATR%)
- Formule Short: Stop = min(BE, Close + m × ATR%)
- Valeur m: 0.6-0.8 selon nervosité marché

**4. Time-Stop:**
- Limites: 3 barres (1m), 5 barres (5m), 10 barres (1H)
- Sortie si pas de progression

### Gestion des Spikes Volatilité

**Définition Spike:** TR ≥ 2× ATR_précédent

**Actions:**
- Spike dans sens: geler stop 1 bougie
- Spike contre + DX↑: sortie immédiate  
- Spike contre + DX↓: geler 1 bougie puis réévaluer

### Ratio de Respiration

**Calcul:** R = (|CCI| + |MFI|) / DX

**Interprétation:**
- R baisse > 20% → marché respire → geler stop
- R + DX en baisse → sortie ou réduction

---

## 🎯 Logique de Trading

### Signal Contrarien (Respiration)

**Conditions Contexte:**
- DX en baisse OU ATR en baisse (respiration active)
- MFI contexte stable ou décroissant

**Conditions Exécution:**
- CCI extrême (≥ +200 ou ≤ -200)
- MFI extrême (< 30 ou > 70) avec inflexion
- Stochastique croise à contre-sens du DI dominant
- DX < ADX (impulsion affaiblie)

**Validation:** Toutes conditions doivent être remplies

### Signal Directionnel (Impulsion)

**Conditions Contexte:**
- DX en hausse ET ATR en hausse (impulsion claire)
- DI dominant identifié (DI+ ou DI-)

**Conditions Exécution:**
- CCI du même côté que 0 (dans le sens DI)
- Stochastique croise dans le sens du DI
- DX > ADX (force réelle)
- MFI confirme (> 50 si haussier, < 50 si baissier)

**Validation:** Toutes conditions doivent être remplies

---

## 🎨 Spécifications Produits

### Produit 1: Scalping Crypto 5m/1m

**Configuration:**
- Timeframes: Contexte 5m / Exécution 1m
- Cibles TP: 0.2-0.5% par trade
- SL maximum: 0.35%
- Actifs: SOL, SUI, AVAX, LINK, ARB

**Indicateurs Contexte (5m):**
- DMI(48,6): structure tendance
- ATR(48): régime volatilité  
- MFI(60): pression flux
- CCI(60): écart structurel

**Indicateurs Exécution (1m):**
- CCI(14-20): extrêmes locaux
- MFI(14): pression instantanée
- Stoch(9,3,3): déclencheur
- DMI(14,3): filtre local
- ATR(24): calibrage SL/TP

### Produit 2: Investissement Actif (4H/1H)

**Configuration:**
- Timeframes: Contexte 4H / Exécution 1H
- Cibles TP: 1-2× ATR (1-5 jours)
- SL: 1× ATR
- Actifs: TSLAon, NVDAon, AAPLon, METAon, MSFTon

**Principe Capitalisation:**
- Vendre montant investi initial
- Conserver bénéfices en actions
- Réinvestir profits dans nouvelles positions

**Indicateurs Contexte (4H):**
- DMI(24,6): tendance fond
- ATR(24): volatilité moyenne
- MFI(30): flux de fond
- CCI(30): excès structurels

**Indicateurs Exécution (1H):**
- CCI(20): timing précis
- MFI(14): validation flux
- Stoch(9,3,3): déclencheur
- DMI(14,3): filtre local
- ATR(14): dimensionnement

---

## 🧪 Sélection et Validation des Actifs

### Critères Quantitatifs de Compatibilité

**Test CCI:** Oscillations symétriques (pas collé à ±200)
- Score > 0.7 = bon

**Test MFI:** Corrélation avec prix
- Corrélation > 0.5 = flux cohérent

**Test DMI:** Alternance DI+/DI-
- DX moyen entre 25-55 = tendance exploitable
- Alternance > 0.6 = structure saine

**Test ATR:** Variance faible
- Variance < 0.0005 = volatilité stable

**Score de Compatibilité:** 0-5 points
- 5/5: Actif optimal
- 3-4/5: Exploitable avec prudence
- <3/5: À exclure

### Actifs Validés

**Crypto (Scalping):**
- **SOL/USDT**: k=0.9-1.1, m=0.6-0.8, score=5/5
- **SUI/USDT**: k=1.0-1.3, m=0.8, score=5/5  
- **AVAX/USDT**: k=0.9-1.1, m=0.7-0.8, score=4/5
- **LINK/USDT**: k=0.8-1.0, m=0.6-0.7, score=4/5
- **ARB/USDT**: k=0.9-1.2, m=0.7-0.8, score=4/5

**Actions Tokenisées (Investissement):**
- **TSLAon**: Priorité #1, ATR×1.2, score=5/5
- **NVDAon**: Priorité #2, ATR×1.0, score=4/5
- **AAPLon**: Priorité #3, ATR×0.8, score=3/5
- **METAon**: Priorité #4, ATR×0.9, score=3/5
- **MSFTon**: Priorité #5, ATR×0.8, score=3/5

---

## 📈 Système de Métriques

### Métriques par Phase

**Métriques Contexte (avant trade):**
- Score contexte: 0-1 (DX↓, ATR normal, MFI pression, structure alignée, bruit faible)
- Régime volatilité: compression/normal/expansion
- Pente DX: ΔDX sur 3 bougies
- Pression flux: zone MFI
- Index bruit: ATR%/DX

**Métriques Entrée (signal):**
- Efficacité entrée: distance entrée/extrême en ATR
- Distance CCI: |CCI|/200
- Divergence MFI: ΔCCI/ΔMFI
- Alignement Stoch: angle croisement

**Métriques Dynamique (pendant):**
- Ratio vitesse: barres peak/total
- Efficacité retour: max gain/max DD
- Persistance momentum: durée avant inversion CCI
- Réactivité spikes: temps réaction

**Métriques Sortie:**
- Type sortie: TP/Trail/BE/Spike/Time/ContextFlip
- Efficacité sortie: (% gain final/gain max) × 100
- Timing CCI: précision vs inflexion

### Score Global de Setup

**Calcul pondéré:**
- Contexte: 25%
- Entrée: 25%  
- Dynamique: 30%
- Sortie: 20%

**Échelle:** 0-1 (0.7+ = setup qualité)

---

## 🔧 Workflow d'Implémentation

### Pipeline de Trading Complet

**Étape 1: Évaluation Contexte**
- Analyser DX/ATR sur TF supérieur
- Calculer score contexte (0-1)
- Autoriser/Interdire signaux selon score (>0.6)

**Étape 2: Détection Signaux**
- Si FLAT: chercher setups contrarien/directionnel
- Valider toutes conditions contexte + exécution
- Générer signal avec type et direction

**Étape 3: Entrée Position**
- Calculer SL initial: k × ATR% selon régime
- Poser ordre avec SL immédiat
- Basculer état OPEN_PROTECT

**Étape 4: Gestion Dynamique**
- Surveillance gain vs seuil BE
- Gestion spikes (gel/sortie immédiate)
- Lock BE si conditions remplies
- Activation trailing si progression

**Étape 5: Sortie**
- Perte pente CCI/MFI
- Time-stop dépassé
- Stop touché
- Flip contexte DX
- Spike contre + DX↑

**Étape 6: Logging**
- Enregistrer toutes métriques
- Calculer score setup
- Analyser performance post-trade

---

## 📊 Spécifications de Performance

### Objectifs par Produit

**Scalping Crypto:**
- Capture mensuelle: 10% variation
- Avec levier 10: 80-120% brut/mois
- Durée moyenne trade: 2-3 minutes
- Fréquence: 10-30 trades/jour
- Win rate cible: 55%
- Profit factor cible: 1.8

**Investissement Actions:**
- Retour mensuel: 4-6% brut
- Annuel composé: 70-100%
- Durée moyenne: 1-5 jours
- Fréquence: 4-10 trades/mois
- Win rate cible: 65%
- Profit factor cible: 2.2

### Métriques de Suivi

**Qualité Setup:**
- Score contexte: 0-1
- Efficacité entrée: unités ATR
- Performance dynamique: ratio vitesse, efficacité retour
- Efficacité sortie: % du gain max

**Métriques Risque:**
- Drawdown maximum: %
- Ratio Sharpe: mean/std
- Ratio Calmar: return/max_dd
- Pertes consécutives: nombre

**Métriques Opérationnelles:**
- Fréquence trades: trades/jour
- Durée détention: barres
- Impact slippage: %
- Latence exécution: ms

---

## ✅ Roadmap d'Implémentation

### Phase 1: Fondations (Sprint 1-2)
- [ ] Implémenter calculs indicateurs (CCI, MFI, Stoch, DMI, ATR)
- [ ] Valider précision sur données historiques
- [ ] Créer machine à états de base
- [ ] Implémenter SL initial dynamique

### Phase 2: MM Avancé (Sprint 3-4)
- [ ] Lock break-even automatique
- [ ] Trailing stop adaptatif
- [ ] Gestion spikes volatilité
- [ ] Time-stop configurable
- [ ] Ratio de respiration

### Phase 3: Signaux (Sprint 5-6)
- [ ] Détection contexte multi-TF
- [ ] Signaux contrarien/directionnel complets
- [ ] Filtrage avancé (DX/ADX, ATR regime)
- [ ] Validation confluence

### Phase 4: Métriques (Sprint 7-8)
- [ ] Système de logging exhaustif
- [ ] Calcul métriques temps réel
- [ ] Setup score pondéré
- [ ] Dashboard visualisation

### Phase 5: Production (Sprint 9-10)
- [ ] Backtesting multi-paires
- [ ] Optimisation paramètres (grid search)
- [ ] Walk-forward analysis
- [ ] Gestion ordres réels + slippage
- [ ] Monitoring + alertes

---

## 📚 Références Techniques

### Formules Clés

**ATR Percentage:**
ATR% = (ATR(période) / Close) × 100

**CCI:**
CCI = (Typical_Price - SMA(Typical_Price, période)) / (0.015 × Mean_Deviation)

**MFI:**
MFI = 100 - (100 / (1 + Money_Flow_Ratio))

**DMI:**
DX = 100 × |DI+ - DI-| / (DI+ + DI-)
ADX = SMA(DX, période)

**Stochastique:**
%K = 100 × (Close - LL(Low, k_période)) / (HH(High, k_période) - LL(Low, k_période))
%D = SMA(%K, d_période)

### Paramètres Optimisés

**Scalping 1m:**
- CCI: période 16
- MFI: période 14
- Stoch: K=9, D=3
- DMI: période 14
- ATR: période 24
- k_SL: 1.0, m_trail: 0.7

**Investissement 1H:**
- CCI: période 20
- MFI: période 14
- Stoch: K=9, D=3
- DMI: période 14
- ATR: période 14
- k_SL: 1.0, m_trail: 0.8

---

## 🏁 Conclusion

Le système "Harmonie Capital" représente une approche **sophistiquée mais implémentable** du trading algorithmique, combinant:

- **Architecture universelle** 2 couches adaptable à tous les timeframes
- **Money Management adaptatif** remplaçant les ratios fixes
- **Deux produits complémentaires** ciblant différents profils d'investisseurs
- **Système de métriques complet** pour optimisation continue
- **Sélection rigoureuse des actifs** basée sur critères quantitatifs

La feuille de route technique permet un déploiement progressif en 10 sprints, avec validation à chaque étape.

---

**Document technique prêt pour implémentation et backtesting**
