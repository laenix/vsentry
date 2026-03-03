package mapper

import (
	"strconv"

	"github.com/laenix/vsentry/pkg/ocsf"
)

// MapFunc 定义了映射函数的签名
type MapFunc func(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent)

// registry 是全局的 Event ID 路由表
var registry = make(map[int]MapFunc)

// Register 供各子模块(如 identity, process)在 init() 中注册自己能处理的 Event ID
func Register(eventIDs []int, fn MapFunc) {
	for _, id := range eventIDs {
		registry[id] = fn
	}
}

// Enrich 是供 Windows Collector 调用的入口。
// 如果找到对应的映射函数，就将原始 XML 提取的 unmapped 数据填充到 OCSF 标准实体中。
func Enrich(eventID int, unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	if fn, exists := registry[eventID]; exists {
		fn(unmapped, entry)
	}
}

// ==========================================
// 辅助提取工具函数 (供具体的映射文件使用)
// ==========================================

// GetStr 安全提取字符串
func GetStr(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// GetInt 安全提取整数
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
