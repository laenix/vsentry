package mapper

import (
	"regexp"
	"strings"

	"github.com/laenix/vsentry/pkg/ocsf"
)

var (
	sshIpRe   = regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	sshUserRe = regexp.MustCompile(`(?:for|user)\s+(?:invalid user\s+)?(\S+)\s+from`)
)

func init() {
	// RegisterHandle "auth" 和 "secure" Type的Log
	RegisterText([]string{"auth", "secure"}, mapSSHAuth)
}

func mapSSHAuth(line string, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategoryIdentity
	entry.ClassName = "Authentication"
	entry.ClassUID = ocsf.ClassAuthentication

	if entry.Metadata == nil {
		entry.Metadata = &ocsf.Metadata{}
	}
	entry.Metadata.Product = "sshd"

	// 行为定性
	if strings.Contains(line, "Accepted password") || strings.Contains(line, "session opened") || strings.Contains(line, "Accepted publickey") {
		entry.ActivityName = ocsf.ActionLogon
		entry.Severity = ocsf.SeverityInfo
		entry.SeverityID = ocsf.SeverityIDInfo
	} else if strings.Contains(line, "Failed password") || strings.Contains(line, "authentication failure") || strings.Contains(line, "Connection closed by authenticating user") {
		entry.ActivityName = ocsf.ActionLogonFailed
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	} else if strings.Contains(line, "BREAK-IN") {
		entry.ActivityName = "Possible Break-in"
		entry.Severity = ocsf.SeverityCritical
		entry.SeverityID = ocsf.SeverityIDCritical
	}

	// 实体提取
	if ipMatch := sshIpRe.FindStringSubmatch(line); len(ipMatch) > 1 {
		entry.SrcEndpoint = &ocsf.Endpoint{IP: ipMatch[1]}
	}

	if userMatch := sshUserRe.FindStringSubmatch(line); len(userMatch) > 1 {
		entry.TargetUser = &ocsf.User{Name: userMatch[1]}
	}
}
