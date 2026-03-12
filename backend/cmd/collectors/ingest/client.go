package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/laenix/vsentry/pkg/ocsf"
)

type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

func NewClient(endpoint, token, streamFields string) *Client {
	streamFields = strings.TrimPrefix(streamFields, "_stream_fields=")

	u, err := url.Parse(endpoint)
	if err == nil {
		q := u.Query()
		if streamFields != "" {
			q.Set("_stream_fields", streamFields)
		}

		q.Set("_msg_field", "raw_data")
		q.Set("_time_field", "time")

		u.RawQuery = q.Encode()
		endpoint = u.String()
	}

	return &Client{
		endpoint: endpoint,
		token:    token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) SendBatch(logs []ocsf.VSentryOCSFEvent) (success int, failed int) {
	if len(logs) == 0 {
		return 0, 0
	}

	// 【核心改造】：不使用 Marshal 生成大数Group，而是逐行 Encode 形成 JSONL
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	for _, logEntry := range logs {
		if err := encoder.Encode(logEntry); err != nil {
			failed++
			continue
		}
		success++
	}

	req, err := http.NewRequest("POST", c.endpoint, &buf)
	if err != nil {
		return 0, len(logs)
	}

	// 明确声明我们Send的是 NDJSON 流
	req.Header.Set("Content-Type", "application/x-ndjson")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, len(logs)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return success, failed
	}

	return 0, len(logs)
}
