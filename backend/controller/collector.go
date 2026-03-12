package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// ListCollectorConfigs GetCollect器配置List
func ListCollectorConfigs(ctx *gin.Context) {
	db := database.GetDB()
	var configs []model.CollectorConfig
	db.Find(&configs)
	ctx.JSON(200, gin.H{"code": 200, "data": configs})
}

// GetCollectorTemplates GetCollect器TemplateList
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
		// 【New增】Application层通用CollectTemplate，方便User在界面上直接Select
		{
			"id":          "app_layer",
			"name":        "Application & Web Logs",
			"type":        "linux", // 通常跑在 Linux 上，但由于我们的 AppCollector 是跨平台的，其实两边都能跑
			"description": "Collect Nginx, Apache, Tomcat, MySQL, and Redis logs via file tailing.",
			"icon":        "server",
		},
	}
	ctx.JSON(200, gin.H{"code": 200, "data": templates})
}

// AddCollectorConfig AddCollect器配置
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

// UpdateCollectorConfig UpdateCollect器配置
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

// DeleteCollectorConfig DeleteCollect器配置
func DeleteCollectorConfig(ctx *gin.Context) {
	id := ctx.Query("id")
	database.GetDB().Delete(&model.CollectorConfig{}, id)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Deleted successfully"})
}

// GetAvailableChannels Get指定Type可用的Collect通道及Filter预设
func GetAvailableChannels(ctx *gin.Context) {
	collectorType := ctx.Query("type")

	// 使用 map[string]interface{} 以支持嵌套的 presets 数Group
	channels := map[string][]map[string]interface{}{
		"windows": {
			{
				"type": "System", "path": "System", "label": "System",
				"presets": []map[string]string{{"name": "Service Install", "ids": "7045"}},
			},
			{
				"type": "Security", "path": "Security", "label": "Security (requires admin)",
				"presets": []map[string]string{
					{"name": "Auth & Logon", "ids": "4624, 4625, 4634, 4648"},
					{"name": "Account Mgmt", "ids": "4720, 4722, 4723, 4724, 4725, 4726"},
					{"name": "Group Mgmt & Lockout", "ids": "4732, 4733, 4730, 4740"},
					{"name": "Process Exec", "ids": "4688, 4689"},
					{"name": "AD/Kerberos", "ids": "4768, 4769, 4771, 4776"},
					{"name": "Critical & Tampering", "ids": "1102, 4719, 4672, 4765, 4766, 4794"},
					{"name": "File Access", "ids": "4663"},
					{"name": "Scheduled Tasks", "ids": "4698, 4699, 4700, 4702"},
				},
			},
			{
				"type": "PowerShell", "path": "Microsoft-Windows-PowerShell/Operational", "label": "PowerShell",
				"presets": []map[string]string{
					{"name": "Script Blocks (Fileless)", "ids": "4104"},
					{"name": "Module & Runspace", "ids": "4103, 4105, 4106"},
				},
			},
			{
				"type": "WFP_Network", "path": "Security", "label": "Network Connections (WFP)",
				"presets": []map[string]string{{"name": "Allowed Connections", "ids": "5156"}},
			},
			{
				"type": "Sysmon", "path": "Microsoft-Windows-Sysmon/Operational", "label": "Sysmon (Advanced)",
				"presets": []map[string]string{
					{"name": "Process", "ids": "1, 5"},
					{"name": "Network", "ids": "3"},
					{"name": "File Activity", "ids": "11, 23"},
					{"name": "DNS Queries", "ids": "22"},
					{"name": "Registry Activity", "ids": "12, 13, 14"},
					{"name": "WMI & Pipe", "ids": "17, 18, 19, 20, 21"},
					{"name": "Process Injection", "ids": "8, 10"},
				},
			},
			{
				"type": "Defender", "path": "Microsoft-Windows-Windows Defender/Operational", "label": "Windows Defender",
				"presets": []map[string]string{{"name": "Malware Detections", "ids": "1116, 1117"}},
			},
		},
		"linux": {
			// OS 层Log
			{"type": "syslog", "path": "/var/log/syslog", "label": "Syslog (OS)"},
			{"type": "auth", "path": "/var/log/auth.log", "label": "Auth Log (SSH)"},
			{"type": "secure", "path": "/var/log/secure", "label": "Secure Log (RedHat)"},

			// 【New增】Web 容器Log
			{"type": "nginx_access", "path": "/var/log/nginx/access.log", "label": "Nginx Access Log"},
			{"type": "nginx_error", "path": "/var/log/nginx/error.log", "label": "Nginx Error Log"},
			{"type": "apache_access", "path": "/var/log/apache2/access.log", "label": "Apache Access Log"},
			{"type": "apache_error", "path": "/var/log/apache2/error.log", "label": "Apache Error Log"},
			{"type": "tomcat_access", "path": "/opt/tomcat/logs/localhost_access_log.txt", "label": "Tomcat Access Log"},
			{"type": "tomcat_catalina", "path": "/opt/tomcat/logs/catalina.out", "label": "Tomcat Catalina Log"},

			// 【New增】Data库Log
			{"type": "mysql_error", "path": "/var/log/mysql/error.log", "label": "MySQL Error Log"},
			{"type": "redis_log", "path": "/var/log/redis/redis-server.log", "label": "Redis Log"},
		},
		"macos": {
			{"type": "darwin_unified", "path": "system", "label": "macOS System Log"},
			{"type": "darwin_unified", "path": "system.net", "label": "macOS Network Log"},
		},
	}

	if ch, ok := channels[collectorType]; ok {
		ctx.JSON(200, gin.H{"code": 200, "data": ch})
	} else {
		ctx.JSON(200, gin.H{"code": 200, "data": []interface{}{}})
	}
}

// 定义编译产物的全局Storage目录
const BuildOutputDir = "./data/builds"

func init() {
	// 确保程序Start时，构建目录存在
	os.MkdirAll(BuildOutputDir, 0755)
}

// BuildCollector 触发编译跨平台Collect器 (仅编译，不Download)
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

	// 提取 Ingest 凭证
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

	// 1. UpdateStatus为构建Medium
	db.Model(&config).Update("build_status", "building")

	// 2. 生成嵌入式 JSON 配置
	embeddedConfigJSON := generateEmbeddedJSON(config, endpoint, token, streamFields)

	// 3. 确定最终的持久化File名和Path
	fileName := fmt.Sprintf("vsentry-agent-%d-%s", config.ID, config.Type)
	if config.Type == "windows" {
		fileName += ".exe"
	}
	finalBinaryPath := filepath.Join(BuildOutputDir, fileName)

	// 4. Execute动态编译 (将最终Path传入)
	err := compileAgentDynamic(config.Type, embeddedConfigJSON, finalBinaryPath)
	if err != nil {
		// 编译Failed，记录ErrorLog
		db.Model(&config).Updates(map[string]interface{}{
			"build_status": "failed",
			"build_output": err.Error(),
		})
		ctx.JSON(500, gin.H{"msg": "Compilation failed: " + err.Error()})
		return
	}

	// 5. 编译Success，UpdateData库Status
	db.Model(&config).Updates(map[string]interface{}{
		"build_status": "completed",
		"build_output": "Compilation successful",
	})

	ctx.JSON(200, gin.H{
		"code": 200,
		"msg":  "Build completed successfully",
		"data": map[string]string{
			"status": "completed",
			"file":   fileName,
		},
	})
}

// DownloadCollector Download已编译的Collect器 (纯粹的FileService器功能)
func DownloadCollector(ctx *gin.Context) {
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

	// 拦截未编译完成的Request
	if config.BuildStatus != "completed" {
		ctx.JSON(400, gin.H{"msg": "Agent has not been built successfully yet. Please build it first."})
		return
	}

	// 推导FilePath
	fileName := fmt.Sprintf("vsentry-agent-%d-%s", config.ID, config.Type)
	if config.Type == "windows" {
		fileName += ".exe"
	}
	finalBinaryPath := filepath.Join(BuildOutputDir, fileName)

	// 检查File是否存在于磁盘上
	if _, err := os.Stat(finalBinaryPath); os.IsNotExist(err) {
		// 修复File丢失导致的Data库Status不一致
		db.Model(&config).Update("build_status", "pending")
		ctx.JSON(404, gin.H{"msg": "Binary file is missing from server disk. Please rebuild."})
		return
	}

	// 触发浏览器Download
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(finalBinaryPath)
}

// compileAgentDynamic 核心编译逻辑
// targetOS: 目标SystemType; configJSON: 嵌入的配置File; finalPath: 最终持久化Storage的Path
func compileAgentDynamic(targetOS model.CollectorType, configJSON []byte, finalPath string) error {
	tempBuildDir, err := os.MkdirTemp("", "vsentry-build-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	// 确保编译结束后清理掉那些占用几百兆空间的临时源码目录
	defer os.RemoveAll(tempBuildDir)

	// 拷贝源码到沙盒目录
	copyFile("./go.mod", filepath.Join(tempBuildDir, "go.mod"))
	copyFile("./go.sum", filepath.Join(tempBuildDir, "go.sum"))
	copyDir("./pkg", filepath.Join(tempBuildDir, "pkg"))
	copyDir("./cmd", filepath.Join(tempBuildDir, "cmd"))

	// 注入配置
	configFilePath := filepath.Join(tempBuildDir, "cmd", "collectors", "config", "config.json")
	if err := os.WriteFile(configFilePath, configJSON, 0644); err != nil {
		return fmt.Errorf("failed to write embedded config: %w", err)
	}

	goos := "linux"
	if targetOS == "windows" {
		goos = "windows"
	} else if targetOS == "macos" {
		goos = "darwin"
	}

	outputFile := filepath.Join(tempBuildDir, "vsentry-agent-build")
	if goos == "windows" {
		outputFile += ".exe"
	}

	// Execute go build
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags", "-s -w", "-o", outputFile, ".")
	cmd.Dir = filepath.Join(tempBuildDir, "cmd", "collectors")
	// 关闭 CGO，确保 Agent 跨平台免依赖 (静态链接)
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH=amd64", "CGO_ENABLED=0")

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go build failed: %s\nOutput: %s", err.Error(), string(output))
	}

	// 编译Success后，将沙盒Medium的二进制File拷贝到永久Storage目录
	if err := copyFile(outputFile, finalPath); err != nil {
		return fmt.Errorf("failed to move compiled binary to permanent storage: %w", err)
	}

	return nil
}

// generateEmbeddedJSON Receive从Data库提取的真实配置参数
func generateEmbeddedJSON(config model.CollectorConfig, endpoint, token, streamFields string) []byte {
	var sources []map[string]interface{}

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

// ================= 工具Function区 =================

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
