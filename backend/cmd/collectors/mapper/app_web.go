package mapper

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/laenix/vsentry/pkg/ocsf"
)

var (
	// 完美匹配 Nginx / Apache 的 Combined 格式
	// 提取: 1:IP 2:Ident 3:Auth 4:Time 5:Method 6:URL 7:Protocol 8:Status 9:Bytes 10:Referer 11:User-Agent
	webAccessRe = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+\[([^\]]+)\]\s+"(\S+)\s+(.+?)\s+(HTTP/[0-9.]+)"\s+(\d{3})\s+(\d+|-)\s+"([^"]*)"\s+"([^"]*)"`)

	// Nginx ErrorLog格式: 2026/03/03 10:00:00 [error] 12345#0: *6789 open() failed...
	nginxErrorRe = regexp.MustCompile(`^\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+\[(\w+)\]\s+(.*)$`)

	// Apache ErrorLog格式: [Tue Mar 03 10:00:00.123 2026] [core:error] [pid 1234] [client 1.2.3.4:5678] message
	apacheErrorRe = regexp.MustCompile(`^\[[^\]]+\]\s+\[(?:[^:]+:)?(\w+)\]\s+\[pid\s+\d+(?::tid\s+\d+)?\]\s+(?:\[client\s+([^\]]+)\])?\s+(.*)$`)
)

func init() {
	RegisterText([]string{"nginx_access", "apache_access"}, mapWebAccess)
	RegisterText([]string{"nginx_error"}, mapNginxError)
	RegisterText([]string{"apache_error"}, mapApacheError)
}

func mapWebAccess(line string, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryNetwork
	entry.ClassName = "HTTP Activity"
	entry.ClassUID = 4002 // OCSF HTTP Activity

	matches := webAccessRe.FindStringSubmatch(line)
	if len(matches) >= 12 {
		entry.SrcEndpoint = &ocsf.Endpoint{IP: matches[1]}

		// 提取 HTTP Request的详细指标 (猎杀 Web 攻击的核心)
		entry.Unmapped["http_request"] = map[string]interface{}{
			"http_method": matches[5],
			"url":         matches[6],
			"protocol":    matches[7],
			"referrer":    matches[10],
			"user_agent":  matches[11],
		}

		status, _ := strconv.Atoi(matches[8])
		entry.Unmapped["http_response"] = map[string]interface{}{
			"code": status,
		}

		if matches[9] != "-" {
			bytes, _ := strconv.Atoi(matches[9])
			entry.Unmapped["bytes_out"] = bytes
		}

		// --- 威胁定性Engine ---
		url := strings.ToLower(matches[6])
		ua := strings.ToLower(matches[11])

		// 1. 拦截明显的 Web 攻击特征 (SQLi, XSS, 目录穿越, 敏感File读取)
		if strings.Contains(url, "../") || strings.Contains(url, "%2e%2e") ||
			strings.Contains(url, "union select") || strings.Contains(url, "/etc/passwd") ||
			strings.Contains(ua, "curl/") || strings.Contains(ua, "nmap") || strings.Contains(ua, "sqlmap") {

			entry.CategoryName = ocsf.CategoryFindings
			entry.ClassName = "Security Finding"
			entry.ClassUID = 2001
			entry.ActivityName = "Possible Web Attack / Scanning"
			entry.Severity = ocsf.SeverityHigh
			entry.SeverityID = ocsf.SeverityIDHigh

			// 2. 常规Status码分Class
		} else if status >= 500 {
			entry.Severity = ocsf.SeverityMedium
			entry.SeverityID = ocsf.SeverityIDMedium
			entry.ActivityName = "Server Error (5xx)"
		} else if status >= 400 {
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
			entry.ActivityName = "Client Error (4xx)"
		} else {
			entry.Severity = ocsf.SeverityInfo
			entry.SeverityID = ocsf.SeverityIDInfo
			entry.ActivityName = "Web Access"
		}
	}
}

func mapNginxError(line string, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryApp
	entry.ClassName = "Application Error"
	entry.ClassUID = 1000

	matches := nginxErrorRe.FindStringSubmatch(line)
	if len(matches) >= 3 {
		logLevel := matches[1]
		entry.Message = matches[2]
		mapWebErrorLevel(logLevel, entry)
	}
}

func mapApacheError(line string, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryApp
	entry.ClassName = "Application Error"
	entry.ClassUID = 1000

	matches := apacheErrorRe.FindStringSubmatch(line)
	if len(matches) >= 4 {
		logLevel := matches[1] // 如 error, warn
		clientIP := matches[2] // 如 192.168.1.100:54321
		entry.Message = matches[3]

		if clientIP != "" {
			ipParts := strings.Split(clientIP, ":")
			entry.SrcEndpoint = &ocsf.Endpoint{IP: ipParts[0]}
		}

		mapWebErrorLevel(logLevel, entry)
	}
}

// 统一的 Web Error等级映射器
func mapWebErrorLevel(level string, entry *ocsf.VSentryOCSFEvent) {
	switch level {
	case "emerg", "alert", "crit", "error":
		entry.Severity = ocsf.SeverityHigh
		entry.SeverityID = ocsf.SeverityIDHigh
		entry.ActivityName = "Critical Web Error"
	case "warn":
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
		entry.ActivityName = "Web Warning"
	default:
		entry.Severity = ocsf.SeverityInfo
		entry.SeverityID = ocsf.SeverityIDInfo
		entry.ActivityName = "Web Notice"
	}
}
