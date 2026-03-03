package forensic

import (
	"fmt"
	"strings"
)

// ForensicEvent 标准化输出格式
type ForensicEvent map[string]interface{}

// Parser 解析器接口
type Parser interface {
	Parse(filePath string) ([]ForensicEvent, error)
}

// GetParser 工厂方法：根据文件后缀动态分配解析器
func GetParser(fileType string) (Parser, error) {
	fileType = strings.ToLower(fileType)
	switch fileType {
	case "evtx":
		return &EVTXParser{}, nil
	case "pcap", "pcapng":
		return &PCAPParser{}, nil
	case "log", "txt", "csv":
		return &TextParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}