# 🚀 Installation Nomad Server (Singapour)

**Serveur:** 31.57.224.79 (Singapour)  
**Rôle:** Nomad Server + Client  
**OS:** Ubuntu 24.04 LTS

---

## 📋 Prérequis

- ✅ SSH root configuré
- ✅ Ubuntu 24.04 LTS
- ✅ Ports ouverts : 4646 (HTTP), 4647 (RPC), 4648 (Serf)

---

## 🔧 Installation Nomad

### **1. Installer HashiCorp Repository**

```bash
# Update system
apt update && apt upgrade -y

# Install dependencies
apt install -y wget curl unzip gpg

# Add HashiCorp GPG key
wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg

# Add HashiCorp repository
echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list

# Update and install Nomad
apt update
apt install -y nomad
```

### **2. Vérifier Installation**

```bash
nomad version
# Expected: Nomad v1.8+ (latest)
```

---

## ⚙️ Configuration Nomad Server

### **1. Créer Configuration**

```bash
mkdir -p /etc/nomad.d
mkdir -p /opt/nomad/data
```

Fichier : `/etc/nomad.d/nomad.hcl`

```hcl
# Nomad Server Configuration
# Server: 31.57.224.79 (Singapore)

datacenter = "sg1"
data_dir   = "/opt/nomad/data"
bind_addr  = "0.0.0.0"

# Server mode
server {
  enabled          = true
  bootstrap_expect = 1
  
  # Server RPC address
  server_join {
    retry_join = ["31.57.224.79"]
  }
}

# Client mode (run jobs on same server)
client {
  enabled = true
  
  # Host volumes for data persistence
  host_volume "scalping-data" {
    path      = "/root/data/scalping-live-bybit/data"
    read_only = false
  }
  
  host_volume "scalping-logs" {
    path      = "/root/data/scalping-live-bybit/logs"
    read_only = false
  }
  
  host_volume "scalping-state" {
    path      = "/root/data/scalping-live-bybit/state"
    read_only = false
  }
}

# UI
ui {
  enabled = true
  
  # Accessible on port 4646
  # Access: http://31.57.224.79:4646
}

# Telemetry (optional)
telemetry {
  publish_allocation_metrics = true
  publish_node_metrics       = true
}

# ACL (disabled for now, can enable later)
acl {
  enabled = false
}
```

### **2. Créer Service Systemd**

Fichier : `/etc/systemd/system/nomad.service`

```ini
[Unit]
Description=Nomad
Documentation=https://www.nomadproject.io/docs/
Wants=network-online.target
After=network-online.target

[Service]
Type=exec
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=/usr/bin/nomad agent -config=/etc/nomad.d
KillMode=process
KillSignal=SIGINT
LimitNOFILE=65536
LimitNPROC=infinity
Restart=on-failure
RestartSec=2
TasksMax=infinity

[Install]
WantedBy=multi-user.target
```

### **3. Créer Dossiers Données**

```bash
mkdir -p /root/data/scalping-live-bybit/{data,logs,state,config}
```

---

## 🚀 Démarrage Nomad

```bash
# Reload systemd
systemctl daemon-reload

# Enable Nomad
systemctl enable nomad

# Start Nomad
systemctl start nomad

# Check status
systemctl status nomad
```

---

## ✅ Vérification

### **1. Vérifier Cluster**

```bash
nomad server members
# Expected: 1 server alive

nomad node status
# Expected: 1 client ready
```

### **2. Vérifier UI**

```bash
# Test local
curl http://localhost:4646/v1/status/leader

# Access UI (from local machine)
# http://31.57.224.79:4646
```

### **3. Vérifier Logs**

```bash
journalctl -u nomad -f
```

---

## 🔒 Sécurité Firewall

### **Configuration UFW**

```bash
# Install UFW if not present
apt install -y ufw

# SSH (important - do this first!)
ufw allow 22/tcp

# Nomad ports
ufw allow 4646/tcp  # HTTP API + UI
ufw allow 4647/tcp  # RPC
ufw allow 4648/tcp  # Serf (gossip)

# WireGuard (to configure later)
ufw allow 51820/udp

# Enable firewall
ufw --force enable

# Check status
ufw status
```

---

## 📊 Configuration ACL (Optionnel)

Pour activer l'ACL plus tard :

```bash
# Bootstrap ACL
nomad acl bootstrap

# Save the Secret ID (management token)
# Export for CLI use
export NOMAD_TOKEN="<secret-id>"
```

---

## 🧪 Test Déploiement Simple

Créer : `test-job.nomad`

```hcl
job "test-echo" {
  datacenters = ["sg1"]
  type = "service"

  group "echo-group" {
    count = 1

    task "echo-task" {
      driver = "raw_exec"

      config {
        command = "/bin/echo"
        args    = ["Hello from Nomad Singapore!"]
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}
```

```bash
# Deploy test job
nomad job run test-job.nomad

# Check status
nomad job status test-echo

# View logs
nomad alloc logs $(nomad job allocs test-echo -json | jq -r '.[0].ID')

# Stop job
nomad job stop test-echo
```

---

## 📝 Notes Importantes

1. **Datacenter:** Configuré comme `sg1` (Singapore 1)
2. **Single Server:** Configuration pour 1 serveur (bootstrap_expect = 1)
3. **Modes combinés:** Server + Client sur même machine (économique pour démarrer)
4. **Volumes:** Host volumes configurés pour persistence données

---

## 🔄 Prochaines Étapes

1. ✅ Nomad Server installé
2. ⏳ Configurer WireGuard VPN (doc 03-wireguard-vpn.md)
3. ⏳ Déployer application Scalping (doc 04-deployment-workflow.md)

---

## 🆘 Troubleshooting

### **Nomad ne démarre pas**

```bash
# Vérifier config
nomad agent -config=/etc/nomad.d -dev-connect

# Vérifier logs
journalctl -u nomad -n 50
```

### **Port déjà utilisé**

```bash
# Vérifier port 4646
netstat -tlnp | grep 4646

# Kill process si nécessaire
sudo kill -9 <PID>
```

### **Permissions dossiers**

```bash
# Fix permissions
chown -R root:root /opt/nomad
chmod -R 755 /opt/nomad
```
