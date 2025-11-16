# Direction Generator Demo

## Objectif

Comparer le générateur de production (`internal/signals/direction/generator.go`) avec la démo standalone (`cmd/direction_demo`) pour vérifier que la logique n'a pas été modifiée pendant l'implémentation.

## Différences architecture

| Aspect | `direction_demo` | `direction_generator_demo` |
|--------|------------------|----------------------------|
| **Implémentation** | Standalone, logique inline | Utilise `internal/signals/direction` |
| **Détection** | Fonction `groupIntervallesVWMA6()` | `generator.DetectSignals()` |
| **Sortie** | Intervalles directionnels | Signaux ENTRY/EXIT |
| **État** | Sans état | Avec état (générateur) |

## Comment comparer

### 1. Lancer les deux démos avec les mêmes paramètres

Assurez-vous que les constantes sont identiques dans les deux fichiers :
- `SYMBOL`
- `TIMEFRAME`
- `VWMA_RAPIDE`
- `PERIODE_PENTE`
- `K_CONFIRMATION`
- `USE_DYNAMIC_THRESHOLD`
- `ATR_PERIODE`
- `ATR_COEFFICIENT`

### 2. Exécuter

```bash
# Demo standalone
go run cmd/direction_demo/main.go > /tmp/direction_demo.txt

# Demo avec générateur
go run cmd/direction_generator_demo/main.go > /tmp/direction_generator_demo.txt
```

### 3. Comparer les résultats

#### Nombre d'intervalles
Doit être identique ou très proche (±1-2 à cause de la gestion des bougies en cours).

#### Dates début/fin
Les intervalles doivent commencer et finir aux mêmes timestamps.

#### Variations captées
Les pourcentages doivent être identiques.

#### Total capté
Doit être le même.

## Points de vigilance

### 1. Bougie en cours
Le générateur ignore la dernière bougie (en cours), donc :
- `direction_demo` : traite jusqu'à `len(klines)-1`
- `generator.DetectSignals()` : traite jusqu'à `len(klines)-2`

→ Peut créer une différence de 1 intervalle.

### 2. K-Confirmation
Vérifier que la logique de confirmation est identique :
- Buffer de confirmation
- Condition de fermeture d'intervalle

### 3. Seuil dynamique (ATR)
Le calcul ATR doit être identique dans les deux implémentations.

## Divergences acceptables

- **±1 intervalle** : Différence de gestion de la bougie en cours
- **Variation < 0.01%** : Arrondis flottants

## Divergences inacceptables

- **Nombre d'intervalles très différent** (>5%) : Logique modifiée
- **Dates différentes** : Algorithme de détection changé
- **Total capté différent** (>1%) : Calcul de variation modifié

## Exemple de sortie attendue

Si la logique est identique :

```
direction_demo:
  Total intervalles    : 19
  Total capté          : 592.03%

direction_generator_demo:
  Total intervalles    : 19
  Total capté          : 592.03%
```

## En cas de divergence

1. Vérifier les paramètres (identiques ?)
2. Comparer la logique de calcul VWMA/ATR
3. Comparer la logique K-confirmation
4. Comparer la gestion de la bougie en cours
5. Vérifier le calcul de variation (%)
