package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log"
	"os"
)

//go:embed config.json
var embeddedConfigBytes []byte

type AgentConfig struct {
	Name         string         `json:"name"`
	Type         string         `json:"type"` // "windows", "linux", "macos"
	Interval     int            `json:"interval"`
	Sources      []SourceConfig `json:"sources"`
	Endpoint     string         `json:"endpoint"`
	Token        string         `json:"token"`
	StreamFields string         `json:"stream_fields"`
	Hostname     string         `json:"-"`
}

type SourceConfig struct {
	Type         string `json:"type"`
	Path         string `json:"path"`
	Format       string `json:"format"`
	Enabled      bool   `json:"enabled"`
	ReadFromHead bool   `json:"read_from_head"` // 新增：是否收集历史存量日志
}

var Global AgentConfig

func Init() {
	if len(bytes.TrimSpace(embeddedConfigBytes)) == 0 {
		log.Fatal("Agent configuration missing! Must be compiled by VSentry Backend.")
	}

	if err := json.Unmarshal(embeddedConfigBytes, &Global); err != nil {
		log.Fatalf("Failed to parse embedded config: %v", err)
	}

	Global.Hostname, _ = os.Hostname()
	if Global.Interval <= 0 {
		Global.Interval = 5 // 默认 5 秒采集一次
	}
}
