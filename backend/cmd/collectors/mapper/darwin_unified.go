package mapper

import (
	"regexp"
	"strings"

	"github.com/laenix/vsentry/pkg/ocsf"
)

var (
	// 匹配 macOS 的 syslog 风格输出
	// 示例: 2026-03-03 10:00:00.123456+0800  localhost sudo[12345]: (pam_unix) session opened for user root
	darwinLogRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\.\d+[+-]\d{4})\s+(\S+)\s+([^\[:]+)(?:\[(\d+)\])?:\s*(.*)$`)
)

func init() {
	RegisterText([]string{"darwin_unified"}, mapDarwinUnified)
}

func mapDarwinUnified(line string, entry *ocsf.VSentryOCSFEvent) {
	matches := darwinLogRe.FindStringSubmatch(line)
	if len(matches) >= 6 {
		// 提取时间 (macOS 自带时区，非常精准)
		timeStr := strings.Replace(matches[1], " ", "T", 1)
		entry.Time = timeStr

		// 提取执行动作的进程 (如 sudo, loginwindow, CoreServices)
		processName := matches[3]
		entry.Process = &ocsf.Process{Name: processName}

		msg := matches[5]
		entry.Message = msg

		// ==========================================
		// macOS 威胁狩猎定性引擎
		// ==========================================
		lowerMsg := strings.ToLower(msg)

		if processName == "sudo" {
			entry.CategoryName = ocsf.CategoryIdentity
			entry.ClassName = "Authentication"
			entry.ClassUID = ocsf.ClassAuthentication

			if strings.Contains(lowerMsg, "incorrect password") {
				entry.ActivityName = ocsf.ActionLogonFailed
				entry.Severity = ocsf.SeverityMedium
				entry.SeverityID = ocsf.SeverityIDMedium
			} else {
				entry.ActivityName = "Privilege Escalation (sudo)"
				entry.Severity = ocsf.SeverityHigh
				entry.SeverityID = ocsf.SeverityIDHigh
			}

		} else if processName == "loginwindow" || processName == "authorizationhost" {
			entry.CategoryName = ocsf.CategoryIdentity
			entry.ClassName = "Authentication"
			entry.ClassUID = ocsf.ClassAuthentication

			if strings.Contains(lowerMsg, "auth failed") || strings.Contains(lowerMsg, "failed to authenticate") {
				entry.ActivityName = ocsf.ActionLogonFailed
				entry.Severity = ocsf.SeverityMedium
				entry.SeverityID = ocsf.SeverityIDMedium
			} else if strings.Contains(lowerMsg, "succeeded") {
				entry.ActivityName = ocsf.ActionLogon
			}

			// Apple Gatekeeper (防毒/恶意软件拦截)
		} else if processName == "syspolicy" || strings.Contains(lowerMsg, "malware") {
			entry.CategoryName = ocsf.CategoryFindings
			entry.ClassName = "Security Finding"
			entry.ClassUID = 2001
			entry.ActivityName = "Gatekeeper Blocked Execution"
			entry.Severity = ocsf.SeverityHigh
			entry.SeverityID = ocsf.SeverityIDHigh
		}
	}
}
