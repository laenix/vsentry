package collector

import (
	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/pkg/ocsf"
)

// Collector 定义了跨平台日志采集器的标准接口
// 所有平台的采集逻辑最终都必须吐出标准化的 OCSF 事件
type Collector interface {
	Collect() ([]ocsf.VSentryOCSFEvent, error)
}

// NewCollector 是暴露给 main.go 调用的统一工厂方法。
// 内部调用的 NewOsCollector 函数，其具体实现由带有 //go:build 标签的
// linux.go, windows.go 或 macos.go 在编译时动态提供。
func NewCollector(cfg config.AgentConfig) (Collector, error) {
	return NewOsCollector(cfg)
}
