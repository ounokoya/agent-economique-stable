job "scalping-live" {
  datacenters = ["dc1"]
  type = "service"

  group "scalping-live-group" {
    count = 1

    # Tâche pour créer les dossiers nécessaires
    task "setup-dirs" {
      driver = "raw_exec"
      lifecycle {
        hook = "prestart"
      }
      
      config {
        command = "/bin/mkdir"
        args    = ["-p", "/root/data/scalping-live/config", "/root/data/scalping-live/logs", "/root/data/scalping-live/state"]
      }
      
      resources {
        cpu    = 50
        memory = 50
      }
    }

    task "scalping-live-5m" {
      driver = "raw_exec"

      config {
        command = "/root/data/scalping-live/scalping_live"
        args    = ["-config", "local/config.yaml"]
      }

      env {
        # 📝 LOGGING
        LOG_LEVEL = "info"
      }

      # Template pour la configuration principale (YAML)
      template {
        data = <<EOH
# Configuration Scalping Live - SOLUSDT 5m
# Version: 1.0.0

# Configuration données Binance
binance_data:
  cache_root: "data/binance"
  symbols: 
    - "SOLUSDT"
  timeframes:
    - "5m"
  data_types:
    - "klines"
  
  # Configuration cache
  cache:
    checksum_validation: true
    
  # Configuration téléchargeur
  downloader:
    base_url: "https://data.binance.vision"
    max_retries: 3
    retry_delay: "5s"
    timeout: "10m"
    max_concurrent: 5
    checksum_verify: true
    
  # Configuration streaming
  streaming:
    buffer_size: 65536
    max_memory_mb: 100
    enable_metrics: true
    
  # Configuration validation
  validation:
    max_price_deviation: 50.0
    max_volume_deviation: 1000.0
    require_monotonic_time: false
    max_time_gap: 900000

# Configuration stratégie
strategy:
  name: "SCALPING"
  
  # Configuration Stratégie Scalping
  scalping:
    # Seuils extrêmes
    cci_surachat: 100.0
    cci_survente: -100.0
    mfi_surachat: 60.0
    mfi_survente: 40.0
    stoch_surachat: 70.0
    stoch_survente: 30.0
    
    # Synchronisation
    sync_mfi_enabled: true
    sync_cci_enabled: true
    
    # Validation
    validation_window: 3
    volume_threshold: 0.25
    volume_period: 5
    volume_max_ext: 100
    
    # Timeframe
    timeframe: "5m"
    
  # Paramètres indicateurs
  indicators:
    macd:
      fast_period: 12
      slow_period: 26  
      signal_period: 9
    stochastic:
      period_k: 14
      smooth_k: 3
      period_d: 3
      oversold: 20
      overbought: 80
    mfi:
      period: 14
      oversold: 20
      overbought: 80
    cci:
      period: 14
      threshold_oversold: -100
      threshold_overbought: 100
    dmi:
      period: 14
      
  # Configuration signaux
  signal_generation:
    min_confidence: 0.7
    premium_confidence: 0.9
    require_bar_confirmation: true
    enable_multi_tf: false
    
  # Configuration gestion position
  position_management:
    # Trailing Stop Base
    base_trailing_percent: 2.0
    trend_trailing_percent: 2.5
    counter_trend_trailing: 1.5
    
    # Ajustements Dynamiques
    enable_dynamic_adjustments: true
    stoch_inverse_adjust: 0.2
    mfi_inverse_adjust: 0.3
    cci_inverse_adjust: 0.4
    triple_inverse_adjust: 0.9
    
    # Limites de Sécurité
    max_cumulative_adjust: 1.0
    min_trailing_percent: 0.3
    max_trailing_percent: 5.0
    
    # Early Exit
    enable_early_exit: true
    min_profit_for_early_exit: 0.5
    early_exit_trailing: 0.5
    stall_bars_threshold: 5
    
  # Logique risk/money management
  risk_management:
    max_position_size_usd: 1000.0
    max_daily_loss_usd: 500.0
    max_open_positions: 1
    
# Configuration Money Management
money_management:
  initial_capital: 10000.0
  risk_per_trade_pct: 2.0
  max_positions: 1
  
# Configuration Backtesting
backtest:
  start_date: "2024-01-01"
  end_date: "2024-12-31"
  commission_pct: 0.1
  slippage_pct: 0.05
EOH
        destination = "local/config.yaml"
        change_mode = "restart"
      }

      resources {
        cpu    = 512
        memory = 1024
      }

      # 🔄 RESTART POLICY pour haute disponibilité
      restart {
        attempts = 10
        interval = "1h"
        delay    = "30s"
        mode     = "fail"
      }
    }
  }
}
