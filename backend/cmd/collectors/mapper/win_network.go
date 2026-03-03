package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 5156: Windows Filtering Platform 允许连接, 3: Sysmon 网络连接
	Register([]int{5156, 3}, mapNetworkConnection)
}

func mapNetworkConnection(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryNetwork
	entry.ClassName = "Network Activity"
	entry.ClassUID = ocsf.ClassNetworkActivity // 假设 OCSF 网络活动大类
	entry.ActivityName = "Connection Allowed"

	// 提取源信息
	srcIP := GetStr(unmapped, "SourceAddress")
	if srcIP == "" {
		srcIP = GetStr(unmapped, "SourceIp") // 兼容 Sysmon
	}
	srcPort := GetInt(unmapped, "SourcePort")
	entry.SrcEndpoint = &ocsf.Endpoint{IP: srcIP, Port: srcPort}

	// 提取目标信息
	dstIP := GetStr(unmapped, "DestAddress")
	if dstIP == "" {
		dstIP = GetStr(unmapped, "DestinationIp") // 兼容 Sysmon
	}
	dstPort := GetInt(unmapped, "DestPort")
	if dstPort == 0 {
		dstPort = GetInt(unmapped, "DestinationPort")
	}
	entry.Target = &ocsf.Endpoint{IP: dstIP, Port: dstPort} // 复用 Target 存储目标 IP

	// 提取发起连接的进程
	procName := GetStr(unmapped, "Application")
	if procName == "" {
		procName = GetStr(unmapped, "Image")
	}
	entry.Process = &ocsf.Process{Name: procName}

	// 协议信息
	protocol := GetStr(unmapped, "Protocol") // 可能返回 6 (TCP), 17 (UDP) 等
	entry.Unmapped["network_protocol"] = protocol
}
