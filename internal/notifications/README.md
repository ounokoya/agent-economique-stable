# 🔔 Module de Notifications Ntfy

Système de notifications pour les signaux de trading via protocole ntfy.

---

## 📋 Configuration

### **Serveur Ntfy**
- **URL** : `https://notifications.koyad.com`
- **Protocole** : ntfy.sh standard (HTTP POST)

### **Canaux (Topics)**

| Mode | Canal | Description |
|------|-------|-------------|
| **Paper** | `scalping-paper` | Notifications trading testnet |
| **Live** | `scalping-live` | Notifications trading production |

---

## 🎯 Types de Notifications

### **1. Signal de Trading**

Envoyé automatiquement quand un signal scalping est détecté.

**Contenu :**
```
🎯 Signal LONG détecté

📊 Signal: LONG
💰 Prix: 185.43 SOLUSDT
⏰ Heure: 14:35:27

📈 Indicateurs:
   • CCI: -105.2
   • MFI: 18.3
   • Stoch K: 15.7
   • Stoch D: 22.1

📦 Volume: 45782.50

🔧 Mode: paper
```

**Priorité :** High (4/5)  
**Tags :** 📈 (LONG) ou 📉 (SHORT)

---

### **2. Notification de Statut**

Envoyée au démarrage/arrêt de l'application.

**Exemple Démarrage :**
```
ℹ️ Status Scalping Engine

🚀 Démarrage Scalping paper

📊 Symbole: SOLUSDT
⏱️ Timeframe: 5m
🔧 Mode: paper
```

**Exemple Arrêt :**
```
ℹ️ Status Scalping Engine

🛑 Arrêt Scalping paper

📊 Signaux détectés: 12
```

**Priorité :** Default (3/5)  
**Tags :** ℹ️

---

### **3. Notification d'Erreur**

Envoyée en cas d'erreur critique.

**Exemple :**
```
⚠️ Erreur Scalping Engine

Erreur chargement initial: HTTP request failed
```

**Priorité :** Max (5/5)  
**Tags :** ⚠️

---

## 🔧 Utilisation dans le Code

### **Initialisation**

```go
import "agent-economique/internal/notifications"

// Créer client
notifier := notifications.NewNtfyClient(
    "https://notifications.koyad.com",
    "scalping-paper", // ou "scalping-live"
)
```

### **Envoyer Signal**

```go
signalInfo := notifications.SignalInfo{
    Type:    "LONG",
    Symbol:  "SOLUSDT",
    Price:   185.43,
    Time:    time.Now(),
    CCI:     -105.2,
    MFI:     18.3,
    StochK:  15.7,
    StochD:  22.1,
    Volume:  45782.50,
    Mode:    "paper",
}

err := notifier.SendSignalNotification(signalInfo)
```

### **Envoyer Statut**

```go
status := "🚀 Démarrage Scalping paper\n\nSymbole: SOLUSDT"
err := notifier.SendStatusNotification(status)
```

### **Envoyer Erreur**

```go
err := notifier.SendErrorNotification("Erreur critique: ...")
```

---

## 📱 Réception des Notifications

### **Web**
```
https://notifications.koyad.com/scalping-paper
https://notifications.koyad.com/scalping-live
```

### **Application Mobile**

1. **Installer ntfy** (iOS/Android)
2. **S'abonner au canal :**
   - Serveur : `https://notifications.koyad.com`
   - Topic : `scalping-paper` ou `scalping-live`

### **CLI**

```bash
# S'abonner (écouter)
ntfy subscribe --from-config https://notifications.koyad.com/scalping-paper

# Tester manuellement
curl -d "Test notification" https://notifications.koyad.com/scalping-paper
```

---

## 🔒 Sécurité

### **Topics Publics**
Les canaux sont publics par défaut. Toute personne connaissant l'URL peut s'abonner.

### **Recommandations**

1. **Topics uniques** : Utiliser des noms difficiles à deviner
2. **Pas de données sensibles** : Ne jamais envoyer :
   - Clés API
   - Montants exacts
   - Informations personnelles
3. **Filtrage côté client** : Valider l'origine des messages

---

## 📊 Format JSON (ntfy)

```json
{
  "topic": "scalping-paper",
  "title": "🎯 Signal LONG détecté",
  "message": "...",
  "priority": 4,
  "tags": ["chart_with_upwards_trend"]
}
```

### **Priorités**

| Valeur | Nom | Utilisation |
|--------|-----|-------------|
| 1 | Min | Logs de debug |
| 2 | Low | Info non urgente |
| 3 | Default | Statut normal |
| 4 | High | Signaux trading |
| 5 | Max | Erreurs critiques |

### **Tags Emoji**

- `chart_with_upwards_trend` → 📈 (LONG)
- `chart_with_downwards_trend` → 📉 (SHORT)
- `warning` → ⚠️ (Erreur)
- `information_source` → ℹ️ (Info)

---

## 🧪 Tests

### **Test Manuel**

```bash
# Tester connexion serveur
curl https://notifications.koyad.com

# Envoyer notification test
curl -H "Title: Test" \
     -d "Message de test" \
     https://notifications.koyad.com/scalping-paper
```

### **Test Code**

```go
// Test basique
notifier := notifications.NewNtfyClient(
    "https://notifications.koyad.com",
    "scalping-test",
)

err := notifier.SendStatusNotification("Test notification")
if err != nil {
    log.Fatal(err)
}
```

---

## 🔍 Dépannage

### **Notifications non reçues**

1. Vérifier URL serveur : `https://notifications.koyad.com`
2. Vérifier topic : `scalping-paper` ou `scalping-live`
3. Tester connexion : `curl https://notifications.koyad.com`
4. Vérifier logs application

### **Erreur HTTP**

```go
// Logs d'erreur détaillés
if err := notifier.SendSignalNotification(signalInfo); err != nil {
    log.Printf("Notification échouée: %v", err)
}
```

### **Timeout**

Le client HTTP a un timeout de **10 secondes**. Si le serveur ne répond pas, la notification échouera silencieusement (non bloquant).

---

## 📚 Références

- **Ntfy Documentation** : https://ntfy.sh/docs
- **API Specification** : https://ntfy.sh/docs/publish
- **Mobile Apps** : https://ntfy.sh/docs/subscribe/phone

---

## ✅ Intégration Actuelle

**Applications utilisant le module :**
- ✅ `cmd/scalping_paper` (Paper + Live trading)

**Points d'envoi :**
1. Démarrage application
2. Détection signal scalping
3. Erreurs critiques
4. Arrêt application
