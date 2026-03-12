package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	//   4104: PowerShell 脚本块Execute (最核心的无File攻击检测点)
	Register([]int{4104}, mapPowerShell)
}

func mapPowerShell(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "Process Activity" //   或者自定义 "Script Activity"
	entry.ClassUID = ocsf.ClassProcessActivity
	entry.ActivityName = "Execute Script"

	// PowerShell - 最核心的是 ScriptBlockText，包含原始Execute的脚本内容
	scriptContent := GetStr(unmapped, "ScriptBlockText")

	entry.Process = &ocsf.Process{
		Name:    "powershell.exe",
		CmdLine: scriptContent, // 脚本内容放进 - 以便检索
	}
}
