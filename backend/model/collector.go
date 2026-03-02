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

// CollectorSource defines a single data source for collection
type CollectorSource struct {
	Type     string `json:"type"`     // syslog, nginx_access, nginx_error, ssh, custom
	Path     string `json:"path"`     // file path
	Format   string `json:"format"`   // syslog, nginx, apache, ssh, custom
	Enabled  bool   `json:"enabled"`  // whether to collect
}

type CollectorConfig struct {
	gorm.Model
	Name           string         `json:"name"`           // Config name
	TemplateID     uint           `json:"template_id"`    // Template ID
	Type           CollectorType  `json:"type"`           // windows/linux/macos
	
	// Collection settings - JSON array of sources
	Sources        string         `json:"sources"`        // JSON array of CollectorSource
	
	// Legacy field for Windows (comma-separated channels)
	Channels       string         `json:"channels"`       // Comma-separated channels (e.g., "System,Application,Security")
	
	// Additional settings
	Interval       int            `json:"interval"`       // Collection interval in seconds
	
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

// Predefined Linux data sources
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

// Predefined Windows data sources (channels)
var WindowsDataSources = []CollectorSource{
	{Type: "System", Path: "System", Format: "windows_event", Enabled: true},
	{Type: "Application", Path: "Application", Format: "windows_event", Enabled: true},
	{Type: "Security", Path: "Security", Format: "windows_event", Enabled: true},
	{Type: "PowerShell", Path: "Microsoft-Windows-PowerShell/Operational", Format: "windows_event", Enabled: false},
	{Type: "DNS", Path: "DNS", Format: "windows_event", Enabled: false},
	{Type: "Sysmon", Path: "Microsoft-Windows-Sysmon/Operational", Format: "windows_event", Enabled: false},
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