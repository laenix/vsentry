package forensic

import (
	"fmt"
	"strings"
)

// ForensicEvent - type ForensicEvent map[string]interface{}

// Parser - type Parser interface {
	Parse(filePath string) ([]ForensicEvent, error)
}

// GetParser - ：根据File后缀动态分配Parse器
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