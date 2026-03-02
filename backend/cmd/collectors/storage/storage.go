package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/ingest"
)

type Storage struct {
	dataDir string
}

func New(appName string) *Storage {
	// 在系统临时目录下建立专属缓存文件夹
	dataDir := filepath.Join(os.TempDir(), fmt.Sprintf(".%s_cache", appName))
	os.MkdirAll(dataDir, 0755)
	return &Storage{dataDir: dataDir}
}

// SaveLogs 将失败的日志批量落盘
func (s *Storage) SaveLogs(entries []ingest.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	filename := filepath.Join(s.dataDir, fmt.Sprintf("dlq_%d.json", time.Now().UnixNano()))
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadAndClearPending 读取缓存的日志并清空目录
func (s *Storage) LoadAndClearPending() []ingest.LogEntry {
	var entries []ingest.LogEntry

	files, err := os.ReadDir(s.dataDir)
	if err != nil || len(files) == 0 {
		return entries
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.dataDir, file.Name())
		data, err := os.ReadFile(path)
		if err == nil {
			var batch []ingest.LogEntry
			if err := json.Unmarshal(data, &batch); err == nil {
				entries = append(entries, batch...)
			}
		}
		// 读取后立即删除文件，防止无限堆积
		os.Remove(path)
	}
	return entries
}