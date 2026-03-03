package mapper

import (
	"strconv"

	"github.com/laenix/vsentry/pkg/ocsf"
)

// =========================================================================
// 1. Windows 字典引擎 (基于 EventID 整数路由)
// =========================================================================

type MapFunc func(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent)

var registry = make(map[int]MapFunc)

func Register(eventIDs []int, fn MapFunc) {
	for _, id := range eventIDs {
		registry[id] = fn
	}
}

func Enrich(eventID int, unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	if fn, exists := registry[eventID]; exists {
		fn(unmapped, entry)
	}
}

// =========================================================================
// 2. Linux/macOS 文本引擎 (基于 SourceType 字符串路由)
// =========================================================================

// TextMapFunc 针对纯文本日志的映射签名
type TextMapFunc func(line string, entry *ocsf.VSentryOCSFEvent)

var textRegistry = make(map[string]TextMapFunc)

// RegisterText 供 Linux 子模块注册自己能处理的日志类型 (如 "auth", "syslog")
func RegisterText(logTypes []string, fn TextMapFunc) {
	for _, t := range logTypes {
		textRegistry[t] = fn
	}
}

// EnrichText 供 Linux/macOS Collector 调用
func EnrichText(logType string, line string, entry *ocsf.VSentryOCSFEvent) {
	if fn, exists := textRegistry[logType]; exists {
		fn(line, entry)
	}
}

// ==========================================
// 辅助提取工具函数
// ==========================================

func GetStr(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func GetInt(m map[string]interface{}, key string) int {
	str := GetStr(m, key)
	if str == "" {
		return 0
	}
	if val, err := strconv.Atoi(str); err == nil {
		return val
	}
	return 0
}
