job "agent-economique-sol-1h" {
  datacenters = ["dc1"]
  type = "service"

  group "sol-1h" {
    count = 1

    # Tâche pour créer les dossiers nécessaires
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

    task "agent-sol-1h" {
      driver = "raw_exec"

      config {
        command = "/root/data/backtest-optimizer/agent_economique_live_notifications"
        args    = ["-config", "local/agent_economique_live.json"]
      }

      env {
        LOG_LEVEL = "info"
      }

      # Template pour la configuration principale
      template {
        data = <<EOH
{
  "symbol": "SOLUSDT",
  "exec_tf": "1h",
  "ntfy_config": "local/ntfy_config.json",
  "state_file": "/root/data/backtest-optimizer/state/agent_eco_sol_1h.json",
  "database": {
    "url": "http://10.0.0.1:8529",
    "name": "agent_economique",
    "username": "root",
    "password": "agent_economique_2025",
    "collection_prefix": "notification"
  },
  "params": {
    "cci_period": 96,
    "cci_extreme_high": 100.0,
    "cci_extreme_low": -100.0,
    "mfi_period": 16,
    "mfi_exit_high": 70.0,
    "mfi_exit_low": 30.0,
    "stoch_k_period": 16,
    "stoch_smooth_k": 3,
    "stoch_d_period": 3,
    "stoch_oversold": 30.0,
    "stoch_overbought": 70.0,
    "leverage": 5.0,
    "trailing_stop_pct": 2.5,
    "take_profit_target_pct": 0.025,
    "window_size": 300
  }
}
EOH
        destination = "local/agent_economique_live.json"
        change_mode = "restart"
      }

      # Template pour la configuration ntfy
      template {
        data = <<EOH
{
  "server": "notifications.koyad.com",
  "topic": "notification-agent-eco",
  "title": "Agent Eco SOL 1h",
  "ping_message": "🚀 SOL 1h Started",
  "signal_long": "🟢 LONG SOL 1h",
  "signal_short": "🔴 SHORT SOL 1h"
}
EOH
        destination = "local/ntfy_config.json"
        change_mode = "restart"
      }

      resources {
        cpu    = 256
        memory = 256
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
