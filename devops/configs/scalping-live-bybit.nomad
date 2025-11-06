job "scalping-live-bybit" {
  datacenters = ["dc1"]
  type = "service"

  group "scalping-live-bybit-group" {
    count = 1

    # Tâche pour créer les dossiers nécessaires
    task "setup-dirs" {
      driver = "raw_exec"
      lifecycle {
        hook = "prestart"
      }
      
      config {
        command = "/bin/mkdir"
        args    = ["-p", "/root/data/scalping-live-bybit/config", "/root/data/scalping-live-bybit/logs", "/root/data/scalping-live-bybit/state"]
      }
      
      resources {
        cpu    = 50
        memory = 50
      }
    }

    # Use host volume
    volume "app-data" {
      type   = "host"
      source = "app-data"
    }

    task "scalping-live-bybit-5m" {
      driver = "exec"

      # Mount volume
      volume_mount {
        volume      = "app-data"
        destination = "/data"
      }

      config {
        command = "/data/scalping-live-bybit/scalping_live_bybit"
        args    = ["-config", "local/config.yaml"]
      }

      env {
        # 📝 LOGGING
        LOG_LEVEL = "info"
      }

      # Template pour la configuration principale (YAML)
      template {
        data = <<EOH
# Configuration Scalping Live Bybit - SOLUSDT 5m
# Version: 1.0.0
# Exchange: Bybit (Singapore server - No geo restrictions)

# Configuration données Bybit
binance_data:
  cache_root: "data/bybit"
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
    base_url: "https://api.bybit.com"
    max_retries: 3
    retry_delay: "5s"
    timeout: "10m"
    max_concurrent: 5
    checksum_verify: false
    
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
    validation_window: 6
    volume_threshold: 0.25
    volume_period: 3
    volume_max_ext: 4
    
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
      oversold: 30
      overbought: 70
    mfi:
      period: 14
      oversold: 40
      overbought: 60
    cci:
      period: 20
      threshold_oversold: -100
      threshold_overbought: 100

# Configuration notifications
notifications:
  ntfy:
    enabled: true
    server_url: "https://notifications.koyad.com"
    topic: "scalping-live-bybit"
    priority: 4
    
# Configuration Exchange Bybit
bybit:
  endpoint: "https://api.bybit.com"
  api_key: ""        # Not needed for market data
  api_secret: ""     # Not needed for market data
  category: "linear" # USDT Perpetual Futures
EOH
        destination = "local/config.yaml"
      }

      # Resources
      resources {
        cpu    = 500  # 0.5 CPU
        memory = 512  # 512 MB
      }

      # Restart policy
      restart {
        attempts = 3
        interval = "5m"
        delay    = "30s"
        mode     = "fail"
      }

      # Logs rotation
      logs {
        max_files     = 10
        max_file_size = 10  # MB
      }
    }
  }
}
