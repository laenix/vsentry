package controller

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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
	// Parse channels into proper list format
	channels := strings.Split(config.Channels, ",")
	var channelsYaml strings.Builder
	for _, ch := range channels {
		ch = strings.TrimSpace(ch)
		if ch != "" {
			channelsYaml.WriteString("  - " + ch + "\n")
		}
	}
	
	// Generate config.yaml
	return fmt.Sprintf(`# VSentry Collector Configuration
# Generated by VSentry

# Collector Name
name: %s

# Collector Type (windows/linux/macos)
type: %s

# Channels to collect (one per line)
channels:
%s
# Ingest Settings (auto-configured by VSentry)
ingest:
  endpoint: %s
  token: %s
  stream_fields: %s

# Collection interval in seconds
interval: 5
`, config.Name, config.Type, channelsYaml.String(), endpoint, config.Token, config.StreamFields)
}

func createCollectorZip(config model.CollectorConfig, configContent string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// 1. Add config.yaml with the embedded Ingest config
	f, err := w.Create("config.yaml")
	if err != nil {
		return nil, err
	}
	f.Write([]byte(configContent))

	// 2. Add README.md
	readme := fmt.Sprintf(`# %s - VSentry Collector

## Quick Start

1. Extract this archive on your Windows machine
2. Edit config.yaml to select the channels you want to collect
3. Run: powershell -ExecutionPolicy Bypass -File collector.ps1

## Requirements

- Windows 10/11 or Windows Server 2016+
- PowerShell 5.1 or later
- Network access to your VSentry server

## Configuration

The config.yaml is pre-configured with your VSentry server settings:
- Endpoint: Already configured
- Token: Already configured
- Channels: Adjust as needed (System, Application, Security, etc.)

## Supported Channels

Windows Event Log Channels:
- System - System events and errors
- Application - Application logs
- Security - Security events (requires admin)
- PowerShell - PowerShell execution logs
- DNS - DNS server queries
- Microsoft-Windows-PowerShell/Operational - Detailed PowerShell logs

## Testing

To test if the collector works:
1. Start VSentry
2. Run the collector: powershell -ExecutionPolicy Bypass -File collector.ps1
3. Check VSentry Logs page for incoming data

## Troubleshooting

- If no logs appear, check Windows Event Log service is running
- For Security channel logs, run PowerShell as Administrator
- Check network connectivity to VSentry server

---
Generated by VSentry Collector Builder
`, config.Name)
	
	f2, _ := w.Create("README.md")
	f2.Write([]byte(readme))

	// 3. Add the PowerShell collector script
	ps1Content := getCollectorScript(config, configContent)
	f3, err := w.Create("collector.ps1")
	if err != nil {
		return nil, err
	}
	f3.Write([]byte(ps1Content))

	// 4. Add a simple batch runner
	runBat := `@echo off
echo VSentry Collector - %s
echo ========================================
echo.
echo Starting collector...
echo Press Ctrl+C to stop
echo.
powershell -ExecutionPolicy Bypass -File collector.ps1
pause
`
	f4, _ := w.Create("run_collector.bat")
	f4.Write([]byte(fmt.Sprintf(runBat, config.Name)))

	w.Close()
	return buf, nil
}

// getCollectorScript generates the PowerShell collector script with embedded config
func getCollectorScript(config model.CollectorConfig, configContent string) string {
	// Parse channels
	channels := strings.Split(config.Channels, ",")
	var channelList []string
	for _, ch := range channels {
		ch = strings.TrimSpace(ch)
		if ch != "" {
			channelList = append(channelList, ch)
		}
	}
	if len(channelList) == 0 {
		channelList = []string{"System", "Application"}
	}

	// Get Ingest info from config
	// Parse from the generated config content
	var endpoint, token, streamFields string
	lines := strings.Split(configContent, "\n")
	for _, line := range lines {
		if strings.Contains(line, "endpoint:") {
			endpoint = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "token:") {
			token = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "stream_fields:") {
			streamFields = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	// Generate channel array for PowerShell
	channelsJson, _ := json.Marshal(channelList)

	return fmt.Sprintf(`# VSentry Collector for Windows
# Version: 1.0
# Auto-generated by VSentry Collector Builder

# === CONFIGURATION (Embedded) ===
$Config = @{
    Name = "%s"
    Type = "%s"
    Channels = %s
    Ingest = @{
        Endpoint = "%s"
        Token = "%s"
        StreamFields = "%s"
    }
    Interval = 5
}

Write-Host "VSentry Collector v1.0"
Write-Host "Channels: $($Config.Channels -join ', ')"
Write-Host "Endpoint: $($Config.Ingest.Endpoint)"

function Send-ToVSentry {
    param([array]$Logs)
    if ($Logs.Count -eq 0) { return }
    try {
        $json = $Logs | ConvertTo-Json -Compress -Depth 3
        $headers = @{
            "Authorization" = "Bearer $($Config.Ingest.Token)"
            "Content-Type" = "application/json"
        }
        Invoke-RestMethod -Uri $Config.Ingest.Endpoint -Method Post -Headers $headers -Body $json -TimeoutSec 30 -ErrorAction Stop | Out-Null
        Write-Host "[OK] Sent $($Logs.Count) logs"
    } catch {
        Write-Host "[ERROR] $($_.Exception.Message)"
    }
}

function Collect-ChannelLogs {
    param([string]$Channel)
    try {
        $events = Get-WinEvent -FilterHashtable @{LogName=$Channel; StartTime=(Get-Date).AddMinutes(-$Config.Interval)} -MaxEvents 30 -ErrorAction SilentlyContinue
        if (-not $events) { return @() }
        
        $logs = @()
        $hostname = $env:COMPUTERNAME
        $levelMap = @{1='critical';2='error';3='warning';4='information'}
        
        foreach ($e in $events) {
            $msg = if ($e.Message) { $e.Message.Substring(0, [Math]::Min(3000, $e.Message.Length)) } else { "" }
            $level = if ($levelMap[$e.Level]) { $levelMap[$e.Level] } else { "info" }
            
            $logs += @{
                _time = $e.TimeCreated.ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ss.fffZ")
                host = $hostname
                source = $Channel
                channel = $Channel
                message = $msg
                level = $level
                event_id = $e.Id
                provider = $e.ProviderName
            }
        }
        return $logs
    } catch {
        return @()
    }
}

# Main loop
Write-Host "Starting collection... (Ctrl+C to stop)"
while ($true) {
    $allLogs = @()
    foreach ($ch in $Config.Channels) {
        $allLogs += Collect-ChannelLogs -Channel $ch
    }
    if ($allLogs.Count -gt 0) {
        Send-ToVSentry -Logs $allLogs
    }
    Start-Sleep -Seconds $Config.Interval
}
`, config.Name, config.Type, string(channelsJson), endpoint, token, streamFields)
}

// GetAvailableChannels 获取指定类型可用的采集通道
func GetAvailableChannels(ctx *gin.Context) {
	collectorType := ctx.Query("type")
	
	channels := map[string][]string{
		"windows": {"System", "Application", "Security", "PowerShell", "Microsoft-Windows-PowerShell/Operational", "DNS", "Microsoft-Windows-Sysmon/Operational", "Microsoft-Windows-Windows Defender/Operational"},
		"linux":   {"auth", "authpriv", "cron", "daemon", "kern", "lpr", "mail", "news", "syslog", "user", "uucp", "local0", "local1", "local2", "local3", "local4", "local5", "local6", "local7"},
		"macos":   {"system", "system.diag", "system.install", "system.net", "system.wifi", "system.configuration", "system.opendirectory", "system.location"},
	}

	if ch, ok := channels[collectorType]; ok {
		ctx.JSON(200, gin.H{"code": 200, "data": ch})
	} else {
		ctx.JSON(200, gin.H{"code": 200, "data": []string{}})
	}
}