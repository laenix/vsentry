//go:build linux

package collector

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/pkg/ocsf"
)

type LinuxCollector struct {
	cfg       config.AgentConfig
	positions map[string]int64
}

func NewOsCollector(cfg config.AgentConfig) (Collector, error) {
	return &LinuxCollector{
		cfg:       cfg,
		positions: make(map[string]int64),
	}, nil
}

func (c *LinuxCollector) Collect() ([]ocsf.VSentryOCSFEvent, error) {
	var allLogs []ocsf.VSentryOCSFEvent

	for _, source := range c.cfg.Sources {
		if !source.Enabled {
			continue
		}

		logs, err := c.tailFile(source, 2000)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, logs...)
	}

	return allLogs, nil
}

func (c *LinuxCollector) tailFile(source config.SourceConfig, batchLimit int) ([]ocsf.VSentryOCSFEvent, error) {
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
		lastPos = 0
	}

	_, err = file.Seek(lastPos, 0)
	if err != nil {
		return nil, err
	}

	var logs []ocsf.VSentryOCSFEvent
	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		logs = append(logs, c.parseLine(source, line))
		count++

		if count >= batchLimit {
			break
		}
	}

	currentPos, _ := file.Seek(0, 1)
	c.positions[source.Path] = currentPos

	return logs, nil
}

func (c *LinuxCollector) parseLine(source config.SourceConfig, line string) ocsf.VSentryOCSFEvent {
	entry := ocsf.VSentryOCSFEvent{
		Time:         time.Now().UTC().Format(time.RFC3339),
		CategoryName: ocsf.CategorySystem,
		ClassName:    "System Log",
		ClassUID:     1000,
		SeverityID:   ocsf.SeverityIDInfo,
		Severity:     ocsf.SeverityInfo,
		RawData:      line,
		Metadata:     &ocsf.Metadata{Product: source.Type}, // 修复点：Product 属于 Metadata
		Observer: &ocsf.Device{
			Hostname: c.cfg.Hostname,
			Vendor:   "Linux",
			OS:       &ocsf.OS{Type: "linux"},
		},
		Unmapped: make(map[string]interface{}),
	}
	entry.Unmapped["source_type"] = source.Type

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

func (c *LinuxCollector) parseSyslog(entry ocsf.VSentryOCSFEvent, line string) ocsf.VSentryOCSFEvent {
	re := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}|\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[^\s]*)\s+(\S+)\s+(\S+?)(?:\[(\d+)\])?:\s*(.*)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 6 {
		entry.Time = c.guessTime(matches[1])
		entry.Process = &ocsf.Process{Name: matches[3]}
		msg := matches[5]

		if strings.Contains(strings.ToLower(msg), "error") || strings.Contains(strings.ToLower(msg), "fail") {
			entry.Severity = ocsf.SeverityHigh // 修复点：映射为 OCSF 的 High
			entry.SeverityID = ocsf.SeverityIDHigh
		}
	}
	return entry
}

func (c *LinuxCollector) parseSSH(entry ocsf.VSentryOCSFEvent, line string) ocsf.VSentryOCSFEvent {
	entry.CategoryName = ocsf.CategoryIdentity
	entry.ClassName = "Authentication"
	entry.ClassUID = ocsf.ClassAuthentication
	entry.Metadata.Product = "sshd" // 修复点：更新 Metadata 中的 Product

	if strings.Contains(line, "Accepted password") || strings.Contains(line, "session opened") {
		entry.ActivityName = ocsf.ActionLogon
		entry.Severity = ocsf.SeverityInfo
		entry.SeverityID = ocsf.SeverityIDInfo
	} else if strings.Contains(line, "Failed password") || strings.Contains(line, "authentication failure") {
		entry.ActivityName = ocsf.ActionLogonFailed
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	} else if strings.Contains(line, "BREAK-IN") {
		entry.ActivityName = "Possible Break-in"
		entry.Severity = ocsf.SeverityCritical
		entry.SeverityID = ocsf.SeverityIDCritical
	}

	ipRe := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	if ipMatch := ipRe.FindStringSubmatch(line); len(ipMatch) > 1 {
		entry.SrcEndpoint = &ocsf.Endpoint{IP: ipMatch[1]}
	}

	userRe := regexp.MustCompile(`(?:for|user)\s+(?:invalid user\s+)?(\S+)\s+from`)
	if userMatch := userRe.FindStringSubmatch(line); len(userMatch) > 1 {
		entry.Target = &ocsf.User{Name: userMatch[1]}
	}

	return entry
}

func (c *LinuxCollector) parseWebAccess(entry ocsf.VSentryOCSFEvent, line string) ocsf.VSentryOCSFEvent {
	entry.CategoryName = ocsf.CategoryNetwork
	entry.ClassName = "HTTP Activity"
	entry.ClassUID = ocsf.ClassHTTPActivity

	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+\[([^\]]+)\]\s+"(\S+)\s+(\S+)\s+\S+"\s+(\d+)\s+(\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 8 {
		entry.SrcEndpoint = &ocsf.Endpoint{IP: matches[1]}
		entry.HTTPRequest = &ocsf.HTTPRequest{
			Method: matches[5],
			URL:    matches[6],
		}

		status, _ := strconv.Atoi(matches[7])
		entry.HTTPResponse = &ocsf.HTTPResponse{Code: status}

		if status >= 500 {
			entry.Severity = ocsf.SeverityHigh // 修复点：映射为 OCSF 的 High
			entry.SeverityID = ocsf.SeverityIDHigh
		} else if status >= 400 {
			entry.Severity = ocsf.SeverityMedium // 修复点：映射为 OCSF 的 Medium
			entry.SeverityID = ocsf.SeverityIDMedium
		}
	}
	return entry
}

func (c *LinuxCollector) parseWebError(entry ocsf.VSentryOCSFEvent, line string) ocsf.VSentryOCSFEvent {
	re := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+\[(\w+)\].*:\s*(.*)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		entry.Message = matches[2]
		switch matches[1] {
		case "error", "crit", "alert", "emerg":
			entry.Severity = ocsf.SeverityCritical
			entry.SeverityID = ocsf.SeverityIDCritical
		case "warn":
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
		}
	}
	return entry
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
