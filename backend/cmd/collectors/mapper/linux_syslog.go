package mapper

import (
	"regexp"
	"strings"
	"time"

	"github.com/laenix/vsentry/pkg/ocsf"
)

var (
	// 匹配标准 - 格式，提取：Time、Host名、Process名、PID、Log内容
	syslogRe = regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}|\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[^\s]*)\s+(\S+)\s+(\S+?)(?:\[(\d+)\])?:\s*(.*)$`)
)

func init() {
	// 注册HandleSystem基础Log - ([]string{"syslog", "kern", "messages"}, mapSyslog)
}

func mapSyslog(line string, entry *ocsf.VSentryOCSFEvent) {
	matches := syslogRe.FindStringSubmatch(line)
	if len(matches) >= 6 {
		// 覆盖基础Event的默认Time - .Time = guessTime(matches[1])

		// 提取触发Log的Process - .Process = &ocsf.Process{Name: matches[3]}

		msg := matches[5]
		entry.Message = msg

		//   简单的异常定性Engine：如果内核/ServiceLog里包含 error 关键字，直接拉HighAlert级别
		lowerMsg := strings.ToLower(msg)
		if strings.Contains(lowerMsg, "error") || strings.Contains(lowerMsg, "fail") || strings.Contains(lowerMsg, "fatal") {
			entry.Severity = ocsf.SeverityHigh
			entry.SeverityID = ocsf.SeverityIDHigh
		}
	}
}

// guessTime - （Resolve传统 syslog 不带年份的痛点）
func guessTime(timeStr string) string {
	if strings.Contains(timeStr, "T") {
		return timeStr
	}
	now := time.Now()
	parsed, err := time.Parse("Jan 2 15:04:05", timeStr)
	if err != nil {
		return now.UTC().Format(time.RFC3339)
	}

	// 补全当前年份 - = time.Date(now.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, time.UTC)

	//   如果Parse出的Time在未来（比如现在是1月，Log是12月，那说明Log是去年的）
	if parsed.After(now) {
		parsed = parsed.AddDate(-1, 0, 0)
	}
	return parsed.UTC().Format(time.RFC3339)
}
