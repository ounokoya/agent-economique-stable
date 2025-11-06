// Package cli provides the CLI application types and interfaces
package cli

import (
	"time"

	"agent-economique/internal/shared"
)

// ExecutionMode defines the mode of operation for the CLI application
type ExecutionMode int

const (
	ModeDefault ExecutionMode = iota
	ModeDownloadOnly
	ModeProcessingOnly
	ModeStreaming
	ModeBatch
)

// ExportFormat defines supported export formats
type ExportFormat int

const (
	FormatCSV ExportFormat = iota
	FormatJSON
	FormatParquet
)

// NetworkSimulation defines network condition simulation for testing
type NetworkSimulation int

const (
	NetworkSimulationNormal NetworkSimulation = iota
	NetworkSimulationUnstable
	NetworkSimulationSlow
)

// CLIApp represents the main CLI application structure
type CLIApp struct {
	config              *shared.Config
	mode                ExecutionMode
	args                *CLIArguments
	networkSimulation   NetworkSimulation
	diskSpaceLimit      int64 // For testing disk space limits
	shutdownRequested   bool
	checkpointManager   *CheckpointManager
}

// CLIArguments holds parsed command line arguments
type CLIArguments struct {
	ConfigPath        string
	Symbols           []string
	Timeframes        []string
	StartDate         string
	EndDate           string
	OutputPath        string
	Mode              string
	MemoryLimit       int
	Verbose           bool
	ForceRedownload   bool
	ExportFormat      string
	EnableMetrics     bool
}

// WorkflowResult contains the results of workflow execution
type WorkflowResult struct {
	Success                bool
	FilesDownloaded        int
	FilesProcessed         int
	TimeframesDownloaded   int
	CacheHits              int
	RetryAttempts          int
	GracefulShutdown       bool
	TempFilesRemaining     int
	Statistics             map[string]interface{}
	PerformanceMetrics     *PerformanceMetrics
	SymbolsProcessed       []string
	ExecutionTime          time.Duration
	Errors                 []error
	Warnings               []string
}

// PerformanceMetrics contains performance metrics for the execution
type PerformanceMetrics struct {
	ProcessingRate         float64 // Records per second
	MemoryUsageMB          float64
	PeakMemoryUsageMB      float64
	CPUUsagePercent        float64
	DiskIOReadMB           float64
	DiskIOWriteMB          float64
	NetworkDownloadMB      float64
	NetworkUploadMB        float64
	CacheHitRatio          float64
	AverageResponseTimeMs  float64
	ErrorRate              float64
}

// FinalReport contains the final execution report
type FinalReport struct {
	Summary            string
	SymbolsProcessed   []string
	TimeframesGenerated []string
	ExecutionTime      time.Duration
	TotalDataVolume    int64
	SuccessRate        float64
	PerformanceMetrics *PerformanceMetrics
	Recommendations    []string
	Errors             []error
	Warnings           []string
	GeneratedAt        time.Time
}

// CheckpointManager handles workflow checkpoints for recovery
type CheckpointManager struct {
	checkpointPath string
	lastCheckpoint *WorkflowCheckpoint
}

// WorkflowCheckpoint represents a point in workflow execution
type WorkflowCheckpoint struct {
	Timestamp           time.Time
	CompletedSymbols    []string
	CompletedTimeframes []string
	ProcessedFiles      []string
	CurrentStage        WorkflowStage
	Statistics          map[string]interface{}
}

// WorkflowStage represents different stages of workflow execution
type WorkflowStage int

const (
	StageInitialization WorkflowStage = iota
	StageDownload
	StageParsing
	StageStatistics
	StageExport
	StageCompleted
)

// ValidationResult contains validation results for configuration and parameters
type ValidationResult struct {
	IsValid   bool
	Errors    []string
	Warnings  []string
	Suggestions []string
}

// String methods for enums
func (m ExecutionMode) String() string {
	switch m {
	case ModeDefault:
		return "default"
	case ModeDownloadOnly:
		return "download-only"
	case ModeProcessingOnly:
		return "processing-only"
	case ModeStreaming:
		return "streaming"
	case ModeBatch:
		return "batch"
	default:
		return "unknown"
	}
}

func (f ExportFormat) String() string {
	switch f {
	case FormatCSV:
		return "csv"
	case FormatJSON:
		return "json"
	case FormatParquet:
		return "parquet"
	default:
		return "unknown"
	}
}

func (s WorkflowStage) String() string {
	switch s {
	case StageInitialization:
		return "initialization"
	case StageDownload:
		return "download"
	case StageParsing:
		return "parsing"
	case StageStatistics:
		return "statistics"
	case StageExport:
		return "export"
	case StageCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

// ParseExecutionMode parses a string into ExecutionMode
func ParseExecutionMode(mode string) ExecutionMode {
	switch mode {
	case "download-only":
		return ModeDownloadOnly
	case "processing-only":
		return ModeProcessingOnly
	case "streaming":
		return ModeStreaming
	case "batch":
		return ModeBatch
	default:
		return ModeDefault
	}
}

// ParseExportFormat parses a string into ExportFormat
func ParseExportFormat(format string) ExportFormat {
	switch format {
	case "csv":
		return FormatCSV
	case "json":
		return FormatJSON
	case "parquet":
		return FormatParquet
	default:
		return FormatCSV // Default to CSV
	}
}

// NewWorkflowResult creates a new WorkflowResult with default values
func NewWorkflowResult() *WorkflowResult {
	return &WorkflowResult{
		Success:              false,
		FilesDownloaded:      0,
		FilesProcessed:       0,
		TimeframesDownloaded: 0,
		CacheHits:            0,
		RetryAttempts:        0,
		GracefulShutdown:     false,
		TempFilesRemaining:   0,
		Statistics:           make(map[string]interface{}),
		PerformanceMetrics:   &PerformanceMetrics{},
		SymbolsProcessed:     make([]string, 0),
		ExecutionTime:        0,
		Errors:               make([]error, 0),
		Warnings:             make([]string, 0),
	}
}

// NewPerformanceMetrics creates a new PerformanceMetrics with default values
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		ProcessingRate:        0,
		MemoryUsageMB:         0,
		PeakMemoryUsageMB:     0,
		CPUUsagePercent:       0,
		DiskIOReadMB:          0,
		DiskIOWriteMB:         0,
		NetworkDownloadMB:     0,
		NetworkUploadMB:       0,
		CacheHitRatio:         0,
		AverageResponseTimeMs: 0,
		ErrorRate:             0,
	}
}

// AddError adds an error to the workflow result
func (wr *WorkflowResult) AddError(err error) {
	if err != nil {
		wr.Errors = append(wr.Errors, err)
		wr.Success = false
	}
}

// AddWarning adds a warning to the workflow result
func (wr *WorkflowResult) AddWarning(warning string) {
	if warning != "" {
		wr.Warnings = append(wr.Warnings, warning)
	}
}

// HasErrors returns true if there are any errors
func (wr *WorkflowResult) HasErrors() bool {
	return len(wr.Errors) > 0
}

// HasWarnings returns true if there are any warnings
func (wr *WorkflowResult) HasWarnings() bool {
	return len(wr.Warnings) > 0
}

// GetErrorCount returns the total number of errors
func (wr *WorkflowResult) GetErrorCount() int {
	return len(wr.Errors)
}

// GetWarningCount returns the total number of warnings
func (wr *WorkflowResult) GetWarningCount() int {
	return len(wr.Warnings)
}

// CalculateSuccessRate calculates the success rate based on processed vs total files
func (wr *WorkflowResult) CalculateSuccessRate() float64 {
	if wr.FilesDownloaded == 0 {
		return 0.0
	}
	successfulFiles := wr.FilesProcessed
	return float64(successfulFiles) / float64(wr.FilesDownloaded) * 100.0
}
