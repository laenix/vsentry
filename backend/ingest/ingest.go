package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Ingest struct {
	url           string
	streamFields  string // 【新增】保存原始配置的 streamFields，用于判断配置是否真的变更
	batchSize     int
	flushInterval time.Duration
	client        *http.Client
	buffer        []interface{}
	logChan       chan interface{}
	wg            sync.WaitGroup
	eventCount    int64
	errorCount    int64
}

func NewIngest(baseURL string, batchSize int, flushInterval time.Duration, fields string) *Ingest {
	// 1. 清理可能带过来的前缀
	cleanFields := strings.TrimPrefix(fields, "_stream_fields=")

	// 2. 优雅且安全地构建带参数的 URL
	finalURL := baseURL
	u, err := url.Parse(baseURL)
	if err == nil {
		q := u.Query()
		if cleanFields != "" {
			q.Set("_stream_fields", cleanFields)
		}
		// 【核心修复】：显式告诉 VictoriaLogs 如何解析 OCSF 标准日志
		q.Set("_msg_field", "raw_data")
		q.Set("_time_field", "time")

		u.RawQuery = q.Encode()
		finalURL = u.String()
	}

	return &Ingest{
		url:           finalURL,
		streamFields:  cleanFields, // 存下来给调度器做比对
		batchSize:     batchSize,
		flushInterval: flushInterval,
		client:        &http.Client{Timeout: 10 * time.Second},
		buffer:        make([]interface{}, 0, batchSize),
		logChan:       make(chan interface{}, 2000),
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

	// 执行发往 VictoriaLogs 的请求
	_ = i.sendBatch(events)
}

// sendBatch sends a batch of events to VictoriaLogs
func (i *Ingest) sendBatch(logs []interface{}) error {
	if len(logs) == 0 {
		return nil
	}

	// 1. Convert events to NDJSON (newline-delimited JSON)
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	for _, logEntry := range logs {
		if err := encoder.Encode(logEntry); err != nil {
			i.errorCount++
			log.Printf("Error encoding event: %v", err)
			continue
		}
	}

	// 2. Send to VictoriaLogs
	req, err := http.NewRequest("POST", i.url, &buf)
	if err != nil {
		i.errorCount += int64(len(logs))
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 推荐使用 application/stream+json 或 application/x-ndjson
	req.Header.Set("Content-Type", "application/stream+json")

	resp, err := i.client.Do(req)
	if err != nil {
		i.errorCount += int64(len(logs))
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 3. 【核心优化】：如果 VictoriaLogs 拒绝写入，必须把底层的报错原因打印出来
	// VictoriaLogs 成功通常返回 204 No Content，偶尔 200 或 202
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted {
		i.errorCount += int64(len(logs))

		// 读取 VL 吐出的详细错误信息
		bodyBytes, _ := io.ReadAll(resp.Body)
		errDetail := string(bodyBytes)

		log.Printf("[ERROR] VictoriaLogs rejected batch. Code: %d, Reason: %s", resp.StatusCode, errDetail)
		return fmt.Errorf("victorialogs error %d: %s", resp.StatusCode, errDetail)
	}

	// 4. 更新统计并打印成功日志
	i.eventCount += int64(len(logs))
	log.Printf("Successfully sent %d events to VictoriaLogs (total: %d)", len(logs), i.eventCount)

	return nil
}
