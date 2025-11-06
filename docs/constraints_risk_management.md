# Contraintes et règles de gestion des risques

**Version:** 0.1  
**Statut:** Règles de risque obligatoires  
**Scope:** Contraintes de l'agent économique et gestion des risques

## Contraintes de risque

### Pertes maximales autorisées
```yaml
risk_constraints:
  max_daily_loss_percent: 5.0       # Max perte journalière en %
  max_monthly_loss_percent: 15.0    # Max perte mensuelle en %
```

### Stop loss obligatoire
```yaml
  mandatory_stop_loss: true         # Toujours avoir un stop loss
  max_stop_loss_percent: 10.0      # Stop loss maximum autorisé
```

### Objectifs de performance
```yaml
  monthly_profit_target_percent: 8.0 # Objectif bénéfice mensuel en %
```

### Actions si contraintes dépassées
```yaml
  # Actions automatiques
  daily_limit_action: "halt_for_day"     # Si perte journalière → arrêt pour la journée
  monthly_limit_action: "halt_daily_retry" # Si perte mensuelle → arrêt pour la journée, essayer le lendemain
```

## Règles de Money Management

### Contraintes de position
- **Taille maximale**: Définie selon volatilité et capital
- **Levier maximum**: Configurable par environnement
- **Exposition maximale**: Pourcentage du capital total

### Règles de diversification
- **Nombre de positions**: Limite simultanées par paire
- **Corrélation**: Évitement positions corrélées
- **Concentration**: Limite par secteur/type d'actif

### Gestion des stops
- **Stop loss**: Obligatoire sur toute position
- **Trailing stop**: Ajustement automatique selon profit
- **Stop d'urgence**: Arrêt immédiat si conditions extrêmes

## Environnements et contraintes

### Backtest
- **Slippage**: Simulation réaliste
- **Commissions**: Intégrées dans calculs
- **Latence**: Prise en compte délais exécution

### Paper trading
- **Contraintes identiques**: Mêmes règles qu'en live
- **Simulation réaliste**: Prix et volumes réels
- **Validation stratégie**: Avant passage en live

### Live trading
- **Contraintes renforcées**: Sécurités additionnelles
- **Monitoring temps réel**: Surveillance continue
- **Circuit breakers**: Arrêt automatique si anomalie
