//go:build windows

package collector

import (
	"encoding/xml"
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/laenix/vsentry/cmd/collectors/config"
	"github.com/laenix/vsentry/cmd/collectors/mapper"
	"github.com/laenix/vsentry/pkg/ocsf"
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

func NewOsCollector(cfg config.AgentConfig) (Collector, error) {
	return &WindowsCollector{cfg: cfg}, nil
}

func (c *WindowsCollector) Collect() ([]ocsf.VSentryOCSFEvent, error) {
	var allLogs []ocsf.VSentryOCSFEvent

	for _, source := range c.cfg.Sources {
		if !source.Enabled {
			continue
		}

		// =========================================================================
		// 动态生成HighPerformance底层 XPath Filter语句 (让 Windows 内核帮我们Filter噪音)
		// =========================================================================
		timeDiffMs := (c.cfg.Interval + 2) * 1000 // 冗余2seconds防止漏漏

		var queryStr string

		if len(source.EventIDs) > 0 {
			// 1. 如果ago端配置了 EventID List，生成精确匹配的 XPath
			var idConditions []string
			for _, id := range source.EventIDs {
				idConditions = append(idConditions, fmt.Sprintf("EventID=%d", id))
			}
			queryStr = fmt.Sprintf(`*[System[(%s) and TimeCreated[timediff(@SystemTime) <= %d]]]`, strings.Join(idConditions, " or "), timeDiffMs)

		} else if source.Query != "" {
			// 2. 如果配置了High级自定义 Query，直接嵌入
			queryStr = fmt.Sprintf(`*[System[(%s) and TimeCreated[timediff(@SystemTime) <= %d]]]`, source.Query, timeDiffMs)

		} else {
			// 3. Default全量Collect：不Limit EventID，只做Time切片
			queryStr = fmt.Sprintf(`*[System[TimeCreated[timediff(@SystemTime) <= %d]]]`, timeDiffMs)
		}

		// 将 Go string Convert为 Windows C-String (UTF-16)
		channel16, _ := syscall.UTF16PtrFromString(source.Path)
		query16, _ := syscall.UTF16PtrFromString(queryStr)

		// 调用 Win32 原生 API 进行极速Query
		handle, _, _ := procEvtQuery.Call(
			0,
			uintptr(unsafe.Pointer(channel16)),
			uintptr(unsafe.Pointer(query16)),
			uintptr(EvtQueryChannelPath),
		)

		if handle == 0 {
			continue // 可能Permission不足(如非Admin读Security)或通道Not found
		}

		events := c.fetchEvents(handle, source)
		allLogs = append(allLogs, events...)

		procEvtClose.Call(handle)
	}
	return allLogs, nil
}

func (c *WindowsCollector) fetchEvents(handle uintptr, source config.SourceConfig) []ocsf.VSentryOCSFEvent {
	var logs []ocsf.VSentryOCSFEvent
	var events [10]uintptr
	var returned uint32

	for {
		ret, _, _ := procEvtNext.Call(
			handle,
			uintptr(10),
			uintptr(unsafe.Pointer(&events[0])),
			uintptr(2000), // Timeout 2000ms
			uintptr(0),
			uintptr(unsafe.Pointer(&returned)),
		)

		if returned == 0 || ret == 0 {
			break
		}

		for i := 0; i < int(returned); i++ {
			var bufferUsed, propCount uint32

			// 第一次调用Get所需的内存大小
			procEvtRender.Call(0, events[i], uintptr(EvtRenderEventXml), 0, 0, uintptr(unsafe.Pointer(&bufferUsed)), uintptr(unsafe.Pointer(&propCount)))

			buf := make([]uint16, bufferUsed)

			// 第二次调用真正将 XML Render到内存Medium
			procEvtRender.Call(0, events[i], uintptr(EvtRenderEventXml), uintptr(bufferUsed), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&bufferUsed)), uintptr(unsafe.Pointer(&propCount)))

			xmlStr := syscall.UTF16ToString(buf)
			if entry := c.parseXmlToLog(xmlStr, source); entry != nil {
				logs = append(logs, *entry)
			}

			procEvtClose.Call(events[i]) // 防内存泄漏：必须关闭Event句柄
		}
	}
	return logs
}

func (c *WindowsCollector) parseXmlToLog(xmlStr string, source config.SourceConfig) *ocsf.VSentryOCSFEvent {
	var evt WinEventXml
	if err := xml.Unmarshal([]byte(xmlStr), &evt); err != nil {
		return nil
	}

	t, err := time.Parse(time.RFC3339Nano, evt.System.TimeCreated.SystemTime)
	if err != nil {
		t = time.Now().UTC()
	}

	sevName, sevID := c.mapLevel(evt.System.Level)

	// =========================================================================
	// 全量提取 XML 内的所有参数并打平放入 unmapped
	// =========================================================================
	unmapped := map[string]interface{}{
		"event_id": evt.System.EventID,
		"channel":  evt.System.Channel,
		"provider": evt.System.Provider.Name,
	}

	for _, d := range evt.EventData.Data {
		if d.Name != "" && d.Value != "" {
			unmapped[d.Name] = d.Value
		}
	}

	// 构造 OCSF 基础骨架 (Default当成一般 System Event)
	entry := &ocsf.VSentryOCSFEvent{
		Time:         t.Format(time.RFC3339),
		CategoryName: ocsf.CategorySystem,
		ClassName:    "System Event",
		ClassUID:     1000,
		Severity:     sevName,
		SeverityID:   sevID,
		RawData:      xmlStr,
		Metadata:     &ocsf.Metadata{Product: evt.System.Provider.Name},
		Observer: &ocsf.Device{
			Hostname: c.cfg.Hostname,
			Vendor:   "Microsoft",
			OS:       &ocsf.OS{Type: "windows"},
		},
		Unmapped: unmapped,
	}

	// =========================================================================
	// 将Group装好的基础 entry 和全部字典丢给大一统的 Mapper Engine进行深度加工
	// =========================================================================
	mapper.Enrich(evt.System.EventID, unmapped, entry)

	return entry
}

func (c *WindowsCollector) mapLevel(level int) (string, int) {
	switch level {
	case 1:
		return ocsf.SeverityCritical, ocsf.SeverityIDCritical
	case 2:
		return ocsf.SeverityHigh, ocsf.SeverityIDHigh
	case 3:
		return ocsf.SeverityMedium, ocsf.SeverityIDMedium
	default:
		return ocsf.SeverityInfo, ocsf.SeverityIDInfo
	}
}
