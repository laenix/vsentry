package collector

import (
	"bufio"
	"os"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/mapper"
	"github.com/laenix/vsentry/pkg/ocsf"
)

// AppCollector - (如 Nginx, MySQL, 业务System)
type AppCollector struct {
	cfg       config.AgentConfig
	positions map[string]int64
}

func NewAppCollector(cfg config.AgentConfig) *AppCollector {
	return &AppCollector{
		cfg:       cfg,
		positions: make(map[string]int64),
	}
}

// Collect - Format 为 "file" 的Data源
func (c *AppCollector) Collect() ([]ocsf.VSentryOCSFEvent, error) {
	var allLogs []ocsf.VSentryOCSFEvent

	for _, source := range c.cfg.Sources {
		if !source.Enabled || source.Format != "file" {
			continue //   Ignore未Enable或不是普通File的源 (如 windows_event)
		}

		logs, err := c.tailFile(source, 2000)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, logs...)
	}

	return allLogs, nil
}

func (c *AppCollector) tailFile(source config.SourceConfig, batchLimit int) ([]ocsf.VSentryOCSFEvent, error) {
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
		lastPos = 0 //   File被轮转截断
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

		logs = append(logs, c.parseAppLine(source, line))
		count++

		if count >= batchLimit {
			break
		}
	}

	currentPos, _ := file.Seek(0, 1)
	c.positions[source.Path] = currentPos

	return logs, nil
}

func (c *AppCollector) parseAppLine(source config.SourceConfig, line string) ocsf.VSentryOCSFEvent {
	// 构造Application层Log的基础 - 骨架
	entry := ocsf.VSentryOCSFEvent{
		Time:         time.Now().UTC().Format(time.RFC3339),
		CategoryName: ocsf.CategoryApp, // 默认归类为Application层 - :    "Application Activity",
		ClassUID:     1000, // 根据后续 - 覆盖
		SeverityID:   ocsf.SeverityIDInfo,
		Severity:     ocsf.SeverityInfo,
		RawData:      line,
		Metadata:     &ocsf.Metadata{Product: source.Type},
		Observer: &ocsf.Device{
			Hostname: c.cfg.Hostname,
		},
		Unmapped: make(map[string]interface{}),
	}
	entry.Unmapped["app_protocol"] = source.Type

	// 移交给大一统的双Engine - (我们之前写的 linux_web.go 里的正则会在这里生效)
	mapper.EnrichText(source.Type, line, &entry)

	return entry
}
