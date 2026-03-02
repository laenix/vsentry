package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LogEntry 统一规范采集上来的数据结构
type LogEntry struct {
	Time    string                 `json:"_time"`
	Host    string                 `json:"host"`
	Source  string                 `json:"source"`
	Channel string                 `json:"channel"`
	Message string                 `json:"message"`
	Level   string                 `json:"level"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

func NewClient(endpoint, token, streamFields string) *Client {
	// 将 streamFields 自动拼接到 URL 中
	if len(streamFields) > 0 {
		endpoint = fmt.Sprintf("%s?_stream_fields=%s", endpoint, streamFields)
	}

	return &Client{
		endpoint: endpoint,
		token:    token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // 防止网络拥塞导致协程卡死
		},
	}
}

// SendBatch 批量发送日志，返回成功和失败的数量
func (c *Client) SendBatch(logs []LogEntry) (success int, failed int) {
	if len(logs) == 0 {
		return 0, 0
	}

	jsonData, err := json.Marshal(logs)
	if err != nil {
		return 0, len(logs)
	}

	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, len(logs)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, len(logs)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return len(logs), 0
	}

	return 0, len(logs)
}