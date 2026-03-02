//go:build darwin

package collector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/ingest"
)

type MacOSCollector struct {
	cfg config.AgentConfig
}

func NewMacOSCollector(cfg config.AgentConfig) *MacOSCollector {
	return &MacOSCollector{cfg: cfg}
}

func (c *MacOSCollector) Collect() ([]ingest.LogEntry, error) {
	// 构建允许采集的 subsystem 列表
	validSources := make(map[string]bool)
	hasEnabled := false
	for _, src := range c.cfg.Sources {
		if src.Enabled {
			validSources[src.Path] = true // mock data 中的 path: "system", "system.wifi"
			hasEnabled = true
		}
	}

	if !hasEnabled {
		return nil, nil // 没有启用的源，直接返回
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

	var allLogs []ingest.LogEntry
	for _, raw := range rawLogs {
		msg := fmt.Sprintf("%v", raw["eventMessage"])
		subsystem := fmt.Sprintf("%v", raw["subsystem"])

		// 精准过滤：只保留用户勾选的日志类型
		if !c.isMatchSource(subsystem, validSources) {
			continue
		}

		level := "info"
		if msgType := fmt.Sprintf("%v", raw["messageType"]); msgType == "Error" || msgType == "Fault" {
			level = "error"
		}

		allLogs = append(allLogs, ingest.LogEntry{
			Time:    time.Now().UTC().Format(time.RFC3339),
			Host:    c.cfg.Hostname,
			Source:  subsystem,
			Channel: "unified_log",
			Message: msg,
			Level:   level,
		})
	}

	return allLogs, nil
}

// isMatchSource 判断苹果底层日志的 subsystem 是否命中用户配置的规则
func (c *MacOSCollector) isMatchSource(subsystem string, validSources map[string]bool) bool {
	// 如果用户勾选了 "system" (对应所有系统底层日志)
	if validSources["system"] && strings.HasPrefix(subsystem, "com.apple.") {
		return true
	}

	// 针对具体类型匹配 (例如 system.wifi 匹配 com.apple.wifi)
	for srcPath := range validSources {
		parts := strings.Split(srcPath, ".")
		if len(parts) > 1 {
			keyword := parts[len(parts)-1] // 取 wifi, net, install
			if strings.Contains(subsystem, keyword) {
				return true
			}
		}
	}
	return false
}

func NewOsCollector(cfg config.AgentConfig) (Collector, error) {
	return NewMacOSCollector(cfg), nil
}
