# 🚀 DevOps Infrastructure - Scalping Live Bybit

Infrastructure complète pour déploiement sécurisé du bot de trading Scalping Live sur serveur Singapour.

> ⚠️ **IMPORTANT :** Si vous rencontrez des problèmes, consultez [CORRECTIONS.md](./CORRECTIONS.md) pour les correctifs critiques appliqués (TLS SANs, mTLS Caddy, driver exec, etc.)

---

## 📁 Structure

```
devops/
├── README.md                    # Cette documentation
├── QUICKSTART.md                # Guide démarrage rapide
├── CORRECTIONS.md               # ⚠️ Corrections critiques (Nov 2025)
├── docs/                        # Documentation détaillée
│   ├── 01-nomad-server-setup.md # Setup Nomad Server
│   ├── 02-wireguard-vpn.md      # Configuration VPN
│   └── 03-deployment-workflow.md # Workflow déploiement
├── scripts/                     # Scripts automatisés
│   ├── setup-server.sh          # Installation complète (master)
│   ├── install-nomad.sh         # Installation Nomad
│   ├── setup-wireguard.sh       # Setup WireGuard VPN
│   ├── generate-nomad-certs.sh  # Génération certificats TLS
│   ├── get-nomad-certs.sh       # Récupération certs local
│   ├── install-caddy.sh         # Installation Caddy
│   ├── deploy-binary.sh         # Déploiement binaire
│   ├── deploy-nomad-job.sh      # Déploiement job Nomad
│   ├── full-deploy.sh           # Déploiement complet
│   ├── manage-job.sh            # Gestion job (logs, restart)
│   └── check-server.sh          # Vérification infrastructure
└── configs/                     # Configurations
    ├── nomad-server.hcl         # Config Nomad Server
    └── scalping-live-bybit.nomad # Job definition
```

---

## 🌍 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Machine Locale (Dev)                                       │
│  193.29.62.96 (USA)                                         │
│  ├─ VPN IP: 10.8.0.2                                        │
│  └─ Accès: Nomad UI, Déploiement                           │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   │ WireGuard VPN (Encrypted)
                   │ Port: 51820/udp
                   │
┌──────────────────▼──────────────────────────────────────────┐
│  Serveur Production (Singapore)                             │
│  31.57.224.79                                               │
│  ├─ VPN IP: 10.8.0.1                                        │
│  ├─ Nomad Server: http://10.8.0.1:4646                     │
│  ├─ Nomad Client: Same server                              │
│  └─ Application: Scalping Live Bybit                       │
│     ├─ Binaire: /root/data/scalping-live-bybit/            │
│     ├─ Logs: /root/data/scalping-live-bybit/logs/          │
│     └─ State: /root/data/scalping-live-bybit/state/        │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  │ HTTPS API
                  │
┌─────────────────▼───────────────────────────────────────────┐
│  Bybit Exchange                                             │
│  api.bybit.com                                              │
│  ├─ Market Data (Klines)                                    │
│  └─ Category: linear (USDT Perpetual)                      │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start

### **1. Setup Infrastructure (One-Time)**

```bash
# Sur serveur Singapour (31.57.224.79)
ssh root@31.57.224.79

# Download scripts
git clone <repo> /root/agent_economique_stable
cd /root/agent_economique_stable/devops

# Install Nomad
./scripts/install-nomad.sh

# Setup WireGuard Server
./scripts/setup-wireguard.sh server
```

### **2. Setup VPN Client (Machine Locale)**

```bash
# Sur machine locale
cd devops

# Setup WireGuard Client
./scripts/setup-wireguard.sh client

# Test connexion
ping 10.8.0.1
```

### **3. Deploy Application**

```bash
# Sur machine locale (via VPN)
export NOMAD_ADDR="http://10.8.0.1:4646"

# Déploiement complet
./devops/scripts/full-deploy.sh
```

---

## 📖 Documentation Détaillée

### **Setup Infrastructure**

| Doc | Description | Temps |
|-----|-------------|-------|
| [01-nomad-server-setup.md](docs/01-nomad-server-setup.md) | Installation Nomad Server + Client | 30 min |
| [02-wireguard-vpn.md](docs/02-wireguard-vpn.md) | Configuration VPN sécurisé | 20 min |
| [03-deployment-workflow.md](docs/03-deployment-workflow.md) | Workflow déploiement app | 10 min |
| [04-certificates-setup.md](docs/04-certificates-setup.md) | Setup certificats TLS Nomad | 10 min |

### **Total Setup Time:** ~1h10

---

## 🔧 Scripts Disponibles

### **Infrastructure**

```bash
# Install Nomad on server
./scripts/install-nomad.sh

# Generate TLS certificates (server)
./scripts/generate-nomad-certs.sh

# Setup WireGuard VPN
./scripts/setup-wireguard.sh [server|client]

# Get Nomad certificates (local machine)
./scripts/get-nomad-certs.sh
```

### **Deployment**

```bash
# Deploy binary only
./scripts/deploy-binary.sh

# Deploy Nomad job only
./scripts/deploy-nomad-job.sh

# Full deployment (binary + config + job)
./scripts/full-deploy.sh

# Manage running job
./scripts/manage-job.sh [status|logs|restart|stop]
```

---

## ✅ Prérequis

### **Serveur Production (31.57.224.79)**

- ✅ OS: Ubuntu 24.04 LTS
- ✅ RAM: 2GB minimum
- ✅ Disk: 20GB minimum
- ✅ Root access
- ✅ Ports ouverts: 22 (SSH), 4646-4648 (Nomad), 51820 (WireGuard)

### **Machine Locale**

- ✅ SSH access to server
- ✅ Go 1.21+ installé
- ✅ WireGuard installé
- ✅ jq installé (pour parsing JSON)

---

## 🔒 Sécurité

### **VPN WireGuard**

- ✅ Chiffrement moderne (ChaCha20, Poly1305)
- ✅ Clés asymétriques (pas de secrets partagés)
- ✅ Keep-alive pour NAT traversal
- ✅ Nomad accessible uniquement via VPN

### **Firewall (UFW)**

```bash
# Ports autorisés
22/tcp     # SSH
4646/tcp   # Nomad UI/API (VPN only)
4647/tcp   # Nomad RPC (VPN only)
4648/tcp   # Nomad Serf (VPN only)
51820/udp  # WireGuard
```

### **ACL Nomad (Optionnel)**

```bash
# Bootstrap ACL for production
nomad acl bootstrap
```

---

## 📊 Monitoring

### **Nomad UI**

```
URL: http://10.8.0.1:4646
(Accessible via VPN)
```

### **Logs**

```bash
# Via Nomad
export NOMAD_ADDR="http://10.8.0.1:4646"
ALLOC_ID=$(nomad job allocs scalping-live-bybit -json | jq -r '.[0].ID')
nomad alloc logs -f $ALLOC_ID

# Via SSH
ssh root@31.57.224.79 "tail -f /root/data/scalping-live-bybit/logs/scalping.log"
```

### **Notifications**

Application envoie notifications via ntfy:
- ✅ Démarrage
- ✅ Signaux trading
- ✅ Erreurs

Topic: `scalping-live-bybit`

---

## 🆘 Troubleshooting

### **VPN ne connecte pas**

```bash
# Server: check WireGuard status
wg show

# Client: check connection
ping 10.8.0.1

# Check firewall
ufw status
```

### **Nomad inaccessible**

```bash
# Check Nomad service
ssh root@31.57.224.79 "systemctl status nomad"

# Check logs
ssh root@31.57.224.79 "journalctl -u nomad -n 50"

# Verify VPN active
wg show
```

### **Job ne démarre pas**

```bash
# Check job status
nomad job status scalping-live-bybit

# Check allocation
nomad alloc status <ALLOC_ID>

# View logs
nomad alloc logs <ALLOC_ID>
```

### **Bybit API bloquée**

```bash
# Test from server
ssh root@31.57.224.79 'curl -s "https://api.bybit.com/v5/market/kline?category=linear&symbol=SOLUSDT&interval=5&limit=1"'

# Should return JSON with retCode: 0
# If 403: Geographic restriction (check server location)
```

---

## 📝 Configuration Files

### **Nomad Server** (`configs/nomad-server.hcl`)

```hcl
datacenter = "sg1"
data_dir   = "/opt/nomad/data"

server {
  enabled = true
  bootstrap_expect = 1
}

client {
  enabled = true
}
```

### **Nomad Job** (`configs/scalping-live-bybit.nomad`)

```hcl
job "scalping-live-bybit" {
  datacenters = ["sg1"]
  type = "service"
  
  group "scalping-group" {
    task "scalping-task" {
      driver = "raw_exec"
      config {
        command = "/root/data/scalping-live-bybit/scalping_live_bybit"
        args = ["-config", "local/config.yaml"]
      }
    }
  }
}
```

---

## 🔄 Workflow Déploiement Standard

```bash
# 1. Develop locally
cd /root/projects/.../agent_economique_stable
# ... modify code ...

# 2. Compile for Linux
GOOS=linux GOARCH=amd64 go build -o scalping_live_bybit ./cmd/scalping_live_bybit

# 3. Upload binary
scp scalping_live_bybit root@31.57.224.79:/root/data/scalping-live-bybit/

# 4. Restart Nomad job
export NOMAD_ADDR="http://10.8.0.1:4646"
nomad job stop scalping-live-bybit
sleep 2
nomad job run devops/configs/scalping-live-bybit.nomad

# 5. Monitor
nomad alloc logs -f $(nomad job allocs scalping-live-bybit -json | jq -r '.[0].ID')
```

---

## 📚 Ressources

### **HashiCorp Nomad**

- [Official Docs](https://www.nomadproject.io/docs)
- [Job Specification](https://www.nomadproject.io/docs/job-specification)
- [CLI Reference](https://www.nomadproject.io/docs/commands)

### **WireGuard**

- [Official Site](https://www.wireguard.com/)
- [Quick Start](https://www.wireguard.com/quickstart/)
- [Ubuntu Guide](https://ubuntu.com/server/docs/wireguard-vpn)

---

## 🎯 Checklist Production

- [ ] Serveur Singapour provisionné (31.57.224.79)
- [ ] Nomad Server installé et opérationnel
- [ ] WireGuard VPN configuré (serveur + client)
- [ ] VPN testé et fonctionnel (ping 10.8.0.1)
- [ ] Bybit API accessible depuis serveur
- [ ] Binaire compilé et uploadé
- [ ] Configuration uploadée
- [ ] Job Nomad déployé
- [ ] Application démarre sans erreur
- [ ] Notifications fonctionnelles
- [ ] Monitoring actif
- [ ] Backup state configuré

---

## 👥 Support

**Questions / Issues:**  
Voir documentation détaillée dans `docs/`

**Logs importants:**
- Nomad: `journalctl -u nomad -f`
- WireGuard: `journalctl -u wg-quick@wg0 -f`
- Application: `/root/data/scalping-live-bybit/logs/`

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-06  
**Serveur:** 31.57.224.79 (Singapore)  
**Status:** ✅ Ready for Production
