package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 登录/注销事件
	Register([]int{4624, 4625, 4634, 4648}, mapAuthentication)

	// 账号管理事件 (创建、启用、禁用、删除、改密)
	Register([]int{4720, 4722, 4723, 4724, 4725, 4726, 4738}, mapAccountManagement)

	// 组管理 (删除组、加人、踢人)
	Register([]int{4730, 4732, 4733}, mapGroupManagement)

	// 账号锁定 (防爆破核心)
	Register([]int{4740}, mapAccountLockout)
}

func mapAuthentication(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryIdentity
	entry.ClassName = "Authentication"
	entry.ClassUID = ocsf.ClassAuthentication

	eventID := entry.Unmapped["event_id"].(int)

	// 提取源 IP 和端口
	ip := GetStr(unmapped, "IpAddress")
	if ip != "" && ip != "-" {
		entry.SrcEndpoint = &ocsf.Endpoint{IP: ip}
		if port := GetInt(unmapped, "IpPort"); port > 0 {
			entry.SrcEndpoint.Port = port
		}
	}

	// 提取目标用户 (必须使用 TargetUser 避免与 Endpoint 冲突)
	entry.TargetUser = &ocsf.User{
		Name:   GetStr(unmapped, "TargetUserName"),
		Domain: GetStr(unmapped, "TargetDomainName"),
	}

	// 细分活动类型和严重等级
	switch eventID {
	case 4625: // 登录失败
		entry.ActivityName = ocsf.ActionLogonFailed
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
		entry.Unmapped["failure_status"] = GetStr(unmapped, "Status") // 记录具体的失败原因，如 0xC000006A
	case 4624, 4648: // 登录成功 / 使用显式凭据登录
		entry.ActivityName = ocsf.ActionLogon
		entry.Unmapped["logon_type"] = GetStr(unmapped, "LogonType") // 极其重要：区分本地还是 RDP
	case 4634: // 注销
		entry.ActivityName = ocsf.ActionLogoff
	}
}

func mapAccountManagement(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryIdentity
	entry.ClassName = "Account Change"
	entry.ClassUID = 3001 // OCSF Account Change Class

	eventID := entry.Unmapped["event_id"].(int)

	// 被操作的目标账号
	entry.TargetUser = &ocsf.User{
		Name:   GetStr(unmapped, "TargetUserName"),
		Domain: GetStr(unmapped, "TargetDomainName"),
	}

	// 执行操作的管理员/系统账号
	entry.Actor = &ocsf.User{
		Name:   GetStr(unmapped, "SubjectUserName"),
		Domain: GetStr(unmapped, "SubjectDomainName"),
	}

	switch eventID {
	case 4720: // 创建用户
		entry.ActivityName = ocsf.ActionCreate
		entry.Severity = ocsf.SeverityHigh
		entry.SeverityID = ocsf.SeverityIDHigh
	case 4722: // 启用用户
		entry.ActivityName = "Enable"
		entry.Severity = ocsf.SeverityHigh
		entry.SeverityID = ocsf.SeverityIDHigh
	case 4725: // 禁用用户
		entry.ActivityName = "Disable"
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	case 4726: // 删除用户
		entry.ActivityName = "Delete"
		entry.Severity = ocsf.SeverityHigh
		entry.SeverityID = ocsf.SeverityIDHigh
	case 4723, 4724: // 修改密码
		entry.ActivityName = "Password Change"
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	case 4738: // 账户变更
		entry.ActivityName = ocsf.ActionUpdate
	}
}

func mapGroupManagement(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryIdentity
	entry.ClassName = "Account Change"
	entry.ClassUID = 3001

	eventID := entry.Unmapped["event_id"].(int)

	// 执行操作的管理员
	entry.Actor = &ocsf.User{
		Name:   GetStr(unmapped, "SubjectUserName"),
		Domain: GetStr(unmapped, "SubjectDomainName"),
	}

	// 被操作的人 (如被拉入管理员组的用户)
	entry.TargetUser = &ocsf.User{
		Name: GetStr(unmapped, "MemberName"),
	}

	// 被操作的组
	entry.Unmapped["group_name"] = GetStr(unmapped, "TargetUserName")

	switch eventID {
	case 4730:
		entry.ActivityName = "Group Deleted"
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	case 4732:
		entry.ActivityName = "Member Added to Group"
		entry.Severity = ocsf.SeverityHigh // 加入组（尤其是特权组）属于高危关注点
		entry.SeverityID = ocsf.SeverityIDHigh
	case 4733:
		entry.ActivityName = "Member Removed from Group"
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	}
}

func mapAccountLockout(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryIdentity
	entry.ClassName = "Authentication"
	entry.ClassUID = ocsf.ClassAuthentication
	entry.ActivityName = "Account Locked"
	entry.Severity = ocsf.SeverityHigh // 账号被锁说明正在遭遇高频爆破，严重性必须设为 High
	entry.SeverityID = ocsf.SeverityIDHigh

	entry.TargetUser = &ocsf.User{
		Name:   GetStr(unmapped, "TargetUserName"),
		Domain: GetStr(unmapped, "TargetDomainName"),
	}

	entry.Actor = &ocsf.User{
		Name:   GetStr(unmapped, "SubjectUserName"),
		Domain: GetStr(unmapped, "SubjectDomainName"),
	}
}
