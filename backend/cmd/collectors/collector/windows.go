//go:build windows

package collector

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/laenix/vsentry/cmd/collectors/config"
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

		queryStr := fmt.Sprintf(`*[System[TimeCreated[timediff(@SystemTime) <= %d]]]`, (c.cfg.Interval+2)*1000)

		channel16, _ := syscall.UTF16PtrFromString(source.Path)
		query16, _ := syscall.UTF16PtrFromString(queryStr)

		handle, _, _ := procEvtQuery.Call(
			0,
			uintptr(unsafe.Pointer(channel16)),
			uintptr(unsafe.Pointer(query16)),
			uintptr(EvtQueryChannelPath),
		)

		if handle == 0 {
			continue
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
			uintptr(2000),
			uintptr(0),
			uintptr(unsafe.Pointer(&returned)),
		)

		if returned == 0 || ret == 0 {
			break
		}

		for i := 0; i < int(returned); i++ {
			var bufferUsed, propCount uint32
			procEvtRender.Call(0, events[i], uintptr(EvtRenderEventXml), 0, 0, uintptr(unsafe.Pointer(&bufferUsed)), uintptr(unsafe.Pointer(&propCount)))

			buf := make([]uint16, bufferUsed)
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

	entry := &ocsf.VSentryOCSFEvent{
		Time:         t.Format(time.RFC3339),
		CategoryName: ocsf.CategorySystem,
		ClassName:    "System Event",
		ClassUID:     1000,
		Severity:     sevName,
		SeverityID:   sevID,
		RawData:      xmlStr,
		Metadata:     &ocsf.Metadata{Product: evt.System.Provider.Name}, // 修复点：Product 放进 Metadata
		Observer: &ocsf.Device{
			Hostname: c.cfg.Hostname,
			Vendor:   "Microsoft",
			OS:       &ocsf.OS{Type: "windows"},
		},
		Unmapped: map[string]interface{}{
			"event_id": evt.System.EventID,
			"channel":  evt.System.Channel,
		},
	}

	getVal := func(name string) string {
		for _, d := range evt.EventData.Data {
			if d.Name == name {
				return d.Value
			}
		}
		return ""
	}

	switch evt.System.EventID {
	case 4624, 4625, 4634:
		entry.CategoryName = ocsf.CategoryIdentity
		entry.ClassName = "Authentication"
		entry.ClassUID = ocsf.ClassAuthentication

		entry.SrcEndpoint = &ocsf.Endpoint{IP: getVal("IpAddress")}
		if port, err := strconv.Atoi(getVal("IpPort")); err == nil && port > 0 {
			entry.SrcEndpoint.Port = port
		}

		entry.Target = &ocsf.User{
			Name:   getVal("TargetUserName"),
			Domain: getVal("TargetDomainName"),
		}

		if evt.System.EventID == 4625 {
			entry.ActivityName = ocsf.ActionLogonFailed
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
		} else if evt.System.EventID == 4624 {
			entry.ActivityName = ocsf.ActionLogon
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
		} else {
			entry.ActivityName = ocsf.ActionLogoff
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
		}

	case 4688, 1:
		entry.CategoryName = ocsf.CategorySystem
		entry.ClassName = "Process Activity"
		entry.ClassUID = ocsf.ClassProcessActivity
		entry.ActivityName = ocsf.ActionCreate

		procName := getVal("NewProcessName")
		if procName == "" {
			procName = getVal("Image")
		}

		entry.Process = &ocsf.Process{
			Name:    procName,
			CmdLine: getVal("CommandLine"),
		}

		userName := getVal("SubjectUserName")
		if userName == "" {
			userName = getVal("User")
		}
		entry.Actor = &ocsf.User{Name: userName}
	}

	return entry
}

func (c *WindowsCollector) mapLevel(level int) (string, int) {
	switch level {
	case 1:
		return ocsf.SeverityCritical, ocsf.SeverityIDCritical
	case 2:
		return ocsf.SeverityHigh, ocsf.SeverityIDHigh // 修复点：Error 映射为 High
	case 3:
		return ocsf.SeverityMedium, ocsf.SeverityIDMedium // 修复点：Warning 映射为 Medium
	default:
		return ocsf.SeverityInfo, ocsf.SeverityIDInfo
	}
}
