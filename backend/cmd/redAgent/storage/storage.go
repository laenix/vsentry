package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/laenix/redAgent/ingest"
)

// Storage provides local file-based caching
type Storage struct {
	dataDir string
}

// New creates a new storage instance
func New(dataDir string) (*Storage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &Storage{
		dataDir: dataDir,
	}, nil
}

// SaveLog saves a log entry to local storage for retry
func (s *Storage) SaveLog(entry ingest.LogEntry) error {
	// Use timestamp as filename
	filename := filepath.Join(s.dataDir, fmt.Sprintf("%d.json", time.Now().UnixNano()))
	
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// LoadPending loads all pending logs for retry
func (s *Storage) LoadPending() ([]ingest.LogEntry, error) {
	var entries []ingest.LogEntry

	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.dataDir, file.Name()))
		if err != nil {
			continue
		}

		var entry ingest.LogEntry
		if err := json.Unmarshal(data, &entry); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// Clear removes all cached logs
func (s *Storage) Clear() error {
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		os.Remove(filepath.Join(s.dataDir, file.Name()))
	}

	return nil
}