package signals

import "time"

// SignalAction définit le type d'action du signal
type SignalAction string

const (
	SignalActionEntry SignalAction = "ENTRY" // Signal d'entrée en position
	SignalActionExit  SignalAction = "EXIT"  // Signal de sortie de position
)

// SignalType définit la direction du signal
type SignalType string

const (
	SignalTypeLong  SignalType = "LONG"  // Position longue (achat)
	SignalTypeShort SignalType = "SHORT" // Position courte (vente)
)

// Signal unifié pour tous les générateurs
type Signal struct {
	Timestamp  time.Time              // Timestamp du signal
	Action     SignalAction           // ENTRY ou EXIT
	Type       SignalType             // LONG ou SHORT
	Price      float64                // Prix au moment du signal
	Confidence float64                // 0.0 à 1.0 (qualité du signal)
	Metadata   map[string]interface{} // Métadonnées spécifiques au générateur

	// Pour les signaux EXIT : référence à l'entrée
	EntryPrice *float64   // Prix d'entrée (si EXIT)
	EntryTime  *time.Time // Time d'entrée (si EXIT)
}

// Config commune pour tous les générateurs
type GeneratorConfig struct {
	Symbol      string
	Timeframe   string
	HistorySize int // Nombre de klines à maintenir
}

// Generator interface que tous les générateurs doivent implémenter
type Generator interface {
	// Nom du générateur
	Name() string

	// Initialiser avec configuration
	Initialize(config GeneratorConfig) error

	// Calculer indicateurs sur historique complet
	CalculateIndicators(klines []Kline) error

	// Détecter signaux (retourne nouveaux signaux depuis dernier appel)
	DetectSignals(klines []Kline) ([]Signal, error)

	// Obtenir métriques du générateur
	GetMetrics() GeneratorMetrics
}

// Kline format unifié
type Kline struct {
	OpenTime time.Time
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
}

// GeneratorMetrics métriques communes
type GeneratorMetrics struct {
	TotalSignals   int
	EntrySignals   int
	ExitSignals    int
	LongSignals    int
	ShortSignals   int
	AvgConfidence  float64
	LastSignalTime time.Time
}
