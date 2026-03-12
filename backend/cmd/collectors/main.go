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
	"github.com/laenix/vsentry/pkg/ocsf" // 引入 ocsf 包以定义合并用的Event数Group
)

func main() {
	// 1. Initialize内嵌配置
	config.Init()
	log.Printf("Starting VSentry Collector [%s] on %s", config.Global.Name, config.Global.Hostname)

	// 2. Initialize核心Group件
	client := ingest.NewClient(
		config.Global.Endpoint,
		config.Global.Token,
		config.Global.StreamFields,
	)

	// 2.1 Initialize底层操作SystemCollect器 (Windows EventLog / Linux Syslog)
	osCol, err := collector.NewOsCollector(config.Global)
	if err != nil {
		log.Fatalf("Failed to initialize OS collector: %v", err)
	}

	// 2.2 Initialize跨平台Application层Collect器 (Nginx, MySQL 等通用文本Log)
	appCol := collector.NewAppCollector(config.Global)

	// Initialize本地FailedRetry队列 (DLQ)
	dlq := storage.New(config.Global.Name)

	// 3. RegisterSystem信号，实现优雅Exit (Graceful Shutdown)
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
			// 这里可以留出空间做最后一次 flush 或者Save书签
			return

		case <-ticker.C:
			// A. ExecuteNewLogCollect
			var allLogs []ocsf.VSentryOCSFEvent

			// Collect OS 级Log
			osLogs, err := osCol.Collect()
			if err != nil {
				log.Printf("OS Collection error: %v", err)
			}
			allLogs = append(allLogs, osLogs...)

			// Collect App 级Log
			appLogs, err := appCol.Collect()
			if err != nil {
				log.Printf("App Collection error: %v", err)
			}
			allLogs = append(allLogs, appLogs...)

			networkIsUp := false

			// B. 优先SendNewCollect的Log (OS 和 App 合并Send)
			if len(allLogs) > 0 {
				success, failed := client.SendBatch(allLogs)
				if failed > 0 {
					log.Printf("Network error, saving %d new logs to local dead-letter queue", failed)
					dlq.SaveLogs(allLogs)
					networkIsUp = false
				} else {
					// Log比较多时不建议every 5 seconds打印一次，这里仅在Debug期间保留
					log.Printf("Flushed %d new logs successfully", success)
					networkIsUp = true
				}
			} else {
				// 本次没有NewLog产生，假定网络是通畅的，给予Handle积压Log的机会
				networkIsUp = true
			}

			// C. 【核心防丢机制】：只有在网络Confirm畅通的情况下，才去Handle历史死信队列
			if networkIsUp {
				pendingLogs := dlq.LoadAndClearPending()
				if len(pendingLogs) > 0 {
					log.Printf("Network restored, attempting to flush %d pending logs from cache", len(pendingLogs))

					pSuccess, pFailed := client.SendBatch(pendingLogs)
					if pFailed > 0 {
						// 如果再次Failed，重New存回本地
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
