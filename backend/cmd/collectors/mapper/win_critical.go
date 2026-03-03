package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 1102: 审计日志被清除 (极度高危，掩盖痕迹)
	// 4719: 审计策略被更改 (高危，关闭监控)
	// 4672: 特殊权限分配 (高危，提权)
	// 4765, 4766: SID History 操作 (跨域提权后门)
	// 4794: 尝试设置 DSRM (目录服务恢复模式) 密码
	Register([]int{1102, 4719, 4672, 4765, 4766, 4794}, mapCriticalEvents)
}

func mapCriticalEvents(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	eventID := entry.Unmapped["event_id"].(int)

	// 这些高危操作通常都有一个确定的执行者 (Actor)
	actorName := GetStr(unmapped, "SubjectUserName")
	if actorName != "" {
		entry.Actor = &ocsf.User{
			Name:   actorName,
			Domain: GetStr(unmapped, "SubjectDomainName"),
		}
	}

	switch eventID {
	case 1102: // 日志被清空
		entry.CategoryName = ocsf.CategoryFindings
		entry.ClassName = "Security Finding"
		entry.ClassUID = ocsf.ClassSecurityFinding
		entry.ActivityName = "Audit Log Cleared"
		entry.Severity = ocsf.SeverityCritical
		entry.SeverityID = ocsf.SeverityIDCritical

	case 4719: // 审计策略更改
		entry.CategoryName = ocsf.CategorySystem
		entry.ClassName = "System Activity"
		entry.ClassUID = 1000
		entry.ActivityName = "Audit Policy Changed"
		entry.Severity = ocsf.SeverityCritical
		entry.SeverityID = ocsf.SeverityIDCritical

	case 4672: // 敏感权限分配 (如 SeDebugPrivilege)
		entry.CategoryName = ocsf.CategoryIdentity
		entry.ClassName = "Authorization"
		entry.ClassUID = ocsf.ClassAuthorization
		entry.ActivityName = "Special Privileges Assigned"
		entry.Severity = ocsf.SeverityHigh
		entry.SeverityID = ocsf.SeverityIDHigh
		// 将分配的具体权限列表提取到外层，方便搜索
		entry.Unmapped["privileges"] = GetStr(unmapped, "PrivilegeList")

	case 4765, 4766: // SID History 注入
		entry.CategoryName = ocsf.CategoryIdentity
		entry.ClassName = "Account Change"
		entry.ClassUID = ocsf.ClassAccountChange
		entry.ActivityName = "SID History Added"
		entry.Severity = ocsf.SeverityCritical
		entry.SeverityID = ocsf.SeverityIDCritical
		entry.TargetUser = &ocsf.User{Name: GetStr(unmapped, "TargetUserName")}

	case 4794: // 设置 DSRM 密码
		entry.CategoryName = ocsf.CategoryIdentity
		entry.ClassName = "Account Change"
		entry.ClassUID = ocsf.ClassAccountChange
		entry.ActivityName = "DSRM Password Set Attempt"
		entry.Severity = ocsf.SeverityCritical
		entry.SeverityID = ocsf.SeverityIDCritical
	}
}
