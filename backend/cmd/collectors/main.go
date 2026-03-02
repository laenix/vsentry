package main

import (
	"log"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/collector"
	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/ingest"
	"github.com/laenix/vsentry/cmd/collectors/storage"
)

func main() {
	// 1. 初始化内嵌配置
	config.Init()
	log.Printf("Starting VSentry Collector [%s] on %s", config.Global.Name, config.Global.Hostname)

	// 2. 初始化核心组件
	client := ingest.NewClient(
		config.Global.Endpoint,
		config.Global.Token,
		config.Global.StreamFields,
	)

	col, err := collector.NewOsCollector(config.Global)
	if err != nil {
		log.Fatalf("Failed to initialize collector: %v", err)
	}

	// 初始化本地失败重试队列 (DLQ)
	dlq := storage.New(config.Global.Name)

	// 3. 进入主循环
	ticker := time.NewTicker(time.Duration(config.Global.Interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// A. 执行采集
		logs, err := col.Collect()
		if err != nil {
			log.Printf("Collection error: %v", err)
			continue
		}

		// B. 尝试加载之前失败的日志合并发送
		pendingLogs := dlq.LoadAndClearPending()
		if len(pendingLogs) > 0 {
			logs = append(logs, pendingLogs...)
			log.Printf("Loaded %d pending logs from local cache", len(pendingLogs))
		}

		if len(logs) == 0 {
			continue
		}

		// C. 发送至 VSentry 服务端
		success, failed := client.SendBatch(logs)
		log.Printf("Flushed %d logs (Success: %d, Failed: %d)", len(logs), success, failed)

		// D. 发送失败，将全部日志存回本地队列
		if failed > 0 {
			log.Printf("Network error, saving logs to local dead-letter queue")
			dlq.SaveLogs(logs)
		}
	}
}
