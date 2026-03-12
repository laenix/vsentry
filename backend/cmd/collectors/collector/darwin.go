//go:build darwin

package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/mapper"
	"github.com/laenix/vsentry/pkg/ocsf"
)

type DarwinCollector struct {
	cfg config.AgentConfig
}

// NewOsCollector 当编译目标为 GOOS=darwin 时自动选Medium
func NewOsCollector(cfg config.AgentConfig) (Collector, error) {
	return &DarwinCollector{
		cfg: cfg,
	}, nil
}

func (c *DarwinCollector) Collect() ([]ocsf.VSentryOCSFEvent, error) {
	var allLogs []ocsf.VSentryOCSFEvent

	// 构建 log show 命令，Get过去 Interval seconds的Log
	// --style syslog: 让 Apple 复杂的二进制LogOutput为我们熟悉的单行文本格式
	// --predicate: Filter出我们关心的Security子System，避免Log风暴引发 CPU 飙升
	timeWindow := fmt.Sprintf("%ds", c.cfg.Interval)
	predicate := `process == "sudo" OR process == "loginwindow" OR subsystem == "com.apple.securityd" OR subsystem == "com.apple.syspolicy"`

	cmd := exec.Command("log", "show", "--last", timeWindow, "--style", "syslog", "--predicate", predicate)

	output, err := cmd.Output()
	if err != nil {
		// 如果ExecuteFailed，可能没有匹配到Log，直接Return空
		return allLogs, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore log show 的头尾说明Info (如 "Filtering the log data using...")
		if line == "" || line[0] == 'F' || line[0] == 'L' {
			continue
		}

		entry := ocsf.VSentryOCSFEvent{
			Time:         time.Now().UTC().Format(time.RFC3339),
			CategoryName: ocsf.CategorySystem,
			ClassName:    "System Activity",
			ClassUID:     1000,
			SeverityID:   ocsf.SeverityIDInfo,
			Severity:     ocsf.SeverityInfo,
			RawData:      line,
			Metadata:     &ocsf.Metadata{Product: "macOS Unified Log"},
			Observer: &ocsf.Device{
				Hostname: c.cfg.Hostname,
				Vendor:   "Apple",
				OS:       &ocsf.OS{Type: "macos"},
			},
			Unmapped: make(map[string]interface{}),
		}

		// 将Handle权交给双Engine大一统 Mapper
		mapper.EnrichText("darwin_unified", line, &entry)

		allLogs = append(allLogs, entry)
	}

	return allLogs, nil
}
