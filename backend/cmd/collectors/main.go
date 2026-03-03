package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/laenix/vsentry/cmd/collectors/collector"
	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/ingest"
	"github.com/laenix/vsentry/cmd/collectors/storage"
)

func main() {
	// 1. 初始化内嵌配置 (确保你的 config 包里有 Init 函数加载了 config.json)
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

	// 3. 注册系统信号，实现优雅退出 (Graceful Shutdown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 4. 进入主循环
	ticker := time.NewTicker(time.Duration(config.Global.Interval) * time.Second)
	defer ticker.Stop()

	log.Printf("Collector initialized successfully. Interval: %ds", config.Global.Interval)

	for {
		select {
		case <-sigChan:
			log.Println("Received shutdown signal. Shutting down collector safely...")
			// 这里可以留出空间做最后一次 flush 或者保存书签
			return

		case <-ticker.C:
			// A. 执行新日志采集
			logs, err := col.Collect()
			if err != nil {
				log.Printf("Collection error: %v", err)
			}

			networkIsUp := false

			// B. 优先发送新采集的日志
			if len(logs) > 0 {
				success, failed := client.SendBatch(logs)
				if failed > 0 {
					log.Printf("Network error, saving %d new logs to local dead-letter queue", failed)
					dlq.SaveLogs(logs)
					networkIsUp = false
				} else {
					log.Printf("Flushed %d new logs successfully", success)
					networkIsUp = true
				}
			} else {
				// 本次没有新日志产生，假定网络是通畅的，给予处理积压日志的机会
				networkIsUp = true
			}

			// C. 【核心修复】：只有在网络确认畅通的情况下，才去处理历史死信队列
			if networkIsUp {
				pendingLogs := dlq.LoadAndClearPending()
				if len(pendingLogs) > 0 {
					log.Printf("Network restored, attempting to flush %d pending logs from cache", len(pendingLogs))

					pSuccess, pFailed := client.SendBatch(pendingLogs)
					if pFailed > 0 {
						// 如果再次失败，重新存回本地
						log.Printf("Failed to flush pending logs, returning to dead-letter queue")
						dlq.SaveLogs(pendingLogs)
					} else {
						log.Printf("Pending logs flushed successfully: %d", pSuccess)
					}
				}
			}
		}
	}
}
