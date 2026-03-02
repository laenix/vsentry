package ingest

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Time      string                 `json:"_time"`
	Host      string                 `json:"host"`
	Source    string                 `json:"source"`
	Channel   string                 `json:"channel"`
	Message   string                 `json:"message"`
	Level     string                 `json:"level"`
	EventID   int                    `json:"event_id,omitempty"`
	Provider  string                 `json:"provider,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
	Sent      bool                   `json:"-"`
}

// Client handles sending logs to VSentry
type Client struct {
	endpoint    string
	token       string
	streamFields string
	httpClient  *http.Client
}

// NewClient creates a new ingest client
func NewClient(endpoint, token, streamFields string) *Client {
	return &Client{
		endpoint:    endpoint,
		token:       token,
		streamFields: streamFields,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendBatch sends a batch of logs to VSentry
func (c *Client) SendBatch(logs []map[string]interface{}) (success, failed int) {
	if len(logs) == 0 {
		return 0, 0
	}

	// Convert to JSON
	jsonData, err := json.Marshal(logs)
	if err != nil {
		log.Printf("Error marshaling logs: %v", err)
		return 0, len(logs)
	}

	// Create gzipped request body
	var bodyBuf bytes.Buffer
	gw := gzip.NewWriter(&bodyBuf)
	gw.Write(jsonData)
	gw.Close()

	req, err := http.NewRequest("POST", c.endpoint, &bodyBuf)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return 0, len(logs)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Error sending logs: %v", err)
		return 0, len(logs)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		return len(logs), 0
	}

	// Read error response
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Server returned error: %d - %s", resp.StatusCode, string(body))

	return 0, len(logs)
}

// Helper to convert []map[string]interface{} to []LogEntry
func toLogEntries(logs []map[string]interface{}) []LogEntry {
	var entries []LogEntry
	for _, m := range logs {
		entry := LogEntry{
			Time:    getString(m, "_time"),
			Host:    getString(m, "host"),
			Source:  getString(m, "source"),
			Channel: getString(m, "channel"),
			Message: getString(m, "message"),
			Level:   getString(m, "level"),
		}
		entries = append(entries, entry)
	}
	return entries
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}