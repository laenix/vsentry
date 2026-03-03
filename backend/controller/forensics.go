package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

const ForensicUploadDir = "./data/forensics"

func init() {
	os.MkdirAll(ForensicUploadDir, 0755)
}

// CreateForensicTask 创建一个新的取证沙箱任务
func CreateForensicTask(ctx *gin.Context) {
	var req model.ForensicTask
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid parameters"})
		return
	}
	database.GetDB().Create(&req)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": req})
}

// UploadForensicFile 上传证据文件并触发异步解析
func UploadForensicFile(ctx *gin.Context) {
	taskID := ctx.PostForm("task_id")
	if taskID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "task_id is required"})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "File upload error"})
		return
	}

	// 生成安全的文件路径
	ext := strings.ToLower(filepath.Ext(file.Filename))
	safeFileName := fmt.Sprintf("task_%s_%d%s", taskID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(ForensicUploadDir, safeFileName)

	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to save file"})
		return
	}

	// 记录到数据库
	db := database.GetDB()
	var tID uint
	fmt.Sscanf(taskID, "%d", &tID)

	forensicFile := model.ForensicFile{
		TaskID:       tID,
		FileName:     safeFileName,
		OriginalName: file.Filename,
		FileType:     strings.TrimPrefix(ext, "."),
		FileSize:     file.Size,
		FilePath:     savePath,
		ParseStatus:  "pending",
	}
	db.Create(&forensicFile)

	// 🔥 核心：启动异步解析协程，不阻塞前端响应
	go processForensicFile(forensicFile)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "File uploaded successfully. Parsing started in background.",
		"data": forensicFile,
	})
}

// processForensicFile 异步文件解析分发器
func processForensicFile(f model.ForensicFile) {
	db := database.GetDB()
	db.Model(&f).Update("parse_status", "parsing")

	var parsedEvents []map[string]interface{}
	var err error

	// 策略模式：根据文件后缀调用不同的解析引擎
	switch f.FileType {
	case "log", "txt", "csv":
		parsedEvents, err = parseTextLog(f.FilePath)
	case "evtx":
		parsedEvents, err = parseEVTX(f.FilePath)
	case "pcap", "pcapng":
		parsedEvents, err = parsePCAP(f.FilePath)
	default:
		err = fmt.Errorf("unsupported file type: %s", f.FileType)
	}

	if err != nil {
		db.Model(&f).Updates(map[string]interface{}{
			"parse_status":  "failed",
			"parse_message": err.Error(),
		})
		return
	}

	// 🔥 隔离注入：将解析出的数据打入 VictoriaLogs，并强制附加沙箱标签
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}
	// 注意这里：我们把 env 设置为 forensics，并把 task_id 作为流字段！
	ingestURL := fmt.Sprintf("%s/insert/jsonline?_stream_fields=env,task_id&_time_field=time&_msg_field=raw_data", vlURL)

	var jsonlBuffer bytes.Buffer
	for _, event := range parsedEvents {
		// 强制注入沙箱隔离标签
		event["env"] = "forensics"
		event["task_id"] = fmt.Sprintf("%d", f.TaskID)
		event["forensic_file_id"] = fmt.Sprintf("%d", f.ID)

		// 确保有时间戳
		if _, ok := event["time"]; !ok {
			event["time"] = time.Now().UTC().Format(time.RFC3339)
		}

		jsonData, _ := json.Marshal(event)
		jsonlBuffer.Write(jsonData)
		jsonlBuffer.WriteString("\n")
	}

	// 发送到 VictoriaLogs
	resp, postErr := http.Post(ingestURL, "application/x-ndjson", &jsonlBuffer)
	if postErr != nil || resp.StatusCode >= 400 {
		db.Model(&f).Updates(map[string]interface{}{
			"parse_status":  "failed",
			"parse_message": "Failed to inject into VictoriaLogs",
		})
		return
	}

	// 更新成功状态
	db.Model(&f).Updates(map[string]interface{}{
		"parse_status":  "completed",
		"event_count":   len(parsedEvents),
		"parse_message": "Successfully parsed and injected.",
	})
}

// ==================== 解析引擎实现区 ====================

// parseTextLog 解析普通文本日志
func parseTextLog(filePath string) ([]map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []map[string]interface{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// 这里你可以复用之前的 Mapper，为了演示，我们先包裹成基础 OCSF 格式
		events = append(events, map[string]interface{}{
			"raw_data":      line,
			"category_name": "Findings",
			"class_name":    "Forensic Evidence",
		})
	}
	return events, nil
}

// parseEVTX 解析 Windows EVTX (二期扩展点)
func parseEVTX(filePath string) ([]map[string]interface{}, error) {
	// TODO: 引入 github.com/0xrawsec/golang-evtx 进行真实解析
	// 目前先返回 Mock 数据保证流程畅通
	var events []map[string]interface{}
	events = append(events, map[string]interface{}{
		"raw_data":      fmt.Sprintf("Mock parsed EVTX from %s", filePath),
		"event_id":      4624,
		"activity_name": "Logon",
	})
	return events, nil
}

// parsePCAP 解析网络抓包 (二期扩展点)
func parsePCAP(filePath string) ([]map[string]interface{}, error) {
	// TODO: 引入 github.com/google/gopacket 进行真实解析
	var events []map[string]interface{}
	events = append(events, map[string]interface{}{
		"raw_data": fmt.Sprintf("Mock parsed PCAP from %s", filePath),
		"src_ip":   "192.168.1.100",
		"dst_ip":   "8.8.8.8",
		"protocol": "DNS",
	})
	return events, nil
}
