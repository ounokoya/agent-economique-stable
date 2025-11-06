// Binance Backtest Live - Real-time simulation using Binance API data
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/engine"
	"agent-economique/internal/shared"
)

const (
	// Application metadata
	AppName     = "Binance Backtest Live"
	AppVersion  = "1.0.0"
	AppDesc     = "🔄 Simulation temps réel avec données API Binance (10 derniers jours)"
	
	// Default configuration
	DefaultSymbol     = "SOLUSDT"
	DefaultTimeframe1 = "5m"
	DefaultTimeframe2 = "15m"
	DefaultDaysBack   = 10
)

// AppConfig holds application configuration
type AppConfig struct {
	ConfigPath   string
	Symbol       string
	Timeframe1   string // Primary timeframe (5m)
	Timeframe2   string // Secondary timeframe (15m)
	DaysBack     int    // Number of days to look back
	Verbose      bool
	DryRun       bool
}

// ApplicationState manages the running state
type ApplicationState struct {
	config         *AppConfig
	yamlConfig     *shared.Config
	binanceClient  *binance.Client
	temporalEngine *engine.TemporalEngine
	dataFetcher    *DataFetcher
	simEngine      *SimulationEngine
	dataset        *HistoricalDataSet
	ctx            context.Context
	cancel         context.CancelFunc
	running        bool
}

func main() {
	// Parse command line arguments
	config := parseFlags()
	
	// Display application banner
	displayBanner(config)
	
	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}
	
	// Initialize application state
	app, err := initializeApplication(config)
	if err != nil {
		log.Fatalf("❌ Initialization failed: %v", err)
	}
	defer app.cleanup()
	
	// Setup graceful shutdown
	app.setupGracefulShutdown()
	
	// Run the main application loop
	if err := app.run(); err != nil {
		log.Fatalf("❌ Application error: %v", err)
	}
	
	fmt.Println("👋 Binance Backtest Live terminé avec succès")
}

// parseFlags parses command line arguments
func parseFlags() *AppConfig {
	config := &AppConfig{}
	
	flag.StringVar(&config.ConfigPath, "config", "config/config.yaml", 
		"Path to YAML configuration file")
	flag.StringVar(&config.Symbol, "symbol", DefaultSymbol, 
		"Trading symbol (e.g., SOLUSDT)")
	flag.StringVar(&config.Timeframe1, "tf1", DefaultTimeframe1, 
		"Primary timeframe (5m, 15m, 1h)")
	flag.StringVar(&config.Timeframe2, "tf2", DefaultTimeframe2, 
		"Secondary timeframe (15m, 1h, 4h)")
	flag.IntVar(&config.DaysBack, "days", DefaultDaysBack, 
		"Number of days to look back")
	flag.BoolVar(&config.Verbose, "verbose", false, 
		"Enable verbose logging")
	flag.BoolVar(&config.DryRun, "dry-run", false, 
		"Dry run mode (no actual trading)")
	
	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s - %s\n\n", AppName, AppDesc)
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --symbol SOLUSDT --tf1 5m --tf2 15m --days 10\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --config config/config.yaml --verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --symbol ETHUSDT --days 7 --dry-run\n", os.Args[0])
	}
	
	flag.Parse()
	return config
}

// displayBanner shows application startup information
func displayBanner(config *AppConfig) {
	fmt.Printf("🔥 %s - Version %s\n", AppName, AppVersion)
	fmt.Printf("%s\n", AppDesc)
	fmt.Printf("================================================================================\n")
	fmt.Printf("📈 Symbol: %s | TF1: %s | TF2: %s | Days: %d\n", 
		config.Symbol, config.Timeframe1, config.Timeframe2, config.DaysBack)
	fmt.Printf("📁 Config: %s | Verbose: %t | DryRun: %t\n", 
		config.ConfigPath, config.Verbose, config.DryRun)
	fmt.Printf("================================================================================\n\n")
}

// validateConfig validates the application configuration
func validateConfig(config *AppConfig) error {
	// Validate symbol format
	if len(config.Symbol) < 6 {
		return fmt.Errorf("invalid symbol format: %s (expected format: SOLUSDT)", config.Symbol)
	}
	
	// Validate timeframes
	validTimeframes := map[string]bool{"1m": true, "5m": true, "15m": true, "1h": true, "4h": true, "1d": true}
	if !validTimeframes[config.Timeframe1] {
		return fmt.Errorf("invalid timeframe1: %s", config.Timeframe1)
	}
	if !validTimeframes[config.Timeframe2] {
		return fmt.Errorf("invalid timeframe2: %s", config.Timeframe2)
	}
	
	// Validate days back
	if config.DaysBack < 1 || config.DaysBack > 30 {
		return fmt.Errorf("invalid days back: %d (must be between 1-30)", config.DaysBack)
	}
	
	// Check config file existence
	if _, err := os.Stat(config.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", config.ConfigPath)
	}
	
	return nil
}

// initializeApplication sets up all application components
func initializeApplication(config *AppConfig) (*ApplicationState, error) {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	app := &ApplicationState{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Load YAML configuration
	yamlConfig, err := shared.LoadConfig(config.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load YAML config: %w", err)
	}
	app.yamlConfig = yamlConfig
	
	fmt.Printf("✅ YAML configuration loaded: %s\n", config.ConfigPath)
	
	// Initialize Binance client
	app.binanceClient = binance.NewClient()
	fmt.Printf("✅ Binance client initialized\n")
	
	// Initialize data fetcher
	app.dataFetcher = NewDataFetcher(
		app.binanceClient,
		config.Symbol,
		config.Timeframe1,
		config.Timeframe2,
		config.Verbose,
	)
	fmt.Printf("✅ Data fetcher initialized (%s, %s/%s)\n", 
		config.Symbol, config.Timeframe1, config.Timeframe2)
	
	// Initialize temporal engine
	engineConfig := engine.EngineConfig{
		WindowSize:    500,  // Keep 500 candles in memory
		AntiLookAhead: true, // Prevent look-ahead bias
		TrailingStop: engine.TrailingStopConfig{
			TrendPercent:        2.0, // 2% trailing stop for trend
			CounterTrendPercent: 3.0, // 3% trailing stop for counter-trend
		},
		AdjustmentGrid: []engine.AdjustmentLevel{
			{ProfitMin: 0.0, ProfitMax: 5.0, TrailingPercent: 3.0},
			{ProfitMin: 5.0, ProfitMax: 10.0, TrailingPercent: 2.0},
			{ProfitMin: 10.0, ProfitMax: 100.0, TrailingPercent: 1.0},
		},
		Zones: convertToEngineZones(),
	}
	
	temporalEngine, err := engine.NewTemporalEngineFromYAML(
		engine.BacktestMode, 
		yamlConfig, 
		engineConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize temporal engine: %w", err)
	}
	app.temporalEngine = temporalEngine
	
	fmt.Printf("✅ Temporal engine initialized (backtest mode)\n")
	
	return app, nil
}

// setupGracefulShutdown configures signal handling for clean shutdown
func (app *ApplicationState) setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Printf("\n🛑 Shutdown signal received, cleaning up...\n")
		app.shutdown()
	}()
}

// run executes the main application logic
func (app *ApplicationState) run() error {
	fmt.Printf("🚀 Starting Binance Backtest Live simulation...\n\n")
	
	app.running = true
	
	// Step 1: Fetch historical data from Binance
	if err := app.fetchHistoricalData(); err != nil {
		return fmt.Errorf("failed to fetch historical data: %w", err)
	}
	
	// Step 2: Start temporal engine
	if err := app.temporalEngine.Start(); err != nil {
		return fmt.Errorf("failed to start temporal engine: %w", err)
	}
	
	// Step 3: Run simulation loop
	if err := app.runSimulationLoop(); err != nil {
		return fmt.Errorf("simulation loop failed: %w", err)
	}
	
	return nil
}

// fetchHistoricalData retrieves data from Binance API
func (app *ApplicationState) fetchHistoricalData() error {
	// Use the data fetcher to retrieve historical data
	dataset, err := app.dataFetcher.FetchHistoricalData(app.ctx, app.config.DaysBack)
	if err != nil {
		return fmt.Errorf("failed to fetch historical data: %w", err)
	}
	
	app.dataset = dataset
	
	// Initialize simulation engine
	app.simEngine = NewSimulationEngine(
		app.temporalEngine,
		app.dataset,
		app.config,
	)
	
	// Initialize simulation engine
	if err := app.simEngine.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize simulation engine: %w", err)
	}
	
	fmt.Printf("✅ Simulation engine initialized and ready\n")
	
	return nil
}

// runSimulationLoop executes the main simulation logic
func (app *ApplicationState) runSimulationLoop() error {
	// Set up keyboard interrupt handling for interactive control
	go app.handleKeyboardInterrupt()
	
	// Run the simulation engine
	if err := app.simEngine.Run(app.ctx); err != nil {
		return fmt.Errorf("simulation engine error: %w", err)
	}
	
	return nil
}

// handleKeyboardInterrupt handles user input for simulation control
func (app *ApplicationState) handleKeyboardInterrupt() {
	fmt.Printf("\n💡 SIMULATION CONTROLS:\n")
	fmt.Printf("   Press Ctrl+C to stop simulation\n")
	fmt.Printf("   Simulation running in background...\n\n")
	
	// Note: In a full implementation, we could add:
	// - 'p' to pause/resume
	// - 's' to show statistics
	// - 'q' to quit gracefully
	// - '+'/'-' to adjust speed
	// For now, we rely on the graceful shutdown signal handling
}

// shutdown performs cleanup and stops all components
func (app *ApplicationState) shutdown() {
	app.running = false
	
	if app.simEngine != nil && app.simEngine.IsRunning() {
		app.simEngine.Stop()
		fmt.Printf("✅ Simulation engine stopped\n")
	}
	
	if app.temporalEngine != nil && app.temporalEngine.IsRunning() {
		app.temporalEngine.Stop()
		fmt.Printf("✅ Temporal engine stopped\n")
	}
	
	app.cancel()
}

// cleanup releases resources
func (app *ApplicationState) cleanup() {
	app.shutdown()
}

// convertToEngineZones converts YAML zones to engine format
func convertToEngineZones() engine.ZoneConfig {
	// Return default zones configuration
	return engine.ZoneConfig{
		CCIInverse: engine.ZoneSettings{
			Enabled:         true,
			Monitoring:      "continuous",
			ProfitThreshold: 5.0,
		},
		MACDInverse: engine.ZoneSettings{
			Enabled:         true,
			Monitoring:      "event",
			ProfitThreshold: 3.0,
		},
		DICounter: engine.ZoneSettings{
			Enabled:         true,
			Monitoring:      "continuous",
			ProfitThreshold: 4.0,
		},
	}
}
