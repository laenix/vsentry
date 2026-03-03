package mapper

import (
	"github.com/laenix/vsentry/pkg/ocsf"
)

func init() {
	Register([]int{4663, 11, 23}, mapFileActivity)
}

func mapFileActivity(unmapped map[string]interface{}, entry *ocsf.VSentryOCSFEvent) {
	entry.CategoryName = ocsf.CategorySystem
	entry.ClassName = "File Activity"
	entry.ClassUID = 1001 // OCSF File Activity

	eventID := entry.Unmapped["event_id"].(int)

	filePath := GetStr(unmapped, "ObjectName")
	if filePath == "" {
		filePath = GetStr(unmapped, "TargetFilename") // Sysmon
	}

	entry.Unmapped["file_path"] = filePath

	procName := GetStr(unmapped, "ProcessName")
	if procName == "" {
		procName = GetStr(unmapped, "Image")
	}
	entry.Process = &ocsf.Process{Name: procName}

	actor := GetStr(unmapped, "SubjectUserName")
	if actor == "" {
		actor = GetStr(unmapped, "User")
	}
	entry.Actor = &ocsf.User{Name: actor}

	switch eventID {
	case 4663:
		accessMask := GetStr(unmapped, "AccessMask")
		// 0x2 写入, 0x1 读取, 0x10000 删除
		if accessMask == "0x2" || accessMask == "0x6" {
			entry.ActivityName = "Write"
		} else if accessMask == "0x10000" {
			entry.ActivityName = "Delete"
		} else {
			entry.ActivityName = "Access"
		}
	case 11:
		entry.ActivityName = ocsf.ActionCreate
	case 23:
		entry.ActivityName = "Delete"
	}
}
