package money_management

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditLogger handles compliance logging for Money Management
type AuditLogger struct {
	logDir  string
	mutex   sync.Mutex
	
	// Current log file
	currentLogFile *os.File
	currentDate    string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logDir string) (*AuditLogger, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	logger := &AuditLogger{
		logDir: logDir,
	}
	
	// Initialize current log file
	if err := logger.rotateLogFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize log file: %w", err)
	}
	
	return logger, nil
}

// LogCircuitBreaker logs circuit breaker events
func (al *AuditLogger) LogCircuitBreaker(eventType string, data map[string]interface{}, success bool, impact string) {
	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "CIRCUIT_BREAKER_" + eventType,
		EventData: data,
		Success:   success,
		Impact:    impact,
	}
	
	if !success && data["error"] != nil {
		entry.ErrorMsg = fmt.Sprintf("%v", data["error"])
	}
	
	al.writeLogEntry(entry)
}

// LogPositionSizing logs position sizing decisions
func (al *AuditLogger) LogPositionSizing(request PositionSizingRequest, result PositionSizingResult, success bool, impact string) {
	data := map[string]interface{}{
		"request": request,
		"result":  result,
	}
	
	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "POSITION_SIZING",
		EventData: data,
		Success:   success,
		Impact:    impact,
	}
	
	if !success {
		entry.ErrorMsg = result.ValidationError
	}
	
	al.writeLogEntry(entry)
}

// LogConfigChange logs configuration changes
func (al *AuditLogger) LogConfigChange(component string, oldConfig, newConfig interface{}, impact string) {
	data := map[string]interface{}{
		"component":   component,
		"old_config":  oldConfig,
		"new_config":  newConfig,
	}
	
	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "CONFIG_CHANGE",
		EventData: data,
		Success:   true,
		Impact:    impact,
	}
	
	al.writeLogEntry(entry)
}

// LogMetricsSnapshot logs periodic metrics snapshots
func (al *AuditLogger) LogMetricsSnapshot(metrics GlobalMetrics, impact string) {
	data := map[string]interface{}{
		"metrics": metrics,
	}
	
	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "METRICS_SNAPSHOT",
		EventData: data,
		Success:   true,
		Impact:    impact,
	}
	
	al.writeLogEntry(entry)
}

// LogEmergencyAction logs emergency actions (position closures, etc.)
func (al *AuditLogger) LogEmergencyAction(action string, data map[string]interface{}, success bool, impact string) {
	entry := AuditLogEntry{
		Timestamp: time.Now().UTC(),
		EventType: "EMERGENCY_" + action,
		EventData: data,
		Success:   success,
		Impact:    impact,
	}
	
	if !success && data["error"] != nil {
		entry.ErrorMsg = fmt.Sprintf("%v", data["error"])
	}
	
	al.writeLogEntry(entry)
}

// writeLogEntry writes a log entry to current log file
func (al *AuditLogger) writeLogEntry(entry AuditLogEntry) {
	al.mutex.Lock()
	defer al.mutex.Unlock()
	
	// Rotate log file if date changed
	currentDate := time.Now().UTC().Format("2006-01-02")
	if currentDate != al.currentDate {
		al.rotateLogFile()
	}
	
	// Serialize entry to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("ERROR: Failed to marshal audit log entry: %v\n", err)
		return
	}
	
	// Write to log file
	if al.currentLogFile != nil {
		jsonData = append(jsonData, '\n') // Add newline for JSONL format
		if _, err := al.currentLogFile.Write(jsonData); err != nil {
			fmt.Printf("ERROR: Failed to write audit log entry: %v\n", err)
		} else {
			// Ensure immediate write to disk
			al.currentLogFile.Sync()
		}
	}
}

// rotateLogFile creates new log file for current date
func (al *AuditLogger) rotateLogFile() error {
	// Close current file if open
	if al.currentLogFile != nil {
		al.currentLogFile.Close()
	}
	
	// Create new log file for current date
	currentDate := time.Now().UTC().Format("2006-01-02")
	filename := fmt.Sprintf("money_management_audit_%s.jsonl", currentDate)
	filepath := filepath.Join(al.logDir, filename)
	
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	
	al.currentLogFile = file
	al.currentDate = currentDate
	
	return nil
}

// Close closes the audit logger and its resources
func (al *AuditLogger) Close() error {
	al.mutex.Lock()
	defer al.mutex.Unlock()
	
	if al.currentLogFile != nil {
		return al.currentLogFile.Close()
	}
	
	return nil
}

// GetLogFileForDate returns path to log file for specific date
func (al *AuditLogger) GetLogFileForDate(date time.Time) string {
	dateStr := date.Format("2006-01-02")
	filename := fmt.Sprintf("money_management_audit_%s.jsonl", dateStr)
	return filepath.Join(al.logDir, filename)
}

// ReadAuditLogs reads audit logs for a specific date range
func (al *AuditLogger) ReadAuditLogs(startDate, endDate time.Time) ([]AuditLogEntry, error) {
	var entries []AuditLogEntry
	
	// Iterate through each day in range
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		logFile := al.GetLogFileForDate(currentDate)
		
		// Read log file if exists
		if data, err := os.ReadFile(logFile); err == nil {
			// Parse JSONL format (one JSON object per line)
			lines := splitLines(string(data))
			for _, line := range lines {
				if line == "" {
					continue
				}
				
				var entry AuditLogEntry
				if err := json.Unmarshal([]byte(line), &entry); err != nil {
					fmt.Printf("Warning: Failed to parse audit log line: %v\n", err)
					continue
				}
				
				entries = append(entries, entry)
			}
		}
		
		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}
	
	return entries, nil
}

// GenerateComplianceReport generates compliance report for audit
func (al *AuditLogger) GenerateComplianceReport(startDate, endDate time.Time) (*ComplianceReport, error) {
	entries, err := al.ReadAuditLogs(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit logs: %w", err)
	}
	
	report := &ComplianceReport{
		StartDate:           startDate,
		EndDate:             endDate,
		TotalEvents:         len(entries),
		CircuitBreakerEvents: 0,
		ConfigChanges:       0,
		EmergencyActions:    0,
		FailedEvents:        0,
		EventsByType:        make(map[string]int),
	}
	
	// Analyze entries
	for _, entry := range entries {
		report.EventsByType[entry.EventType]++
		
		if !entry.Success {
			report.FailedEvents++
		}
		
		switch {
		case entry.EventType[:15] == "CIRCUIT_BREAKER":
			report.CircuitBreakerEvents++
		case entry.EventType == "CONFIG_CHANGE":
			report.ConfigChanges++
		case entry.EventType[:9] == "EMERGENCY":
			report.EmergencyActions++
		}
	}
	
	return report, nil
}

// ComplianceReport represents audit compliance summary
type ComplianceReport struct {
	StartDate            time.Time         `json:"start_date"`
	EndDate              time.Time         `json:"end_date"`
	TotalEvents          int               `json:"total_events"`
	CircuitBreakerEvents int               `json:"circuit_breaker_events"`
	ConfigChanges        int               `json:"config_changes"`
	EmergencyActions     int               `json:"emergency_actions"`
	FailedEvents         int               `json:"failed_events"`
	EventsByType         map[string]int    `json:"events_by_type"`
	ComplianceScore      float64           `json:"compliance_score"`
}

// splitLines splits string by newlines (helper function)
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	
	var lines []string
	start := 0
	
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	
	// Add last line if doesn't end with newline
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	
	return lines
}
