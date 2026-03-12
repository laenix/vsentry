package ingest

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/laenix/vsentry/database"
	"github.com/spf13/viper"
)

// LogPayload - type LogPayload struct {
	Config database.IngestCache
	Data   interface{}
}

// LogQueue - var LogQueue = make(chan LogPayload, 10000)

type workerEntry struct {
	instance *Ingest
	lastSeen time.Time
}

var (
	activeWorkers = make(map[uint]*workerEntry)
	workerMu      sync.RWMutex
)

// StartDispatcher - func StartDispatcher() {
	log.Println("Log Dispatcher started")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case payload := <-LogQueue:
			processPayload(payload)
		case <-ticker.C:
			cleanIdleWorkers()
		}
	}
}

// processPayload - func processPayload(payload LogPayload) {
	id := payload.Config.ID
	vLogsAddr := viper.GetString("victorialogs.url")
	if vLogsAddr == "" {
		vLogsAddr = "http://  victorialogs:9428"
	}
	localEndpoint := vLogsAddr + "/insert/jsonline"

	// 提取出纯净的Config字段进行比对 - := strings.TrimPrefix(payload.Config.StreamFields, "_stream_fields=")

	workerMu.RLock()
	w, ok := activeWorkers[id]
	workerMu.RUnlock()

	//   1. 【核心修复】：如果Config的 StreamFields 发生实质变更，才Stop旧Instance
	if ok && w.instance.streamFields != cleanFields {
		workerMu.Lock()
		if w, ok = activeWorkers[id]; ok && w.instance.streamFields != cleanFields {
			log.Printf("Config changed for IngestID %d (fields: %s -> %s), restarting worker...", id, w.instance.streamFields, cleanFields)
			w.instance.Stop()
			delete(activeWorkers, id)
			ok = false
		}
		workerMu.Unlock()
	}

	//   2. 如果Instance不存在，则CreateNewInstance
	if !ok {
		workerMu.Lock()
		if w, ok = activeWorkers[id]; !ok {
			ins := NewIngest(localEndpoint, 100, 5*time.Second, payload.Config.StreamFields)
			ins.Start()
			w = &workerEntry{instance: ins}
			activeWorkers[id] = w
			log.Printf("Started new worker for IngestID %d (%s)", id, payload.Config.StreamFields)
		}
		workerMu.Unlock()
	}

	//   3. 投递Log到Instance私有通道 (完全非Block)
	w.lastSeen = time.Now()
	w.instance.Send(payload.Data)
}

// cleanIdleWorkers - (需加写锁)
func cleanIdleWorkers() {
	workerMu.Lock()
	defer workerMu.Unlock()
	for id, w := range activeWorkers {
		if time.Since(w.lastSeen) > 10*time.Minute {
			log.Printf("IngestID %d idle for 10m, shutting down worker...", id)
			w.instance.Stop()
			delete(activeWorkers, id)
		}
	}
}

func StopAllWorkers() {
	workerMu.Lock()
	defer workerMu.Unlock()
	var wg sync.WaitGroup
	for _, w := range activeWorkers {
		wg.Add(1)
		go func(entry *workerEntry) {
			defer wg.Done()
			entry.instance.Stop()
		}(w)
	}
	wg.Wait()
	log.Println("All workers stopped.")
}
