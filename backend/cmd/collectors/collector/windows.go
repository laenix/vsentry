//go:build windows

package collector

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/ingest"
)

var (
	modwevtapi    = syscall.NewLazyDLL("wevtapi.dll")
	procEvtQuery  = modwevtapi.NewProc("EvtQuery")
	procEvtNext   = modwevtapi.NewProc("EvtNext")
	procEvtRender = modwevtapi.NewProc("EvtRender")
	procEvtClose  = modwevtapi.NewProc("EvtClose")
)

const (
	EvtQueryChannelPath = 1
	EvtRenderEventXml   = 1
)

// WinEventXml 用于映射底层 EvtRender 输出的 XML 数据
type WinEventXml struct {
	System struct {
		Provider struct {
			Name string `xml:"Name,attr"`
		} `xml:"Provider"`
		EventID     int `xml:"EventID"`
		Level       int `xml:"Level"`
		TimeCreated struct {
			SystemTime string `xml:"SystemTime,attr"`
		} `xml:"TimeCreated"`
		Channel  string `xml:"Channel"`
		Computer string `xml:"Computer"`
	} `xml:"System"`
	EventData struct {
		Data []struct {
			Name  string `xml:"Name,attr"`
			Value string `xml:",chardata"`
		} `xml:"Data"`
	} `xml:"EventData"`
}

type WindowsCollector struct {
	cfg config.AgentConfig
}

// NewOsCollector 暴露给 main 函数调用
func NewOsCollector(cfg config.AgentConfig) (Collector, error) {
	return &WindowsCollector{cfg: cfg}, nil
}

func (c *WindowsCollector) Collect() ([]ingest.LogEntry, error) {
	var allLogs []ingest.LogEntry

	for _, source := range c.cfg.Sources {
		if !source.Enabled {
			continue
		}

		// 核心突破：使用 EvtQuery 的原生 XPath 语法，按毫秒级精确时间差拉取增量日志
		// timediff(@SystemTime) 计算的是事件发生时间与当前系统的毫秒差值
		queryStr := fmt.Sprintf(`*[System[TimeCreated[timediff(@SystemTime) <= %d]]]`, (c.cfg.Interval+2)*1000)

		channel16, _ := syscall.UTF16PtrFromString(source.Path)
		query16, _ := syscall.UTF16PtrFromString(queryStr)

		// 调用 Win32 API: EvtQuery
		handle, _, _ := procEvtQuery.Call(
			0,
			uintptr(unsafe.Pointer(channel16)),
			uintptr(unsafe.Pointer(query16)),
			uintptr(EvtQueryChannelPath),
		)

		if handle == 0 {
			continue // 找不到 Channel 或没有权限 (如 Security 需管理员)
		}

		events := c.fetchEvents(handle, source)
		allLogs = append(allLogs, events...)

		procEvtClose.Call(handle)
	}
	return allLogs, nil
}

func (c *WindowsCollector) fetchEvents(handle uintptr, source config.SourceConfig) []ingest.LogEntry {
	var logs []ingest.LogEntry
	var events [10]uintptr
	var returned uint32

	for {
		// 每次抓取 10 条事件句柄
		ret, _, _ := procEvtNext.Call(
			handle,
			uintptr(10),
			uintptr(unsafe.Pointer(&events[0])),
			uintptr(2000), // 超时 2000ms
			uintptr(0),
			uintptr(unsafe.Pointer(&returned)),
		)

		if returned == 0 || ret == 0 {
			break
		}

		for i := 0; i < int(returned); i++ {
			var bufferUsed, propCount uint32

			// 第一次 Call 获取所需内存 Buffer 大小
			procEvtRender.Call(0, events[i], uintptr(EvtRenderEventXml), 0, 0, uintptr(unsafe.Pointer(&bufferUsed)), uintptr(unsafe.Pointer(&propCount)))

			buf := make([]uint16, bufferUsed)

			// 第二次 Call 执行真正的 XML 渲染写入内存
			procEvtRender.Call(0, events[i], uintptr(EvtRenderEventXml), uintptr(bufferUsed), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&bufferUsed)), uintptr(unsafe.Pointer(&propCount)))

			xmlStr := syscall.UTF16ToString(buf)
			if entry := c.parseXmlToLog(xmlStr, source); entry != nil {
				logs = append(logs, *entry)
			}

			procEvtClose.Call(events[i])
		}
	}
	return logs
}

func (c *WindowsCollector) parseXmlToLog(xmlStr string, source config.SourceConfig) *ingest.LogEntry {
	var evt WinEventXml
	if err := xml.Unmarshal([]byte(xmlStr), &evt); err != nil {
		return nil
	}

	t, err := time.Parse(time.RFC3339Nano, evt.System.TimeCreated.SystemTime)
	if err != nil {
		t = time.Now().UTC()
	}

	// 将底层杂乱的 EventData 整理为 SIEM LogSQL 最喜欢的结构化 JSON
	extraData := map[string]interface{}{
		"event_id": evt.System.EventID,
		"provider": evt.System.Provider.Name,
	}
	for _, d := range evt.EventData.Data {
		if d.Name != "" {
			extraData[d.Name] = d.Value
		}
	}

	// 由于拿不到本地化翻译的长句（比如“系统已从异常关机中恢复”），将 JSON 实体序列化为 Message
	msgBytes, _ := json.Marshal(extraData)

	return &ingest.LogEntry{
		Time:    t.Format(time.RFC3339),
		Host:    c.cfg.Hostname,
		Source:  source.Type,
		Channel: evt.System.Channel,
		Message: string(msgBytes),
		Level:   c.mapLevel(evt.System.Level),
		Extra:   extraData,
	}
}

func (c *WindowsCollector) mapLevel(level int) string {
	// Win32 API 的严重程度映射: 1=Critical, 2=Error, 3=Warning, 4=Info
	switch level {
	case 1:
		return "critical"
	case 2:
		return "error"
	case 3:
		return "warning"
	default:
		return "info"
	}
}
