package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 7045, 4697: 服务安装
	Register([]int{7045, 4697}, mapServiceInstallation)
	// 4698: 计划任务创建
	Register([]int{4698}, mapScheduledTask)
}

func mapServiceInstallation(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Service Activity"
	entry.ClassUID = 3004 // OCSF Service Extension
	entry.ActivityName = ocsf.ActionCreate
	entry.Severity = ocsf.SeverityHigh
	entry.SeverityID = ocsf.SeverityIDHigh

	serviceName := GetStr(unmapped, "ServiceName")
	imagePath := GetStr(unmapped, "ImagePath")
	if imagePath == "" {
		imagePath = GetStr(unmapped, "ServiceFileName") // 4697 字段不同
	}

	entry.Process = &ocsf.Process{
		Name:    serviceName,
		CmdLine: imagePath, // 将启动路径放在 CmdLine
	}

	entry.Actor = &ocsf.User{
		Name: GetStr(unmapped, "SubjectUserName"),
	}
}

func mapScheduledTask(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Scheduled Job Activity"
	entry.ClassUID = 3005 // OCSF Scheduled Task
	entry.ActivityName = ocsf.ActionCreate

	taskName := GetStr(unmapped, "TaskName")

	entry.Process = &ocsf.Process{
		Name: taskName,
	}
	entry.Actor = &ocsf.User{
		Name: GetStr(unmapped, "SubjectUserName"),
	}
}
