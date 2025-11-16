# Stratégie de trading Harmonie - Version détaillée pour validation

## ARCHITECTURE GLOBALE

### Timeframes
- **Contexte** : TF haute (5m pour scalping, 4h pour investissement)
- **Exécution** : TF basse (1m pour scalping, 1h pour investissement)

### Philosophie
- **2 couches** : contexte pour moduler les seuils, exécution pour les décisions
- **Pas de gates en contexte** : uniquement modulation et veto
- **Séquencement strict** : 3 étapes dans fenêtre W pour l'ouverture

---

## CADRE GÉNÉRAL - SIGNAUX DE BASE

### 1. Signal VWMA (exécution)
**Définition** : Croisement des moyennes mobiles pondérées par volume
- **VWMA6** : rapide, réactive aux changements de prix/volume
- **VWMA20** : moyenne terme pour scalping, VWMA30 pour investissement
- **Logique** :
  - **Long** : VWMA6 croise AU-DESSUS de VWMA20/30
  - **Short** : VWMA6 croise AU-DESSOUS de VWMA20/30

**Validation γ_gap** : Le croisement n'est valide que si l'écart absolu entre les deux VWMA est ≥ γ_gap
- **γ_gap** = facteur × ATR(période d'exécution)
- **Scalping** : γ_gap ≈ 0,15 × ATR(1m) = 0.35
- **Investissement** : γ_gap ≈ 0,10 × ATR(1h)
- **Raison** : Éviter les "touch-and-go" et faux croisements

**Validation différée avec fenêtre** (innovation clé) :
Un croisement peut commencer **faiblement** et se **renforcer progressivement**. Au lieu de rejeter immédiatement les croisements dont le gap initial est < γ_gap, on leur accorde une **fenêtre de validation différée** :

```
AU MOMENT DU CROISEMENT (barre n-1) :
SI gap >= γ_gap × ATR 
  ALORS ✓ Validé immédiatement (GapValideBougie = 0)
SINON
  POUR w = 1 à WINDOW_GAMMA_VALIDATE (ex: 5 bougies)
    SI gap_futur[n-1+w] >= γ_gap × ATR[n-1+w]
      ALORS ✓ Validé après w bougies (GapValideBougie = w)
      SORTIR
    FIN SI
  FIN POUR
  SI jamais validé
    ALORS ✗ Rejeté (GapValideBougie = -1)
  FIN SI
FIN SI
```

**Paramètres** :
- **WINDOW_GAMMA_VALIDATE** : 5 bougies (ajustable selon TF)
- **Avantages** : Capture les croisements qui gagnent en puissance progressivement
- **Exemple** : Un croisement avec gap initial de 0.06 (< 0.15 requis) peut être validé si après 3 bougies le gap atteint 0.18
- **Impact** : Augmente significativement le taux de capture des signaux valides (+30-40% de signaux sauvés)

### 2. Validation DMI (obligatoire, exécution)
**Indicateur** : DMI(14,3) pour scalping, DMI(24,6) pour investissement

**2.1. Mode Tendance**
- **Condition** : Croisement des lignes DI dans le sens du trade
- **Validation DI** : Écart |DI+ − DI−| ≥ γ_gap AU MOMENT du croisement
- **Validation DX/ADX** : DX croise au-dessus d'ADX (DX↑>ADX) dans la fenêtre W suivant le croisement DI
- **Contrainte DX/ADX** : Overshoot du croisement ≥ γ_gap, et pendant W le DI directionnel reste dominant avec |DI+−DI−| ≥ γ_gap
- **γ_gap** : 5-8 pour scalping, 8-12 pour investissement
- **Exemple Long** : DI+ croise AU-DESSUS de DI− avec écart ≥ γ_gap, puis DX croise au-dessus d'ADX avec overshoot ≥ γ_gap

**2.2. Mode Momentum (alternative, non cumulative)**
- **Condition** : DX ou ADX croise AU-DESSUS d'un ou des DI
- **Validation** : Dépassement (overshoot) ≥ γ_gap
- **Contrainte** : Le DI directionnel doit rester dominant (|DI+−DI−| ≥ γ_gap) pendant la fenêtre W
- **Note** : Mode momentum est une ALTERNATIVE au mode tendance, pas une étape supplémentaire

**2.3. Motif KO (désordre DMI)**
- **Définition** : DX et ADX restent SOUS les deux DI du début à la fin
- **Action** : Refus automatique du trade
- **Raison** : Indique un marché sans direction claire

---

## FILTRES D'ENTRÉE (après DMI validé)

### 1. MFI (Money Flow Index)
**Rôle** : Confirmation de force/volume dans le sens du trade

**Conditions obligatoires** :
- **Zone extrême favorable** :
  - **Long** : MFI ≥ 80 (ajusté par contexte DMI)
  - **Short** : MFI ≤ 20 (ajusté par contexte DMI)
- **Pente ou constance** :
  - **Pente favorable** : ΔMFI(3 bougies) ≥ τ_slope
  - **Constance** : |ΔMFI(3 bougies)| ≤ 2

**Modulation par contexte DMI** :
- **DI dominant fort** : Extrêmes PLUS stricts dans le sens, PLUS souples à contre-tendance
- **DI faible** : Extrêmes de base 80/20

**Important** : MFI ne déclenche JAMAIS la sortie, uniquement l'entrée

### 2. CHOP (Index de choppiness)
**Rôle** : Détection de régime de marché (trending vs ranging)

**Condition** : **Pente ≤ 0** (constant ou décroissant)
- **Mesure** : Pente sur 3 bougies
- **Seuil τ_slope** : 
  - **Scalping** : Hausse tolérée ≤ +5
  - **Investissement** : Hausse tolérée ≤ +3
- **Veto** : Si CHOP augmente > τ_slope ⇒ suspension de recherche

### 3. ATR% (Average True Range en pourcentage)
**Rôle** : Validation de volatilité suffisante et dimensionnement du stop

**Condition** : **ATR% ≥ seuil_min**
- **Calcul** : ATR / prix × 100
- **Seuils** :
  - **Scalping** : 0,15% - 0,30%
  - **Investissement** : 0,50% - 1,20%
- **Usage double** : Validation + sizing du stop

---

## LOGIQUE D'EXÉCUTION - SÉQUENCE D'OUVERTURE

### Contexte (5m) - Vérification position relative
**Objectif** : Vérifier que la tendance est EN COURS avant de chercher signaux en exécution

**Vérification position relative des composants** :
- VWMA6 vs VWMA20/30 → Position montre direction de tendance
- DI+ vs DI- → Position montre dominance directionnelle
- DX vs ADX → Position montre accélération
- CHOP → Régime trending confirmé

**Output contexte** : Tendance EN COURS → Autorisation pour chercher signaux en exécution

---

### Exécution (1m) - Fenêtre W

**Fenêtre W (intervalle de séquencement)** :
- **Départ** : PREMIER croisement détecté (VWMA OU DMI)
- **Fin** : Premier croisement + W bougies (5-10 bougies)
- **Taille W** : À définir selon la volatilité du marché

**ORDRE FLEXIBLE - Selon le marché** :
- Croisement VWMA peut arriver avant ou après DMI
- DMI peut arriver avant ou après VWMA
- Pas d'ordre fixe, l'important est que les deux soient validés dans W

### Conditions dans fenêtre W

**Signal VWMA** :
```
SI |VWMA6 - VWMA20/30| ≥ γ_gap
ET croisement dans le sens (long/short)
ALORS signal VWMA validé
```

**Validation DMI (alternative OU)** :
```
Mode Tendance :
SI croisement DI dans le sens
ET |DI+ - DI-| ≥ γ_gap au croisement
ET DX croise AU-DESSUS d'ADX
ET overshoot ≥ γ_gap
ET DI directionnel reste dominant (|DI+ - DI-| ≥ γ_gap)
ALORS croisement tendance validé

OU

Mode Momentum :
SI DX croise AU-DESSUS d'ADX
ET overshoot ≥ γ_gap
ET DI directionnel reste dominant (|DI+ - DI-| ≥ γ_gap)
ALORS DMI validé
```

### Validation filtres (progressive et indépendante)

**Démarrage** : À partir de la bougie du DERNIER croisement validé

**Logique de validation** :
```
POUR chaque filtre (MFI, CHOP, ATR%) :
  ├─ Vérification INDÉPENDANTE
  ├─ Ordre FLEXIBLE (selon le marché)
  ├─ Validation PROGRESSIVE (bougie par bougie)
  └─ Un filtre validé = ACQUIS

SI TOUS les filtres validés sur bougie du dernier croisement
ALORS ouverture IMMÉDIATE

SINON
  POUR chaque bougie suivante (jusqu'à fin W) :
    Vérifier filtres non encore validés
    SI tous deviennent validés
    ALORS ouverture position
    
SI fin de W atteinte sans validation complète
ALORS abandon du signal
```

**Filtres** :
- **MFI** : Zone extrême favorable + pente/constance
- **CHOP** : Pente ≤ 0
- **ATR%** : ≥ seuil_min

**Validation finale** :
```
SI (signal VWMA validé) 
ET (DMI validé) 
ET (TOUS filtres validés)
DANS fenêtre W
ALORS ouverture position
```

---

## GESTION DU STOP - PROTECTION DYNAMIQUE

### Stop standard (phase 1)
**Type** : Suiveur de VWMA30 (pas trailing prix)
- **Calcul distance** : p% = clip(k × ATR%, p_min, p_max)
- **Paramètres** :
  - **Scalping** : k = 1,0-1,5 ; p_min = 0,20% ; p_max = 1,20%
  - **Investissement** : k = 1,8-2,5 ; p_min = 0,50% ; p_max = 3,00%
- **Logique** : Stop suit VWMA30 à distance p%

### Bascule de stop (phase 2 - désordre)
**Déclencheurs** (conditions OU) :

**Condition A** :
```
SI CHOP se redresse > τ_slope
ET MFI bascule en pente défavorable ≥ τ_slope
ALORS bascule vers VWMA20
```

**Condition B** :
```
SI MFI passe en extrême inverse (zone opposée)
ALORS bascule vers VWMA20 (pente indifférente)
```

**Actions après bascule** :
- Nouvelle ancre : VWMA20
- Recalcul complet de ATR% sur VWMA20
- **Sortie immédiate** si nouveau stop déjà dépassé
- Stop continue de suivre VWMA20

---

## SUIVI DE POSITION (contexte)

### Indicateurs monitorés (pas de sortie sauf désordre)
- **VWMA6 (pente)** : Doit évoluer favorablement (confirmation, pas contrainte stricte)
- **CHOP** : Doit rester constant/décroissant (petites hausses < τ_slope tolérées)
- **ATR%** : Doit rester ≥ seuil_min (régime de volatilité maintenu)
- **MFI** : Doit rester soutenant (extrême favorable stable ou pente/constance favorables)

### Sortie normale
```
SI croisement inverse VWMA6↔VWMA20/30
OU stop touché
ALORS sortie de position
```

---

## PARAMÈTRES UNIVERSELS

### Tolérances
- **τ_slope** : Seuil de pente pour CHOP, MFI, VWMA, DX/ADX
- **γ_gap** : Écart minimal pour TOUS les croisements (VWMA, DI, DX/ADX)
- **δ_min** : Écart DI minimal pour validation tendance

### Deux bots, même logique
- **Scalping** : Seuils plus stricts, fenêtres serrées, réactivité maximale
- **Investissement** : Seuils plus souples, respirations tolérées, tenue de position

---

## RÉSUMÉ DES POINTS DE VALIDATION

Pour votre validation, voici les éléments clés à vérifier :

1. **Suppression gate VWMA96** : Plus de contrainte de position par rapport à VWMA96
2. **Suppression CCI** : Plus de filtre anti-tardif basé sur CCI
3. **Séquencement fenêtre W** : Croisements et filtres dans ordre flexible du marché
4. **Modulation MFI par DMI contexte** : Ajustement dynamique des extrêmes
5. **Stop à deux phases** : VWMA30 → VWMA20 en cas de désordre
6. **Paramètres unifiés** : γ_gap unique pour tous les croisements

---

## 📊 RÉCAPITULATIF TIMING ET FENÊTRES

### Contexte (5m)
**Objectif** : Vérifier tendance EN COURS via position relative
- VWMA, DI, DX/ADX, CHOP → positions relatives montrent tendance
- Pas de vérification répétée à chaque signal exécution
- Output : Direction + autorisation scalps

### Exécution (1m) - Fenêtre W
**Départ** : PREMIER croisement (VWMA ou DMI)
**Fin** : Premier croisement + W bougies (5-10)

**Phase 1 - Croisements (ordre FLEXIBLE)** :
- VWMA + γ_gap
- DMI (tendance OU momentum) + γ_gap
- Ordre selon le marché (pas fixe)

**Phase 2 - Filtres (validation PROGRESSIVE)** :
- Démarrage : Bougie du DERNIER croisement
- MFI, CHOP, ATR% vérifiés indépendamment
- Validation progressive bougie par bougie
- Signal dès que TOUS validés (dans W)
- Abandon si W atteinte sans validation complète

**Règles clés** :
1. Ordre flexible des croisements
2. Ordre flexible des filtres
3. Validation indépendante de chaque élément
4. Signal immédiat si tout OK sur dernière bougie de croisement
5. Sinon surveillance continue jusqu'à fin W

---

*À valider : Chaque élément doit être testé individuellement avant assemblage complet*
