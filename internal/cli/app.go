// Package cli - CLI Application Implementation
package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"agent-economique/internal/datasource/binance"
	"agent-economique/internal/shared"
)

// NewCLIApp creates a new CLI application instance
func NewCLIApp() *CLIApp {
	return &CLIApp{
		config:              nil,
		mode:                ModeDefault,
		args:                &CLIArguments{},
		networkSimulation:   NetworkSimulationNormal,
		diskSpaceLimit:      -1, // No limit by default
		shutdownRequested:   false,
		checkpointManager:   nil,
	}
}

// LoadConfiguration loads and validates the configuration file
func (app *CLIApp) LoadConfiguration(configPath string) error {
	if configPath == "" {
		return fmt.Errorf("configuration file path cannot be empty")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Load configuration
	config, err := shared.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	validation := app.validateConfiguration(config)
	if !validation.IsValid {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(validation.Errors, ", "))
	}

	app.config = config
	return nil
}

// GetConfiguration returns the loaded configuration
func (app *CLIApp) GetConfiguration() *shared.Config {
	return app.config
}

// ParseArguments parses command line arguments
func (app *CLIApp) ParseArguments(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: %s --config <config.yaml> [options]", args[0])
	}

	app.args = &CLIArguments{}

	// Simple argument parsing
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--config":
			if i+1 >= len(args) {
				return fmt.Errorf("--config requires a value")
			}
			app.args.ConfigPath = args[i+1]
			i++
		case "--symbols":
			if i+1 >= len(args) {
				return fmt.Errorf("--symbols requires a value")
			}
			symbols := strings.Split(args[i+1], ",")
			for _, symbol := range symbols {
				if !app.isValidSymbol(strings.TrimSpace(symbol)) {
					return fmt.Errorf("invalid symbol format: %s", symbol)
				}
			}
			app.args.Symbols = symbols
			i++
		case "--timeframes":
			if i+1 >= len(args) {
				return fmt.Errorf("--timeframes requires a value")
			}
			timeframes := strings.Split(args[i+1], ",")
			// Store timeframes for later validation after config is loaded
			app.args.Timeframes = timeframes
			i++
		case "--mode":
			if i+1 >= len(args) {
				return fmt.Errorf("--mode requires a value")
			}
			app.args.Mode = args[i+1]
			i++
		case "--memory-limit":
			if i+1 >= len(args) {
				return fmt.Errorf("--memory-limit requires a value")
			}
			var memLimit int
			if _, err := fmt.Sscanf(args[i+1], "%d", &memLimit); err != nil {
				return fmt.Errorf("invalid memory limit: %s", args[i+1])
			}
			app.args.MemoryLimit = memLimit
			i++
		case "--force-redownload":
			app.args.ForceRedownload = true
		case "--verbose":
			app.args.Verbose = true
		case "--enable-metrics":
			app.args.EnableMetrics = true
		default:
			return fmt.Errorf("unknown argument: %s", args[i])
		}
	}

	// Load configuration if provided
	if app.args.ConfigPath != "" {
		err := app.LoadConfiguration(app.args.ConfigPath)
		if err != nil {
			return err
		}
		
		// Apply CLI configuration from YAML if not overridden by arguments
		app.applyCLIConfigFromYAML()
		
		// Validate timeframes after config is loaded
		if len(app.args.Timeframes) > 0 {
			for _, tf := range app.args.Timeframes {
				if !app.isValidTimeframe(strings.TrimSpace(tf)) {
					return fmt.Errorf("unsupported timeframe: %s", tf)
				}
			}
		}
	}

	return nil
}

// SetMode sets the execution mode
func (app *CLIApp) SetMode(mode ExecutionMode) error {
	app.mode = mode
	return nil
}

// ExecuteWorkflow executes the main workflow
func (app *CLIApp) ExecuteWorkflow() (*WorkflowResult, error) {
	if app.config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	result := NewWorkflowResult()
	startTime := time.Now()

	// Initialize components
	components, err := app.initializeComponents()
	if err != nil {
		result.AddError(err)
		return result, err
	}

	// Execute workflow stages
	switch app.mode {
	case ModeDownloadOnly:
		err = app.executeDownloadStage(components, result)
	case ModeStreaming:
		err = app.executeStreamingMode(components, result)
	default:
		err = app.executeFullWorkflow(components, result)
	}

	result.ExecutionTime = time.Since(startTime)

	if err != nil {
		result.AddError(err)
		result.Success = false
		return result, err
	}

	result.Success = true
	return result, nil
}

// ValidateMemoryConstraints validates memory constraints for streaming mode
func (app *CLIApp) ValidateMemoryConstraints() (bool, error) {
	if app.config == nil {
		return false, fmt.Errorf("configuration not loaded")
	}

	// Create temporary streaming reader to validate constraints
	tempDir := os.TempDir()
	cache, err := binance.InitializeCache(tempDir)
	if err != nil {
		return false, fmt.Errorf("failed to initialize temporary cache: %w", err)
	}

	streamingConfig := app.config.BinanceData.Streaming.ToStreamingConfig()
	streaming, err := binance.NewStreamingReader(cache, streamingConfig)
	if err != nil {
		return false, fmt.Errorf("failed to create streaming reader: %w", err)
	}

	return streaming.ValidateMemoryConstraints()
}

// GracefulShutdown initiates graceful shutdown
func (app *CLIApp) GracefulShutdown() error {
	app.shutdownRequested = true
	return nil
}

// GenerateFinalReport generates the final execution report
func (app *CLIApp) GenerateFinalReport(result *WorkflowResult) (*FinalReport, error) {
	if result == nil {
		return nil, fmt.Errorf("workflow result is nil")
	}

	report := &FinalReport{
		Summary:             app.generateSummary(result),
		SymbolsProcessed:    result.SymbolsProcessed,
		TimeframesGenerated: app.getProcessedTimeframes(),
		ExecutionTime:       result.ExecutionTime,
		TotalDataVolume:     app.calculateDataVolume(result),
		SuccessRate:         result.CalculateSuccessRate(),
		PerformanceMetrics:  result.PerformanceMetrics,
		Recommendations:     app.generateRecommendations(result),
		Errors:              result.Errors,
		Warnings:            result.Warnings,
		GeneratedAt:         time.Now(),
	}

	return report, nil
}

// SetNetworkSimulation sets network simulation for testing
func (app *CLIApp) SetNetworkSimulation(simulation NetworkSimulation) {
	app.networkSimulation = simulation
}

// SetDiskSpaceLimit sets disk space limit for testing
func (app *CLIApp) SetDiskSpaceLimit(limitMB int64) {
	app.diskSpaceLimit = limitMB
}

// validateConfiguration validates the loaded configuration
func (app *CLIApp) validateConfiguration(config *shared.Config) *ValidationResult {
	result := &ValidationResult{
		IsValid:     true,
		Errors:      make([]string, 0),
		Warnings:    make([]string, 0),
		Suggestions: make([]string, 0),
	}

	// Validate basic fields
	if len(config.BinanceData.Symbols) == 0 {
		result.Errors = append(result.Errors, "no symbols specified in configuration")
		result.IsValid = false
	}

	if len(config.BinanceData.Timeframes) == 0 {
		result.Errors = append(result.Errors, "no timeframes specified in configuration")
		result.IsValid = false
	}

	// Validate symbols
	for _, symbol := range config.BinanceData.Symbols {
		if !app.isValidSymbol(symbol) {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid symbol: %s", symbol))
			result.IsValid = false
		}
	}

	// Validate timeframes
	for _, tf := range config.BinanceData.Timeframes {
		if !app.isValidTimeframe(tf) {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid timeframe: %s", tf))
			result.IsValid = false
		}
	}

	// Validate date range
	if config.DataPeriod.StartDate == "" || config.DataPeriod.EndDate == "" {
		result.Errors = append(result.Errors, "start_date and end_date must be specified")
		result.IsValid = false
	}

	// Validate cache path
	if config.BinanceData.CacheRoot == "" {
		result.Warnings = append(result.Warnings, "cache_root not specified, using default")
		result.Suggestions = append(result.Suggestions, "Consider specifying a dedicated cache directory")
	}

	return result
}

// isValidSymbol validates symbol format
func (app *CLIApp) isValidSymbol(symbol string) bool {
	// Basic validation: should be uppercase, end with USDT
	if len(symbol) < 5 {
		return false
	}
	
	return strings.ToUpper(symbol) == symbol && 
		   (strings.HasSuffix(symbol, "USDT") || strings.HasSuffix(symbol, "BUSD"))
}

// isValidTimeframe validates timeframe format
func (app *CLIApp) isValidTimeframe(timeframe string) bool {
	// First check basic format
	basicValidTimeframes := []string{"1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d"}
	isBasicFormat := false
	for _, valid := range basicValidTimeframes {
		if timeframe == valid {
			isBasicFormat = true
			break
		}
	}
	
	if !isBasicFormat {
		return false
	}
	
	// If configuration is loaded, validate against supported timeframes
	if app.config != nil && len(app.config.BinanceData.Timeframes) > 0 {
		for _, configTf := range app.config.BinanceData.Timeframes {
			if timeframe == configTf {
				return true
			}
		}
		return false // Not in config's supported timeframes
	}
	
	// If no config loaded yet, accept basic format
	return true
}

// generateSummary generates a summary of the execution
func (app *CLIApp) generateSummary(result *WorkflowResult) string {
	if result.Success {
		return fmt.Sprintf("Successfully processed %d files for %d symbols in %v",
			result.FilesProcessed, len(result.SymbolsProcessed), result.ExecutionTime)
	}
	
	return fmt.Sprintf("Execution failed with %d errors after %v",
		len(result.Errors), result.ExecutionTime)
}

// getProcessedTimeframes returns list of processed timeframes
func (app *CLIApp) getProcessedTimeframes() []string {
	// Use command line args if provided, otherwise use config
	if len(app.args.Timeframes) > 0 {
		return app.args.Timeframes
	}
	
	if app.config == nil {
		return []string{}
	}
	
	return app.config.BinanceData.Timeframes
}

// calculateDataVolume calculates total data volume processed
func (app *CLIApp) calculateDataVolume(result *WorkflowResult) int64 {
	// Simplified calculation
	return int64(result.FilesProcessed * 1024 * 1024) // Assume 1MB per file average
}

// generateRecommendations generates recommendations based on results
func (app *CLIApp) generateRecommendations(result *WorkflowResult) []string {
	recommendations := make([]string, 0)
	
	if result.RetryAttempts > 10 {
		recommendations = append(recommendations, "High number of retries detected. Consider checking network connectivity.")
	}
	
	if result.PerformanceMetrics.MemoryUsageMB > 500 {
		recommendations = append(recommendations, "High memory usage detected. Consider reducing buffer sizes or using streaming mode.")
	}
	
	if len(result.Errors) > 0 {
		recommendations = append(recommendations, "Errors detected. Check logs for detailed error information.")
	}
	
	return recommendations
}

// applyCLIConfigFromYAML applies CLI configuration from YAML if not overridden by command line
func (app *CLIApp) applyCLIConfigFromYAML() {
	if app.config == nil {
		return
	}

	// Apply execution mode if not specified in command line
	if app.args.Mode == "" {
		app.mode = ParseExecutionMode(app.config.CLI.ExecutionMode)
	}

	// Apply memory limit if not specified in command line
	if app.args.MemoryLimit == 0 && app.config.CLI.MemoryLimitMB > 0 {
		app.args.MemoryLimit = app.config.CLI.MemoryLimitMB
	}

	// Apply force redownload if not specified in command line
	if !app.args.ForceRedownload && app.config.CLI.ForceRedownload {
		app.args.ForceRedownload = app.config.CLI.ForceRedownload
	}

	// Apply verbose if not specified in command line
	if !app.args.Verbose && app.config.CLI.Verbose {
		app.args.Verbose = app.config.CLI.Verbose
	}

	// Apply enable metrics if not specified in command line
	if !app.args.EnableMetrics && app.config.CLI.EnableMetrics {
		app.args.EnableMetrics = app.config.CLI.EnableMetrics
	}
}
