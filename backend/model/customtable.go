package model

import "gorm.io/gorm"

type CustomTable struct {
	gorm.Model
	Name        string `json:"name"`        // Table - display
	StreamFields string `json:"stream_fields"` // _stream_fields - , e.g., "host,service"
	Description string `json:"description"` // Table - Query       string `json:"query"`       // Default - for this table
	IsActive    bool   `json:"is_active" gorm:"default:true"`
}