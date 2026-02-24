package model

import "gorm.io/gorm"

type CustomTable struct {
	gorm.Model
	Name        string `json:"name"`        // Table name display
	StreamFields string `json:"stream_fields"` // _stream_fields value, e.g., "host,service"
	Description string `json:"description"` // Table description
	Query       string `json:"query"`       // Default query for this table
	IsActive    bool   `json:"is_active" gorm:"default:true"`
}