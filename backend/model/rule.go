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
