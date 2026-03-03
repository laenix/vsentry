package forensic

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xrawsec/golang-evtx/evtx"
)

type EVTXParser struct{}

func (p *EVTXParser) Parse(filePath string) ([]ForensicEvent, error) {
	ef, err := evtx.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EVTX: %v", err)
	}
	defer ef.Close()

	var events []ForensicEvent

	for e := range ef.FastEvents() {
		// ✅ 修复：e 本身就是 GoEvtxMap，直接序列化它即可
		eventJSON, err := json.Marshal(e)
		if err != nil {
			continue
		}

		var rawMap map[string]interface{}
		json.Unmarshal(eventJSON, &rawMap)

		forensicEvt := ForensicEvent{
			"class_name": "Windows Event",
			"raw_data":   string(eventJSON),
		}

		// 提取 EVTX 核心元数据 (兼容不同层级的 System 字段)
		var systemData map[string]interface{}
		
		// 0xrawsec 解析出来的结构通常形如 {"Event": {"System": {...}, "EventData": {...}}}
		if eventNode, ok := rawMap["Event"].(map[string]interface{}); ok {
			if sys, ok := eventNode["System"].(map[string]interface{}); ok {
				systemData = sys
			}
		} else if sys, ok := rawMap["System"].(map[string]interface{}); ok {
			// 兜底：如果 System 直接在最外层
			systemData = sys
		}

		// 如果成功拿到了 System 节点，提取 EventID 和 Time
		if systemData != nil {
			if eventID, ok := systemData["EventID"].(map[string]interface{}); ok {
				forensicEvt["event_id"] = eventID["Value"]
			} else if id, ok := systemData["EventID"].(float64); ok {
				forensicEvt["event_id"] = id
			}

			if timeData, ok := systemData["TimeCreated"].(map[string]interface{}); ok {
				if sysTime, ok := timeData["SystemTime"].(string); ok {
					if t, err := time.Parse(time.RFC3339Nano, sysTime); err == nil {
						forensicEvt["time"] = t.UTC().Format(time.RFC3339)
					}
				}
			}
		}
		
		events = append(events, forensicEvt)
	}

	return events, nil
}