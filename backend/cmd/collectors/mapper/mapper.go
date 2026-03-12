package mapper

import (
	"strconv"

	"github.com/laenix/vsentry/pkg/ocsf"
)

//   =========================================================================
//   1. Windows 字典Engine (基于 EventID 整数路由)
//   =========================================================================

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

//   =========================================================================
//   2. Linux/macOS 文本Engine (基于 SourceType 字符串路由)
//   =========================================================================

// TextMapFunc - type TextMapFunc func(line string, entry *ocsf.VSentryOCSFEvent)

var textRegistry = make(map[string]TextMapFunc)

// RegisterText - Linux 子模块注册自己能Handle的LogType (如 "auth", "syslog")
func RegisterText(logTypes []string, fn TextMapFunc) {
	for _, t := range logTypes {
		textRegistry[t] = fn
	}
}

// EnrichText - Linux/macOS Collector 调用
func EnrichText(logType string, line string, entry *ocsf.VSentryOCSFEvent) {
	if fn, exists := textRegistry[logType]; exists {
		fn(line, entry)
	}
}

//   ==========================================
//   辅助提取工具函数
//   ==========================================

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
