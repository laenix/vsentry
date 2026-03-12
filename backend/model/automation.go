package model

import (
	"time"

	"gorm.io/datatypes" // Need go get gorm.io/datatypes
	"gorm.io/gorm"
)

// Playbook 结构体增强
type Playbook struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active" gorm:"default:false"`
	// 触发Type修改为枚举支持
	// "manual", "incident", "timer"
	TriggerType string         `json:"trigger_type"`
	Definition  datatypes.JSON `json:"definition"`

	// 多对多关联：一个Playbook可以绑定多个Rule，一个Rule也可以触发多个Playbook
	Rules []Rule `json:"rules" gorm:"many2many:rule_playbooks;"`
}

// PlaybookExecution Execute历史
type PlaybookExecution struct {
	gorm.Model
	PlaybookID       uint   `json:"playbook_id"`
	Status           string `json:"status"`             // "running", "success", "failed"
	TriggerContextID uint   `json:"trigger_context_id"` // 关联的 Incident ID

	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int64     `json:"duration_ms"` // 毫seconds

	// 记录every个节点的详细Log (Input、Output、Error)
	// 格式: { "node_1": { "status": "success", "output": ... } }
	Logs datatypes.JSON `json:"logs"`
}
