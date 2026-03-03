package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// Windows Defender Operational Logs
	Register([]int{1116, 1117}, mapDefender)
}

func mapDefender(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryFindings
	entry.ClassName = "Security Finding"
	entry.ClassUID = 2001 // OCSF Security Finding
	entry.Severity = ocsf.SeverityCritical
	entry.SeverityID = ocsf.SeverityIDCritical

	eventID := entry.Unmapped["event_id"].(int)

	threatName := GetStr(unmapped, "Threat Name")
	targetFile := GetStr(unmapped, "Path")

	if eventID == 1116 {
		entry.ActivityName = "Malware Detected"
	} else if eventID == 1117 {
		entry.ActivityName = "Malware Remediation Action"
	}

	// 将威胁名称和感染路径提取出来
	entry.Unmapped["threat_name"] = threatName
	entry.Unmapped["infected_path"] = targetFile
}
