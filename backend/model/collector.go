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
	Name        string         `json:"name"`         // Template - Type        CollectorType  `json:"type"`         //   windows/linux/macos
	Description string         `json:"description"`  // Description - string         `json:"icon"`         // icon - Features    string         `json:"features"`     // JSON - of features
}

// CollectorSource - a single data source for collection
type CollectorSource struct {
	Type     string `json:"type"`     //   syslog, nginx_access, nginx_error, ssh, custom
	Path     string `json:"path"`     // file - Format   string `json:"format"`   //   syslog, nginx, apache, ssh, custom
	Enabled  bool   `json:"enabled"`  // whether - collect
}

type CollectorConfig struct {
	gorm.Model
	Name           string         `json:"name"`           // Config - TemplateID     uint           `json:"template_id"`    // Template - Type           CollectorType  `json:"type"`           //   windows/linux/macos
	
	// Collection - - JSON array of sources
	Sources        string         `json:"sources"`        // JSON - of CollectorSource
	
	// Legacy - for Windows (comma-separated channels)
	Channels       string         `json:"channels"`       //   Comma-separated channels (e.g., "System,Application,Security")
	
	// Additional - Interval       int            `json:"interval"`       // Collection - in seconds
	
	// Ingest - IngestID       uint           `json:"ingest_id"`      // Linked - ID
	IngestEndpoint string         `json:"endpoint"`       // Override - Token          string         `json:"token"`          // Ingest - // Build - StreamFields   string         `json:"stream_fields"`  // _stream_fields - VL
	IsEnabled      bool           `json:"is_enabled" gorm:"default:false"`
	
	// Build - BuildStatus    string         `json:"build_status"`   //   pending/building/completed/failed
	BuildOutput    string         `json:"build_output"`   // Build - or download URL
}

// Predefined - data sources
var LinuxDataSources = []CollectorSource{
	{Type: "syslog", Path: "/var/log/syslog", Format: "syslog", Enabled: true},
	{Type: "auth", Path: "/var/log/auth.log", Format: "auth", Enabled: false},
	{Type: "secure", Path: "/var/log/secure", Format: "ssh", Enabled: false},
	{Type: "nginx_access", Path: "/var/log/nginx/access.log", Format: "nginx_access", Enabled: false},
	{Type: "nginx_error", Path: "/var/log/nginx/error.log", Format: "nginx_error", Enabled: false},
	{Type: "apache_access", Path: "/var/log/apache2/access.log", Format: "apache", Enabled: false},
	{Type: "kern", Path: "/var/log/kern.log", Format: "syslog", Enabled: false},
	{Type: "messages", Path: "/var/log/messages", Format: "syslog", Enabled: false},
}

// Predefined - data sources (channels)
var WindowsDataSources = []CollectorSource{
	{Type: "System", Path: "System", Format: "windows_event", Enabled: true},
	{Type: "Application", Path: "Application", Format: "windows_event", Enabled: true},
	{Type: "Security", Path: "Security", Format: "windows_event", Enabled: true},
	{Type: "PowerShell", Path: "Microsoft-Windows-PowerShell/Operational", Format: "windows_event", Enabled: false},
	{Type: "DNS", Path: "DNS", Format: "windows_event", Enabled: false},
	{Type: "Sysmon", Path: "Microsoft-Windows-Sysmon/Operational", Format: "windows_event", Enabled: false},
}

// Predefined - templates
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