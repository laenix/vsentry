package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/laenix/vsentry/pkg/ocsf" // 引入标准契约
)

type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

func NewClient(endpoint, token, streamFields string) *Client {
	if len(streamFields) > 0 {
		endpoint = fmt.Sprintf("%s?_stream_fields=%s", endpoint, streamFields)
	}

	return &Client{
		endpoint: endpoint,
		token:    token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SendBatch 的签名改为接收 OCSF 事件数组
func (c *Client) SendBatch(logs []ocsf.VSentryOCSFEvent) (success int, failed int) {
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
