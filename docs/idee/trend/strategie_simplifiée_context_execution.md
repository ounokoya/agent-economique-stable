# Stratégie Simplifiée - Architecture Context/Execution

## 📋 Vision stratégique

**Principe** : 
- **Contexte** : Identifier et valider une tendance solide
- **Exécution** : Scalper les momentum dans le sens de la tendance

---

## 🔵 CONTEXTE (5m) - Filtre tendance

### Objectif
Établir une direction trending claire et forte avant d'autoriser les scalps

### Règles contexte

#### 1. Signal VWMA
- **Croisement VWMA6↔VWMA20/30** avec validation γ_gap
- **Validation** : |VWMA6 - VWMA20/30| ≥ γ_gap
- **Rôle** : Identifier la direction de la tendance (long/short)

#### 2. DMI Mode Tendance (complet)
**Étape A - Croisement DI**
- Croisement DI dans le sens du trade
- Écart |DI+ − DI−| ≥ γ_gap au moment du croisement

**Étape B - Validation DX/ADX**
- DX croise au-dessus d'ADX (DX↑>ADX)
- Overshoot du croisement ≥ γ_gap
- DI directionnel reste dominant (|DI+−DI−| ≥ γ_gap)

**Rôle** : Confirmer la force et l'accélération de la tendance

#### 3. CHOP (Choppiness Index)
- **Condition** : Pente ≤ 0 (constant ou décroissant)
- **Mesure** : Pente sur 3 bougies
- **Seuil variation pente** : τ_slope (hausse tolérée ≤ +5 pour scalping)
- **Veto** : Si CHOP augmente > τ_slope ⇒ pas de scalps
- **Rôle** : Vérifier que le marché est en régime trending, pas ranging

### Fenêtre W_context - Validation progressive

**Fenêtre W_context** :
- **Départ** : PREMIER croisement détecté (VWMA, DI, ou DX/ADX)
- **Fin** : Premier croisement + W bougies (quelques bougies 5m)

**ORDRE FLEXIBLE - Selon le marché** :
- Croisements peuvent arriver dans n'importe quel ordre
- VWMA → DI → DX/ADX
- DI → VWMA → DX/ADX
- DX/ADX → VWMA → DI
- ... etc.

**Logique de validation** :
```
POUR chaque condition (VWMA, DI, DX/ADX, CHOP) :
  ├─ Vérification INDÉPENDANTE
  ├─ Ordre FLEXIBLE (selon le marché)
  ├─ Validation PROGRESSIVE (bougie par bougie)
  └─ Une condition validée = ACQUISE

SI TOUTES les conditions validées sur bougie du dernier croisement
ALORS contexte IMMÉDIATEMENT validé

SINON
  POUR chaque bougie suivante (jusqu'à fin W_context) :
    Vérifier conditions non encore validées
    SI toutes deviennent validées
    ALORS contexte établi
    
SI fin de W_context atteinte sans validation complète
ALORS pas de contexte, attendre nouveau cycle
```

**Validation contexte** :
```
SI croisement VWMA valide (γ_gap)
ET DMI tendance validé (DI + DX/ADX avec γ_gap)
ET CHOP décroissant/constant (pente ≤ 0)
DANS fenêtre W_context
ALORS contexte trending établi
→ Autoriser recherche de scalps en exécution
```

**Vérification position relative** :
Une fois contexte établi, pas de revérification continue
- Position relative des composants montre tendance EN COURS
- Contexte reste valide tant que position relative maintenue
- Pas de vérification à chaque signal d'exécution

### Output contexte
- **Direction** : Long ou Short
- **Statut** : Tendance EN COURS (validée) ou refusée
- **Autorisation** : Scalps autorisés ou interdits

---

## 🟢 EXÉCUTION (1m) - Scalper momentum

### Objectif
Détecter et exploiter les impulsions momentum dans le sens du contexte

### Règles exécution

#### 1. DMI Mode Momentum UNIQUEMENT

**Deux types de momentum selon l'alignement avec le contexte :**

**📈 Cas 1 : Momentum aligné (accélération tendance)**
- **Contexte LONG** + DI+ > DI- en exécution
  - Croisement : DX↑ > ADX (accélération haussière)
  - Overshoot ≥ γ_gap
  - DX ≥ seuil_DX au moment du croisement
  - DI+ reste dominant (|DI+ - DI-| ≥ γ_gap)
  - **Logique** : Scalper l'accélération dans le sens de la tendance

- **Contexte SHORT** + DI- > DI+ en exécution
  - Croisement : DX↑ > ADX (accélération baissière)
  - Overshoot ≥ γ_gap
  - DX ≥ seuil_DX au moment du croisement
  - DI- reste dominant (|DI- - DI+| ≥ γ_gap)
  - **Logique** : Scalper l'accélération dans le sens de la tendance

**📉 Cas 2 : Momentum pullback (ralentissement contre-tendance)**
- **Contexte LONG** + DI- > DI+ en exécution (pullback baissier local)
  - Croisement : DX↓ < ADX (ralentissement du pullback)
  - Undershoot ≥ γ_gap
  - DX ≥ seuil_DX au moment du croisement
  - DI- dominant mais s'affaiblissant
  - **Logique** : Scalper la fin du pullback pour reprendre la tendance LONG

- **Contexte SHORT** + DI+ > DI- en exécution (pullback haussier local)
  - Croisement : DX↓ < ADX (ralentissement du pullback)
  - Undershoot ≥ γ_gap
  - DX ≥ seuil_DX au moment du croisement
  - DI+ dominant mais s'affaiblissant
  - **Logique** : Scalper la fin du pullback pour reprendre la tendance SHORT

**Rôle** : Détecter les impulsions momentum = points d'entrée pour scalps (accélération OU fin de pullback)

#### 2. MFI (Money Flow Index)
- **Zone extrême favorable** :
  - Long : MFI ≥ 80 (ajusté par contexte DMI)
  - Short : MFI ≤ 20 (ajusté par contexte DMI)
- **Pente ou constance** :
  - Pente favorable : ΔMFI(3 bougies) ≥ τ_slope (variation minimale pour être considérée)
  - Constance : |ΔMFI(3 bougies)| ≤ 2 (variation négligeable)
- **Modulation** : DI dominant fort ⇒ extrêmes plus stricts dans le sens
- **Rôle** : Confirmer la force du momentum avec volume

#### 3. CHOP local (exécution)
- **Condition** : Pente ≤ 0 (constant/décroissant)
- **Seuil variation pente** : τ_slope (hausse tolérée ≤ +5)
- **Veto local** : Si CHOP augmente > τ_slope ⇒ pas d'entrée
- **Rôle** : Vérifier que le momentum n'est pas dans du bruit

#### 4. ATR% (volatilité)
- **Condition** : ATR% ≥ seuil_min
- **Seuils** : 0,15-0,30% pour scalping 1m
- **Rôle** : Volatilité suffisante pour scalper + sizing du stop

### Fenêtre W_exec - Validation progressive

**PRÉREQUIS** : Contexte trending EN COURS (validé et maintenu)

**Fenêtre W_exec** :
- **Départ** : Détection croisement DX/ADX
- **Fin** : Croisement DX/ADX + W bougies (quelques bougies 1m)

**Logique de validation filtres** :
```
Bougie M : Momentum DX/ADX détecté
  ├─ DX/ADX croisement validé ✓
  ├─ DX ≥ seuil_DX ✓
  ├─ DI dominant maintenu ✓
  │
  └─ Vérifier FILTRES sur CETTE BOUGIE :
     
     POUR chaque filtre (MFI, CHOP, ATR%) :
       ├─ Vérification INDÉPENDANTE
       ├─ Ordre FLEXIBLE (selon le marché)
       ├─ Validation PROGRESSIVE
       └─ Un filtre validé = ACQUIS
     
     SI TOUS validés sur bougie M
     ALORS scalp IMMÉDIAT
     
     SINON
       POUR chaque bougie suivante (M+1, M+2, ... jusqu'à fin W_exec) :
         Vérifier filtres non encore validés
         SI tous deviennent validés
         ALORS ouverture scalp
       
       SI fin de W_exec sans validation complète
       ALORS abandon du signal, attendre prochain momentum
```

### Validation exécution

**Cas 1 - Momentum aligné (accélération) :**
```
PRÉREQUIS : Contexte trending établi (direction validée)

SIGNAL MOMENTUM :
SI DI exécution aligné avec contexte (même direction)
ET DX croise au-dessus d'ADX (DX↑>ADX, overshoot ≥ γ_gap)
ET DX ≥ seuil_DX au moment du croisement
ET DI directionnel reste dominant (≥ γ_gap)
ALORS momentum accélération détecté → Vérifier filtres

FILTRES (validation progressive dans W_exec) :
├─ MFI extrême favorable + pente/constance (≥ τ_slope)
├─ CHOP exécution ≤ 0 (variation < τ_slope)
└─ ATR% ≥ seuil_min

SI momentum validé ET TOUS filtres validés DANS W_exec
ALORS entrée scalp ACCÉLÉRATION
```

**Cas 2 - Momentum pullback (fin de contre-tendance) :**
```
PRÉREQUIS : Contexte trending établi (direction validée)

SIGNAL MOMENTUM :
SI DI exécution OPPOSÉ au contexte (pullback local)
ET DX croise en dessous d'ADX (DX↓<ADX, undershoot ≥ γ_gap)
ET DX ≥ seuil_DX au moment du croisement
ET DI contre-tendance s'affaiblit
ALORS momentum pullback détecté → Vérifier filtres

FILTRES (validation progressive dans W_exec) :
├─ MFI extrême favorable pour REPRISE + pente/constance (≥ τ_slope)
├─ CHOP exécution ≤ 0 (variation < τ_slope)
└─ ATR% ≥ seuil_min

SI momentum validé ET TOUS filtres validés DANS W_exec
ALORS entrée scalp PULLBACK
```

### Output exécution
- **Entrée** : Oui ou Non
- **Stop** : Suiveur VWMA30 à p%(ATR%)
- **Sortie** : Stop touché ou croisement inverse VWMA6↔20/30

---

## 🎯 Tableau récapitulatif des règles

### CONTEXTE (5m → valide direction)
| Règle | Indicateur | Validation | Rôle |
|-------|-----------|-----------|------|
| Direction tendance | VWMA6↔20/30 | γ_gap | Sens long/short |
| Force DI | DI croisement | γ_gap | Tendance DI |
| Accélération | DX/ADX > ADX | overshoot γ_gap | Momentum tendance |
| Régime trending | CHOP | pente ≤ 0 | Anti-ranging |

**Output** : Direction (long/short) + autorisation scalps

---

### EXÉCUTION (1m → scalpe momentum)

**Cas 1 - Momentum aligné (accélération tendance)**
| Règle | Indicateur | Validation | Rôle |
|-------|-----------|-----------|------|
| Détection momentum | DX↑>ADX + DX≥seuil | overshoot γ_gap | Point d'entrée accélération |
| Alignement DI | DI exécution | même sens contexte | Cohérence directionnelle |
| Force momentum | MFI | extrême + pente | Volume confirmé |
| Pas de bruit | CHOP local | pente ≤ 0 | Qualité signal |
| Volatilité OK | ATR% | ≥ seuil_min | Sizing stop |

**Cas 2 - Momentum pullback (fin contre-tendance)**
| Règle | Indicateur | Validation | Rôle |
|-------|-----------|-----------|------|
| Détection ralentissement | DX↓<ADX + DX≥seuil | undershoot γ_gap | Point d'entrée pullback |
| Pullback local | DI exécution | opposé au contexte | Détection pullback |
| Affaiblissement | DI contre-tendance | s'affaiblit | Fin pullback |
| Force reprise | MFI | extrême reprise + pente | Volume reprise |
| Pas de bruit | CHOP local | pente ≤ 0 | Qualité signal |
| Volatilité OK | ATR% | ≥ seuil_min | Sizing stop |

**Output** : Entrée scalp si tous validés (accélération OU pullback)

---

## ✅ Avantages de cette architecture

### 1. Séparation claire des rôles
- **Contexte** = Filtre tendance (solide, lent)
- **Exécution** = Scalps momentum (rapide, réactif)

### 2. Pas de redondance
- **VWMA uniquement en contexte** (trop lent pour scalps 1m)
- **DMI tendance en contexte** (établir direction)
- **DMI momentum en exécution** (détecter impulsions)

### 3. Protection multicouche
- **Contexte refuse** si pas trending ⇒ pas de scalps
- **Exécution filtre** momentum faibles ou bruités
- **CHOP double** : contexte ET local
- **Deux types d'opportunités** : Accélération tendance ET fin de pullback

### 4. Exploitation complète des mouvements
- **Accélération** : Entrées dans le sens du momentum fort
- **Pullback** : Entrées à la reprise après correction
- **Couverture totale** : Ne manque aucune opportunité trending

### 5. Scalabilité
- Même logique pour investissement 4h/1h
- Ajustement des paramètres γ_gap et τ_slope uniquement

---

## 🔧 Paramètres

### Contexte 5m
- **VWMA** : 6/20 ou 6/30
- **DMI** : 14,3
- **γ_gap (VWMA)** : ≈ 0,15 × ATR(5m)
- **γ_gap (DMI)** : 5-8
- **CHOP** : len 14, τ_slope = +5

### Exécution 1m
- **DMI** : 14,3
- **γ_gap (momentum)** : 5-8
- **seuil_DX** : valeur minimale de DX au croisement DX/ADX (à définir)
- **MFI** : len 14, extrêmes 80/20 ajustés
- **τ_slope** : +5 (seuil variation pente pour MFI et CHOP)
- **CHOP** : len 14, τ_slope = +5
- **ATR%** : len 24, min 0,15-0,30%
- **Stop** : VWMA30, k=1,0-1,5, p_min=0,20%, p_max=1,20%

---

## 📊 Explication détaillée des deux cas de momentum

### 🎯 Comprendre les croisements DX/ADX

**DX/ADX ne donne PAS la direction, seulement l'accélération de la directionnalité**

- **DX > ADX** = La directionnalité (quelle qu'elle soit) s'accélère
- **DX < ADX** = La directionnalité ralentit/s'essouffle
- **Direction** donnée par : DI+ dominant (hausse) ou DI- dominant (baisse)

### 📈 Exemple LONG - Contexte haussier

**Cas 1 : Accélération tendance**
```
Contexte 5m : LONG validé (VWMA6>20, DI+>DI-, DX>ADX)
Exécution 1m :
- DI+ > DI- (tendance locale haussière alignée)
- DX croise AU-DESSUS d'ADX (DX↑>ADX)
→ Interprétation : La tendance haussière ACCÉLÈRE
→ Entrée LONG pour scalper l'accélération
```

**Cas 2 : Fin de pullback**
```
Contexte 5m : LONG validé (tendance globale haussière)
Exécution 1m :
- DI- > DI+ (pullback baissier LOCAL)
- DX croise EN DESSOUS d'ADX (DX↓<ADX)
→ Interprétation : Le pullback baissier RALENTIT/s'essouffle
→ Entrée LONG pour scalper la REPRISE de tendance
```

### 📉 Exemple SHORT - Contexte baissier

**Cas 1 : Accélération tendance**
```
Contexte 5m : SHORT validé (VWMA6<20, DI->DI+, DX>ADX)
Exécution 1m :
- DI- > DI+ (tendance locale baissière alignée)
- DX croise AU-DESSUS d'ADX (DX↑>ADX)
→ Interprétation : La tendance baissière ACCÉLÈRE
→ Entrée SHORT pour scalper l'accélération
```

**Cas 2 : Fin de pullback**
```
Contexte 5m : SHORT validé (tendance globale baissière)
Exécution 1m :
- DI+ > DI- (pullback haussier LOCAL)
- DX croise EN DESSOUS d'ADX (DX↓<ADX)
→ Interprétation : Le pullback haussier RALENTIT/s'essouffle
→ Entrée SHORT pour scalper la REPRISE de tendance
```

### 🎯 Tableau des 4 configurations possibles

| Contexte | DI exécution | Situation | Croisement DX/ADX | Action |
|----------|--------------|-----------|-------------------|--------|
| **LONG** | DI+ > DI- | Aligné | DX↑>ADX | ✅ Scalp accélération LONG |
| **LONG** | DI- > DI+ | Pullback | DX↓<ADX | ✅ Scalp fin pullback → LONG |
| **SHORT** | DI- > DI+ | Aligné | DX↑>ADX | ✅ Scalp accélération SHORT |
| **SHORT** | DI+ > DI- | Pullback | DX↓<ADX | ✅ Scalp fin pullback → SHORT |

---

## 📊 RÉCAPITULATIF TIMING ET FENÊTRES

### Contexte (5m) - Fenêtre W_context
**Départ** : PREMIER croisement (VWMA, DI, ou DX/ADX)
**Fin** : Premier croisement + W bougies

**Phase 1 - Croisements (ordre FLEXIBLE)** :
- VWMA + γ_gap
- DI + γ_gap
- DX/ADX + γ_gap
- CHOP pente ≤ 0
- Ordre selon le marché (pas fixe)

**Validation progressive** :
- Chaque condition vérifiée indépendamment
- Validation bougie par bougie
- Signal contexte dès que TOUTES validées
- Position relative maintenue = tendance EN COURS

**Output** : Direction + autorisation scalps (pas de revérification)

---

### Exécution (1m) - Fenêtre W_exec
**PRÉREQUIS** : Contexte EN COURS

**Départ** : Détection momentum DX/ADX
**Fin** : Momentum + W bougies

**Phase 1 - Signal momentum** :
- DX/ADX croisement (↑ ou ↓)
- DX ≥ seuil_DX
- DI dominant maintenu
- Alignement ou pullback identifié

**Phase 2 - Filtres (validation PROGRESSIVE)** :
- Démarrage : Bougie du croisement momentum
- MFI, CHOP, ATR% vérifiés indépendamment
- Validation progressive bougie par bougie
- Signal dès que TOUS validés (dans W_exec)
- Abandon si W_exec atteinte sans validation complète

**Règles clés** :
1. Ordre flexible des filtres
2. Validation indépendante de chaque filtre
3. Signal immédiat si tout OK sur bougie momentum
4. Sinon surveillance continue jusqu'à fin W_exec
5. Deux types : Accélération (DX↑>ADX) ou Pullback (DX↓<ADX)

---

## 📝 Notes importantes

1. **Pas de VWMA en exécution** : Le croisement VWMA est trop lent pour scalper efficacement sur 1m
2. **Deux types de momentum** : Accélération (DX↑>ADX aligné) OU Fin de pullback (DX↓<ADX opposé)
3. **Double CHOP** : Contexte vérifie régime global, exécution vérifie bruit local
4. **MFI modulé** : Les extrêmes MFI sont ajustés dynamiquement selon la force du DI contexte
5. **γ_gap unique** : Même paramètre pour tous les croisements (VWMA, DI, DX/ADX)
6. **τ_slope** : Seuil de variation de pente pour considérer qu'il y a une vraie variation (CHOP, MFI)
7. **seuil_DX** : Valeur minimale de DX au moment du croisement DX/ADX pour valider le momentum (accélération OU ralentissement)
8. **Pullback = opportunité** : Un pullback contre la tendance globale n'est pas un danger, c'est une opportunité d'entrée quand il s'essouffle
9. **Ordre flexible** : Croisements et filtres se valident dans l'ordre du marché, pas d'ordre fixe imposé
10. **Validation progressive** : Chaque élément validé indépendamment, signal généré dès validation complète

---

*Architecture simplifiée : Contexte solide + Exécution réactive = Scalps dans la tendance*
