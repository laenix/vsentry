package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 5156: Windows Filtering Platform AllowConnection, 3: Sysmon 网络Connection
	Register([]int{5156, 3}, mapNetworkConnection)
}

func mapNetworkConnection(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryNetwork
	entry.ClassName = "Network Activity"
	entry.ClassUID = ocsf.ClassNetworkActivity // 假设 OCSF 网络活动大Class
	entry.ActivityName = "Connection Allowed"

	// 提取源Info
	srcIP := GetStr(unmapped, "SourceAddress")
	if srcIP == "" {
		srcIP = GetStr(unmapped, "SourceIp") // 兼容 Sysmon
	}
	srcPort := GetInt(unmapped, "SourcePort")
	entry.SrcEndpoint = &ocsf.Endpoint{IP: srcIP, Port: srcPort}

	// 提取目标Info
	dstIP := GetStr(unmapped, "DestAddress")
	if dstIP == "" {
		dstIP = GetStr(unmapped, "DestinationIp") // 兼容 Sysmon
	}
	dstPort := GetInt(unmapped, "DestPort")
	if dstPort == 0 {
		dstPort = GetInt(unmapped, "DestinationPort")
	}
	entry.Target = &ocsf.Endpoint{IP: dstIP, Port: dstPort} // 复用 Target Storage目标 IP

	// 提取发起Connection的Process
	procName := GetStr(unmapped, "Application")
	if procName == "" {
		procName = GetStr(unmapped, "Image")
	}
	entry.Process = &ocsf.Process{Name: procName}

	// ProtocolInfo
	protocol := GetStr(unmapped, "Protocol") // 可能Return 6 (TCP), 17 (UDP) 等
	entry.Unmapped["network_protocol"] = protocol
}
