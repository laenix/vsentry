package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 4768: Kerberos TGT 请求 (登录域)
	// 4769: Kerberos 服务票据请求 (访问域内资源)
	// 4771: Kerberos 预认证失败 (密码错误/爆破)
	// 4776: NTLM 凭据验证 (哈希传递攻击)
	Register([]int{4768, 4769, 4771, 4776}, mapKerberosEvents)
}

func mapKerberosEvents(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryIdentity
	entry.ClassName = "Authentication"
	entry.ClassUID = ocsf.ClassAuthentication

	eventID := entry.Unmapped["event_id"].(int)

	// IP 地址在不同事件里字段名不同
	ip := GetStr(unmapped, "IpAddress")
	if ip == "" {
		ip = GetStr(unmapped, "Workstation") // 4776 专属
	}
	if ip != "" && ip != "-" {
		entry.SrcEndpoint = &ocsf.Endpoint{IP: ip}
	}

	// 被请求的用户
	userName := GetStr(unmapped, "TargetUserName")
	if userName == "" {
		userName = GetStr(unmapped, "TargetAccountName") // 4776 专属
	}
	entry.TargetUser = &ocsf.User{Name: userName}

	switch eventID {
	case 4768:
		entry.ActivityName = "Kerberos TGT Request"
		status := GetStr(unmapped, "Status")
		if status == "0x0" {
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
		} else {
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
		}

	case 4769:
		entry.ActivityName = "Kerberos Service Ticket Request"
		entry.Severity = ocsf.SeverityInfo
		entry.SeverityID = ocsf.SeverityIDInfo
		entry.Unmapped["service_name"] = GetStr(unmapped, "ServiceName")

	case 4771:
		entry.ActivityName = ocsf.ActionLogonFailed // 预认证失败通常意味着密码错误或域账号爆破
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
		entry.Unmapped["failure_code"] = GetStr(unmapped, "Status")

	case 4776:
		entry.ActivityName = "NTLM Credential Validation"
		status := GetStr(unmapped, "Status") // 0x0 成功, 0xC000006A 密码错误
		if status == "0x0" {
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
		} else {
			entry.ActivityName = ocsf.ActionLogonFailed
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
			entry.Unmapped["failure_code"] = status
		}
	}
}
