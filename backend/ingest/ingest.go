package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Ingest struct {
	url           string
	batchSize     int
	flushInterval time.Duration
	client        *http.Client
	buffer        []interface{}
	mu            sync.Mutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
	eventCount    int64
	errorCount    int64
}

func NewIngest(url string, batchSize int, flushInterval time.Duration, fields string) *Ingest {
	if len(url) > 0 {
		if strings.Contains(url, "?") {
			url += "&"
		} else {
			url += "?"
		}
		url += "_stream_fields=" + fields
	}
	return &Ingest{
		url:           url,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		client:        &http.Client{Timeout: 10 * time.Second},
		buffer:        make([]interface{}, 0, batchSize),
		stopChan:      make(chan struct{}),
	}
}

// Start starts the shipper's flush timer
func (i *Ingest) Start() {
	i.wg.Add(1)
	go i.flushLoop()
	log.Printf("VictoriaLogs shipper started, sending to: %s", i.url)
}

// Stop stops the shipper and flushes remaining events
func (i *Ingest) Stop() {
	close(i.stopChan)
	i.wg.Wait()
	i.Flush() // Final flush
	log.Printf("VictoriaLogs shipper stopped. Total events: %d, errors: %d", i.eventCount, i.errorCount)
}

// Send adds an event to the buffer
func (i *Ingest) Send(event interface{}) error {
	i.mu.Lock()
	i.buffer = append(i.buffer, event)
	shouldFlush := len(i.buffer) >= i.batchSize
	i.mu.Unlock()

	if shouldFlush {
		return i.Flush()
	}

	return nil
}

// Flush sends all buffered events to VictoriaLogs
func (i *Ingest) Flush() error {
	i.mu.Lock()
	if len(i.buffer) == 0 {
		i.mu.Unlock()
		return nil
	}

	// Copy buffer and clear it
	events := make([]interface{}, len(i.buffer))
	copy(events, i.buffer)
	i.buffer = i.buffer[:0]
	i.mu.Unlock()

	return i.sendBatch(events)
}

// flushLoop periodically flushes the buffer
func (i *Ingest) flushLoop() {
	defer i.wg.Done()
	ticker := time.NewTicker(i.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := i.Flush(); err != nil {
				log.Printf("Error flushing events: %v", err)
			}
		case <-i.stopChan:
			return
		}
	}
}

// sendBatch sends a batch of events to VictoriaLogs
func (i *Ingest) sendBatch(logs []interface{}) error {
	if len(logs) == 0 {
		return nil
	}

	// Convert events to NDJSON (newline-delimited JSON)
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	for _, logEntry := range logs {
		if err := encoder.Encode(logEntry); err != nil {
			i.errorCount++
			log.Printf("Error encoding event: %v", err)
			continue
		}
	}

	// Send to VictoriaLogs
	req, err := http.NewRequest("POST", i.url, &buf)
	if err != nil {
		i.errorCount += int64(len(logs))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := i.client.Do(req)
	if err != nil {
		i.errorCount += int64(len(logs))
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		i.errorCount += int64(len(logs))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	i.eventCount += int64(len(logs))
	log.Printf("Successfully sent %d events to VictoriaLogs (total: %d)", len(logs), i.eventCount)

	return nil
}
