//go:build darwin

package collector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/pkg/ocsf"
)

type MacOSCollector struct {
	cfg config.AgentConfig
}

func NewOsCollector(cfg config.AgentConfig) (Collector, error) {
	return &MacOSCollector{cfg: cfg}, nil
}

func (c *MacOSCollector) Collect() ([]ocsf.VSentryOCSFEvent, error) {
	validSources := make(map[string]bool)
	hasEnabled := false
	for _, src := range c.cfg.Sources {
		if src.Enabled {
			validSources[src.Path] = true
			hasEnabled = true
		}
	}

	if !hasEnabled {
		return nil, nil
	}

	timeFilter := fmt.Sprintf("%ds", c.cfg.Interval+2)
	cmd := exec.Command("log", "show", "--style", "json", "--last", timeFilter)

	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return nil, err
	}

	var rawLogs []map[string]interface{}
	if err := json.Unmarshal(output, &rawLogs); err != nil {
		return nil, err
	}

	var allLogs []ocsf.VSentryOCSFEvent
	for _, raw := range rawLogs {
		subsystem := fmt.Sprintf("%v", raw["subsystem"])

		if !c.isMatchSource(subsystem, validSources) {
			continue
		}

		allLogs = append(allLogs, c.parseToOCSF(raw))
	}

	return allLogs, nil
}

func (c *MacOSCollector) isMatchSource(subsystem string, validSources map[string]bool) bool {
	if validSources["system"] && strings.HasPrefix(subsystem, "com.apple.") {
		return true
	}

	for srcPath := range validSources {
		parts := strings.Split(srcPath, ".")
		if len(parts) > 1 {
			keyword := parts[len(parts)-1]
			if strings.Contains(subsystem, keyword) {
				return true
			}
		}
	}
	return false
}

func (c *MacOSCollector) parseToOCSF(raw map[string]interface{}) ocsf.VSentryOCSFEvent {
	subsystem := fmt.Sprintf("%v", raw["subsystem"])
	msgType := fmt.Sprintf("%v", raw["messageType"])

	rawBytes, _ := json.Marshal(raw)

	entry := ocsf.VSentryOCSFEvent{
		Time:         time.Now().UTC().Format(time.RFC3339),
		CategoryName: ocsf.CategorySystem,
		ClassName:    "System Log",
		ClassUID:     1000,
		RawData:      string(rawBytes),
		Metadata:     &ocsf.Metadata{Product: subsystem}, // 修复点：Product 属于 Metadata
		Observer: &ocsf.Device{
			Hostname: c.cfg.Hostname,
			Vendor:   "Apple",
			OS:       &ocsf.OS{Type: "macos"},
		},
		Unmapped: make(map[string]interface{}),
	}

	switch msgType {
	case "Fault", "Error":
		entry.Severity = ocsf.SeverityHigh // 修复点
		entry.SeverityID = ocsf.SeverityIDHigh
	case "Warning":
		entry.Severity = ocsf.SeverityMedium // 修复点
		entry.SeverityID = ocsf.SeverityIDMedium
	default:
		entry.Severity = ocsf.SeverityInfo
		entry.SeverityID = ocsf.SeverityIDInfo
	}

	if msg, ok := raw["eventMessage"].(string); ok && msg != "" {
		entry.Message = msg
	}

	if strings.Contains(subsystem, "auth") || strings.Contains(subsystem, "opendirectory") {
		entry.CategoryName = ocsf.CategoryIdentity
		entry.ClassName = "Authentication"
		entry.ClassUID = ocsf.ClassAuthentication

		if entry.SeverityID >= ocsf.SeverityIDHigh {
			entry.ActivityName = ocsf.ActionLogonFailed
		}
	}

	if proc, ok := raw["processImagePath"].(string); ok && proc != "" {
		entry.Process = &ocsf.Process{Path: proc}
	}

	return entry
}
