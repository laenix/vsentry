package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// ListCollectorConfigs 获取采集器配置列表
func ListCollectorConfigs(ctx *gin.Context) {
	db := database.GetDB()
	var configs []model.CollectorConfig
	db.Find(&configs)
	ctx.JSON(200, gin.H{"code": 200, "data": configs})
}

// GetCollectorTemplates 获取采集器模板列表
func GetCollectorTemplates(ctx *gin.Context) {
	templates := []map[string]interface{}{
		{
			"id":          "windows_event",
			"name":        "Windows Event Collector",
			"type":        "windows",
			"description": "Deploy a native Go agent to collect Windows Event Logs with zero dependencies.",
			"icon":        "windows",
		},
		{
			"id":          "linux_syslog",
			"name":        "Linux Syslog Collector",
			"type":        "linux",
			"description": "Collect Linux auth, syslog, and application logs via direct file tailing.",
			"icon":        "linux",
		},
		{
			"id":          "macos_unified",
			"name":        "macOS Unified Logging",
			"type":        "macos",
			"description": "Collect macOS unified logging directly from the Apple log datastore.",
			"icon":        "apple",
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

	if req.StreamFields == "" {
		req.StreamFields = "observer.hostname,observer.vendor,class_uid"
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

// GetAvailableChannels 获取指定类型可用的采集通道及过滤预设
func GetAvailableChannels(ctx *gin.Context) {
	collectorType := ctx.Query("type")

	// 使用 map[string]interface{} 以支持嵌套的 presets 数组
	channels := map[string][]map[string]interface{}{
		"windows": {
			{
				"type": "System", "path": "System", "label": "System",
				"presets": []map[string]string{
					{"name": "Service Install", "ids": "7045"},
				},
			},
			{
				"type": "Security", "path": "Security", "label": "Security (requires admin)",
				"presets": []map[string]string{
					{"name": "Auth & Logon", "ids": "4624, 4625, 4634, 4648"},
					{"name": "Account Mgmt", "ids": "4720, 4722, 4723, 4724, 4725, 4726"},
					{"name": "Process Exec", "ids": "4688, 4689"},
					{"name": "File Access", "ids": "4663"},
				},
			},
			{
				"type": "PowerShell", "path": "Microsoft-Windows-PowerShell/Operational", "label": "PowerShell",
				"presets": []map[string]string{
					{"name": "Script Blocks (Fileless)", "ids": "4104"},
				},
			},
			{
				"type": "WFP_Network", "path": "Security", "label": "Network Connections (WFP)",
				"presets": []map[string]string{
					{"name": "Allowed Connections", "ids": "5156"},
				},
			},
			{
				"type": "Sysmon", "path": "Microsoft-Windows-Sysmon/Operational", "label": "Sysmon (Advanced)",
				"presets": []map[string]string{
					{"name": "Process", "ids": "1, 5"},
					{"name": "Network", "ids": "3"},
					{"name": "File Activity", "ids": "11, 23"},
				},
			},
			{
				"type": "Defender", "path": "Microsoft-Windows-Windows Defender/Operational", "label": "Windows Defender",
				"presets": []map[string]string{
					{"name": "Malware Detections", "ids": "1116, 1117"},
				},
			},
		},
		// Linux 和 macOS 保持简单的文件路径即可，不需要 Event ID 预设
		"linux": {
			{"type": "syslog", "path": "/var/log/syslog", "label": "Syslog"},
			{"type": "auth", "path": "/var/log/auth.log", "label": "Auth Log"},
			{"type": "secure", "path": "/var/log/secure", "label": "Secure (SSH)"},
		},
		"macos": {
			{"type": "system", "path": "system", "label": "System Log"},
			{"type": "network", "path": "system.net", "label": "Network Log"},
		},
	}

	if ch, ok := channels[collectorType]; ok {
		ctx.JSON(200, gin.H{"code": 200, "data": ch})
	} else {
		ctx.JSON(200, gin.H{"code": 200, "data": []interface{}{}})
	}
}

// BuildCollector 动态编译跨平台采集器
func BuildCollector(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(400, gin.H{"msg": "ID is required"})
		return
	}

	db := database.GetDB()
	var config model.CollectorConfig
	if err := db.First(&config, id).Error; err != nil {
		ctx.JSON(404, gin.H{"msg": "Config not found"})
		return
	}

	var endpoint, token, streamFields string

	if config.IngestID > 0 {
		var ingest model.Ingest
		var auth model.IngestAuth
		if err := db.First(&ingest, config.IngestID).Error; err == nil {
			endpoint = ingest.Endpoint
			streamFields = ingest.StreamFields
			if err := db.Where("ingest_id = ?", ingest.ID).First(&auth).Error; err == nil {
				token = auth.SecretKey
			}
		}
	}

	// 更新状态为构建中
	db.Model(&config).Update("build_status", "building")

	// 生成嵌入式 JSON 配置
	embeddedConfigJSON := generateEmbeddedJSON(config, endpoint, token, streamFields)

	// 执行动态编译
	binaryPath, err := compileAgentDynamic(config.Type, embeddedConfigJSON)
	if err != nil {
		db.Model(&config).Updates(map[string]interface{}{
			"build_status": "failed",
			"build_output": err.Error(),
		})
		ctx.JSON(500, gin.H{"msg": "Compilation failed: " + err.Error()})
		return
	}
	defer os.Remove(binaryPath)

	db.Model(&config).Updates(map[string]interface{}{
		"build_status": "completed",
		"build_output": "Compilation successful",
	})

	fileName := fmt.Sprintf("vsentry-agent-%d", config.ID)
	if config.Type == "windows" {
		fileName += ".exe"
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(binaryPath)
}

// compileAgentDynamic 核心编译逻辑
func compileAgentDynamic(targetOS model.CollectorType, configJSON []byte) (string, error) {
	tempBuildDir, err := os.MkdirTemp("", "vsentry-build-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempBuildDir)

	copyFile("./go.mod", filepath.Join(tempBuildDir, "go.mod"))
	copyFile("./go.sum", filepath.Join(tempBuildDir, "go.sum"))
	copyDir("./pkg", filepath.Join(tempBuildDir, "pkg"))
	copyDir("./cmd", filepath.Join(tempBuildDir, "cmd"))

	configFilePath := filepath.Join(tempBuildDir, "cmd", "collectors", "config", "config.json")
	if err := os.WriteFile(configFilePath, configJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write embedded config: %w", err)
	}

	goos := "linux"
	if targetOS == "windows" {
		goos = "windows"
	} else if targetOS == "macos" {
		goos = "darwin"
	}

	outputFile := filepath.Join(tempBuildDir, "vsentry-agent")
	if goos == "windows" {
		outputFile += ".exe"
	}

	cmd := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", outputFile, ".")
	cmd.Dir = filepath.Join(tempBuildDir, "cmd", "collectors")
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH=amd64", "CGO_ENABLED=0")

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("go build failed: %s\nOutput: %s", err.Error(), string(output))
	}

	finalPath := filepath.Join(os.TempDir(), fmt.Sprintf("agent_%d_%s", time.Now().UnixNano(), filepath.Base(outputFile)))
	if err := copyFile(outputFile, finalPath); err != nil {
		return "", fmt.Errorf("failed to move compiled binary: %w", err)
	}

	return finalPath, nil
}

// generateEmbeddedJSON 接收从数据库提取的真实配置参数
func generateEmbeddedJSON(config model.CollectorConfig, endpoint, token, streamFields string) []byte {
	var sources []map[string]interface{}

	// 【核心修复】：全面拥抱前端发来的 JSON 格式的 Sources，抛弃旧的 Channels 字符串
	if config.Sources != "" {
		json.Unmarshal([]byte(config.Sources), &sources)
	}

	payload := map[string]interface{}{
		"name":          config.Name,
		"type":          config.Type,
		"interval":      config.Interval,
		"sources":       sources,
		"endpoint":      endpoint,
		"token":         token,
		"stream_fields": streamFields,
	}

	data, _ := json.MarshalIndent(payload, "", "  ")
	return data
}

// ================= 工具函数区 =================

func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}
