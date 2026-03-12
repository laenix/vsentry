package model

import (
	"time"

	"gorm.io/gorm"
)

// model/incident.go
type Incident struct {
	gorm.Model
	RuleID     uint      `json:"rule_id"`
	Name       string    `json:"name"`
	Severity   string    `json:"severity"`
	Status     string    `json:"status"` // New, Acknowledged, Resolved
	AlertCount int       `json:"alert_count"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`

	// 处置字段
	Assignee              uint   `json:"assignee"`
	ClosingClassification string `json:"closing_classification"`
	ClosingComment        string `json:"closing_comment"`

	// 关联的Rule（用于GetRuleType）
	Rule Rule `json:"rule,omitempty" gorm:"foreignKey:RuleID"`

	// 核心：一对多关联。ago端在拿到 Incident Detail时，可以直接拿到这个数Group
	Alerts []Alert `json:"alerts" gorm:"foreignKey:IncidentID"`
}

// model/alert.go
type Alert struct {
	gorm.Model
	IncidentID  uint   `json:"incident_id"` // 外键
	RuleID      uint   `json:"rule_id"`
	Content     string `json:"content"` // Storage VictoriaLogs 搜出的原始 JSON Data
	Fingerprint string `gorm:"uniqueIndex" json:"fingerprint"`
}
