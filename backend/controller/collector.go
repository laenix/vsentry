package controller

import (
	"archive/zip"
	"bytes"
	"fmt"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

var collectorTemplates = []model.ConnectorTemplate{
	{ID: "windows_event", Name: "Windows Event Collector", Type: "windows", Protocol: "api", DefaultPort: 0, Description: "Collect Windows Event Logs", Icon: "windows"},
	{ID: "linux_syslog", Name: "Linux Syslog Collector", Type: "linux", Protocol: "syslog", DefaultPort: 514, Description: "Collect Linux syslog", Icon: "linux"},
	{ID: "macos_log", Name: "macOS Unified Logging", Type: "macos", Protocol: "api", DefaultPort: 0, Description: "Collect macOS unified logging", Icon: "apple"},
}

// ListCollectorConfigs 获取采集器配置列表
func ListCollectorConfigs(ctx *gin.Context) {
	db := database.GetDB()
	var configs []model.CollectorConfig
	db.Find(&configs)
	ctx.JSON(200, gin.H{"code": 200, "data": configs})
}

// GetCollectorTemplates 获取采集器模板列表
func GetCollectorTemplates(ctx *gin.Context) {
	// Return our predefined templates
	templates := []map[string]interface{}{
		{
			"id":          "windows_event",
			"name":        "Windows Event Collector",
			"type":        "windows",
			"description": "Collect Windows Event Logs (Application, System, Security, etc.)",
			"icon":        "windows",
			"channels":    []string{"System", "Application", "Security", "PowerShell", "DNS", "Microsoft-Windows-PowerShell/Operational"},
		},
		{
			"id":          "linux_syslog",
			"name":        "Linux Syslog Collector",
			"type":        "linux",
			"description": "Collect Linux syslog via rsyslog/syslog-ng",
			"icon":        "linux",
			"channels":    []string{"auth", "authpriv", "cron", "daemon", "kern", "mail", "syslog"},
		},
		{
			"id":          "macos_unified",
			"name":        "macOS Unified Logging",
			"type":        "macos",
			"description": "Collect macOS unified logging",
			"icon":        "apple",
			"channels":    []string{"system", "network", "wifi", "install"},
		},
	}
	ctx.JSON(200, gin.H{"code": 200, "data": templates})
}

// AddCollectorConfig 添加采集器配置
func AddCollectorConfig(ctx *gin.Context) {
	var req model.CollectorConfig
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	// Get external URL from config
	externalURL := viper.GetString("server.external_url")
	if externalURL == "" {
		externalURL = "http://localhost:8088"
	}
	
	// Set default values
	if req.StreamFields == "" {
		req.StreamFields = "channel,source,host"
	}
	req.BuildStatus = "pending"

	database.GetDB().Create(&req)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Created successfully", "data": req})
}

// UpdateCollectorConfig 更新采集器配置
func UpdateCollectorConfig(ctx *gin.Context) {
	var req model.CollectorConfig
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	if req.ID == 0 {
		ctx.JSON(400, gin.H{"msg": "ID is required"})
		return
	}

	var existing model.CollectorConfig
	if err := database.GetDB().First(&existing, req.ID).Error; err != nil {
		ctx.JSON(404, gin.H{"msg": "Not found"})
		return
	}

	database.GetDB().Model(&existing).Updates(req)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Updated successfully"})
}

// DeleteCollectorConfig 删除采集器配置
func DeleteCollectorConfig(ctx *gin.Context) {
	id := ctx.Query("id")
	database.GetDB().Delete(&model.CollectorConfig{}, id)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Deleted successfully"})
}

// BuildCollector 构建采集器
func BuildCollector(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(400, gin.H{"msg": "ID is required"})
		return
	}

	var config model.CollectorConfig
	if err := database.GetDB().First(&config, id).Error; err != nil {
		ctx.JSON(404, gin.H{"msg": "Config not found"})
		return
	}

	// Update status to building
	database.GetDB().Model(&config).Update("build_status", "building")

	// Get external URL
	externalURL := viper.GetString("server.external_url")
	if externalURL == "" {
		externalURL = "http://localhost:8088"
	}

	// Use config's token or derive from ingest
	ingestEndpoint := externalURL + "/api/ingest/collect"
	if config.IngestEndpoint != "" {
		ingestEndpoint = config.IngestEndpoint
	}

	// Generate config.yaml content
	configContent := generateCollectorConfig(config, ingestEndpoint)

	// Create ZIP with config
	zipBuffer, err := createCollectorZip(config, configContent)
	if err != nil {
		database.GetDB().Model(&config).Update("build_status", "failed")
		database.GetDB().Model(&config).Update("build_output", err.Error())
		ctx.JSON(500, gin.H{"msg": "Build failed: " + err.Error()})
		return
	}

	// Update status
	database.GetDB().Model(&config).Update("build_status", "completed")
	database.GetDB().Model(&config).Update("build_output", "Build completed")

	// Return the ZIP
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=collector_%d_%s.zip", config.ID, config.Name))
	ctx.Header("Content-Type", "application/zip")
	ctx.Data(200, "application/zip", zipBuffer.Bytes())
}

func generateCollectorConfig(config model.CollectorConfig, endpoint string) string {
	// Generate a config.yaml file content
	tpl := `# Collector Configuration
name: {{.Name}}
type: {{.Type}}

# Collection Settings
channels:
{{range split .Channels ","}}
  - {{trim .}}
{{end}}

# Ingest Settings
ingest:
  endpoint: {{.Endpoint}}
  token: {{.Token}}
  stream_fields: {{.StreamFields}}
  
# Collection interval (seconds)
interval: 5
`
	
	data := map[string]string{
		"Name":         config.Name,
		"Type":         string(config.Type),
		"Channels":     config.Channels,
		"Endpoint":     endpoint,
		"Token":        config.Token,
		"StreamFields": config.StreamFields,
	}

	var buf bytes.Buffer
	tmpl, _ := template.New("config").Parse(tpl)
	tmpl.Execute(&buf, data)
	
	return buf.String()
}

func createCollectorZip(config model.CollectorConfig, configContent string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Add config.yaml
	f, err := w.Create("config.yaml")
	if err != nil {
		return nil, err
	}
	f.Write([]byte(configContent))

	// Add README.md
	readme := fmt.Sprintf(`# %s

## Installation

1. Extract this archive
2. Edit config.yaml with your settings
3. For Windows: Compile redAgent first (requires CGO), then run
   - Install Go 1.25+ for Windows
   - Set GOOS=windows GOARCH=amd64 CGO_ENABLED=1
   - Build: go build -o redAgent.exe ./cmd/redAgent
4. Run the collector

## Configuration

Edit config.yaml to customize:
- Collection channels (System, Application, Security, etc.)
- Ingest endpoint URL (automatically configured)
- Token for authentication (automatically configured)

## Supported Channels

Windows Event Log Channels:
- System
- Application  
- Security
- PowerShell
- DNS
- Active Directory
- Microsoft-Windows-PowerShell/Operational

## Support

Generated by VSentry Collector Builder

## Build from Source

This package includes redAgent source code. To compile:

## Build Requirements
- Go 1.25+
- CGO enabled (for Windows APIs)
- Windows SDK

## Compile on Windows
go build -o redAgent.exe ./cmd/redAgent

## Notes
The Ingest endpoint and Token are pre-configured based on your VSentry setup.
`, config.Name)
	
	f2, _ := w.Create("README.md")
	f2.Write([]byte(readme))

	// Add redAgent source code files
	agentFiles := map[string]string{
		"cmd/redAgent/main.go":        redAgentMainContent,
		"pkg/collector/windowsevent.go": redAgentWindowsEventContent,
		"pkg/ingest/ingest.go":         redAgentIngestContent,
		"pkg/storage/storage.go":       redAgentStorageContent,
		"go.mod":                       redAgentGoMod,
		"go.sum":                       redAgentGoSum,
	}

	for filename, content := range agentFiles {
		f, err := w.Create("redAgent/" + filename)
		if err != nil {
			return nil, err
		}
		f.Write([]byte(content))
	}

	// Add run script for Windows
	runBat := `@echo off
echo Starting %s...
echo.
echo Reading configuration from config.yaml...
echo.
echo Note: You need to compile redAgent.exe first:
echo   go build -o redAgent.exe ./cmd/redAgent
echo.
echo Please edit config.yaml and run: redAgent.exe
pause
`
	f3, _ := w.Create("run.bat")
	f3.Write([]byte(fmt.Sprintf(runBat, config.Name)))

	// Add run script for Linux/Mac
	runSh := `#!/bin/bash
# %s - Linux/macOS (requires adaptation)
echo "Starting collector..."
echo "This package is designed for Windows."
echo "For Linux/Mac, check VSentry for alternative collectors."
`
	f4, _ := w.Create("run.sh")
	f4.Write([]byte(fmt.Sprintf(runSh, config.Name)))

	w.Close()
	return buf, nil
}

// redAgent source code placeholders
var redAgentMainContent = `// redAgent - Windows Event Log Collector
// This is a placeholder. Use the complete source from VSentry project.
package main

func main() {
    // Implementation from redAgent source
}
`

var redAgentWindowsEventContent = `// Windows Event Collector
// See redAgent project for full implementation
`

var redAgentIngestContent = `// Ingest module
// See redAgent project for full implementation
`

var redAgentStorageContent = `// Storage module (BadgerDB)
// See redAgent project for full implementation
`

var redAgentGoMod = `module github.com/laneix/redAgent

go 1.21
`

var redAgentGoSum = ``

// GetAvailableChannels 获取指定类型可用的采集通道
func GetAvailableChannels(ctx *gin.Context) {
	collectorType := ctx.Query("type")
	
	channels := map[string][]string{
		"windows": {"System", "Application", "Security", "PowerShell", "Microsoft-Windows-PowerShell/Operational", "DNS", "Active Directory", "TerminalServices", "Windows Defender"},
		"linux":   {"auth", "authpriv", "cron", "daemon", "kern", "lpr", "mail", "news", "syslog", "user", "uucp", "local0", "local1", "local2", "local3", "local4", "local5", "local6", "local7"},
		"macos":   {"system", "system.diag", "system.install", "system.net", "system.wifi", "system.configuration", "system.opendirectory", "system.location"},
	}

	if ch, ok := channels[collectorType]; ok {
		ctx.JSON(200, gin.H{"code": 200, "data": ch})
	} else {
		ctx.JSON(200, gin.H{"code": 200, "data": []string{}})
	}
}