//  go:build linux

package collector

import (
	"bufio"
	"os"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/mapper"
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
	//   如果File被轮转（Log Rotation）或者截断，重置读取位置
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
	//   1. 构造保底的 OCSF 基础Event
	// 如果这是一条Unable - ，它依然会被Security地上报
	entry := ocsf.VSentryOCSFEvent{
		Time:         time.Now().UTC().Format(time.RFC3339),
		CategoryName: ocsf.CategorySystem,
		ClassName:    "System Log",
		ClassUID:     1000,
		SeverityID:   ocsf.SeverityIDInfo,
		Severity:     ocsf.SeverityInfo,
		RawData:      line,
		Metadata:     &ocsf.Metadata{Product: source.Type},
		Observer: &ocsf.Device{
			Hostname: c.cfg.Hostname,
			Vendor:   "Linux",
			OS:       &ocsf.OS{Type: "linux"},
		},
		Unmapped: make(map[string]interface{}),
	}
	entry.Unmapped["source_type"] = source.Type

	//   =========================================================================
	//   2. 将基础Event传递给大一统的 Mapper 文本Engine进行深度正则ParseSum字段覆写
	//   =========================================================================
	mapper.EnrichText(source.Type, line, &entry)

	return entry
}
