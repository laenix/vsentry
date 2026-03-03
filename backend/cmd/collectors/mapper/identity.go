package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 注册登录/注销事件
	Register([]int{4624, 4625, 4634, 4648}, mapAuthentication)

	// 注册账号管理事件 (创建、启用、禁用、删除、改密)
	Register([]int{4720, 4722, 4723, 4724, 4725, 4726, 4738}, mapAccountManagement)
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

	// 【核心修复】：将 Target 改为 TargetUser
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
	case 4624, 4648: // 登录成功 / 使用显式凭据登录
		entry.ActivityName = ocsf.ActionLogon
	case 4634: // 注销
		entry.ActivityName = ocsf.ActionLogoff
	}
}

func mapAccountManagement(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryIdentity
	entry.ClassName = "Account Change"
	entry.ClassUID = 3001 // OCSF Account Change Class

	eventID := entry.Unmapped["event_id"].(int)

	// 【核心修复】：将 Target 改为 TargetUser，表示被操作的目标账号
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
	case 4722: // 启用用户
		entry.ActivityName = "Enable"
	case 4725: // 禁用用户
		entry.ActivityName = "Disable"
	case 4726: // 删除用户
		entry.ActivityName = "Delete"
	case 4723, 4724: // 修改密码
		entry.ActivityName = "Password Change"
	}
}
