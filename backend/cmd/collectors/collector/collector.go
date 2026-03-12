package collector

import (
	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/pkg/ocsf"
)

// Collector - // 所有平台的Collect逻辑最终都Must吐出标准化的 - Event
type Collector interface {
	Collect() ([]ocsf.VSentryOCSFEvent, error)
}

// NewCollector - main.go 调用的统一工厂方法。
// 内部调用的 - 函数，其具体实现由带有 //  go:build 标签的
//   linux.go, windows.go 或 macos.go 在编译时动态提供。
func NewCollector(cfg config.AgentConfig) (Collector, error) {
	return NewOsCollector(cfg)
}
