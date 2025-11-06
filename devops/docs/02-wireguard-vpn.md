# 🔒 WireGuard VPN Setup

**Objectif:** Sécuriser les communications entre machines locales et serveur Singapour

---

## 🌐 Architecture VPN

```
Machine Locale (Dev)          WireGuard VPN          Serveur Singapour
193.29.62.96 (US)      <--------------------->     31.57.224.79 (SG)
10.8.0.1 (VPN)                                     10.8.0.2 (VPN)

- Accès sécurisé Nomad UI
- Déploiement via VPN
- Communication cryptée
```

---

## 📋 Prérequis

- ✅ Root access sur les deux serveurs
- ✅ Ports UDP ouverts (51820)
- ✅ Ubuntu/Debian sur les deux machines

---

## 🚀 Installation WireGuard

### **Sur SERVEUR SINGAPOUR (31.57.224.79)**

```bash
# Update system
apt update

# Install WireGuard
apt install -y wireguard wireguard-tools

# Enable IP forwarding
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
echo "net.ipv6.conf.all.forwarding=1" >> /etc/sysctl.conf
sysctl -p
```

### **Sur MACHINE LOCALE (si Linux)**

```bash
# Same steps
apt update
apt install -y wireguard wireguard-tools
```

---

## 🔑 Génération Clés

### **Serveur Singapour (31.57.224.79)**

```bash
# Navigate to config directory
cd /etc/wireguard

# Generate server keys
wg genkey | tee server_private.key | wg pubkey > server_public.key

# Secure keys
chmod 600 server_private.key

# Display keys (save them)
echo "Server Private Key:"
cat server_private.key

echo "Server Public Key:"
cat server_public.key
```

### **Machine Locale**

```bash
# Generate client keys
cd /etc/wireguard
wg genkey | tee client_private.key | wg pubkey > client_public.key

chmod 600 client_private.key

echo "Client Private Key:"
cat client_private.key

echo "Client Public Key:"
cat client_public.key
```

---

## ⚙️ Configuration WireGuard

### **SERVEUR SINGAPOUR** : `/etc/wireguard/wg0.conf`

```ini
[Interface]
# Server VPN IP
Address = 10.8.0.1/24

# Server private key (replace with your generated key)
PrivateKey = SERVER_PRIVATE_KEY_HERE

# WireGuard port
ListenPort = 51820

# Post-up/down rules for NAT
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

# Peer: Machine Locale (US)
[Peer]
# Client public key
PublicKey = CLIENT_PUBLIC_KEY_HERE

# Client VPN IP
AllowedIPs = 10.8.0.2/32

# Keep alive (optional, for NAT traversal)
PersistentKeepalive = 25
```

**⚠️ Remplace :**
- `SERVER_PRIVATE_KEY_HERE` → Clé privée serveur
- `CLIENT_PUBLIC_KEY_HERE` → Clé publique client
- `eth0` → Interface réseau serveur (vérifier avec `ip a`)

### **MACHINE LOCALE** : `/etc/wireguard/wg0.conf`

```ini
[Interface]
# Client VPN IP
Address = 10.8.0.2/24

# Client private key
PrivateKey = CLIENT_PRIVATE_KEY_HERE

# DNS (optional)
DNS = 8.8.8.8

[Peer]
# Server public key
PublicKey = SERVER_PUBLIC_KEY_HERE

# Server endpoint (public IP + port)
Endpoint = 31.57.224.79:51820

# Traffic to route through VPN
# 10.8.0.0/24 = VPN traffic only
# 0.0.0.0/0 = All traffic (full tunnel)
AllowedIPs = 10.8.0.0/24

# Keep alive
PersistentKeepalive = 25
```

**⚠️ Remplace :**
- `CLIENT_PRIVATE_KEY_HERE` → Clé privée client
- `SERVER_PUBLIC_KEY_HERE` → Clé publique serveur

---

## 🔒 Sécuriser les Configs

```bash
# Sur les deux machines
chmod 600 /etc/wireguard/wg0.conf
```

---

## 🚀 Démarrage WireGuard

### **Serveur Singapour**

```bash
# Start WireGuard
wg-quick up wg0

# Enable at boot
systemctl enable wg-quick@wg0

# Check status
wg show
```

### **Machine Locale**

```bash
# Start WireGuard
wg-quick up wg0

# Enable at boot
systemctl enable wg-quick@wg0

# Check status
wg show
```

---

## ✅ Vérification Connexion

### **Test Ping**

```bash
# From local machine → ping server VPN IP
ping 10.8.0.1

# From server → ping client VPN IP
ping 10.8.0.2
```

### **Test Nomad via VPN**

```bash
# From local machine
curl http://10.8.0.1:4646/v1/status/leader
```

---

## 🔧 Configuration UFW (Firewall)

### **Serveur Singapour**

```bash
# Allow WireGuard port
ufw allow 51820/udp

# Allow Nomad only from VPN
ufw allow from 10.8.0.0/24 to any port 4646
ufw allow from 10.8.0.0/24 to any port 4647
ufw allow from 10.8.0.0/24 to any port 4648

# Reload firewall
ufw reload
```

---

## 📊 Monitoring WireGuard

```bash
# Show interface status
wg show

# Show detailed stats
wg show wg0

# Monitor traffic
watch -n 1 'wg show wg0 transfer'
```

---

## 🔄 Redémarrage WireGuard

```bash
# Restart
wg-quick down wg0
wg-quick up wg0

# Or via systemd
systemctl restart wg-quick@wg0
```

---

## 🆘 Troubleshooting

### **Connexion ne s'établit pas**

```bash
# 1. Vérifier clés publiques/privées
wg show wg0

# 2. Vérifier firewall
ufw status

# 3. Vérifier port UDP ouvert
nc -u -v 31.57.224.79 51820

# 4. Check logs
journalctl -u wg-quick@wg0 -f
```

### **Handshake fails**

```bash
# Vérifier clés correctes dans configs
grep PublicKey /etc/wireguard/wg0.conf

# Vérifier endpoint
grep Endpoint /etc/wireguard/wg0.conf

# Restart both ends
systemctl restart wg-quick@wg0
```

### **IP forwarding not working**

```bash
# Verify IP forwarding enabled
sysctl net.ipv4.ip_forward
# Should return: 1

# If not, enable it
echo 1 > /proc/sys/net/ipv4/ip_forward
```

---

## 🎯 Configuration Nomad avec VPN

### **Machine Locale - Access Nomad via VPN**

```bash
# Set Nomad address to VPN IP
export NOMAD_ADDR="http://10.8.0.1:4646"

# Test
nomad server members
```

### **Scripts de Déploiement**

Modifier les scripts pour utiliser VPN IP :

```bash
# In deploy scripts
NOMAD_ADDR="http://10.8.0.1:4646"
REMOTE_HOST="10.8.0.1"  # Use VPN IP instead of public IP
```

---

## 📝 Notes de Sécurité

1. ✅ **Clés privées** : Jamais partager, chmod 600
2. ✅ **Firewall** : Bloquer Nomad ports depuis internet public
3. ✅ **VPN only** : Nomad accessible uniquement via VPN
4. ✅ **Keep Alive** : Maintient connexion active (NAT traversal)

---

## 🔄 Prochaines Étapes

1. ✅ Nomad Server installé
2. ✅ WireGuard VPN configuré
3. ⏳ Déployer application Scalping (doc 04-deployment-workflow.md)

---

## 📚 Ressources

- [WireGuard Quick Start](https://www.wireguard.com/quickstart/)
- [WireGuard Ubuntu Guide](https://ubuntu.com/server/docs/wireguard-vpn)
- [DigitalOcean WireGuard Tutorial](https://www.digitalocean.com/community/tutorials/how-to-set-up-wireguard-on-ubuntu-20-04)
