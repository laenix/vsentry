package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/laenix/redAgent/ingest"
	"github.com/laenix/redAgent/storage"
)

// Collector is the interface for all collectors
type Collector interface {
	Collect() ([]ingest.LogEntry, error)
	GetClient() *ingest.Client
	GetStorage() *storage.Storage
}

// WindowsEventCollector collects Windows Event Logs
type WindowsEventCollector struct {
	channels []string
	store    *storage.Storage
	client   *ingest.Client
	lastTime time.Time
}

// NewWindowsEventCollector creates a new Windows Event Log collector
func NewWindowsEventCollector(channels []string, store *storage.Storage, client *ingest.Client) *WindowsEventCollector {
	return &WindowsEventCollector{
		channels: channels,
		store:    store,
		client:   client,
		lastTime: time.Now().Add(-5 * time.Minute),
	}
}

// Collect collects logs from all configured channels
func (c *WindowsEventCollector) Collect() ([]ingest.LogEntry, error) {
	hostname, _ := os.Hostname()
	var allLogs []ingest.LogEntry

	for _, channel := range c.channels {
		logs, err := c.collectChannel(channel, hostname)
		if err != nil {
			log.Printf("Warning: failed to collect from %s: %v", channel, err)
			continue
		}
		allLogs = append(allLogs, logs...)
	}

	return allLogs, nil
}

// collectChannel collects from a specific Windows Event Log channel
func (c *WindowsEventCollector) collectChannel(channel, hostname string) ([]ingest.LogEntry, error) {
	// PowerShell script to query Windows Event Log
	psScript := fmt.Sprintf(`
$startTime = (Get-Date).AddMinutes(-5)
$events = Get-WinEvent -FilterHashtable @{
    LogName='%s'
    StartTime=$startTime
} -MaxEvents 30 -ErrorAction SilentlyContinue

if (-not $events) { Write-Output "[]"; exit }

$results = @()
foreach ($e in $events) {
    $levelMap = @{1='critical';2='error';3='warning';4='information';0='info'}
    $level = if ($levelMap[$e.Level]) { $levelMap[$e.Level] } else { 'info' }
    
    $msg = if ($e.Message) { $e.Message.Substring(0, [Math]::Min(3000, $e.Message.Length)) } else { "" }
    
    $results += @{
        _time = $e.TimeCreated.ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ss.fffZ')
        host = $env:COMPUTERNAME
        source = '%s'
        channel = '%s'
        message = $msg
        level = $level
        event_id = $e.Id
        provider = $e.ProviderName
    }
}
$results | ConvertTo-Json -Compress
`, channel, channel, channel)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, nil // No events or error
	}

	result := strings.TrimSpace(string(output))
	if result == "" || result == "[]" {
		return nil, nil
	}

	// Parse JSON
	var logs []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &logs); err != nil {
		// Try single object
		var single map[string]interface{}
		if err := json.Unmarshal([]byte(result), &single); err == nil {
			logs = []map[string]interface{}{single}
		} else {
			return nil, nil
		}
	}

	// Convert to LogEntry
	var entries []ingest.LogEntry
	for _, raw := range logs {
		entry := ingest.LogEntry{
			Time:    getString(raw, "_time"),
			Host:    getString(raw, "host"),
			Source:  getString(raw, "source"),
			Channel: getString(raw, "channel"),
			Message: getString(raw, "message"),
			Level:   getString(raw, "level"),
		}

		if v, ok := raw["event_id"].(float64); ok {
			entry.EventID = int(v)
		}
		if v, ok := raw["provider"].(string); ok {
			entry.Provider = v
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// SyslogCollector for Linux systems
type SyslogCollector struct {
	channels []string
	store    *storage.Storage
	client   *ingest.Client
}

// NewSyslogCollector creates a new syslog collector
func NewSyslogCollector(channels []string, store *storage.Storage, client *ingest.Client) *SyslogCollector {
	return &SyslogCollector{
		channels: channels,
		store:    store,
		client:   client,
	}
}

// Collect is not implemented for Linux yet
func (c *SyslogCollector) Collect() ([]ingest.LogEntry, error) {
	return nil, fmt.Errorf("Linux collector not implemented yet")
}

// GetClient returns the HTTP client
func (c *WindowsEventCollector) GetClient() *ingest.Client {
	return c.client
}

// GetStorage returns the storage
func (c *WindowsEventCollector) GetStorage() *storage.Storage {
	return c.store
}

// GetClient for SyslogCollector
func (c *SyslogCollector) GetClient() *ingest.Client {
	return c.client
}

// GetStorage for SyslogCollector
func (c *SyslogCollector) GetStorage() *storage.Storage {
	return c.store
}
// SourceConfigFromJSON parses JSON string to []SourceConfig
func SourceConfigFromJSON(jsonStr string) ([]SourceConfig, error) {
	if jsonStr == "" {
		return []SourceConfig{}, nil
	}
	var sources []SourceConfig
	if err := json.Unmarshal([]byte(jsonStr), &sources); err != nil {
		return nil, err
	}
	return sources, nil
}

// SourceConfigToJSON converts []SourceConfig to JSON string
func SourceConfigToJSON(sources []SourceConfig) string {
	data, _ := json.Marshal(sources)
	return string(data)
}
