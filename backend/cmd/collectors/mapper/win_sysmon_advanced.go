package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 22: DNS 查询
	Register([]int{22}, mapSysmonDNS)
	// 12, 13, 14: 注册表操作
	Register([]int{12, 13, 14}, mapSysmonRegistry)
	// 8, 10: 进程注入与异常访问 (窃取哈希)
	Register([]int{8, 10}, mapSysmonInjection)
}

func mapSysmonDNS(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryNetwork
	entry.ClassName = "DNS Activity"
	entry.ClassUID = 4003 // OCSF DNS Activity
	entry.ActivityName = "Query"

	procName := GetStr(unmapped, "Image")
	entry.Process = &ocsf.Process{Name: procName}

	query := GetStr(unmapped, "QueryName")
	status := GetStr(unmapped, "QueryStatus")

	// 在 Unmapped 里记录查询详情
	entry.Unmapped["dns_query"] = query
	entry.Unmapped["dns_status"] = status
}

func mapSysmonRegistry(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Registry Activity"
	entry.ClassUID = 1010 // OCSF Registry

	eventID := entry.Unmapped["event_id"].(int)

	entry.Process = &ocsf.Process{Name: GetStr(unmapped, "Image")}

	entry.Registry = &ocsf.Registry{
		Key:   GetStr(unmapped, "TargetObject"),
		Value: GetStr(unmapped, "Details"),
	}

	switch eventID {
	case 12:
		entry.ActivityName = ocsf.ActionCreate
	case 13:
		entry.ActivityName = ocsf.ActionUpdate
	case 14:
		entry.ActivityName = "Rename"
	}
}

func mapSysmonInjection(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Process Activity"
	entry.ClassUID = ocsf.ClassProcessActivity
	entry.Severity = ocsf.SeverityHigh
	entry.SeverityID = ocsf.SeverityIDHigh

	eventID := entry.Unmapped["event_id"].(int)

	// 发起攻击的进程
	entry.Actor = &ocsf.User{Name: GetStr(unmapped, "SourceImage")}
	// 被注入/被读取的受害进程 (如 lsass.exe)
	entry.Process = &ocsf.Process{Name: GetStr(unmapped, "TargetImage")}

	if eventID == 8 {
		entry.ActivityName = "Create Remote Thread"
	} else if eventID == 10 {
		entry.ActivityName = "Process Accessed"
		entry.Unmapped["granted_access"] = GetStr(unmapped, "GrantedAccess") // 如 0x1FFFFF (全部权限)
	}
}
