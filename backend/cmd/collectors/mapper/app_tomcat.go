package mapper

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/laenix/vsentry/pkg/ocsf"
)

var (
	// Tomcat - Access Log 格式: %h %l %u %t "%r" %s %b
	//   Example: 192.168.1.100 - - [25/Oct/2026:10:00:00 +0800] "GET /api/v1/users HTTP/1.1" 200 1024
	tomcatAccessRe = regexp.MustCompile(`^(\S+)\s+\S+\s+\S+\s+\[([^\]]+)\]\s+"(\S+)\s+(\S+)\s+\S+"\s+(\d+)\s+(\d+|-)`)

	// Tomcat - RunLog格式
	//   Example: 25-Oct-2026 10:00:00.123 INFO [main] org.apache.catalina.startup.Catalina.start Server startup in [1234] milliseconds
	tomcatCatalinaRe = regexp.MustCompile(`^(\d{2}-[a-zA-Z]{3}-\d{4}\s+\d{2}:\d{2}:\d{2}\.\d{3})\s+(SEVERE|WARNING|INFO|CONFIG|FINE|FINER|FINEST)\s+\[([^\]]+)\]\s+(\S+)\s+(.*)$`)
)

func init() {
	RegisterText([]string{"tomcat_access"}, mapTomcatAccess)
	RegisterText([]string{"tomcat_catalina", "java_app"}, mapTomcatCatalina)
}

func mapTomcatAccess(line string, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryNetwork
	entry.ClassName = "HTTP Activity"
	entry.ClassUID = 4002 // OCSF - Activity

	matches := tomcatAccessRe.FindStringSubmatch(line)
	if len(matches) >= 7 {
		entry.SrcEndpoint = &ocsf.Endpoint{IP: matches[1]}

		entry.Unmapped["http_request"] = map[string]interface{}{
			"http_method": matches[3],
			"url":         matches[4],
		}

		status, _ := strconv.Atoi(matches[5])
		entry.Unmapped["http_response"] = map[string]interface{}{
			"code": status,
		}

		bytesSent := matches[6]
		if bytesSent != "-" {
			size, _ := strconv.Atoi(bytesSent)
			entry.Unmapped["bytes_out"] = size
		}

		if status >= 500 {
			entry.Severity = ocsf.SeverityHigh
			entry.SeverityID = ocsf.SeverityIDHigh
			entry.ActivityName = "Server Error"
		} else if status >= 400 {
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
			entry.ActivityName = "Client Error"
		} else {
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
			entry.ActivityName = "Web Access"
		}
	}
}

func mapTomcatCatalina(line string, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryApp
	entry.ClassName = "Application Activity"
	entry.ClassUID = 1000

	matches := tomcatCatalinaRe.FindStringSubmatch(line)
	if len(matches) >= 6 {
		// Java - threadName := matches[3]
		// 抛出Log的 - 类名
		className := matches[4]
		entry.Process = &ocsf.Process{Name: threadName}
		entry.Unmapped["java_class"] = className

		entry.Message = matches[5]
		level := matches[2]

		switch level {
		case "SEVERE":
			entry.Severity = ocsf.SeverityCritical
			entry.SeverityID = ocsf.SeverityIDCritical
			entry.ActivityName = "Application Crash / Severe Error"
		case "WARNING":
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
			entry.ActivityName = "Application Warning"
		default:
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
		}

		// 检查是否有典型的 - 漏洞利用痕迹 (如 Log4j JNDI 注入尝试)
		if strings.Contains(line, "${jndi:") {
			entry.CategoryName = ocsf.CategoryFindings
			entry.ClassName = "Security Finding"
			entry.ClassUID = ocsf.ClassSecurityFinding
			entry.ActivityName = "Possible JNDI Injection Exploit"
			entry.Severity = ocsf.SeverityCritical
			entry.SeverityID = ocsf.SeverityIDCritical
		}
	}
}
