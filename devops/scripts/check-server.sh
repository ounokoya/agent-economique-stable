#!/bin/bash
# 🔍 Vérification Infrastructure Serveur
# Check WireGuard, TLS, Nomad, Caddy

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

echo -e "${BLUE}🔍 VÉRIFICATION INFRASTRUCTURE SERVEUR${NC}"
echo "================================================"
echo ""

# Check 1: WireGuard
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  WireGuard VPN"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v wg &> /dev/null; then
    log_success "WireGuard installé"
    
    if wg show wg0 &> /dev/null; then
        log_success "Interface wg0 active"
        echo ""
        wg show wg0
        echo ""
    else
        log_error "Interface wg0 non active"
        log_info "Démarrer avec: wg-quick up wg0"
    fi
else
    log_error "WireGuard non installé"
fi

echo ""

# Check 2: TLS Certificates
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  Certificats TLS Nomad"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -d "/etc/nomad.d/certs" ]; then
    log_success "Dossier certificats existe"
    
    CERTS_REQUIRED=("ca.pem" "server.pem" "server-key.pem" "client.pem" "client-key.pem" "cli.pem" "cli-key.pem")
    CERTS_MISSING=0
    
    for cert in "${CERTS_REQUIRED[@]}"; do
        if [ -f "/etc/nomad.d/certs/$cert" ]; then
            echo "  ✅ $cert"
        else
            echo "  ❌ $cert (manquant)"
            CERTS_MISSING=$((CERTS_MISSING + 1))
        fi
    done
    
    if [ $CERTS_MISSING -eq 0 ]; then
        log_success "Tous les certificats présents"
    else
        log_error "$CERTS_MISSING certificats manquants"
    fi
    
    # Check client certs package
    if [ -d "/tmp/nomad-client-certs" ]; then
        log_success "Package client disponible (/tmp/nomad-client-certs)"
    else
        log_warning "Package client non trouvé (normal si déjà récupéré)"
    fi
else
    log_error "Dossier certificats manquant"
    log_info "Exécuter: ./generate-nomad-certs.sh"
fi

echo ""

# Check 3: Nomad
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  Nomad Server"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v nomad &> /dev/null; then
    NOMAD_VERSION=$(nomad version | head -1)
    log_success "Nomad installé: $NOMAD_VERSION"
    
    if systemctl is-active --quiet nomad; then
        log_success "Nomad service actif"
        
        # Check config
        if [ -f "/etc/nomad.d/nomad.hcl" ]; then
            log_success "Config Nomad présente"
            
            # Check TLS in config
            if grep -q "tls {" /etc/nomad.d/nomad.hcl; then
                log_success "TLS configuré"
            else
                log_warning "TLS non configuré"
            fi
            
            # Check wg0 interface
            if grep -q "network_interface.*wg0" /etc/nomad.d/nomad.hcl; then
                log_success "Interface wg0 configurée"
            else
                log_warning "Interface wg0 non configurée"
            fi
            
            # Check host volume
            if grep -q "host_volume" /etc/nomad.d/nomad.hcl; then
                log_success "Volumes host configurés"
            else
                log_warning "Volumes host non configurés"
            fi
        else
            log_error "Config Nomad manquante"
        fi
        
        echo ""
        log_info "Status Nomad:"
        nomad server members 2>/dev/null || log_error "Impossible de contacter Nomad"
        echo ""
        nomad node status 2>/dev/null || log_error "Aucun node"
        
    else
        log_error "Nomad service non actif"
        log_info "Démarrer avec: systemctl start nomad"
    fi
else
    log_error "Nomad non installé"
    log_info "Installer avec: ./install-nomad.sh"
fi

echo ""

# Check 4: Caddy
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4️⃣  Caddy Reverse Proxy"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v caddy &> /dev/null; then
    CADDY_VERSION=$(caddy version | head -1)
    log_success "Caddy installé: $CADDY_VERSION"
    
    if systemctl is-active --quiet caddy; then
        log_success "Caddy service actif"
        
        # Test endpoints
        echo ""
        log_info "Test endpoints:"
        
        if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q "200"; then
            log_success "Health check OK (http://10.8.0.1:8080/health)"
        else
            log_error "Health check échoué"
        fi
        
    else
        log_error "Caddy service non actif"
        log_info "Démarrer avec: systemctl start caddy"
    fi
else
    log_error "Caddy non installé"
    log_info "Installer avec: ./install-caddy.sh"
fi

echo ""

# Check 5: Firewall
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5️⃣  Firewall UFW"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v ufw &> /dev/null; then
    if ufw status | grep -q "Status: active"; then
        log_success "UFW actif"
        echo ""
        ufw status | grep -E "(51820|4646|80|8080)"
    else
        log_warning "UFW non actif"
    fi
else
    log_warning "UFW non installé"
fi

echo ""

# Check 6: Ports
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6️⃣  Ports en écoute"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

PORTS_TO_CHECK=("51820" "4646" "4647" "4648" "80" "8080")

for port in "${PORTS_TO_CHECK[@]}"; do
    if ss -tuln | grep -q ":$port "; then
        SERVICE=$(ss -tulnp | grep ":$port " | awk '{print $7}' | cut -d'"' -f2 | head -1)
        log_success "Port $port: $SERVICE"
    else
        log_warning "Port $port: non utilisé"
    fi
done

echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 RÉSUMÉ"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Count services
SERVICES_OK=0
SERVICES_TOTAL=4

systemctl is-active --quiet wg-quick@wg0 && SERVICES_OK=$((SERVICES_OK + 1))
systemctl is-active --quiet nomad && SERVICES_OK=$((SERVICES_OK + 1))
systemctl is-active --quiet caddy && SERVICES_OK=$((SERVICES_OK + 1))
[ -d "/etc/nomad.d/certs" ] && SERVICES_OK=$((SERVICES_OK + 1))

echo "Services actifs: $SERVICES_OK/$SERVICES_TOTAL"
echo ""

if [ $SERVICES_OK -eq $SERVICES_TOTAL ]; then
    log_success "Infrastructure complète et opérationnelle!"
    echo ""
    echo "🌐 Endpoints disponibles (via VPN 10.8.0.1):"
    echo "   https://10.8.0.1:4646  → Nomad UI (TLS)"
    echo "   http://10.8.0.1:80     → Nomad UI (via Caddy)"
    echo "   http://10.8.0.1:8080/health → Health check"
    echo ""
    echo "📝 Prochaines étapes:"
    echo "   1. Configurer VPN client en local"
    echo "   2. Récupérer certificats: ./get-nomad-certs.sh"
    echo "   3. Déployer application: ./full-deploy.sh"
else
    log_warning "Infrastructure incomplète ($SERVICES_OK/$SERVICES_TOTAL)"
    echo ""
    echo "📝 Actions requises:"
    [ ! -f "/etc/wireguard/wg0.conf" ] && echo "   - Configurer WireGuard: ./setup-wireguard.sh server"
    [ ! -d "/etc/nomad.d/certs" ] && echo "   - Générer certificats: ./generate-nomad-certs.sh"
    ! systemctl is-active --quiet nomad && echo "   - Démarrer Nomad: systemctl start nomad"
    ! systemctl is-active --quiet caddy && echo "   - Démarrer Caddy: systemctl start caddy"
fi

echo ""
