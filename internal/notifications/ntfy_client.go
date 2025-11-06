// Package notifications provides notification services for trading signals
package notifications

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

// NtfyClient manages notifications via ntfy.sh protocol
type NtfyClient struct {
	serverURL string
	topic     string
	client    *http.Client
}

// NtfyMessage represents a notification message
type NtfyMessage struct {
	Topic    string   `json:"topic"`
	Title    string   `json:"title"`
	Message  string   `json:"message"`
	Priority int      `json:"priority,omitempty"` // 1=min, 3=default, 5=max
	Tags     []string `json:"tags,omitempty"`
	Actions  []Action `json:"actions,omitempty"`
}

// Action represents a clickable action in the notification
type Action struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	URL    string `json:"url,omitempty"`
}

// NewNtfyClient creates a new ntfy notification client
func NewNtfyClient(serverURL, topic string) *NtfyClient {
	return &NtfyClient{
		serverURL: serverURL,
		topic:     topic,
		client: &http.Client{
			Timeout: 30 * time.Second, // Timeout plus long pour première requête HTTP/DNS/TLS
		},
	}
}

// SendSignalNotification sends a trading signal notification
func (n *NtfyClient) SendSignalNotification(signal SignalInfo) error {
	// Format message
	title := fmt.Sprintf("🎯 Signal %s détecté", signal.Type)
	message := n.formatSignalMessage(signal)

	// Déterminer priorité selon type
	priority := 4 // High
	if signal.Type == "LONG" {
		priority = 4 // High (bullish)
	} else {
		priority = 4 // High (bearish)
	}

	// Tags
	tags := []string{"chart_with_upwards_trend"}
	if signal.Type == "SHORT" {
		tags = []string{"chart_with_downwards_trend"}
	}

	// Créer message ntfy
	msg := NtfyMessage{
		Topic:    n.topic,
		Title:    title,
		Message:  message,
		Priority: priority,
		Tags:     tags,
	}

	return n.send(msg)
}

// SendErrorNotification sends an error notification
func (n *NtfyClient) SendErrorNotification(errorMsg string) error {
	msg := NtfyMessage{
		Topic:    n.topic,
		Title:    "⚠️ Erreur Scalping Engine",
		Message:  errorMsg,
		Priority: 5, // Max
		Tags:     []string{"warning"},
	}

	return n.send(msg)
}

// SendStatusNotification sends a status notification
func (n *NtfyClient) SendStatusNotification(status string) error {
	msg := NtfyMessage{
		Topic:    n.topic,
		Title:    "ℹ️ Status Scalping Engine",
		Message:  status,
		Priority: 3, // Default
		Tags:     []string{"information_source"},
	}

	return n.send(msg)
}

// send sends a notification to the ntfy server in TEXT format
func (n *NtfyClient) send(msg NtfyMessage) error {
	// URL complète
	url := fmt.Sprintf("%s/%s", n.serverURL, n.topic)

	// Créer requête HTTP POST avec le MESSAGE en text/plain
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(msg.Message))
	if err != nil {
		return fmt.Errorf("request creation error: %w", err)
	}

	// Headers ntfy (format TEXT)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("Title", msg.Title)
	
	// Priority
	if msg.Priority > 0 {
		req.Header.Set("Priority", fmt.Sprintf("%d", msg.Priority))
	}
	
	// Tags (comma-separated)
	if len(msg.Tags) > 0 {
		tags := ""
		for i, tag := range msg.Tags {
			if i > 0 {
				tags += ","
			}
			tags += tag
		}
		req.Header.Set("Tags", tags)
	}

	// Envoyer
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request error: %w", err)
	}
	defer resp.Body.Close()

	// Vérifier status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy server returned status %d", resp.StatusCode)
	}

	return nil
}

// formatSignalMessage formate un message de signal
func (n *NtfyClient) formatSignalMessage(signal SignalInfo) string {
	// Format date/heure cohérent avec scalping_engine
	// Format: 2006-01-02 15:04:05
	dateTime := signal.Time.Format("2006-01-02 15:04:05")

	msg := fmt.Sprintf("📊 Signal: %s\n", signal.Type)
	msg += fmt.Sprintf("💰 Prix: %.2f %s\n", signal.Price, signal.Symbol)
	msg += fmt.Sprintf("📅 Date: %s UTC\n", dateTime)
	msg += "\n📈 Indicateurs:\n"
	msg += fmt.Sprintf("   • CCI: %.1f\n", signal.CCI)
	msg += fmt.Sprintf("   • MFI: %.1f\n", signal.MFI)
	msg += fmt.Sprintf("   • Stoch K: %.1f\n", signal.StochK)
	msg += fmt.Sprintf("   • Stoch D: %.1f\n", signal.StochD)
	msg += fmt.Sprintf("\n📦 Volume: %.2f\n", signal.Volume)

	if signal.Mode != "" {
		msg += fmt.Sprintf("\n🔧 Mode: %s\n", signal.Mode)
	}

	return msg
}

// SignalInfo holds information about a trading signal
type SignalInfo struct {
	Type   string    // "LONG" ou "SHORT"
	Symbol string    // "SOLUSDT"
	Price  float64   // Prix du signal
	Time   time.Time // Heure du signal
	CCI    float64   // Valeur CCI
	MFI    float64   // Valeur MFI
	StochK float64   // Valeur Stochastic K
	StochD float64   // Valeur Stochastic D
	Volume float64   // Volume
	Mode   string    // "paper" ou "live"
}
