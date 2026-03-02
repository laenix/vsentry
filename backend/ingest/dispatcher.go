package ingest

import (
	"log"
	"sync"
	"time"

	"github.com/laenix/vsentry/database"
	"github.com/spf13/viper"
)

// LogPayload 包装日志数据及其元数据
type LogPayload struct {
	Config database.IngestCache
	Data   interface{}
}

// LogQueue 全局日志队列
var LogQueue = make(chan LogPayload, 10000)

type workerEntry struct {
	instance *Ingest
	lastSeen time.Time
}

var (
	activeWorkers = make(map[uint]*workerEntry)
	workerMu      sync.RWMutex // 改为读写锁，提升并发性能
)

// StartDispatcher 启动后台调度器
func StartDispatcher() {
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

// processPayload 现已变为纯内存非阻塞操作
func processPayload(payload LogPayload) {
	id := payload.Config.ID
	vLogsAddr := viper.GetString("victorialogs.url")
	if vLogsAddr == "" {
		vLogsAddr = "http://victorialogs:9428"
	}
	localEndpoint := vLogsAddr + "/insert/jsonline"

	workerMu.RLock()
	w, ok := activeWorkers[id]
	workerMu.RUnlock()

	// 1. 如果配置变更，先停止旧实例 (双重检查锁定)
	if ok && w.instance.url != localEndpoint {
		workerMu.Lock()
		if w, ok = activeWorkers[id]; ok && w.instance.url != localEndpoint {
			log.Printf("Config changed for IngestID %d, restarting worker...", id)
			w.instance.Stop()
			delete(activeWorkers, id)
			ok = false
		}
		workerMu.Unlock()
	}

	// 2. 如果实例不存在，则创建新实例
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

	// 3. 投递日志到实例私有通道 (完全非阻塞)
	w.lastSeen = time.Now()
	w.instance.Send(payload.Data)
}

// cleanIdleWorkers 回收逻辑保持不变 (需加写锁)
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
