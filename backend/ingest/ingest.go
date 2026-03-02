package ingest

import (
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
	logChan       chan interface{} // 核心：用于接收调度器日志的缓冲通道
	wg            sync.WaitGroup
	eventCount    int64
	errorCount    int64
}

func NewIngest(url string, batchSize int, flushInterval time.Duration, fields string) *Ingest {
	if len(url) > 0 {
		sep := "?"
		if strings.Contains(url, "?") {
			sep = "&"
		}
		fields = strings.TrimPrefix(fields, "_stream_fields=")
		url += sep + "_stream_fields=" + fields
	}
	return &Ingest{
		url:           url,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		client:        &http.Client{Timeout: 10 * time.Second},
		buffer:        make([]interface{}, 0, batchSize),
		logChan:       make(chan interface{}, 2000), // 为每个 Worker 配置私有背压缓冲
	}
}

func (i *Ingest) Start() {
	i.wg.Add(1)
	go i.runShipper()
	log.Printf("VictoriaLogs shipper started, sending to: %s", i.url)
}

func (i *Ingest) Stop() {
	close(i.logChan) // 关闭通道，通知 runShipper 排空残留数据并退出
	i.wg.Wait()
	log.Printf("VictoriaLogs shipper stopped. Total events: %d, errors: %d", i.eventCount, i.errorCount)
}

func (i *Ingest) Send(event interface{}) {
	i.logChan <- event // 调度器只负责放入通道，极速返回
}

// runShipper 是该实例专属的后台工作协程
func (i *Ingest) runShipper() {
	defer i.wg.Done()
	ticker := time.NewTicker(i.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-i.logChan:
			if !ok {
				// 通道关闭，执行最终刷写
				i.Flush()
				return
			}
			i.buffer = append(i.buffer, event)
			if len(i.buffer) >= i.batchSize {
				i.Flush()
			}
		case <-ticker.C:
			i.Flush()
		}
	}
}

func (i *Ingest) Flush() {
	if len(i.buffer) == 0 {
		return
	}
	events := make([]interface{}, len(i.buffer))
	copy(events, i.buffer)
	i.buffer = i.buffer[:0] // 复用底层数组

	// 这里可以考虑加入重试机制 (Retries)
	_ = i.sendBatch(events)
}

func (i *Ingest) sendBatch(logs []interface{}) error {
	// 保持原有的 JSON 编码和 HTTP POST 逻辑不变
	// ...
	// (此处省略原有 sendBatch 的内部实现，保持不变即可)
	return nil
}
