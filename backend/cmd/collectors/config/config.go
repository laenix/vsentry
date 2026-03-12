package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log"
	"os"
)

//  go:embed config.json
var embeddedConfigBytes []byte

type AgentConfig struct {
	Name         string         `json:"name"`
	Type         string         `json:"type"` //   "windows", "linux", "macos"
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
	ReadFromHead bool   `json:"read_from_head"` //   是否Collect历史存量Log

	//   【New增】：High级Filter能力 (Windows EventLog 专用)
	EventIDs []int  `json:"event_ids,omitempty"` // 只Collect指定的 - (如: [4624, 4625, 4688])
	Query    string `json:"query,omitempty"`     // 供High级User直接编写的原生 - FilterCondition
}

var Global AgentConfig

func Init() {
	//   CheckConfig是否为空 (防止Development者在本地直接 go run 误Start)
	if len(bytes.TrimSpace(embeddedConfigBytes)) == 0 || string(bytes.TrimSpace(embeddedConfigBytes)) == "{}" {
		log.Fatal("Agent configuration missing! Must be compiled by VSentry Backend.")
	}

	if err := json.Unmarshal(embeddedConfigBytes, &Global); err != nil {
		log.Fatalf("Failed to parse embedded config: %v", err)
	}

	Global.Hostname, _ = os.Hostname()
	if Global.Hostname == "" {
		Global.Hostname = "unknown-host"
	}

	if Global.Interval <= 0 {
		Global.Interval = 5 // 默认 - 秒Collect一次
	}
}
