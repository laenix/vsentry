package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	//   7045, 4697: Service安装 (常用于Permission维持或勒索软件投递)
	Register([]int{7045, 4697}, mapServiceInstallation)

	//   4698-4702: PlanTask全生命周期 (High级 APT 常用后门驻留方式)
	Register([]int{4698, 4699, 4700, 4701, 4702}, mapScheduledTask)
}

func mapServiceInstallation(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Service Activity"
	entry.ClassUID = 3004 // OCSF - Extension
	entry.ActivityName = ocsf.ActionCreate
	entry.Severity = ocsf.SeverityHigh
	entry.SeverityID = ocsf.SeverityIDHigh

	serviceName := GetStr(unmapped, "ServiceName")
	imagePath := GetStr(unmapped, "ImagePath")
	if imagePath == "" {
		imagePath = GetStr(unmapped, "ServiceFileName") // 4697 - 7045 字段名差异兼容
	}

	entry.Process = &ocsf.Process{
		Name:    serviceName,
		CmdLine: imagePath, // 将StartPath放在 - 以便快速检索恶意Path
	}

	entry.Actor = &ocsf.User{
		Name: GetStr(unmapped, "SubjectUserName"),
	}

	entry.Unmapped["service_type"] = GetStr(unmapped, "ServiceType")
	entry.Unmapped["service_start_type"] = GetStr(unmapped, "ServiceStartType")
}

func mapScheduledTask(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Scheduled Job Activity"
	entry.ClassUID = 3005 // OCSF - Task

	eventID := entry.Unmapped["event_id"].(int)

	entry.Process = &ocsf.Process{
		Name: GetStr(unmapped, "TaskName"),
	}

	entry.Actor = &ocsf.User{
		Name: GetStr(unmapped, "SubjectUserName"),
	}

	switch eventID {
	case 4698:
		entry.ActivityName = ocsf.ActionCreate
		entry.Severity = ocsf.SeverityHigh // CreatePlanTask需重点Audit - .SeverityID = ocsf.SeverityIDHigh
	case 4699:
		entry.ActivityName = "Delete"
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	case 4700:
		entry.ActivityName = "Enable"
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	case 4701:
		entry.ActivityName = "Disable"
		entry.Severity = ocsf.SeverityMedium
		entry.SeverityID = ocsf.SeverityIDMedium
	case 4702:
		entry.ActivityName = ocsf.ActionUpdate
		entry.Severity = ocsf.SeverityHigh //   Update现有Task（如将合法Task替换为恶意脚本）
		entry.SeverityID = ocsf.SeverityIDHigh
	}
}
