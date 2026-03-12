package mapper

import (
	"regexp"
	"strings"

	"github.com/laenix/vsentry/pkg/ocsf"
)

var (
	// MySQL 8.x ErrorLog格式: 2026-10-25T10:00:00.123456Z 0 [ERROR] [MY-012345] [Server] message
	mysqlErrRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+\d+\s+\[([^\]]+)\]\s+\[([^\]]+)\]\s+\[([^\]]+)\]\s+(.*)$`)

	// Redis Log格式: 12345:M 25 Oct 2026 10:00:00.123 * message
	// PID:Role (M=Master, S=Slave, C=Child) Date Time Level (., -, *, #) Message
	redisLogRe = regexp.MustCompile(`^(\d+):([A-Z])\s+(\d{1,2}\s+[a-zA-Z]{3}\s+\d{4}\s+\d{2}:\d{2}:\d{2}\.\d+)\s+([.\-*#])\s+(.*)$`)
)

func init() {
	RegisterText([]string{"mysql_error"}, mapMysqlError)
	RegisterText([]string{"redis_log"}, mapRedisLog)
}

func mapMysqlError(line string, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryApp
	entry.ClassName = "Database Activity"
	entry.ClassUID = 1000

	matches := mysqlErrRe.FindStringSubmatch(line)
	if len(matches) >= 6 {
		level := strings.TrimSpace(matches[2])
		errCode := matches[3]
		subsystem := matches[4]
		entry.Message = matches[5]

		entry.Unmapped["db_error_code"] = errCode
		entry.Unmapped["db_subsystem"] = subsystem

		switch level {
		case "ERROR", "Fatal":
			entry.Severity = ocsf.SeverityHigh
			entry.SeverityID = ocsf.SeverityIDHigh
			entry.ActivityName = "Database Error"
		case "Warning":
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
			entry.ActivityName = "Database Warning"
		default:
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
		}

		// Monitor暴破或PermissionReject
		if strings.Contains(entry.Message, "Access denied for user") {
			entry.CategoryName = ocsf.CategoryIdentity
			entry.ClassName = "Authentication"
			entry.ClassUID = ocsf.ClassAuthentication
			entry.ActivityName = ocsf.ActionLogonFailed
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
		}
	}
}

func mapRedisLog(line string, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryApp
	entry.ClassName = "Database Activity"
	entry.ClassUID = 1000

	matches := redisLogRe.FindStringSubmatch(line)
	if len(matches) >= 6 {
		// matches[1]=PID, matches[2]=Role, matches[4]=Level, matches[5]=Message
		entry.Process = &ocsf.Process{
			Name: "redis-server",
		}

		// RoleParse：M=Master, S=Slave, C=Child, X=Sentinel
		role := matches[2]
		entry.Unmapped["redis_role"] = role

		levelIndicator := matches[4]
		entry.Message = matches[5]

		// Redis Log等级符号: . (Debug), - (Verbose), * (Notice), # (Warning)
		switch levelIndicator {
		case "#":
			entry.Severity = ocsf.SeverityHigh
			entry.SeverityID = ocsf.SeverityIDHigh
			entry.ActivityName = "Redis Warning / Error"
		case "*":
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
			entry.ActivityName = "Redis Notice"
		default:
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
		}
	}
}
