package model

import (
	"time"

	"gorm.io/gorm"
)

// Rule - gorm.Model
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

	//   RuleType: alert(报警Rule) / forensic(ForensicsRule) / investigation(InvestigationRule)
	Type string `json:"type" gorm:"default:alert"`

	//   回溯Config（仅报警Rule使用）
	EnableBacktrace bool   `json:"enable_backtrace"`
	BacktraceCron   string `json:"backtrace_cron"`
	BacktraceStart  string `json:"backtrace_start"`
}

// RuleResponse - API Return，包含正确的 id 字段
type RuleResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Query       string    `json:"query"`
	Interval    string    `json:"interval"`
	Severity    string    `json:"severity"`
	Enabled     bool      `json:"enabled"`
	Version     int64     `json:"version"`
	AuthorID    uint      `json:"author_id"`
	Source      string    `json:"source"`
	Type        string    `json:"type"`
}

// ToResponse - Rule Convert为 RuleResponse
func (r *Rule) ToResponse() RuleResponse {
	return RuleResponse{
		ID:          r.ID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		Name:        r.Name,
		Description: r.Description,
		Query:       r.Query,
		Interval:    r.Interval,
		Severity:    r.Severity,
		Enabled:     r.Enabled,
		Version:     r.Version,
		AuthorID:    r.AuthorID,
		Source:      r.Source,
		Type:        r.Type,
	}
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
