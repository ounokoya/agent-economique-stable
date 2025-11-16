# Configurations détaillées des bots - Paramètres chiffrés

## Bot SCALPING — Contexte 5m / Exécution 1m

### Paramètres opérationnels

* **VWMA** : **6 / 20** (par défaut, plus vif) *ou* 6/30.
  • **γ_gap (VWMA)** : **≈ 0,15 × ATR(1m)** (écart mini pour valider un croisement).
* **DMI (exéc.)** : **DMI(14,3)** ; **δ_min (DI)** : **5–8**.
  • Momentum : **DX/ADX > DI** de **≥ γ_gap**.
* **MFI (exéc.)** : **len 14** ; **extrêmes de base 80/20**, **ajustés** par **DI contexte** (DI dominant fort ⇒ plus **strict** dans le sens, plus **souple** à contre-tendance).
  • **Pente** (Δ3 bougies) **≥ +5** *ou* **plateau** (|Δ| ≤ 2).
* **CHOP** : **len 14** ; pente mesurée sur **3 bougies** ; **hausse tolérée ≤ +5** (sinon veto).
* **ATR** : **len 24** ; **ATR%_min** = **0,15–0,30 %**.
* **Stop p%(ATR%)** : **k = 1,0–1,5** ; **p_min = 0,20 %** ; **p_max = 1,20 %**.

### Justifs TF & réglages

* **5m** : contexte assez stable sans inertie lourde ;
* **1m** : exécution fine, bruit maîtrisé par **DMI** + **MFI/CHOP** ;
* Paramètres **courts** pour capter **coups rapides** et limiter la donation via le **stop ancré VWMA**.

---

## Bot INVESTISSEMENT — Contexte 4h / Exécution 1h

### Paramètres opérationnels

* **VWMA** : **6 / 30** (par défaut, tenue) *ou* 6/20.
  • **γ_gap (VWMA)** : **≈ 0,10 × ATR(1h)**.
* **DMI (exéc.)** : **DMI(24,6)** ; **δ_min (DI)** : **8–12**.
  • Momentum : **DX/ADX > DI** de **≥ γ_gap**.
* **MFI (exéc.)** : **len 30** ; extrêmes **80/20** **ajustés** par **DI contexte**.
  • Pente (Δ3 bougies) **≥ +4** *ou* **plateau** (|Δ| ≤ 2).
* **CHOP** : **len 14–21** ; pente sur **3 bougies** ; **hausse tolérée ≤ +3** (sinon veto).
* **ATR** : **len 14–24** ; **ATR%_min** = **0,50–1,20 %**.
* **Stop p%(ATR%)** : **k = 1,8–2,5** ; **p_min = 0,50 %** ; **p_max = 3,00 %**.

### Justifs TF & réglages

* **4h** : lit bien les **régimes** et **tendances** ;
* **1h** : exécutions propres avec **moins de bruit** ;
* Paramètres **plus lents** pour **tenir** les mouvements, **bascule VWMA20** pour sécuriser en **début de désordre**.

---

## Tolérances universelles (communes, appliquées partout)

* **γ_gap** : **seul écart** pour **tous les croisements** (VWMA, DI, DX/ADX).
* **τ_slope** : tolérance **de pente** (CHOP, MFI, VWMA, DX/ADX) — **non utilisée** pendant l’**ouverture**.
* **Fenêtre W** : intervalle borné où doivent se produire **Signal → DI → DX>ADX** (pas "même bougie").

> Ces trois tolérances évitent les **faux signaux** (micro-variations / "touch-and-go").

---

*Note : Ces configurations complètent le framework stratégique décrit dans `strategie.md`*
