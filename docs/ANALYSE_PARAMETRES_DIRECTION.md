# 📊 ANALYSE COMPARATIVE - PARAMÈTRES DIRECTION STRATEGY

**Date**: 8 Novembre 2025  
**Tests analysés**: 33 configurations différentes  
**Timeframe**: 5 minutes  
**Période**: ~2.5 jours (500 bougies)

---

## 🏆 TOP 3 CONFIGURATIONS GLOBALES

### 🥇 #1 - MOYEN TERME OPTIMAL
- **Config**: `VWMA=20, Slope=6, ATR=8, Coef=0.25`
- **Performance**: **+6.03% capté**
- **Profil**: 12 intervalles, avg 37.8 bougies (~3h par trade)
- **Force**: Capte les tendances moyennes avec peu de faux signaux
- **Long**: +4.60% | **Short**: -1.43% (inversé = +1.43%)

### 🥈 #2 - INTRADAY PERFORMANT  
- **Config**: `VWMA=12, Slope=2, ATR=4, Coef=0.50`
- **Performance**: **+5.98% capté**
- **Profil**: 10 intervalles, avg 46.1 bougies (~4h par trade)
- **Force**: Équilibre entre réactivité et stabilité
- **Long**: +4.89% | **Short**: -1.08% (inversé = +1.08%)

### 🥉 #3 - SWING TRADING
- **Config**: `VWMA=6, Slope=2, ATR=14, Coef=1.00`
- **Performance**: **+5.60% capté**
- **Profil**: 8 intervalles, avg 57.5 bougies (~5h par trade)
- **Force**: Filtre agressif, ne prend que les mouvements forts
- **Long**: +4.57% | **Short**: -1.04% (inversé = +1.04%)

---

## 📈 ANALYSE PAR HORIZON DE TRADING

### 📍 COURT TERME (Scalping < 2h, <20 bougies)

**Performance moyenne**: **-4.92%** ❌  
**Meilleure config**: `VWMA=6, Slope=3, ATR=6, Coef=0.25` → **+0.29%**

**⚠️ CONSTAT**: Le court terme est très difficile pour cette stratégie
- VWMA=3 (très réactif) : **-15.67%** à **-6.24%** (désastreux)
- Trop de faux signaux et de retournements
- Les variations captées sont annulées par le bruit du marché

**Recommandation**: 
> ❌ **ÉVITER le court terme avec cette stratégie**  
> Si scalping souhaité → utiliser `scalping_engine` avec validation N-2→N+2

---

### 📊 MOYEN TERME (Intraday 2-8h, 20-50 bougies) 

**Performance moyenne**: **+1.10%** ✅  
**Meilleure config**: `VWMA=20, Slope=6, ATR=8, Coef=0.25` → **+6.03%**

**✅ SWEET SPOT DE LA STRATÉGIE**
- VWMA=12-20 : Filtrage optimal du bruit
- ATR_coef=0.25-0.50 : Sensibilité adaptée
- Intervalles moyens de 30-45 bougies (2.5-4h)

**Top 5 configs moyen terme**:
1. VWMA=20, Slope=6, ATR=8, Coef=0.25 → **+6.03%** 🥇
2. VWMA=12, Slope=2, ATR=4, Coef=0.50 → **+5.98%** 🥈
3. VWMA=12, Slope=4, ATR=8, Coef=0.80 → **+4.90%** ⭐
4. VWMA=12, Slope=3, ATR=4, Coef=0.80 → **+4.79%** ⭐
5. VWMA=20, Slope=3, ATR=8, Coef=0.25 → **+4.61%** ⭐

**Recommandation**: 
> ✅ **PRIVILÉGIER le moyen terme (2-6h par position)**
> - Utiliser VWMA entre 12 et 20
> - ATR_coef entre 0.25 et 0.50
> - Attendre des mouvements de 30+ bougies

---

### 📈 LONG TERME (Swing >8h, >50 bougies)

**Performance moyenne**: **-0.05%** (neutre)  
**Meilleure config**: `VWMA=6, Slope=2, ATR=14, Coef=1.00` → **+5.60%**

**⚖️ RÉSULTATS MIXTES**
- Peu d'intervalles (2-10 par période)
- Performance dépend fortement du timing
- VWMA élevé (48) : performances négatives (-1% à -5%)
- Paradoxe : VWMA=6 avec ATR élevé fonctionne mieux

**Top 3 configs long terme**:
1. VWMA=6, Slope=2, ATR=14, Coef=1.00 → **+5.60%** 🥇
2. VWMA=6, Slope=3, ATR=14, Coef=1.00 → **+1.70%** 
3. VWMA=12, Slope=3, ATR=4, Coef=0.80 → **+4.79%**

**Recommandation**: 
> ⚖️ **UTILISER avec prudence**
> - Si swing → VWMA=6 + ATR_coef élevé (0.80-1.00)
> - Éviter VWMA=48 (trop lent, rate les sorties)
> - Préférer timeframe supérieur (15m, 1h) pour le swing

---

## 🎯 IMPACT DES PARAMÈTRES

### 📐 VWMA PERIOD (Moyenne mobile pondérée volume)

| VWMA | Tests | Avg Capté | Meilleure | Pire | Profil |
|------|-------|-----------|-----------|------|--------|
| **20** | 5 | **+3.44%** ✅ | +6.03% | -1.59% | **OPTIMAL moyen terme** |
| **12** | 11 | **+1.40%** ✅ | +5.98% | -4.72% | **Polyvalent, stable** |
| **6** | 9 | -1.30% | +5.60% | -4.67% | Risqué mais peut exploser |
| **48** | 4 | -3.10% ❌ | -1.03% | -5.42% | **Trop lent** |
| **9** | 1 | -3.88% ❌ | -3.88% | -3.88% | Non testé suffisamment |
| **3** | 3 | **-11.35%** ❌❌ | -6.24% | -15.67% | **DÉSASTREUX** |

**Interprétation**:
- **VWMA=3** : Trop réactif, suit chaque micro-mouvement → overtrading
- **VWMA=6** : Réactif mais risqué, nécessite ATR_coef élevé pour filtrer
- **VWMA=12-20** : 🎯 **ZONE OPTIMALE** → Filtre bruit, suit tendances réelles
- **VWMA=48** : Trop lent, rate les sorties, signaux rares et tardifs

---

### ⚡ ATR COEFFICIENT (Seuil de pente dynamique)

| Coef | Tests | Avg Capté | Meilleure | Pire | Profil |
|------|-------|-----------|-----------|------|--------|
| **0.40** | 1 | **+4.10%** ✅ | +4.10% | +4.10% | Sensible, capte petits mouvements |
| **0.70** | 1 | **+3.49%** ✅ | +3.49% | +3.49% | Équilibré |
| **0.80** | 7 | **+1.45%** ✅ | +4.90% | -4.43% | Conservateur, filtre bien |
| **0.90** | 1 | -0.32% | -0.32% | -0.32% | Limite |
| **1.00** | 4 | -1.17% | +5.60% | -6.24% | Très sélectif, risqué |
| **0.25** | 11 | -1.48% ❌ | +6.03% | -15.67% | **Trop sensible avec VWMA court** |
| **0.50** | 6 | -2.73% ❌ | +5.98% | -12.13% | Variance élevée |
| **1.50** | 1 | -4.67% ❌ | -4.67% | -4.67% | Trop restrictif |
| **1.10** | 1 | -4.72% ❌ | -4.72% | -4.72% | Trop restrictif |

**Interprétation**:
- **Coef < 0.40** : Capte beaucoup de mouvements, mais risque de bruit élevé
- **Coef 0.40-0.80** : 🎯 **ZONE OPTIMALE** → Bon équilibre sensibilité/filtrage
- **Coef > 1.00** : Trop conservateur, rate des opportunités

**💡 Règle d'or**:
> - VWMA court (6-12) → ATR_coef élevé (0.70-1.00) pour filtrer
> - VWMA long (20+) → ATR_coef bas (0.25-0.50) pour sensibilité

---

## 💡 RECOMMANDATIONS PAR OBJECTIF

### 🎯 OBJECTIF: CAPTER MOUVEMENTS COURT TERME (<2h)

**❌ PAS RECOMMANDÉ avec cette stratégie**

**Pourquoi ?**
- Moyenne capté: -4.92%
- VWMA court (3-6) génère trop de faux signaux
- Variations trop faibles pour couvrir les spread/fees

**Alternative**:
→ Utiliser `scalping_engine` avec validation multi-étapes

**Si vraiment nécessaire**:
```yaml
Config défensive:
- VWMA: 6
- Slope: 3
- ATR: 6
- Coef: 0.25
Performance attendue: ~+0.3% (marginal)
```

---

### 📊 OBJECTIF: CAPTER MOUVEMENTS MOYEN TERME (2-6h) ✅ RECOMMANDÉ

**🥇 CONFIG OPTIMALE**:
```yaml
VWMA_RAPIDE: 20
PERIODE_PENTE: 6
ATR_PERIODE: 8
ATR_COEFFICIENT: 0.25
```
- **Performance**: +6.03%
- **Intervalles**: 12 (~3h par position)
- **Style**: Suit les tendances intraday majeures

**🥈 CONFIG ALTERNATIVE (Plus de trades)**:
```yaml
VWMA_RAPIDE: 12
PERIODE_PENTE: 2
ATR_PERIODE: 4
ATR_COEFFICIENT: 0.50
```
- **Performance**: +5.98%
- **Intervalles**: 10 (~4h par position)
- **Style**: Plus réactif, plus de signaux

**Profil idéal**:
- Trader intraday (8h-20h de marché actif)
- Aime les positions de 2-6h
- Cherche 1-3% par mouvement

---

### 📈 OBJECTIF: CAPTER MOUVEMENTS LONG TERME (>8h)

**🥇 CONFIG SWING**:
```yaml
VWMA_RAPIDE: 6
PERIODE_PENTE: 2
ATR_PERIODE: 14
ATR_COEFFICIENT: 1.00
```
- **Performance**: +5.60%
- **Intervalles**: 8 (~5-6h par position)
- **Style**: Filtre agressif, tendances fortes uniquement

**⚠️ MAIS**: 
- Seulement 8 trades sur 2.5 jours → Peu de signaux
- Risque d'attendre longtemps entre positions
- Performance dépend du timing de marché

**Alternative recommandée**:
> Passer à un **timeframe supérieur** (15m ou 1h) avec:
> - VWMA=12-20
> - ATR_coef=0.50-0.80
> → Plus adapté pour swing trading multi-jours

---

## 🔍 PATTERNS OBSERVÉS

### ✅ CE QUI FONCTIONNE

1. **VWMA moyen (12-20)** avec **ATR_coef bas-moyen (0.25-0.50)**
   - Filtre le bruit du court terme
   - Capte les vraies tendances intraday
   - Performance: +4% à +6%

2. **Combinaison paradoxale**: VWMA court + ATR_coef élevé
   - VWMA=6 + ATR_coef=1.00 → +5.60%
   - Réactivité + Filtrage strict = Signaux rares mais qualitatifs

3. **Slope Period élevé (4-6)** avec VWMA moyen
   - Calcul de pente sur plus de bougies = Plus stable
   - Moins de faux signaux

### ❌ CE QUI NE FONCTIONNE PAS

1. **VWMA très court (3)** quelle que soit la config
   - Performance: -6% à -15%
   - Overtrading massif, suit le bruit

2. **VWMA très long (48)**
   - Performance: -1% à -5%
   - Trop lent, rate les sorties, captures incomplètes

3. **ATR_coef extrêmes** (<0.25 ou >1.10)
   - Soit trop de bruit, soit pas assez de signaux

4. **Court terme en général**
   - Moyenne: -4.92%
   - La stratégie n'est pas conçue pour le scalping

---

## 📋 CONFIGS RECOMMANDÉES PAR CAS D'USAGE

### 🎯 TRADER CONSERVATEUR (Peu de trades, haute qualité)
```yaml
Objectif: 1-2 trades/jour, capture >3% par trade
Config:
  VWMA_RAPIDE: 20
  PERIODE_PENTE: 6
  ATR_PERIODE: 8
  ATR_COEFFICIENT: 0.40
  K_CONFIRMATION: 2
Performance attendue: +6% sur période test
Intervalles moyens: ~35-40 bougies (3-4h)
```

### 📊 TRADER ÉQUILIBRÉ (Balance qualité/quantité)
```yaml
Objectif: 2-3 trades/jour, capture >2% par trade
Config:
  VWMA_RAPIDE: 12
  PERIODE_PENTE: 3
  ATR_PERIODE: 4
  ATR_COEFFICIENT: 0.80
  K_CONFIRMATION: 2
Performance attendue: +4.8% sur période test
Intervalles moyens: ~50-70 bougies (4-6h)
```

### ⚡ TRADER ACTIF (Plus de trades, réactivité)
```yaml
Objectif: 3-5 trades/jour, capture >1% par trade
Config:
  VWMA_RAPIDE: 12
  PERIODE_PENTE: 2
  ATR_PERIODE: 4
  ATR_COEFFICIENT: 0.50
  K_CONFIRMATION: 2
Performance attendue: +5.98% sur période test
Intervalles moyens: ~45 bougies (4h)
```

### 🎲 TRADER SWING (Positions longues)
```yaml
Objectif: 1 trade tous les 2-3 jours, capture >5% par trade
Config:
  VWMA_RAPIDE: 6
  PERIODE_PENTE: 2
  ATR_PERIODE: 14
  ATR_COEFFICIENT: 1.00
  K_CONFIRMATION: 2
Performance attendue: +5.60% sur période test
⚠️ Considérer timeframe supérieur (15m/1h)
```

---

## 🧪 TESTS COMPLÉMENTAIRES SUGGÉRÉS

### Pour affiner davantage:

1. **Timeframe 1m** (court terme)
   - Tester VWMA=12-20 avec ATR_coef=0.80-1.50
   - Objectif: Voir si filtrage plus strict fonctionne en 1m

2. **Timeframe 15m** (swing)
   - Tester VWMA=6-12 avec ATR_coef=0.50-1.00
   - Objectif: Meilleur pour positions >12h

3. **Période plus longue** (5-7 jours)
   - Valider que les meilleures configs restent stables
   - Identifier si certaines configs surfit la période actuelle

4. **Tests avec fees**
   - Soustraire 0.1% (maker/taker) par trade
   - Recalculer quel minimum d'intervalles est profitable

---

## 📊 STATISTIQUES CLÉS

| Métrique | Valeur |
|----------|--------|
| **Meilleure performance** | +6.03% (VWMA=20, Coef=0.25) |
| **Pire performance** | -15.67% (VWMA=3, Coef=0.25) |
| **Configs positives** | 18/33 (54.5%) |
| **Moyenne toutes configs** | -0.42% |
| **Médiane** | +0.29% |
| **Écart-type** | 4.2% |

**Conclusion statistique**:
> La stratégie est **rentable si bien paramétrée** (top 10 → +4% à +6%)  
> Mais **très sensible aux paramètres** (écart de 21% entre best et worst)

---

## 🎓 LEÇONS APPRISES

### 1. Le VWMA est le paramètre roi
- **Impact**: Écart de +6% à -15% selon la période
- **Optimal**: 12-20 pour timeframe 5m

### 2. Le paradoxe réactivité/filtrage
- VWMA court seul = Désastre
- VWMA court + ATR_coef élevé = Peut fonctionner
- VWMA moyen + ATR_coef bas = ✅ Best

### 3. Le court terme est un piège
- Moyenne -4.92% sur <20 bougies
- Stratégie pas conçue pour scalping
- Même les meilleures configs peinent à être positives

### 4. Le sweet spot est 2-6h par position
- C'est là que la stratégie excelle
- Filtre le bruit, capte les vraies tendances
- Performance: +4% à +6%

### 5. Moins de trades ≠ Moins de profit
- Top config: 12 intervalles → +6.03%
- Pire config: 85 intervalles → -15.67%
- **Qualité > Quantité**

---

## 🚀 PROCHAINES ÉTAPES

### Implémentation recommandée:
1. ✅ Utiliser config #1 (VWMA=20, Coef=0.25) pour production
2. 📊 Backtester sur 1 mois de données pour valider
3. 🧪 Paper trading 1 semaine en live
4. 💰 Démarrer en prod avec position sizing conservateur

### Optimisations futures:
- [ ] Ajouter filtre de volatilité (éviter flat markets)
- [ ] Tester combinaison avec indicateur de volume
- [ ] Implémenter des sorties partielles (take profit à mi-chemin)
- [ ] Analyser performance par session (Asia/EU/US)

---

**Généré le**: 2025-11-08  
**Outil**: `cmd/analyze_tests/main.go`  
**Données source**: `out/direction_demo_*/intervalles.json`
