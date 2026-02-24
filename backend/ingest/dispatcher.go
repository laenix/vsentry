package ingest

import (
	"log"
	"sync"
	"time"

	"github.com/laenix/vsentry/database"
)

// LogPayload 包装日志数据及其元数据，用于在 Channel 中传输
type LogPayload struct {
	Config database.IngestCache
	Data   interface{}
}

// LogQueue 全局日志队列，设置缓冲区以处理背压
var LogQueue = make(chan LogPayload, 10000)

// workerEntry 封装实例及其最后活跃时间
type workerEntry struct {
	instance *Ingest
	lastSeen time.Time
}

var (
	activeWorkers = make(map[uint]*workerEntry)
	workerMu      sync.RWMutex
)

// StartDispatcher 启动后台调度器，负责分发日志并回收闲置实例
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

// processPayload 处理单个日志载荷
func processPayload(payload LogPayload) {
	workerMu.Lock()
	defer workerMu.Unlock()

	id := payload.Config.ID
	w, ok := activeWorkers[id]

	// Use local VictoriaLogs for forwarding (configured in config.yaml)
	localEndpoint := "http://localhost:9428/insert/jsonline"

	// 1. 如果配置发生变更（例如 StreamFields 变了），重启实例以应用新标签
	if ok && (w.instance.url != localEndpoint) {
		log.Printf("Config changed for IngestID %d, restarting worker...", id)
		w.instance.Stop()
		ok = false
	}

	// 2. 如果实例不存在或已被关闭，则创建新实例
	if !ok {
		// 创建新实例：100条批处理，5秒强制刷新
		ins := NewIngest(
			localEndpoint,
			100,
			5*time.Second,
			payload.Config.StreamFields,
		)
		ins.Start()
		w = &workerEntry{instance: ins}
		activeWorkers[id] = w
		log.Printf("Started new worker for IngestID %d (%s)", id, payload.Config.StreamFields)
	}

	// 3. 更新活跃时间并发送日志
	w.lastSeen = time.Now()
	w.instance.Send(payload.Data)
}

// cleanIdleWorkers 回收超过 10 分钟未使用的 Worker 以释放资源
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

// StopAllWorkers 优雅关机时的全局清理逻辑，确保缓冲区日志全部发出
func StopAllWorkers() {
	workerMu.Lock()
	defer workerMu.Unlock()

	log.Printf("Shutting down %d active workers...", len(activeWorkers))

	var wg sync.WaitGroup
	for id, w := range activeWorkers {
		wg.Add(1)
		go func(ingestID uint, entry *workerEntry) {
			defer wg.Done()
			// Stop 会触发最终的 Flush 并等待发送完毕
			entry.instance.Stop()
		}(id, w)
	}

	wg.Wait()
	log.Println("All workers stopped.")
}
