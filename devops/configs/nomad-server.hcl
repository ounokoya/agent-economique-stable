# Nomad Server + Client Configuration
# Server: 31.57.224.79 (Singapore)
# Datacenter: sg1

datacenter = "dc1"
data_dir   = "/opt/nomad/data"
bind_addr  = "0.0.0.0"

# Server mode
server {
  enabled          = true
  bootstrap_expect = 1
  
  server_join {
    retry_join = ["127.0.0.1"]
  }
}

# Client mode (jobs s'exécutent sur même serveur)
client {
  enabled = true
  
  # Network interface (WireGuard)
  network_interface = "wg0"
  
  # Options
  options = {
    "driver.exec.enable"     = "1"
    "driver.raw_exec.enable" = "1"
    "docker.cleanup.image"   = "true"
  }
  
  # Host volumes (path générique /root/data)
  host_volume "app-data" {
    path      = "/root/data"
    read_only = false
  }
}

# UI
ui {
  enabled = true
  
  # Accessible via VPN sur port 4646
  # http://10.8.0.1:4646
}

# Telemetry
telemetry {
  publish_allocation_metrics = true
  publish_node_metrics       = true
  prometheus_metrics         = true
}

# TLS Configuration
tls {
  http = true
  rpc  = true

  ca_file   = "/etc/nomad.d/certs/ca.pem"
  cert_file = "/etc/nomad.d/certs/server.pem"
  key_file  = "/etc/nomad.d/certs/server-key.pem"

  verify_server_hostname = true
  verify_https_client    = true
}

# ACL (désactivé pour l'instant)
acl {
  enabled = false
}

# Consul (optionnel, désactivé)
consul {
  address = ""
}

# Vault (optionnel, désactivé)
vault {
  enabled = false
}
