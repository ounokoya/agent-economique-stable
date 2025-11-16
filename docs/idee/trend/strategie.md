Voici le **résumé validé et maintenu** — net, complet, sans rien d’ajout.
================
# Cadre général

* **Signal (exécution)** : **croisement VWMA6↔VWMA20** **ou** **VWMA6↔VWMA30**.
* **Gate VWMA96** : le croisement n’est **valide** que du **bon côté** de **VWMA96**
  — Long : VWMA6 **et** VWMA20/30 **au-dessus** de VWMA96 ; Short : miroir **au-dessous**.
* **Validation DMI (obligatoire, exécution)** :

  1. **Tendance** : croisement **DI** dans le sens **avec** écart **|DI+−DI−| ≥ δ_min** ; **ou**
  2. **Momentum** : **DX/ADX** qui **croisent au-dessus** d’un/des **DI** (dépassement ≥ **τ_gap**).
     – **Motif KO (désordre DMI)** : **DX/ADX** restent **sous** les deux **DI** (de bout en bout) ⇒ **refus**.

# Filtres d’entrée (après DMI validé)

* **MFI** : **en zone extrême favorable** au trade **ET**, dans cette zone, **pente ≥ τ_slope** *ou* **constance** ; **MFI ne déclenche jamais la sortie**.
  – **Contexte DMI** module dynamiquement les extrêmes MFI : DI dominant fort ⇒ extrême **plus strict** dans le sens / **plus souple** à contre-tendance.
* **CCI** : **non-extrême** (anti-entrée tardive).
* **CHOP** : **pente ≤ 0** (constant/décroissant) ; hausse au-delà de **τ_slope** ⇒ **veto**.
* **ATR%** : **≥ seuil_min** ; sert aussi à dimensionner le stop.

# Gestion du stop (ancré VWMA)

* **Stop standard** : **suiveur de VWMA30** (pas un trailing prix).
  – Distance = **p%**, **dynamique via ATR%** : `p% = clip(k × ATR%, p_min, p_max)`.
* **Début de désordre (bascule de stop)** → **VWMA20** :

  1. **CHOP** se redresse **> τ_slope** **ET** **MFI** bascule en **pente défavorable ≥ τ_slope**, **ou**
  2. **MFI** passe en **extrême inverse** (pente indifférente).
     ⇒ Stop devient **suiveur VWMA20** ; **ATR% recalculé** avec VWMA20 ; **sortie immédiate** si le nouveau stop est déjà dépassé.

# Suivi (contexte)

* **VWMA6 (pente)** : suit l’évolution attendue du trade (confirmation, pas contrainte).
* **CHOP** : **constant/décroissant** (petites hausses tolérées < **τ_slope**).
* **ATR%** : ≥ seuil_min (régime maintenu).
* **MFI** : reste **soutenant** (extrême favorable stable, ou pente/constance favorables).
* **Aucun** de ces éléments n’initie la sortie (hors désordre) : **sortie = croisement inverse VWMA6↔VWMA20/30** **ou** **stop**.

# Tolérances universelles

* **τ_slope** : tolérance de **pente** (CHOP, MFI, VWMA6/20/30, ADX, DX).
* **τ_gap** : écart **minimal** pour **tout croisement** (VWMA, DI, DX/ADX>DI).
* **δ_min** : écart **DI** minimal à l’entrée (validation Tendance).

# Deux bots, même logique

* **Scalping** : seuils **plus stricts**, fenêtres d’acceptation **serrées**.
* **Investissement** : seuils **plus souples**, respirations **tolérées**.

Que le **Seigneur Père Céleste**, le **Seigneur Jésus** et le **Seigneur Saint Esprit** nous gardent clairs et disciplinés sur cette base.

===========================

Voici les **configs finales** des **2 bots** avec la stratégie complète, paramètres proposés par TF, et justifications — sans rien changer à ta logique.

# Bot SCALPING — Contexte 5m / Exécution 1m

## Flux décisionnel (inchangé)

1. **Signal (exéc.)** : **croisement VWMA6↔VWMA20** *ou* **VWMA6↔VWMA30**.
2. **Gate VWMA96** : long **seulement si** VWMA6 **et** VWMA20/30 **au-dessus** de VWMA96 (short miroir).
3. **Validation DMI (exéc.)** :
   • **Tendance** : croisement **DI** dans le sens **avec** écart **|DI+−DI−| ≥ δ_min** ; **ou**
   • **Momentum** : **DX/ADX** qui **passe(nt) au-dessus** d’un/des **DI** (**dépassement ≥ τ_gap**).
   • **Motif KO** : **DX/ADX** restent **sous** les 2 **DI** (début→fin) ⇒ **refus**.
4. **Filtres d’entrée (exéc.)** :
   • **MFI** **en zone extrême favorable** **ET** **pente ≥ τ_slope** *ou* **plateau** (constance) ;
   • **CCI** **non extrême** (anti-tardif) ;
   • **CHOP** **pente ≤ 0** (hausse > **τ_slope** ⇒ veto) ;
   • **ATR% ≥ seuil_min**.
5. **Stop & désordre** :
   • **Stop suiveur VWMA30** avec **p% = f(ATR%)** ;
   • **Début de désordre** ⇒ **bascule** sur **VWMA20** si :
   (A) **CHOP** se redresse **> τ_slope** **ET** **MFI** pente **défavorable ≥ τ_slope**, **ou**
   (B) **MFI** passe en **extrême inverse** (pente indifférente).
   • Après bascule : **ATR% recalculé sur VWMA20** ; **sortie immédiate** si le nouveau stop est déjà dépassé.
6. **Sortie** : **croisement inverse VWMA6↔VWMA20/30** **ou** **stop**.

## Paramètres proposés (scalp, réactifs)

* **VWMA** : **6 / 20** (par défaut, plus vif) *ou* 6/30 ; **gate 96**.
  • **τ_gap (VWMA)** : **≈ 0,15 × ATR(1m)** (écart mini pour valider un croisement).
* **DMI (exéc.)** : **DMI(14,3)** ; **δ_min (DI)** : **5–8**.
  • Momentum : **DX/ADX > DI** de **≥ τ_gap**.
* **MFI (exéc.)** : **len 14** ; **extrêmes de base 80/20**, **ajustés** par **DI contexte** (DI dominant fort ⇒ plus **strict** dans le sens, plus **souple** à contre-tendance).
  • **Pente** (Δ3 bougies) **≥ +5** *ou* **plateau** (|Δ| ≤ 2).
* **CCI (exéc.)** : **len 14** ; **non-extrême** = **|CCI| < 150**.
* **CHOP** : **len 14** ; pente mesurée sur **3 bougies** ; **hausse tolérée ≤ +5** (sinon veto).
* **ATR** : **len 24** ; **ATR%_min** = **0,15–0,30 %**.
* **Stop p%(ATR%)** : **k = 1,0–1,5** ; **p_min = 0,20 %** ; **p_max = 1,20 %**.

## Justifs TF & réglages

* **5m** : contexte assez stable sans inertie lourde ;
* **1m** : exécution fine, bruit maîtrisé par **gate 96** + **DMI** + **MFI/CHOP** ;
* Paramètres **courts** pour capter **coups rapides** et limiter la donation via le **stop ancré VWMA**.

---

# Bot INVESTISSEMENT — Contexte 4h / Exécution 1h

## Flux décisionnel (identique)

1. **Signal (exéc.)** : **croisement VWMA6↔VWMA20** *ou* **VWMA6↔VWMA30**.
2. **Gate VWMA96** : long **au-dessus**, short **au-dessous**.
3. **Validation DMI (exéc.)** : **Tendance** (croisement DI + **δ_min**) **ou** **Momentum** (DX/ADX > DI avec **τ_gap**). **Motif KO** idem.
4. **Filtres (exéc.)** : **MFI** extrême favorable + **pente/plateau**, **CCI** non-extrême, **CHOP** pente ≤ 0 (hausse > **τ_slope** ⇒ veto), **ATR% ≥ seuil_min**.
5. **Stop & désordre** : **suiveur VWMA30** → **bascule VWMA20** selon (A) ou (B) ci-dessus ; **ATR%** recalculé sur l’ancre courante ; **sortie immédiate** si dépassé.
6. **Sortie** : **croisement inverse** ou **stop**.

## Paramètres proposés (invest, plus lisses)

* **VWMA** : **6 / 30** (par défaut, tenue) *ou* 6/20 ; **gate 96**.
  • **τ_gap (VWMA)** : **≈ 0,10 × ATR(1h)**.
* **DMI (exéc.)** : **DMI(24,6)** ; **δ_min (DI)** : **8–12**.
  • Momentum : **DX/ADX > DI** de **≥ τ_gap**.
* **MFI (exéc.)** : **len 30** ; extrêmes **80/20** **ajustés** par **DI contexte**.
  • Pente (Δ3 bougies) **≥ +4** *ou* **plateau** (|Δ| ≤ 2).
* **CCI (exéc.)** : **len 20** ; **non-extrême** = **|CCI| < 200**.
* **CHOP** : **len 14–21** ; pente sur **3 bougies** ; **hausse tolérée ≤ +3** (sinon veto).
* **ATR** : **len 14–24** ; **ATR%_min** = **0,50–1,20 %**.
* **Stop p%(ATR%)** : **k = 1,8–2,5** ; **p_min = 0,50 %** ; **p_max = 3,00 %**.

## Justifs TF & réglages

* **4h** : lit bien les **régimes** et **tendances** ;
* **1h** : exécutions propres avec **moins de bruit** ;
* Paramètres **plus lents** pour **tenir** les mouvements, **bascule VWMA20** pour sécuriser en **début de désordre**.

---

## Tolérances universelles (communes, appliquées partout)

* **τ_slope** : seuil de **pente** (CHOP, MFI, VWMA6/20/30, ADX, DX).
* **τ_gap** (ou **γ_gap**) : **écart minimal** pour **tout croisement** (VWMA, DI, DX/ADX>DI).
* **δ_min** : **écart DI** minimal à l’entrée (validation Tendance).

> Ces trois tolérances évitent les **faux signaux** (micro-variations / “touch-and-go”).

### **Innovation : Validation Gamma Différée avec Fenêtre**

Un croisement peut commencer **faiblement** et se **renforcer progressivement**. Au lieu de rejeter immédiatement les croisements dont le gap initial est < **γ_gap**, on leur accorde une **fenêtre de validation différée** :

**Logique** :
```
AU CROISEMENT (barre n-1) :
├─ SI gap >= γ_gap × ATR 
│  └─ ✓ Validé IMMÉDIATEMENT (GapValideBougie = 0)
│
└─ SINON : Vérifier sur WINDOW_GAMMA_VALIDATE bougies (ex: 5)
* **DMI (contexte)** : sert **uniquement** à **ajuster** les seuils d’extrême **MFI** (plus strict dans le sens DI dominant, plus souple en contre-tendance).
* **CHOP** : si la **pente** ↑ **au-delà de τ_slope** ⇒ **veto** (on suspend la recherche).
* **ATR%** : doit rester **≥ seuil_min** (régime suffisant).
  *(Pas de gate ici.)*

## Exécution (TF basse)

### Ouverture (3 étapes dans la fenêtre W)

1. **Signal** : **croisement VWMA6↔VWMA20** **ou** **VWMA6↔VWMA30** **valide seulement si**
   |VWMA6 − VWMA20/30| ≥ **γ_gap** **et** le croisement se fait **du bon côté de VWMA96** (long au-dessus / short au-dessous).
2. **DI** : **croisement DI** dans le sens, **valide si** |DI+ − DI−| ≥ **γ_gap** **au croisement**.
3. **DX/ADX** : **DX** croise **au-dessus** d’**ADX** dans **W**, **valide si** l’overshoot ≥ **γ_gap** **et** le **DI directionnel** reste **dominant** (|DI+−DI−| ≥ **γ_gap**) dans **W**.
   ➡ Si (1)+(2)+(3) vrais **dans W** ⇒ **ouverture validée**.

### Filtres après ouverture validée

* **MFI** : **extrême favorable obligatoire** + **pente favorable** *ou* **plateau** (MFI ne fait **jamais** sortir).
* **CCI** : **non-extrême** (anti-tardif).
* **CHOP (exéc.)** : **pente ≤ 0** ; si **> τ_slope** ⇒ **veto**.
* **ATR%** : **≥ seuil_min** (sert aussi au sizing du stop).

### Stops & désordre

* **Stop standard** : **suiveur VWMA30** à **p%(ATR%)** (clip).
* **Début de désordre** ⇒ **bascule du stop vers VWMA20** si :
  (A) **CHOP** ↑ **> τ_slope** **ET** **MFI** devient **défavorable** (pente ≥ τ_slope), **ou**
  (B) **MFI** passe en **extrême inverse** (pente indifférente).
  → **ATR%** recalculé sur **VWMA20** ; **sortie immédiate** si le nouveau stop est déjà dépassé.
* **Sortie normale** : **croisement inverse VWMA6↔VWMA20/30** **ou** **stop**.

---

# Bot INVESTISSEMENT

## Contexte (TF haute)

* **DMI (contexte)** : **uniquement** pour **moduler** les extrêmes **MFI** (plus strict pro-tendance, plus souple contre).
* **CHOP** : **pente** ↑ **> τ_slope** ⇒ **veto**.
* **ATR%** : **≥ seuil_min**.
  *(Pas de gate ici.)*

## Exécution (TF basse)

### Ouverture (3 étapes dans la fenêtre W)

1. **Signal** : **VWMA6↔VWMA20** **ou** **VWMA6↔VWMA30** **valide si** |VWMA6 − VWMA20/30| ≥ **γ_gap** **et** croisement **du bon côté de VWMA96**.
2. **DI** : **croisement DI** dans le sens, **valide si** |DI+ − DI−| ≥ **γ_gap** **au croisement**.
3. **DX/ADX** : **DX** croise **au-dessus** d’**ADX** dans **W**, **valide si** overshoot ≥ **γ_gap** **et** **dominance DI** maintenue (|DI+−DI−| ≥ **γ_gap**) dans **W**.
   ➡ Si (1)+(2)+(3) vrais **dans W** ⇒ **ouverture validée**.

### Filtres / Stops

* **MFI** : extrême favorable + pente/plateau ; seuils **ajustés** par **DMI contexte**.
* **CCI** non-extrême ; **CHOP** pente ≤ 0 (sinon veto) ; **ATR% ≥ seuil_min**.
* **Stop suiveur** : **VWMA30** → **bascule VWMA20** en désordre (mêmes règles que scalping) ; **ATR%** recalculé ; sortie immédiate si déjà dépassé.
* **Sortie** : **croisement inverse** ou **stop**.

---

## Tolérances communes

* **γ_gap** : **seul écart** pour **tous les croisements** (VWMA, DI, DX/ADX).
* **τ_slope** : tolérance **de pente** (CHOP, MFI, VWMA, DX/ADX) — **non utilisée** pendant l’**ouverture**.
* **Fenêtre W** : intervalle borné où doivent se produire **Signal → DI → DX>ADX** (pas “même bougie”).

Si autre chose a été déformé, dis-moi **la ligne exacte** à corriger et je rectifie aussitôt, par la grâce du **Seigneur Père Céleste**, du **Seigneur Jésus** et du **Seigneur Saint Esprit**.
==============