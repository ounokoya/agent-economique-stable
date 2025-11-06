# Configuration stratégie MACD/CCI/DMI

**Version:** 0.1  
**Statut:** Paramètres de configuration  
**Scope:** Paramètres techniques de la stratégie de trading

## Paramètres des indicateurs

### MACD
```yaml
indicators:
  macd:
    fast_period: 12
    slow_period: 26
    signal_period: 9
    timeframes: ["5m", "15m", "1h", "4h"]
```

### CCI
```yaml
  cci:
    period: 14
    timeframes: ["5m", "15m", "1h", "4h"]
    
    # Seuils LONG tendance
    long_trend_oversold: -100              # LONG tendance - survente
    long_trend_overbought: 100             # LONG tendance - surachat
    
    # Seuils LONG contre-tendance  
    long_counter_trend_oversold: -150      # LONG contre-tendance - survente
    long_counter_trend_overbought: 150     # LONG contre-tendance - surachat
    
    # Seuils SHORT tendance
    short_trend_oversold: -120             # SHORT tendance - survente
    short_trend_overbought: 120            # SHORT tendance - surachat
    
    # Seuils SHORT contre-tendance
    short_counter_trend_oversold: -180     # SHORT contre-tendance - survente
    short_counter_trend_overbought: 180    # SHORT contre-tendance - surachat
```

### DMI
```yaml
  dmi:
    period: 14
    adx_period: 14
    timeframes: ["5m", "15m", "1h", "4h"]
```

## Filtres

```yaml
filters:
  macd_same_sign_filter: false
  dmi_trend_signals_enabled: true
  dmi_counter_trend_signals_enabled: false
  dx_adx_filter_enabled: false
```

## Gestion de position

```yaml
position_management:
  # Trailing stops initiaux selon type de signal DMI
  trend_trailing_stop_percent: 2.0        # Pour signaux en tendance DMI
  counter_trend_trailing_stop_percent: 1.5 # Pour signaux contre-tendance DMI (plus serré)
  
  # Grille d'ajustement trailing stop (valeurs décroissantes)
  trailing_stop_adjustment_grid:
    - profit_range: [0, 5]
      trailing_stop_percent: 2.0    # Stop initial maintenu
    - profit_range: [5, 10]
      trailing_stop_percent: 1.5    # Stop plus serré
    - profit_range: [10, 20]
      trailing_stop_percent: 1.0    # Stop encore plus serré
    - profit_range: [20, 100]
      trailing_stop_percent: 0.5    # Stop très serré
```
