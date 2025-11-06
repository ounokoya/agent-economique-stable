// Package tests provides tests for CLI application functionality - TDD Approach
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agent-economique/internal/cli"
)

// TestCLIConfigurationLoading tests configuration loading and validation
func TestCLIConfigurationLoading(t *testing.T) {
	t.Run("ValidConfigurationFile", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createValidTestConfig(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Expected valid config to load successfully, got: %v", err)
		}

		config := app.GetConfiguration()
		if config == nil {
			t.Fatal("Expected configuration to be loaded, got nil")
		}

		// Verify key configuration values
		if len(config.BinanceData.Symbols) == 0 {
			t.Error("Expected symbols to be loaded from config")
		}

		if len(config.BinanceData.Timeframes) == 0 {
			t.Error("Expected timeframes to be loaded from config")
		}

		t.Logf("✅ Valid configuration loaded successfully")
	})

	t.Run("InvalidConfigurationFile", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createInvalidTestConfig(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err == nil {
			t.Fatal("Expected invalid config to fail loading, got nil error")
		}

		// Error message should be informative
		if !strings.Contains(err.Error(), "validation") {
			t.Errorf("Expected validation error message, got: %v", err)
		}

		t.Logf("✅ Invalid configuration properly rejected")
	})

	t.Run("MissingConfigurationFile", func(t *testing.T) {
		app := cli.NewCLIApp()
		err := app.LoadConfiguration("non_existent_config.yaml")
		if err == nil {
			t.Fatal("Expected missing config file to fail, got nil error")
		}

		t.Logf("✅ Missing configuration file properly handled")
	})
}

// TestCLIParameterValidation tests command line parameter validation
func TestCLIParameterValidation(t *testing.T) {
	t.Run("ValidCommandLineArguments", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createValidTestConfig(t, tempDir)

		app := cli.NewCLIApp()
		args := []string{
			"agent-economique",
			"--config", configPath,
			"--symbols", "SOLUSDT,ETHUSDT",
			"--timeframes", "5m,15m",
		}

		err := app.ParseArguments(args)
		if err != nil {
			t.Fatalf("Expected valid arguments to be accepted, got: %v", err)
		}

		t.Logf("✅ Valid command line arguments accepted")
	})

	t.Run("InvalidSymbolFormat", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createValidTestConfig(t, tempDir)

		app := cli.NewCLIApp()
		args := []string{
			"agent-economique",
			"--config", configPath,
			"--symbols", "invalid-symbol,ETHUSDT",
		}

		err := app.ParseArguments(args)
		if err == nil {
			t.Fatal("Expected invalid symbol format to be rejected, got nil error")
		}

		t.Logf("✅ Invalid symbol format properly rejected")
	})

	t.Run("UnsupportedTimeframe", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createValidTestConfig(t, tempDir)

		app := cli.NewCLIApp()
		args := []string{
			"agent-economique",
			"--config", configPath,
			"--timeframes", "5m,2h", // 2h not supported
		}

		err := app.ParseArguments(args)
		if err == nil {
			t.Fatal("Expected unsupported timeframe to be rejected, got nil error")
		}

		t.Logf("✅ Unsupported timeframe properly rejected")
	})
}

// TestCLIDownloadFlow tests the complete download workflow
func TestCLIDownloadFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping download flow test in short mode")
	}

	t.Run("DownloadOnlyMode", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigWithCache(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Set download-only mode
		err = app.SetMode(cli.ModeDownloadOnly)
		if err != nil {
			t.Fatalf("Failed to set download-only mode: %v", err)
		}

		// Execute download workflow
		result, err := app.ExecuteWorkflow()
		if err != nil {
			t.Fatalf("Download workflow failed: %v", err)
		}

		if result == nil {
			t.Fatal("Expected workflow result, got nil")
		}

		// Verify files were downloaded
		if result.FilesDownloaded == 0 {
			t.Error("Expected some files to be downloaded")
		}

		// Verify cache was updated
		if result.CacheHits < 0 {
			t.Error("Cache hits should not be negative")
		}

		t.Logf("✅ Download-only workflow completed - downloaded %d files", result.FilesDownloaded)
	})

	t.Run("FullWorkflowMode", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigWithCache(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Execute full workflow (download + processing + aggregation)
		result, err := app.ExecuteWorkflow()
		if err != nil {
			t.Fatalf("Full workflow failed: %v", err)
		}

		// Verify all stages completed
		if result.FilesProcessed == 0 {
			t.Error("Expected some files to be processed")
		}

		if result.TimeframesDownloaded == 0 {
			t.Error("Expected some timeframes to be downloaded")
		}

		if len(result.Statistics) == 0 {
			t.Error("Expected statistics to be calculated")
		}

		t.Logf("✅ Full workflow completed successfully")
	})
}

// TestCLIStreamingMode tests streaming mode functionality
func TestCLIStreamingMode(t *testing.T) {
	t.Run("MemoryConstraintValidation", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigWithMemoryLimit(t, tempDir, 50) // 50MB limit

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		err = app.SetMode(cli.ModeStreaming)
		if err != nil {
			t.Fatalf("Failed to set streaming mode: %v", err)
		}

		// Validate memory constraints before execution
		isValid, err := app.ValidateMemoryConstraints()
		if err != nil {
			t.Fatalf("Memory constraint validation failed: %v", err)
		}

		if !isValid {
			t.Log("Memory constraints validation returned false - may be expected depending on system")
		}

		t.Logf("✅ Memory constraint validation completed")
	})

	t.Run("StreamingPerformanceMetrics", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigWithMetrics(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		err = app.SetMode(cli.ModeStreaming)
		if err != nil {
			t.Fatalf("Failed to set streaming mode: %v", err)
		}

		// Execute with metrics enabled
		result, err := app.ExecuteWorkflow()
		if err != nil {
			t.Fatalf("Streaming workflow failed: %v", err)
		}

		// Verify performance metrics were collected
		metrics := result.PerformanceMetrics
		if metrics == nil {
			t.Fatal("Expected performance metrics, got nil")
		}

		if metrics.ProcessingRate <= 0 {
			t.Error("Processing rate should be positive")
		}

		if metrics.MemoryUsageMB < 0 {
			t.Error("Memory usage should not be negative")
		}

		t.Logf("✅ Streaming mode with metrics completed")
	})
}

// TestCLIErrorHandling tests various error conditions
func TestCLIErrorHandling(t *testing.T) {
	t.Run("NetworkErrorRecovery", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigWithRetry(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Simulate network error scenario
		app.SetNetworkSimulation(cli.NetworkSimulationUnstable)

		result, err := app.ExecuteWorkflow()
		// Should succeed with retries or fail gracefully
		if err != nil {
			// Verify error is informative
			if !strings.Contains(err.Error(), "network") && !strings.Contains(err.Error(), "retry") {
				t.Errorf("Expected network-related error message, got: %v", err)
			}
			t.Logf("Network error handled gracefully: %v", err)
		} else {
			// If successful, verify retry attempts were made
			if result.RetryAttempts == 0 {
				t.Error("Expected retry attempts in unstable network simulation")
			}
			t.Logf("✅ Network error recovery succeeded with %d retries", result.RetryAttempts)
		}
	})

	t.Run("InsufficientDiskSpace", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigLargeData(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Simulate limited disk space
		app.SetDiskSpaceLimit(1) // 1MB limit

		result, err := app.ExecuteWorkflow()
		if err == nil {
			t.Error("Expected disk space error, got nil")
		}

		// Verify proper cleanup occurred
		if result != nil && result.TempFilesRemaining > 0 {
			t.Error("Expected temp files to be cleaned up on disk space error")
		}

		t.Logf("✅ Disk space error handled properly")
	})

	t.Run("GracefulShutdown", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigWithCheckpoints(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Start workflow in goroutine
		resultChan := make(chan *cli.WorkflowResult, 1)
		errorChan := make(chan error, 1)

		go func() {
			result, err := app.ExecuteWorkflow()
			if err != nil {
				errorChan <- err
			} else {
				resultChan <- result
			}
		}()

		// Allow some processing time
		time.Sleep(100 * time.Millisecond)

		// Send graceful shutdown signal
		err = app.GracefulShutdown()
		if err != nil {
			t.Fatalf("Graceful shutdown failed: %v", err)
		}

		// Wait for completion
		select {
		case result := <-resultChan:
			if !result.GracefulShutdown {
				t.Error("Expected graceful shutdown flag to be set")
			}
			t.Logf("✅ Graceful shutdown completed")
		case err := <-errorChan:
			// Shutdown during processing is acceptable
			t.Logf("✅ Graceful shutdown during processing: %v", err)
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for graceful shutdown")
		}
	})
}

// TestCLIReportGeneration tests report and output generation
func TestCLIReportGeneration(t *testing.T) {
	t.Run("FinalReport", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigWithReports(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		result, err := app.ExecuteWorkflow()
		if err != nil {
			t.Fatalf("Workflow failed: %v", err)
		}

		// Generate final report
		report, err := app.GenerateFinalReport(result)
		if err != nil {
			t.Fatalf("Report generation failed: %v", err)
		}

		if report == nil {
			t.Fatal("Expected report, got nil")
		}

		// Verify report contents
		if report.Summary == "" {
			t.Error("Expected report summary")
		}

		if len(report.SymbolsProcessed) == 0 {
			t.Error("Expected symbols processed in report")
		}

		if report.ExecutionTime <= 0 {
			t.Error("Expected positive execution time")
		}

		t.Logf("✅ Final report generated successfully")
	})

	t.Run("ExportFormats", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := createTestConfigWithExports(t, tempDir)

		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		result, err := app.ExecuteWorkflow()
		if err != nil {
			t.Fatalf("Workflow failed: %v", err)
		}

		// Test CSV export
		csvPath := filepath.Join(tempDir, "export.csv")
		err = app.ExportData(result, cli.FormatCSV, csvPath)
		if err != nil {
			t.Fatalf("CSV export failed: %v", err)
		}

		if !fileExists(csvPath) {
			t.Error("Expected CSV file to be created")
		}

		// Test JSON export
		jsonPath := filepath.Join(tempDir, "export.json")
		err = app.ExportData(result, cli.FormatJSON, jsonPath)
		if err != nil {
			t.Fatalf("JSON export failed: %v", err)
		}

		if !fileExists(jsonPath) {
			t.Error("Expected JSON file to be created")
		}

		t.Logf("✅ Data export in multiple formats completed")
	})
}

// Helper functions for creating test configurations

func createValidTestConfig(t *testing.T, tempDir string) string {
	configContent := `
binance_data:
  symbols: ["SOLUSDT", "ETHUSDT"]
  timeframes: ["5m", "15m", "1h"]
  cache_root: "` + tempDir + `/cache"
  downloader:
    base_url: "https://data.binance.vision"
    max_retries: 2
    timeout: "30s"
  streaming:
    buffer_size: 1024
    max_memory_mb: 100
    enable_metrics: true

data_period:
  start_date: "2023-06-01"
  end_date: "2023-06-02"

strategy:
  name: "test-strategy"

monitoring:
  enable_metrics: true
  log_level: "info"
`

	configPath := filepath.Join(tempDir, "test_config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	return configPath
}

func createInvalidTestConfig(t *testing.T, tempDir string) string {
	configContent := `
binance_data:
  symbols: []  # Invalid: empty symbols
  timeframes: ["invalid-timeframe"]  # Invalid timeframe
  cache_root: "` + tempDir + `/cache"

data_period:
  start_date: "invalid-date"  # Invalid date format
  end_date: "2023-06-02"
`

	configPath := filepath.Join(tempDir, "invalid_config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid test config: %v", err)
	}

	return configPath
}

func createTestConfigWithCache(t *testing.T, tempDir string) string {
	config := createValidTestConfig(t, tempDir)
	// Cache config already included in valid config
	return config
}

func createTestConfigWithMemoryLimit(t *testing.T, tempDir string, limitMB int) string {
	configContent := `
app:
  name: "agent-economique-test"
  log_level: "info"

data:
  symbols: ["SOLUSDT"]
  timeframes: ["5m"]
  date_range:
    start: "2023-06-01"
    end: "2023-06-01"
  types: ["klines"]

download:
  base_url: "https://data.binance.vision"
  max_concurrent: 1

cache:
  root_path: "` + tempDir + `/cache"

performance:
  max_memory_mb: ` + string(rune(limitMB)) + `
  enable_metrics: true
`

	configPath := filepath.Join(tempDir, "memory_config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create memory config: %v", err)
	}

	return configPath
}

func createTestConfigWithMetrics(t *testing.T, tempDir string) string {
	return createValidTestConfig(t, tempDir) // Metrics already enabled
}

func createTestConfigWithRetry(t *testing.T, tempDir string) string {
	return createValidTestConfig(t, tempDir) // Retry already configured
}

func createTestConfigLargeData(t *testing.T, tempDir string) string {
	configContent := `
app:
  name: "agent-economique-test"

data:
  symbols: ["SOLUSDT", "ETHUSDT", "BTCUSDT"]  # Multiple symbols for larger data
  timeframes: ["5m", "15m", "1h", "4h"]
  date_range:
    start: "2023-06-01"
    end: "2023-06-30"  # Full month
  types: ["klines", "trades"]

download:
  base_url: "https://data.binance.vision"

cache:
  root_path: "` + tempDir + `/cache"
`

	configPath := filepath.Join(tempDir, "large_config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create large data config: %v", err)
	}

	return configPath
}

func createTestConfigWithCheckpoints(t *testing.T, tempDir string) string {
	return createValidTestConfig(t, tempDir) // Basic config supports checkpoints
}

func createTestConfigWithReports(t *testing.T, tempDir string) string {
	return createValidTestConfig(t, tempDir) // Basic config supports reports
}

func createTestConfigWithExports(t *testing.T, tempDir string) string {
	return createValidTestConfig(t, tempDir) // Basic config supports exports
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// createValidTestConfigForBench creates a valid test config for benchmarks
func createValidTestConfigForBench(tempDir string) string {
	configContent := `
app:
  name: "agent-economique-test"
  version: "1.0.0"
  log_level: "info"

binance_data:
  symbols: ["SOLUSDT", "ETHUSDT"]
  timeframes: ["5m", "15m", "1h"]
  cache_root: "` + tempDir + `/cache"
  downloader:
    base_url: "https://data.binance.vision"
    max_retries: 2
    timeout: "30s"
  streaming:
    buffer_size: 1024
    max_memory_mb: 100
    enable_metrics: true

data_period:
  start_date: "2023-06-01"
  end_date: "2023-06-02"

strategy:
  name: "test-strategy"

monitoring:
  enable_metrics: true
  log_level: "info"
`

	configPath := filepath.Join(tempDir, "test_config.yaml")
	os.WriteFile(configPath, []byte(configContent), 0644)
	return configPath
}

// Benchmark tests for CLI performance
func BenchmarkCLIWorkflow(b *testing.B) {
	tempDir := b.TempDir()
	configPath := createValidTestConfigForBench(tempDir)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		app := cli.NewCLIApp()
		err := app.LoadConfiguration(configPath)
		if err != nil {
			b.Fatalf("Config loading failed: %v", err)
		}

		_, err = app.ExecuteWorkflow()
		if err != nil {
			b.Fatalf("Workflow execution failed: %v", err)
		}
	}
}
