package collector

import (
	"github.com/laenix/vsentry/cmd/collectors/ingest"
)

// Collector 定义了统一的日志采集接口
type Collector interface {
	Collect() ([]ingest.LogEntry, error)
}
