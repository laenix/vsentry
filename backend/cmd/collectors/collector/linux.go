//go:build linux

package collector

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/ingest"
)

type LinuxCollector struct {
	cfg       config.AgentConfig
	positions map[string]int64
}

func NewLinuxCollector(cfg config.AgentConfig) *LinuxCollector {
	return &LinuxCollector{
		cfg:       cfg,
		positions: make(map[string]int64),
	}
}

func (c *LinuxCollector) Collect() ([]ingest.LogEntry, error) {
	var allLogs []ingest.LogEntry

	for _, source := range c.cfg.Sources {
		if !source.Enabled {
			continue
		}

		logs, err := c.tailFile(source)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, logs...)
	}

	return allLogs, nil
}

func (c *LinuxCollector) tailFile(source config.SourceConfig) ([]ingest.LogEntry, error) {
	file, err := os.Open(source.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	lastPos := c.positions[source.Path]
	if info.Size() < lastPos {
		lastPos = 0 // 文件被截断或轮转
	}

	_, err = file.Seek(lastPos, 0)
	if err != nil {
		return nil, err
	}

	var logs []ingest.LogEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// 根据类型进行精准解析
		entry := c.parseLine(source, line)
		logs = append(logs, entry)
	}

	if info.Size() > lastPos {
		c.positions[source.Path] = info.Size()
	}

	return logs, nil
}

// parseLine 根据数据源类型路由到对应的解析器
func (c *LinuxCollector) parseLine(source config.SourceConfig, line string) ingest.LogEntry {
	// 默认基础结构
	entry := ingest.LogEntry{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Host:    c.cfg.Hostname,
		Source:  source.Type,
		Channel: source.Path,
		Message: line,
		Level:   "info",
	}

	switch source.Type {
	case "syslog", "kern", "messages":
		return c.parseSyslog(entry, line)
	case "auth", "secure":
		return c.parseSSH(entry, line)
	case "nginx_access", "apache_access":
		return c.parseWebAccess(entry, line)
	case "nginx_error":
		return c.parseWebError(entry, line)
	}

	return entry
}

func (c *LinuxCollector) parseSyslog(entry ingest.LogEntry, line string) ingest.LogEntry {
	re := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}|\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[^\s]*)\s+(\S+)\s+(\S+?)(?:\[(\d+)\])?:\s*(.*)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 6 {
		entry.Time = c.guessTime(matches[1])
		entry.Message = matches[5]
		entry.Level = c.detectLevel(matches[5])
	}
	return entry
}

func (c *LinuxCollector) parseWebAccess(entry ingest.LogEntry, line string) ingest.LogEntry {
	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+\[([^\]]+)\]\s+"(\S+)\s+(\S+)\s+\S+"\s+(\d+)\s+(\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 8 {
		entry.Extra = map[string]interface{}{
			"ip":     matches[1],
			"method": matches[5],
			"uri":    matches[6],
			"status": matches[7],
			"bytes":  matches[8],
		}
		if matches[7] >= "500" {
			entry.Level = "error"
		} else if matches[7] >= "400" {
			entry.Level = "warning"
		}
	}
	return entry
}

func (c *LinuxCollector) parseWebError(entry ingest.LogEntry, line string) ingest.LogEntry {
	re := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+\[(\w+)\].*:\s*(.*)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		entry.Message = matches[2]
		switch matches[1] {
		case "error", "crit", "alert", "emerg":
			entry.Level = "critical"
		case "warn":
			entry.Level = "warning"
		}
	}
	return entry
}

func (c *LinuxCollector) parseSSH(entry ingest.LogEntry, line string) ingest.LogEntry {
	if strings.Contains(line, "Failed password") || strings.Contains(line, "authentication failure") {
		entry.Level = "warning"
	}
	if strings.Contains(line, "BREAK-IN") || strings.Contains(line, "Too many authentication failures") {
		entry.Level = "critical"
	}
	ipRe := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ipMatch := ipRe.FindStringSubmatch(line)
	if len(ipMatch) > 1 {
		entry.Extra = map[string]interface{}{"ip": ipMatch[1]}
	}
	return entry
}

func (c *LinuxCollector) detectLevel(msg string) string {
	msg = strings.ToLower(msg)
	if strings.Contains(msg, "error") || strings.Contains(msg, "fail") || strings.Contains(msg, "critical") {
		return "error"
	}
	if strings.Contains(msg, "warn") {
		return "warning"
	}
	return "info"
}

func (c *LinuxCollector) guessTime(timeStr string) string {
	if strings.Contains(timeStr, "T") {
		return timeStr
	}
	now := time.Now()
	parsed, err := time.Parse("Jan 2 15:04:05", timeStr)
	if err != nil {
		return now.UTC().Format(time.RFC3339)
	}
	parsed = time.Date(now.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, time.UTC)
	if parsed.After(now) {
		parsed = parsed.AddDate(-1, 0, 0)
	}
	return parsed.UTC().Format(time.RFC3339)
}

func NewOsCollector(cfg config.AgentConfig) (Collector, error) {
	return NewLinuxCollector(cfg), nil
}
