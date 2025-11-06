// Package shared provides configuration loading functionality
package shared

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	BinanceData BinanceDataConfig `yaml:"binance_data"`
	BingXData   BingXDataConfig   `yaml:"bingx_data"`
	Strategy    StrategyConfig    `yaml:"strategy"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	Environment EnvironmentConfig `yaml:"environment"`
	Backtest    BacktestConfig    `yaml:"backtest"`
	DataPeriod  DataPeriodConfig  `yaml:"data_period"`
	CLI         CLIConfig         `yaml:"cli"`
}

// BinanceDataConfig holds Binance-specific configuration
type BinanceDataConfig struct {
	CacheRoot   string              `yaml:"cache_root"`
	Symbols     []string            `yaml:"symbols"`
	Timeframes  []string            `yaml:"timeframes"`
	DataTypes   []string            `yaml:"data_types"`    // klines, trades
	Cache       CacheConfig         `yaml:"cache"`
	Downloader  DownloadConfigYAML  `yaml:"downloader"`
	Streaming   StreamingConfigYAML `yaml:"streaming"`
	Validation  ValidationConfigYAML `yaml:"validation"`
}

// BingXDataConfig holds BingX-specific configuration
type BingXDataConfig struct {
	Environment string             `yaml:"environment"` // "demo" or "live"
	MarketType  string             `yaml:"market_type"` // "spot" or "perpetual"
	Credentials BingXCredentials   `yaml:"credentials"`
	Timeout     string             `yaml:"timeout"`     // "30s"
	RateLimit   BingXRateLimit     `yaml:"rate_limit"`
}

// BingXCredentials holds BingX API credentials
type BingXCredentials struct {
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
}

// BingXRateLimit holds BingX rate limiting configuration
type BingXRateLimit struct {
	MarketData int `yaml:"market_data"` // requests per second
	Trading    int `yaml:"trading"`     // requests per second
	Account    int `yaml:"account"`     // requests per second
}

// DownloadConfigYAML matches the YAML structure for downloader config
type DownloadConfigYAML struct {
	BaseURL        string `yaml:"base_url"`
	MaxRetries     int    `yaml:"max_retries"`
	RetryDelay     string `yaml:"retry_delay"`
	Timeout        string `yaml:"timeout"`
	MaxConcurrent  int    `yaml:"max_concurrent"`
	ChecksumVerify bool   `yaml:"checksum_verify"`
}

// StreamingConfigYAML matches the YAML structure for streaming config
type StreamingConfigYAML struct {
	BufferSize    int  `yaml:"buffer_size"`
	MaxMemoryMB   int  `yaml:"max_memory_mb"`
	EnableMetrics bool `yaml:"enable_metrics"`
}

// ValidationConfigYAML matches the YAML structure for validation config
type ValidationConfigYAML struct {
	MaxPriceDeviation    float64 `yaml:"max_price_deviation"`
	MaxVolumeDeviation   float64 `yaml:"max_volume_deviation"`
	RequireMonotonicTime bool    `yaml:"require_monotonic_time"`
	MaxTimeGap          int64   `yaml:"max_time_gap"`
}

// StrategyConfig holds trading strategy configuration
type StrategyConfig struct {
	Name               string                    `yaml:"name"`
	ScalpingConfig     ScalpingStrategyConfig    `yaml:"scalping"`
	Indicators         IndicatorsConfig          `yaml:"indicators"`
	SignalGeneration   SignalGenerationConfig    `yaml:"signal_generation"`
	PositionManagement PositionManagementConfig  `yaml:"position_management"`
}

// ScalpingStrategyConfig holds scalping strategy configuration
type ScalpingStrategyConfig struct {
	// Seuils extrêmes
	CCISurachat   float64 `yaml:"cci_surachat"`
	CCISurvente   float64 `yaml:"cci_survente"`
	MFISurachat   float64 `yaml:"mfi_surachat"`
	MFISurvente   float64 `yaml:"mfi_survente"`
	StochSurachat float64 `yaml:"stoch_surachat"`
	StochSurvente float64 `yaml:"stoch_survente"`
	
	// Synchronisation
	SyncMFIEnabled bool `yaml:"sync_mfi_enabled"`
	SyncCCIEnabled bool `yaml:"sync_cci_enabled"`
	
	// Validation
	ValidationWindow int     `yaml:"validation_window"`
	VolumeThreshold  float64 `yaml:"volume_threshold"`
	VolumePeriod     int     `yaml:"volume_period"`
	VolumeMaxExt     int     `yaml:"volume_max_ext"`
	
	// Timeframe
	Timeframe string `yaml:"timeframe"`
}

// SignalGenerationConfig holds signal generation configuration
type SignalGenerationConfig struct {
	MinConfidence         float64 `yaml:"min_confidence"`
	PremiumConfidence     float64 `yaml:"premium_confidence"`
	RequireBarConfirmation bool   `yaml:"require_bar_confirmation"`
	EnableMultiTF         bool   `yaml:"enable_multi_tf"`
	HigherTimeframe       string `yaml:"higher_timeframe"`
}

// IndicatorsConfig holds indicator-specific configuration
type IndicatorsConfig struct {
	// Legacy indicators (MACD/CCI/DMI strategy)
	MACD MACDConfig `yaml:"macd"`
	CCI  CCIConfig  `yaml:"cci"`
	DMI  DMIConfig  `yaml:"dmi"`
	
	// New indicators (STOCH/MFI/CCI strategy)
	Stochastic StochasticConfig `yaml:"stochastic"`
	MFI        MFIConfig        `yaml:"mfi"`
}

// MACDConfig holds MACD indicator configuration
type MACDConfig struct {
	FastPeriod   int `yaml:"fast_period"`
	SlowPeriod   int `yaml:"slow_period"`
	SignalPeriod int `yaml:"signal_period"`
}

// CCIConfig holds CCI indicator configuration
type CCIConfig struct {
	Period              int `yaml:"period"`
	ThresholdOversold   int `yaml:"threshold_oversold"`
	ThresholdOverbought int `yaml:"threshold_overbought"`
}

// DMIConfig holds DMI indicator configuration
type DMIConfig struct {
	Period int `yaml:"period"`
}

// StochasticConfig holds Stochastic indicator configuration
type StochasticConfig struct {
	PeriodK     int     `yaml:"period_k"`
	SmoothK     int     `yaml:"smooth_k"`
	PeriodD     int     `yaml:"period_d"`
	Oversold    float64 `yaml:"oversold"`
	Overbought  float64 `yaml:"overbought"`
}

// MFIConfig holds Money Flow Index indicator configuration
type MFIConfig struct {
	Period      int     `yaml:"period"`
	Oversold    float64 `yaml:"oversold"`
	Overbought  float64 `yaml:"overbought"`
}

// PositionManagementConfig holds position management configuration
type PositionManagementConfig struct {
	// Legacy configuration
	TrailingStopDefault  bool `yaml:"trailing_stop_default"`
	ExitOnReverseSignal bool `yaml:"exit_on_reverse_signal"`
	
	// New STOCH/MFI/CCI configuration
	BaseTrailingPercent      float64 `yaml:"base_trailing_percent"`
	TrendTrailingPercent     float64 `yaml:"trend_trailing_percent"`
	CounterTrendTrailing     float64 `yaml:"counter_trend_trailing"`
	
	// Dynamic Adjustments
	EnableDynamicAdjustments bool    `yaml:"enable_dynamic_adjustments"`
	STOCHInverseAdjust       float64 `yaml:"stoch_inverse_adjust"`
	MFIInverseAdjust         float64 `yaml:"mfi_inverse_adjust"`
	CCIInverseAdjust         float64 `yaml:"cci_inverse_adjust"`
	TripleInverseAdjust      float64 `yaml:"triple_inverse_adjust"`
	
	// Safety Limits
	MaxCumulativeAdjust      float64 `yaml:"max_cumulative_adjust"`
	MinTrailingPercent       float64 `yaml:"min_trailing_percent"`
	MaxTrailingPercent       float64 `yaml:"max_trailing_percent"`
	
	// Early Exit
	EnableEarlyExit          bool    `yaml:"enable_early_exit"`
	MinProfitForEarlyExit    float64 `yaml:"min_profit_for_early_exit"`
	TripleInverseEarlyExit   bool    `yaml:"triple_inverse_early_exit"`
}

// ToBingXClientConfig converts BingXDataConfig to BingX SDK ClientConfig
func (c *BingXDataConfig) ToBingXClientConfig() (interface{}, error) {
	// Parse timeout
	timeout, err := time.ParseDuration(c.Timeout)
	if err != nil {
		timeout = 30 * time.Second // Default fallback
	}

	// Convert environment string to SDK environment type
	var environment string
	switch c.Environment {
	case "demo":
		environment = "demo"
	case "live":
		environment = "live"
	default:
		environment = "demo" // Default to demo for safety
	}

	return map[string]interface{}{
		"environment": environment,
		"credentials": map[string]string{
			"api_key":    c.Credentials.APIKey,
			"secret_key": c.Credentials.SecretKey,
		},
		"timeout": timeout,
	}, nil
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	EnableMetrics bool   `yaml:"enable_metrics"`
	LogLevel      string `yaml:"log_level"`
	LogFormat     string `yaml:"log_format"`
}

// EnvironmentConfig holds environment configuration
type EnvironmentConfig struct {
	Mode string `yaml:"mode"` // backtest, paper, live, notification
}

// BacktestConfig holds backtest-specific configuration
type BacktestConfig struct {
	WindowSize                  int           `yaml:"window_size"`                    // Nombre de klines pour calculs indicateurs (default: 300)
	TradesHistorySize           int           `yaml:"trades_history_size"`            // Buffer historique trades (default: 300)
	ExportJSON                  bool          `yaml:"export_json"`                    // Activer export JSON des signaux
	ExportPath                  string        `yaml:"export_path"`                    // Dossier export JSON
	ExportDetailedVerification  bool          `yaml:"export_detailed_verification"`   // Inclure détails vérification dans JSON
	Logging                     LoggingConfig `yaml:"logging"`                        // Configuration logs pour optimisation performance
}

// LoggingConfig holds logging configuration for backtest optimization
type LoggingConfig struct {
	EnableTradeLogs     bool `yaml:"enable_trade_logs"`      // Logger chaque trade (verbeux, ralentit backtest)
	EnableMarkerLogs    bool `yaml:"enable_marker_logs"`     // Logger détection marqueurs 5m
	EnableKlineLogs     bool `yaml:"enable_kline_logs"`      // Logger klines fermées au marqueur
	EnableIndicatorLogs bool `yaml:"enable_indicator_logs"`  // Logger valeurs indicateurs au marqueur
	EnableSignalLogs    bool `yaml:"enable_signal_logs"`     // Logger signaux détectés
	EnableProgressLogs  bool `yaml:"enable_progress_logs"`   // Logger progression par jour
	EnableSummaryLogs   bool `yaml:"enable_summary_logs"`    // Logger résumé final
}

// DataPeriodConfig holds data period configuration
type DataPeriodConfig struct {
	StartDate string `yaml:"start_date"`
	EndDate   string `yaml:"end_date"`
}

// CLIConfig holds CLI-specific configuration
type CLIConfig struct {
	ExecutionMode    string `yaml:"execution_mode"`    // default, download-only, processing-only, streaming, batch
	MemoryLimitMB    int    `yaml:"memory_limit_mb"`   // Memory limit for streaming mode
	ForceRedownload  bool   `yaml:"force_redownload"`  // Force re-download of existing files
	Verbose          bool   `yaml:"verbose"`           // Enable verbose logging
	EnableMetrics    bool   `yaml:"enable_metrics"`    // Enable performance metrics
}

// LoadConfig loads configuration from YAML file
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	return &config, nil
}

// ToDownloadConfig converts YAML config to DownloadConfig
func (d DownloadConfigYAML) ToDownloadConfig() (DownloadConfig, error) {
	config := DownloadConfig{
		BaseURL:        d.BaseURL,
		MaxRetries:     d.MaxRetries,
		MaxConcurrent:  d.MaxConcurrent,
		ChecksumVerify: d.ChecksumVerify,
	}

	// Set defaults for zero values
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 5
	}

	// Parse duration strings
	if d.RetryDelay != "" {
		retryDelay, err := time.ParseDuration(d.RetryDelay)
		if err != nil {
			return config, fmt.Errorf("invalid retry_delay format: %w", err)
		}
		config.RetryDelay = retryDelay
	}

	if d.Timeout != "" {
		timeout, err := time.ParseDuration(d.Timeout)
		if err != nil {
			return config, fmt.Errorf("invalid timeout format: %w", err)
		}
		config.Timeout = timeout
	}

	return config, nil
}

// ToStreamingConfig converts YAML config to StreamingConfig
func (s StreamingConfigYAML) ToStreamingConfig() StreamingConfig {
	return StreamingConfig{
		BufferSize:    s.BufferSize,
		MaxMemoryMB:   s.MaxMemoryMB,
		EnableMetrics: s.EnableMetrics,
	}
}

// ToValidationConfig converts YAML config to ValidationConfig
func (v ValidationConfigYAML) ToValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxPriceDeviation:    v.MaxPriceDeviation,
		MaxVolumeDeviation:   v.MaxVolumeDeviation,
		RequireMonotonicTime: v.RequireMonotonicTime,
		MaxTimeGap:          v.MaxTimeGap,
	}
}

// ToAggregationConfig creates AggregationConfig from the loaded config
func (c *Config) ToAggregationConfig() AggregationConfig {
	return AggregationConfig{
		SourceTimeframes: []string{"5m"}, // Default source
		TargetTimeframes: c.BinanceData.Timeframes,
		ValidationRules:  c.BinanceData.Validation.ToValidationConfig(),
	}
}

// ConvertToStrategyConfig creates a generic interface for strategy configs
func (c STOCHMFICCIStrategyConfig) ConvertToStrategyConfig() interface{} {
	// This will be used for passing to strategy constructors
	return map[string]interface{}{
		"stoch_period_k":             c.StochPeriodK,
		"stoch_smooth_k":             c.StochSmoothK,
		"stoch_period_d":             c.StochPeriodD,
		"stoch_oversold":             c.StochOversold,
		"stoch_overbought":           c.StochOverbought,
		"mfi_period":                 c.MFIPeriod,
		"mfi_oversold":               c.MFIOversold,
		"mfi_overbought":             c.MFIOverbought,
		"cci_threshold":              c.CCIThreshold,
		"min_confidence":             c.MinConfidence,
		"premium_confidence":         c.PremiumConfidence,
		"require_bar_confirmation":   c.RequireBarConfirmation,
		"higher_timeframe":           c.HigherTimeframe,
		"enable_multi_tf":            c.EnableMultiTF,
		"base_trailing_percent":      c.BaseTrailingPercent,
		"trend_trailing_percent":     c.TrendTrailingPercent,
		"counter_trend_trailing":     c.CounterTrendTrailing,
		"enable_dynamic_adjustments": c.EnableDynamicAdjustments,
		"stoch_inverse_adjust":       c.STOCHInverseAdjust,
		"mfi_inverse_adjust":         c.MFIInverseAdjust,
		"cci_inverse_adjust":         c.CCIInverseAdjust,
		"triple_inverse_adjust":      c.TripleInverseAdjust,
		"max_cumulative_adjust":      c.MaxCumulativeAdjust,
		"min_trailing_percent":       c.MinTrailingPercent,
		"max_trailing_percent":       c.MaxTrailingPercent,
		"enable_early_exit":          c.EnableEarlyExit,
		"min_profit_for_early_exit":  c.MinProfitForEarlyExit,
		"triple_inverse_early_exit":  c.TripleInverseEarlyExit,
	}
}

// ToSTOCHMFICCIStrategyConfig converts Config to STOCH/MFI/CCI strategy config
func (c *Config) ToSTOCHMFICCIStrategyConfig() STOCHMFICCIStrategyConfig {
	return STOCHMFICCIStrategyConfig{
		// Indicator Parameters
		StochPeriodK:    c.Strategy.Indicators.Stochastic.PeriodK,
		StochSmoothK:    c.Strategy.Indicators.Stochastic.SmoothK,
		StochPeriodD:    c.Strategy.Indicators.Stochastic.PeriodD,
		StochOversold:   c.Strategy.Indicators.Stochastic.Oversold,
		StochOverbought: c.Strategy.Indicators.Stochastic.Overbought,
		
		MFIPeriod:       c.Strategy.Indicators.MFI.Period,
		MFIOversold:     c.Strategy.Indicators.MFI.Oversold,
		MFIOverbought:   c.Strategy.Indicators.MFI.Overbought,
		
		CCIThreshold:    float64(c.Strategy.Indicators.CCI.ThresholdOverbought), // Use overbought as threshold
		
		// Signal Generation
		MinConfidence:         c.Strategy.SignalGeneration.MinConfidence,
		PremiumConfidence:     c.Strategy.SignalGeneration.PremiumConfidence,
		RequireBarConfirmation: c.Strategy.SignalGeneration.RequireBarConfirmation,
		HigherTimeframe:       c.Strategy.SignalGeneration.HigherTimeframe,
		EnableMultiTF:         c.Strategy.SignalGeneration.EnableMultiTF,
		
		// Position Management
		BaseTrailingPercent:      c.Strategy.PositionManagement.BaseTrailingPercent,
		TrendTrailingPercent:     c.Strategy.PositionManagement.TrendTrailingPercent,
		CounterTrendTrailing:     c.Strategy.PositionManagement.CounterTrendTrailing,
		
		// Dynamic Adjustments
		EnableDynamicAdjustments: c.Strategy.PositionManagement.EnableDynamicAdjustments,
		STOCHInverseAdjust:       c.Strategy.PositionManagement.STOCHInverseAdjust,
		MFIInverseAdjust:         c.Strategy.PositionManagement.MFIInverseAdjust,
		CCIInverseAdjust:         c.Strategy.PositionManagement.CCIInverseAdjust,
		TripleInverseAdjust:      c.Strategy.PositionManagement.TripleInverseAdjust,
		
		// Safety Limits
		MaxCumulativeAdjust:      c.Strategy.PositionManagement.MaxCumulativeAdjust,
		MinTrailingPercent:       c.Strategy.PositionManagement.MinTrailingPercent,
		MaxTrailingPercent:       c.Strategy.PositionManagement.MaxTrailingPercent,
		
		// Early Exit
		EnableEarlyExit:          c.Strategy.PositionManagement.EnableEarlyExit,
		MinProfitForEarlyExit:    c.Strategy.PositionManagement.MinProfitForEarlyExit,
		TripleInverseEarlyExit:   c.Strategy.PositionManagement.TripleInverseEarlyExit,
	}
}

// STOCHMFICCIStrategyConfig defines the configuration structure for STOCH/MFI/CCI strategy
// This should match the structure expected by stoch_mfi_cci.StrategyConfig
type STOCHMFICCIStrategyConfig struct {
	// Indicator Parameters
	StochPeriodK    int     `json:"stoch_period_k"`
	StochSmoothK    int     `json:"stoch_smooth_k"`
	StochPeriodD    int     `json:"stoch_period_d"`
	StochOversold   float64 `json:"stoch_oversold"`
	StochOverbought float64 `json:"stoch_overbought"`
	
	MFIPeriod       int     `json:"mfi_period"`
	MFIOversold     float64 `json:"mfi_oversold"`
	MFIOverbought   float64 `json:"mfi_overbought"`
	
	CCIThreshold    float64 `json:"cci_threshold"`
	
	// Signal Generation
	MinConfidence         float64 `json:"min_confidence"`
	PremiumConfidence     float64 `json:"premium_confidence"`
	RequireBarConfirmation bool   `json:"require_bar_confirmation"`
	HigherTimeframe       string `json:"higher_timeframe"`
	EnableMultiTF         bool   `json:"enable_multi_tf"`
	
	// Position Management
	BaseTrailingPercent      float64 `json:"base_trailing_percent"`
	TrendTrailingPercent     float64 `json:"trend_trailing_percent"`
	CounterTrendTrailing     float64 `json:"counter_trend_trailing"`
	
	// Dynamic Adjustments
	EnableDynamicAdjustments bool    `json:"enable_dynamic_adjustments"`
	STOCHInverseAdjust       float64 `json:"stoch_inverse_adjust"`
	MFIInverseAdjust         float64 `json:"mfi_inverse_adjust"`
	CCIInverseAdjust         float64 `json:"cci_inverse_adjust"`
	TripleInverseAdjust      float64 `json:"triple_inverse_adjust"`
	
	// Safety Limits
	MaxCumulativeAdjust      float64 `json:"max_cumulative_adjust"`
	MinTrailingPercent       float64 `json:"min_trailing_percent"`
	MaxTrailingPercent       float64 `json:"max_trailing_percent"`
	
	// Early Exit
	EnableEarlyExit          bool    `json:"enable_early_exit"`
	MinProfitForEarlyExit    float64 `json:"min_profit_for_early_exit"`
	TripleInverseEarlyExit   bool    `json:"triple_inverse_early_exit"`
}
