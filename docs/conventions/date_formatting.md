# 📅 Conventions de Formatage des Dates et Timestamps

Ce document définit les conventions de formatage des dates et timestamps utilisées dans le projet.

---

## 🎯 Fonction Centrale de Conversion

### **`timestampMsToTime()`**

**Fonction centrale utilisée dans tout le projet** pour convertir les timestamps en millisecondes vers `time.Time`.

```go
// timestampMsToTime convertit un timestamp en millisecondes vers time.Time
// FONCTION CENTRALE : Garantit la cohérence entre tous les modules
func timestampMsToTime(timestampMs int64) time.Time {
    return time.Unix(timestampMs/1000, 0).UTC()
}
```

**Localisation :**
- ✅ `cmd/scalping_engine/app.go`
- ✅ `cmd/scalping_paper/app_paper.go`

**Utilisation :**
```go
// Conversion timestamp Binance (ms) → time.Time
t := timestampMsToTime(kline.Timestamp)
t := timestampMsToTime(trade.Time)
t := timestampMsToTime(signal.Timestamp)
```

---

## 📊 Formats Standard

### **1. Format Logs Console**

**Format :** `15:04:05`  
**Usage :** Affichage court dans les logs temps réel

```go
t := timestampMsToTime(timestamp)
fmt.Printf("🕐 %s | MARQUEUR DÉTECTÉ\n", t.Format("15:04:05"))
```

**Exemple :**
```
🕐 14:35:00 | MARQUEUR DÉTECTÉ
```

---

### **2. Format Logs Détaillés**

**Format :** `15:04`  
**Usage :** Plages horaires dans les logs

```go
prevTime := klineTime.Add(-5 * time.Minute)
fmt.Printf("📊 KLINE FERMÉE (%s→%s)\n",
    prevTime.Format("15:04"),
    klineTime.Format("15:04"))
```

**Exemple :**
```
📊 KLINE FERMÉE (14:30→14:35)
```

---

### **3. Format Export JSON**

**Format :** `time.RFC3339`  
**Usage :** Export JSON (format ISO 8601)

```go
data := map[string]interface{}{
    "timestamp": timestampMsToTime(kline.Timestamp).Format(time.RFC3339),
}
```

**Exemple :**
```json
{
  "timestamp": "2023-06-01T14:35:00Z"
}
```

---

### **4. Format Notifications**

**Format :** `2006-01-02 15:04:05`  
**Usage :** Notifications ntfy (date + heure complète)

```go
dateTime := signal.Time.Format("2006-01-02 15:04:05")
msg := fmt.Sprintf("📅 Date: %s UTC\n", dateTime)
```

**Exemple :**
```
📅 Date: 2023-06-01 14:35:00 UTC
```

---

## 🔧 Conversions Courantes

### **Milliseconds → Time**
```go
t := timestampMsToTime(1685628900000)
// → 2023-06-01 14:35:00 UTC
```

### **Time → RFC3339**
```go
t := timestampMsToTime(timestamp)
str := t.Format(time.RFC3339)
// → "2023-06-01T14:35:00Z"
```

### **Time → HH:MM:SS**
```go
t := timestampMsToTime(timestamp)
str := t.Format("15:04:05")
// → "14:35:00"
```

### **Time → YYYY-MM-DD HH:MM:SS**
```go
t := timestampMsToTime(timestamp)
str := t.Format("2006-01-02 15:04:05")
// → "2023-06-01 14:35:00"
```

---

## 📋 Récapitulatif par Module

| Module | Format Console | Format JSON | Format Notification |
|--------|----------------|-------------|---------------------|
| **scalping_engine** | `15:04:05` | `RFC3339` | N/A |
| **scalping_paper** | `15:04:05` | N/A | `2006-01-02 15:04:05` |
| **ntfy_client** | N/A | N/A | `2006-01-02 15:04:05` |

---

## ⚠️ Important

### **Toujours UTC**

Tous les timestamps sont en **UTC** par défaut.

```go
// ✅ Correct
return time.Unix(timestampMs/1000, 0).UTC()

// ❌ Éviter (timezone locale)
return time.Unix(timestampMs/1000, 0)
```

### **Conversion Milliseconds**

Les timestamps Binance sont en **millisecondes**, pas en secondes.

```go
// ✅ Correct
time.Unix(timestamp/1000, 0)

// ❌ Incorrect
time.Unix(timestamp, 0)  // Donnerait l'an 55000+
```

---

## 🧪 Exemples Complets

### **Signal de Trading**

```go
// Dans scalping_engine
signal := &Signal{
    Time: timestampMsToTime(kline.Timestamp),
}

// Export JSON
data := map[string]interface{}{
    "timestamp": signal.Time.Format(time.RFC3339),
}

// Dans notification
signalInfo := notifications.SignalInfo{
    Time: timestampMsToTime(sig.Timestamp),
}
// Format ntfy: "2023-06-01 14:35:00 UTC"
```

### **Log Marqueur**

```go
markerTime := timestampMsToTime(nextMarker)
fmt.Printf("🕐 %s | MARQUEUR DÉTECTÉ\n", markerTime.Format("15:04:05"))
// Output: 🕐 14:35:00 | MARQUEUR DÉTECTÉ
```

### **Plage Horaire**

```go
startTime := timestampMsToTime(kline.Timestamp)
endTime := startTime.Add(5 * time.Minute)
fmt.Printf("📊 KLINE (%s→%s)\n",
    startTime.Format("15:04"),
    endTime.Format("15:04"))
// Output: 📊 KLINE (14:30→14:35)
```

---

## ✅ Checklist Développement

Lors de l'ajout de nouveau code manipulant des timestamps :

- [ ] Utiliser `timestampMsToTime()` pour les conversions
- [ ] Vérifier que le timestamp est en millisecondes
- [ ] Utiliser `.UTC()` pour garantir le fuseau horaire
- [ ] Choisir le bon format d'affichage selon le contexte :
  - Console logs : `15:04:05`
  - JSON export : `time.RFC3339`
  - Notifications : `2006-01-02 15:04:05`
- [ ] Documenter le format utilisé dans les commentaires

---

## 📚 Références Go

### **Constantes time.Layout**

```go
time.RFC3339      // "2006-01-02T15:04:05Z07:00"
time.RFC3339Nano  // "2006-01-02T15:04:05.999999999Z07:00"
time.Kitchen      // "3:04PM"
time.Stamp        // "Jan _2 15:04:05"
```

### **Formats Personnalisés**

Go utilise `2006-01-02 15:04:05` comme reference date.

| Élément | Format |
|---------|--------|
| Année | `2006` |
| Mois | `01` ou `Jan` |
| Jour | `02` |
| Heure (24h) | `15` |
| Minute | `04` |
| Seconde | `05` |
| Timezone | `Z07:00` |

---

## 🔗 Voir Aussi

- `cmd/scalping_engine/app.go` - Implementation reference
- `internal/notifications/ntfy_client.go` - Format notifications
- Go time package: https://pkg.go.dev/time
