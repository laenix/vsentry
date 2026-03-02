package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/laenix/redAgent/collector"
	"github.com/laenix/redAgent/config"
	"github.com/laenix/redAgent/ingest"
	"github.com/laenix/redAgent/storage"
)

// Version is the collector version
const Version = "1.0.0"

func main() {
	log.Printf("redAgent v%s starting...", Version)

	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Collector: %s (Type: %s)", cfg.Name, cfg.Type)
	log.Printf("Channels: %v", cfg.Channels)
	log.Printf("Endpoint: %s", cfg.Ingest.Endpoint)

	// Initialize storage (for offline caching)
	store, err := storage.New("./data")
	if err != nil {
		log.Printf("Warning: Storage init failed: %v (continuing without cache)", err)
		store = nil
	}

	// Initialize HTTP client
	client := ingest.NewClient(
		cfg.Ingest.Endpoint,
		cfg.Ingest.Token,
		cfg.Ingest.StreamFields,
	)

	// Initialize collector based on type
	var coll collector.Collector
	switch cfg.Type {
	case "windows":
		coll = collector.NewWindowsEventCollector(cfg.Channels, store, client)
	case "linux":
		coll = collector.NewSyslogCollector(cfg.Channels, store, client)
	default:
		log.Fatalf("Unknown collector type: %s", cfg.Type)
	}

	// Create stop channel
	stop := make(chan struct{})
	
	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		close(stop)
	}()

	// Initial collection
	collectOnce(coll)

	// Start collection loop
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	log.Println("Collection started. Press Ctrl+C to stop.")
	
	for {
		select {
		case <-stop:
			log.Println("Stopped.")
			return
		case <-ticker.C:
			collectOnce(coll)
		}
	}
}

func collectOnce(coll collector.Collector) {
	logs, err := coll.Collect()
	if err != nil {
		log.Printf("Collection error: %v", err)
		return
	}

	if len(logs) == 0 {
		return
	}

	log.Printf("Collected %d logs", len(logs))

	// Send to VSentry
	client := coll.GetClient()
	success, failed := client.SendBatch(logs)
	log.Printf("Sent: %d, Failed: %d", success, failed)

	// Store failed logs locally
	if failed > 0 && coll.GetStorage() != nil {
		for _, log := range logs {
			if !log.Sent {
				coll.GetStorage().SaveLog(log)
			}
		}
	}
}

// createGzipJSON creates gzipped JSON from log entries
func createGzipJSON(logs []ingest.LogEntry) ([]byte, error) {
	jsonData, err := json.Marshal(logs)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(jsonData)
	gw.Close()

	return buf.Bytes(), nil
}