package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/laenix/vsentry/pkg/ocsf"
)

// ============================================================================
// 1. Dead Letter Queue (DLQ) - 死信队列缓存System
// 用于在断网或后端宕机时，将Collect到的 OCSF Event暂存到本地磁盘
// ============================================================================

type Storage struct {
	dataDir string
}

// New Initialize本地缓存目录
func New(appName string) *Storage {
	// 在System临时目录下建立专属缓存File夹 (例如: /tmp/.vsentry_agent_cache)
	dataDir := filepath.Join(os.TempDir(), fmt.Sprintf(".%s_cache", appName))
	os.MkdirAll(dataDir, 0755)

	return &Storage{dataDir: dataDir}
}

// SaveLogs 将SendFailed的Log批量落盘
func (s *Storage) SaveLogs(entries []ocsf.VSentryOCSFEvent) error {
	if len(entries) == 0 {
		return nil
	}

	filename := filepath.Join(s.dataDir, fmt.Sprintf("dlq_%d.json", time.Now().UnixNano()))
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	// 写入独立File，防止大FileParse吃内存
	return os.WriteFile(filename, data, 0644)
}

// LoadAndClearPending 读取缓存的Log并立即清空已读File
func (s *Storage) LoadAndClearPending() []ocsf.VSentryOCSFEvent {
	var entries []ocsf.VSentryOCSFEvent

	files, err := os.ReadDir(s.dataDir)
	if err != nil || len(files) == 0 {
		return entries
	}

	for _, file := range files {
		// 只Handle DLQ 队列File，Ignore书签File
		if filepath.Ext(file.Name()) != ".json" || !hasPrefix(file.Name(), "dlq_") {
			continue
		}

		path := filepath.Join(s.dataDir, file.Name())
		data, err := os.ReadFile(path)
		if err == nil {
			var batch []ocsf.VSentryOCSFEvent
			if err := json.Unmarshal(data, &batch); err == nil {
				entries = append(entries, batch...)
			}
		}

		// 读取后立即DeleteFile，防止堆积
		os.Remove(path)
	}

	return entries
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
