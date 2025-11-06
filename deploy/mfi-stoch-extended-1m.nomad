job "mfi-stoch-extended-1m" {
  datacenters = ["dc1"]
  type = "service"

  group "mfi-stoch-1m" {
    count = 1

    task "setup-dirs" {
      driver = "raw_exec"
      lifecycle {
        hook = "prestart"
      }
      
      config {
        command = "/bin/mkdir"
        args    = ["-p", "/root/data/backtest-optimizer/out", "/root/data/backtest-optimizer/state"]
      }
      
      resources {
        cpu    = 50
        memory = 50
      }
    }

    task "mfi-stoch-1m" {
      driver = "raw_exec"

      config {
        command = "/root/data/backtest-optimizer/mfi_stoch_live_notifications"
        args    = ["-config", "local/mfi_stoch_config.json"]
      }

      env {
        LOG_LEVEL = "info"
      }

      template {
        data = <<EOH
{
  "symbol": "SOLUSDT",
  "exec_tf": "1m",
  "ntfy_config": "local/ntfy_config.json",
  "state_file": "/root/data/backtest-optimizer/state/mfi_stoch_1m_extended.json",
  "database": {
    "url": "http://10.0.0.1:8529",
    "name": "agent_economique",
    "username": "root",
    "password": "agent_economique_2025",
    "collection_prefix": "mfi_stoch_1m"
  },
  "money_management": {
    "initial_trailing_stop_pct": 0.25,
    "milestones": [
      {"progress": 0.25, "lock": 0.2},
      {"progress": 0.50, "lock": 0.40},
      {"progress": 0.75, "lock": 0.60},
      {"progress": 1.00, "lock": 0.9}
    ],
    "beyond_trailing_pct": 0.9
  },
  "params": {
    "mfi_period": 16,
    "mfi_extreme_high": 70.0,
    "mfi_extreme_low": 30.0,
    "mfi_extended_high": 60.0,
    "mfi_extended_low": 40.0,
    "stoch_k_period": 16,
    "stoch_smooth_k": 3,
    "stoch_d_period": 3,
    "stoch_extreme_high": 80.0,
    "stoch_extreme_low": 20.0,
    "stoch_extended_high": 70.0,
    "stoch_extended_low": 30.0,
    "max_reentries": 3,
    "use_extended_mm": true,
    "leverage": 5.0,
    "trailing_stop_pct": 0.25,
    "take_profit_target_pct": 0.0025,
    "window_size": 500
  }
}
EOH
        destination = "local/mfi_stoch_config.json"
        change_mode = "restart"
      }

      template {
        data = <<EOH
{
  "server": "notifications.koyad.com",
  "topic": "mfi-stoch-1m",
  "title": "MFI+Stoch 1m Extended",
  "ping_message": "🚀 MFI+Stoch 1m Extended Started (SOLUSDT)",
  "signal_long": "🟢 LONG MFI+Stoch 1m",
  "signal_short": "🔴 SHORT MFI+Stoch 1m"
}
EOH
        destination = "local/ntfy_config.json"
        change_mode = "restart"
      }

      resources {
        cpu    = 512
        memory = 512
      }

      restart {
        attempts = 5
        interval = "30m"
        delay    = "15s"
        mode     = "fail"
      }
    }
  }
}
