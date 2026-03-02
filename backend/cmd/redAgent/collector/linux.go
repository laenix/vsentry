package collector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/laenix/redAgent/ingest"
	"github.com/laenix/redAgent/storage"
)

// LinuxFileCollector collects logs from Linux files
type LinuxFileCollector struct {
	sources []SourceConfig
	store   *storage.Storage
	client  *ingest.Client
	positions map[string]int64 // track file position for tail -f
}

// SourceConfig defines a single log source
type SourceConfig struct {
	Type    string `yaml:"type"`     // syslog, nginx_access, ssh, etc
	Path    string `yaml:"path"`     // file path
	Format  string `yaml:"format"`   // log format
	Enabled bool   `yaml:"enabled"`  // whether to collect
}

// NewLinuxFileCollector creates a new Linux file collector
func NewLinuxFileCollector(sources []SourceConfig, store *storage.Storage, client *ingest.Client) *LinuxFileCollector {
	// Filter only enabled sources
	var enabled []SourceConfig
	for _, s := range sources {
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}
	
	return &LinuxFileCollector{
		sources: enabled,
		store:   store,
		client:  client,
		positions: make(map[string]int64),
	}
}

// Collect reads new lines from all configured files
func (c *LinuxFileCollector) Collect() ([]ingest.LogEntry, error) {
	hostname, _ := os.Hostname()
	var allLogs []ingest.LogEntry

	for _, source := range c.sources {
		// Skip if file doesn't exist
		if _, err := os.Stat(source.Path); os.IsNotExist(err) {
			continue
		}

		logs, err := c.collectFile(source, hostname)
		if err != nil {
			log.Printf("Warning: failed to collect %s: %v", source.Path, err)
			continue
		}
		allLogs = append(allLogs, logs...)
	}

	return allLogs, nil
}

// collectFile reads new lines from a single file
func (c *LinuxFileCollector) collectFile(source SourceConfig, hostname string) ([]ingest.LogEntry, error) {
	// Get previous position
	lastPos := c.positions[source.Path]

	// Open file
	file, err := os.Open(source.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// If file was rotated (size smaller), reset position
	if info.Size() < lastPos {
		lastPos = 0
	}

	// Seek to last position
	file.Seek(lastPos, 0)

	// Read new lines
	var logs []ingest.LogEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry := c.parseLine(source, line, hostname)
		logs = append(logs, entry)
	}

	// Update position
	if info.Size() > lastPos {
		c.positions[source.Path] = info.Size()
	}

	return logs, nil
}

// parseLine converts a log line to LogEntry based on format
func (c *LinuxFileCollector) parseLine(source SourceConfig, line, hostname string) ingest.LogEntry {
	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	entry := ingest.LogEntry{
		Time:    now,
		Host:    hostname,
		Source:  source.Type,
		Channel: source.Type,
		Message: line,
		Level:   "info",
	}

	// Parse based on format
	switch source.Format {
	case "syslog", "auth":
		entry = c.parseSyslog(line, hostname, source)
	case "nginx_access":
		entry = c.parseNginxAccess(line, hostname)
	case "nginx_error":
		entry = c.parseNginxError(line, hostname)
	case "ssh":
		entry = c.parseSSH(line, hostname)
	case "apache":
		entry = c.parseApache(line, hostname)
	}

	return entry
}

// parseSyslog parses standard syslog format
func (c *LinuxFileCollector) parseSyslog(line, hostname string, source SourceConfig) ingest.LogEntry {
	// Common syslog format: Jan  1 12:00:00 hostname process[pid]: message
	// Or: 2026-01-01T12:00:00Z hostname process[pid]: message
	
	re := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}|\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[^\s]*)\s+(\S+)\s+(\S+?)(?:\[(\d+)\])?:\s*(.*)$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 6 {
		return ingest.LogEntry{
			Time:    guessTime(matches[1]),
			Host:    hostname,
			Source:  source.Type,
			Channel: source.Type,
			Message: matches[5],
			Level:   detectLevel(matches[5]),
		}
	}

	return ingest.LogEntry{
		Time:    time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Host:    hostname,
		Source:  source.Type,
		Channel: source.Type,
		Message: line,
		Level:   "info",
	}
}

// parseNginxAccess parses Nginx access log
// Format: 127.0.0.1 - - [01/Jan/2026:12:00:00 +0000] "GET /path HTTP/1.1" 200 1234 "-" "Mozilla/5.0"
func (c *LinuxFileCollector) parseNginxAccess(line, hostname string) ingest.LogEntry {
	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+\[([^\]]+)\]\s+"(\S+)\s+(\S+)\s+\S+"\s+(\d+)\s+(\d+)`)
	matches := re.FindStringSubmatch(line)

	entry := ingest.LogEntry{
		Time:    time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Host:    hostname,
		Source:  "nginx_access",
		Channel: "access_log",
		Message: line,
		Level:   "info",
		Extra:   make(map[string]interface{}),
	}

	if len(matches) >= 8 {
		entry.Extra["ip"] = matches[1]
		entry.Extra["method"] = matches[5]
		entry.Extra["uri"] = matches[6]
		entry.Extra["status"] = matches[7]
		entry.Extra["bytes"] = matches[8]

		// Set level based on status code
		if matches[7] >= "500" {
			entry.Level = "error"
		} else if matches[7] >= "400" {
			entry.Level = "warning"
		}
	}

	return entry
}

// parseNginxError parses Nginx error log
func (c *LinuxFileCollector) parseNginxError(line, hostname string) ingest.LogEntry {
	// Format: 2026/01/01 12:00:00 [error] 12345#12345: *6789 connection closed
	re := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+\[(\w+)\].*:\s*(.*)$`)
	matches := re.FindStringSubmatch(line)

	level := "info"
	if len(matches) >= 3 {
		switch matches[1] {
		case "error", "crit", "alert", "emerg":
			level = "critical"
		case "warn":
			level = "warning"
		}
		return ingest.LogEntry{
			Time:    time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			Host:    hostname,
			Source:  "nginx_error",
			Channel: "error_log",
			Message: matches[2],
			Level:   level,
		}
	}

	return ingest.LogEntry{
		Time:    time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Host:    hostname,
		Source:  "nginx_error",
		Channel: "error_log",
		Message: line,
		Level:   "info",
	}
}

// parseSSH parses SSH login logs
func (c *LinuxFileCollector) parseSSH(line, hostname string) ingest.LogEntry {
	level := "info"

	// Detect failed login
	if strings.Contains(line, "Failed password") || strings.Contains(line, "authentication failure") {
		level = "warning"
	}

	// Detect successful login
	if strings.Contains(line, "Accepted password") || strings.Contains(line, "session opened") {
		level = "info"
	}

	// Detect break-in attempts
	if strings.Contains(line, "BREAK-IN") || strings.Contains(line, "Too many authentication failures") {
		level = "critical"
	}

	// Extract IP if present
	ipRe := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ipMatch := ipRe.FindStringSubmatch(line)

	entry := ingest.LogEntry{
		Time:    time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Host:    hostname,
		Source:  "ssh",
		Channel: "auth_log",
		Message: line,
		Level:   level,
	}

	if len(ipMatch) > 1 {
		entry.Extra = map[string]interface{}{"ip": ipMatch[1]}
	}

	return entry
}

// parseApache parses Apache access log
func (c *LinuxFileCollector) parseApache(line, hostname string) ingest.LogEntry {
	// Similar to nginx
	return c.parseNginxAccess(line, hostname)
}

// detectLevel tries to detect log level from message content
func detectLevel(msg string) string {
	msg = strings.ToLower(msg)
	if strings.Contains(msg, "error") || strings.Contains(msg, "fail") || strings.Contains(msg, "critical") {
		return "error"
	}
	if strings.Contains(msg, "warn") {
		return "warning"
	}
	return "info"
}

// guessTime converts syslog time format to ISO
func guessTime(timeStr string) string {
	// If already ISO format, return as-is
	if strings.Contains(timeStr, "T") {
		return timeStr
	}
	
	// Try to parse syslog format (e.g., "Jan  1 12:00:00")
	now := time.Now()
	parsed, err := time.Parse("Jan 2 15:04:05", timeStr)
	if err != nil {
		return now.Format("2006-01-02T15:04:05.000Z")
	}
	
	// Set to current year
	parsed = time.Date(now.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, time.UTC)
	
	// If in the future, use last year
	if parsed.After(now) {
		parsed = parsed.AddDate(-1, 0, 0)
	}
	
	return parsed.Format("2006-01-02T15:04:05.000Z")
}

// GetClient returns the HTTP client
func (c *LinuxFileCollector) GetClient() *ingest.Client {
	return c.client
}

// GetStorage returns the storage
func (c *LinuxFileCollector) GetStorage() *storage.Storage {
	return c.store
}