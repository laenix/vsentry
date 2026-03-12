package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	//   5156: Windows Filtering Platform AllowConnection, 3: Sysmon NetworkConnection
	Register([]int{5156, 3}, mapNetworkConnection)
}

func mapNetworkConnection(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryNetwork
	entry.ClassName = "Network Activity"
	entry.ClassUID = ocsf.ClassNetworkActivity // 假设 - Network活动大类
	entry.ActivityName = "Connection Allowed"

	// 提取源Info - := GetStr(unmapped, "SourceAddress")
	if srcIP == "" {
		srcIP = GetStr(unmapped, "SourceIp") // 兼容 - }
	srcPort := GetInt(unmapped, "SourcePort")
	entry.SrcEndpoint = &ocsf.Endpoint{IP: srcIP, Port: srcPort}

	// 提取目标Info - := GetStr(unmapped, "DestAddress")
	if dstIP == "" {
		dstIP = GetStr(unmapped, "DestinationIp") // 兼容 - }
	dstPort := GetInt(unmapped, "DestPort")
	if dstPort == 0 {
		dstPort = GetInt(unmapped, "DestinationPort")
	}
	entry.Target = &ocsf.Endpoint{IP: dstIP, Port: dstPort} // 复用 - Storage目标 IP

	// 提取发起Connection的Process - := GetStr(unmapped, "Application")
	if procName == "" {
		procName = GetStr(unmapped, "Image")
	}
	entry.Process = &ocsf.Process{Name: procName}

	// ProtocolInfo - := GetStr(unmapped, "Protocol") // 可能Return - (TCP), 17 (UDP) 等
	entry.Unmapped["network_protocol"] = protocol
}
