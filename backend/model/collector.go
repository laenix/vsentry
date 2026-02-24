package model

import "gorm.io/gorm"

type CollectorType string

const (
	CollectorTypeWindows CollectorType = "windows"
	CollectorTypeLinux   CollectorType = "linux"
	CollectorTypeMacOS   CollectorType = "macos"
)

type CollectorTemplate struct {
	gorm.Model
	Name        string         `json:"name"`         // Template name
	Type        CollectorType  `json:"type"`         // windows/linux/macos
	Description string         `json:"description"`  // Description
	Icon        string         `json:"icon"`         // icon name
	Features    string         `json:"features"`     // JSON array of features
}

type CollectorConfig struct {
	gorm.Model
	Name           string         `json:"name"`           // Config name
	TemplateID     uint           `json:"template_id"`    // Template ID
	Type           CollectorType  `json:"type"`           // windows/linux/macos
	
	// Collection settings
	Channels       string         `json:"channels"`       // Comma-separated channels (e.g., "System,Application,Security")
	
	// Ingest settings
	IngestID       uint           `json:"ingest_id"`      // Linked Ingest ID
	IngestEndpoint string         `json:"endpoint"`       // Override endpoint
	Token          string         `json:"token"`          // Ingest token
	
	// Build settings
	StreamFields   string         `json:"stream_fields"`  // _stream_fields for VL
	IsEnabled      bool           `json:"is_enabled" gorm:"default:false"`
	
	// Build output
	BuildStatus    string         `json:"build_status"`   // pending/building/completed/failed
	BuildOutput    string         `json:"build_output"`   // Build log or download URL
}

// Predefined collector templates
var collectorTemplates = []CollectorTemplate{
	{
		Name:        "Windows Event Collector",
		Type:        CollectorTypeWindows,
		Description: "Collect Windows Event Logs (Application, System, Security)",
		Icon:        "windows",
		Features:    `["System","Application","Security","PowerShell","DNS","AD"]`,
	},
	{
		Name:        "Linux Syslog Collector",
		Type:        CollectorTypeLinux,
		Description: "Collect Linux syslog via syslog-ng or rsyslog",
		Icon:        "linux",
		Features:    `["auth","authpriv","cron","daemon","kern","lpr","mail","news","syslog","user","uucp","local0-7"]`,
	},
	{
		Name:        "macOS Unified Logging",
		Type:        CollectorTypeMacOS,
		Description: "Collect macOS unified logging via log stream",
		Icon:        "apple",
		Features:    `["system","install","network","wifi","configuration","opendirectory","location"]`,
	},
}