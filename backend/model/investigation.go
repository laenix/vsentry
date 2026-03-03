package model

import (
	"time"

	"gorm.io/gorm"
)

// InvestigationTemplate 定义了预置的调查模板
type InvestigationTemplate struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"` // 如: "同主机历史事件"
	Description string    `json:"description" gorm:"type:text"`           // 如: "查询该主机过去7天的所有事件"
	LogSQL      string    `json:"logsql" gorm:"type:text;not null"`       // 如: "observer.hostname=${hostname} AND _time:[${start_time}, ${end_time}]"
	Parameters  string    `json:"parameters" gorm:"type:json"`            // 记录需要哪些变量，如: ["hostname", "start_time", "end_time"]
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// InitDefaultInvestigationTemplates 初始化预置调查模板 (出厂默认值)
func InitDefaultInvestigationTemplates(db *gorm.DB) {
	var count int64
	db.Model(&InvestigationTemplate{}).Count(&count)

	// 如果表里已经有数据了，就不重复插入
	if count > 0 {
		return
	}

	defaultTemplates := []InvestigationTemplate{
		{
			Name:        "同主机历史事件",
			Description: "调查特定主机在事件发生前后的所有行为日志，用于时间线还原。",
			LogSQL:      `observer.hostname="${hostname}" AND _time:[${start_time}, ${end_time}]`,
			Parameters:  `["hostname", "start_time", "end_time"]`,
		},
		{
			Name:        "同用户活动轨迹",
			Description: "追踪特定用户（如被盗用的域账号）在全网的登录和操作记录。",
			LogSQL:      `(target_user.name="${username}" OR actor.user.name="${username}") AND _time:[${start_time}, ${end_time}]`,
			Parameters:  `["username", "start_time", "end_time"]`,
		},
		{
			Name:        "横向移动检测 (同源IP)",
			Description: "检测同一个攻击源 IP 在短时间内对内网其他主机的扫描或登录尝试。",
			LogSQL:      `src_endpoint.ip="${src_ip}" AND _time:[${start_time}, ${end_time}]`,
			Parameters:  `["src_ip", "start_time", "end_time"]`,
		},
		{
			Name:        "进程链与子进程回溯",
			Description: "输入可疑进程名，查询是谁启动了它（父进程），以及它又启动了什么子进程。",
			LogSQL:      `(process.name="${process_name}" OR process.parent.name="${process_name}") AND observer.hostname="${hostname}"`,
			Parameters:  `["process_name", "hostname"]`,
		},
		{
			Name:        "暴力破解历史溯源",
			Description: "统计该攻击源 IP 在过去一段时间内的所有认证失败记录。",
			LogSQL:      `src_endpoint.ip="${src_ip}" AND activity_name="Logon Failed" AND _time:[${start_time}, ${end_time}]`,
			Parameters:  `["src_ip", "start_time", "end_time"]`,
		},
	}

	db.Create(&defaultTemplates)
}
