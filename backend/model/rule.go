package model

import "gorm.io/gorm"

type Rule struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Query       string `json:"query"`
	Interval    string `json:"interval"`
	Severity    string `json:"severity"`
	Enabled     bool   `json:"enabled"`
	Version     int64  `json:"version"`
	AuthorID    uint   `json:"author_id"`
	Source      string `json:"source"`

	// 规则类型: alert(报警规则) / forensic(取证规则) / investigation(调查规则)
	Type string `json:"type" gorm:"default:alert"`

	// 回溯配置（仅报警规则使用）
	EnableBacktrace bool   `json:"enable_backtrace"`
	BacktraceCron   string `json:"backtrace_cron"`
	BacktraceStart  string `json:"backtrace_start"` // 如 "1y", "30d", "2025-01-01"
}

type RuleTag struct {
	gorm.Model
	RuleID uint
	Tag    string `json:"tag"`
}

type RuleAutomation struct {
	gorm.Model
	RuleID       uint
	AutomationID uint
}
