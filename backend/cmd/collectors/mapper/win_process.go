package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	// 4688: Windows原生ProcessCreate, 1: SysmonProcessCreate
	Register([]int{4688, 1}, mapProcessCreation)
	// 4689: WindowsProcess终止, 5: SysmonProcess终止
	Register([]int{4689, 5}, mapProcessTermination)
}

func mapProcessCreation(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Process Activity"
	entry.ClassUID = ocsf.ClassProcessActivity
	entry.ActivityName = ocsf.ActionCreate

	// 兼容 Windows 4688 和 Sysmon 1 的字段差异
	procName := GetStr(unmapped, "NewProcessName")
	if procName == "" {
		procName = GetStr(unmapped, "Image")
	}

	parentProc := GetStr(unmapped, "ParentProcessName")
	if parentProc == "" {
		parentProc = GetStr(unmapped, "ParentImage")
	}

	entry.Process = &ocsf.Process{
		Name:    procName,
		CmdLine: GetStr(unmapped, "CommandLine"),
		// 【核心修复】：字段名是 Parent，虽然序列化成 JSON 后它会变成 parent_process
		Parent: &ocsf.Process{
			Name: parentProc,
		},
	}

	actorName := GetStr(unmapped, "SubjectUserName")
	if actorName == "" {
		actorName = GetStr(unmapped, "User")
	}
	entry.Actor = &ocsf.User{Name: actorName}
}

func mapProcessTermination(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Process Activity"
	entry.ClassUID = ocsf.ClassProcessActivity
	entry.ActivityName = ocsf.ActionTerminate // 顺手把 "Terminate" 改为常量 ocsf.ActionTerminate 更规范

	procName := GetStr(unmapped, "ProcessName")
	if procName == "" {
		procName = GetStr(unmapped, "Image")
	}

	entry.Process = &ocsf.Process{
		Name: procName,
	}
}
