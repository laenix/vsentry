package controller

import (
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
	"github.com/laenix/vsentry/forensic"
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

	// 1. 调用同级包的工厂方法
	p, err := forensic.GetParser(f.FileType)
	if err != nil {
		db.Model(&f).Updates(map[string]interface{}{
			"parse_status":  "failed",
			"parse_message": err.Error(),
		})
		return
	}

	// 2. 执行真正的硬核解析
	parsedEvents, err := p.Parse(f.FilePath)
	if err != nil {
		db.Model(&f).Updates(map[string]interface{}{
			"parse_status":  "failed",
			"parse_message": err.Error(),
		})
		return
	}

	// 3. 将解析出的数据打入 VictoriaLogs
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}
	ingestURL := fmt.Sprintf("%s/insert/jsonline?_stream_fields=env,task_id&_time_field=time&_msg_field=raw_data", vlURL)

	var jsonlBuffer bytes.Buffer
	for _, event := range parsedEvents {
		event["env"] = "forensics"
		event["task_id"] = fmt.Sprintf("%d", f.TaskID)
		event["forensic_file_id"] = fmt.Sprintf("%d", f.ID)

		if _, ok := event["time"]; !ok {
			event["time"] = time.Now().UTC().Format(time.RFC3339)
		}

		jsonData, _ := json.Marshal(event)
		jsonlBuffer.Write(jsonData)
		jsonlBuffer.WriteString("\n")
	}

	resp, postErr := http.Post(ingestURL, "application/x-ndjson", &jsonlBuffer)
	if postErr != nil || resp.StatusCode >= 400 {
		db.Model(&f).Updates(map[string]interface{}{
			"parse_status":  "failed",
			"parse_message": "Failed to inject into VictoriaLogs",
		})
		return
	}

	db.Model(&f).Updates(map[string]interface{}{
		"parse_status":  "completed",
		"event_count":   len(parsedEvents),
		"parse_message": "Successfully parsed and injected.",
	})
}

// ListForensicTasks 获取所有取证案件列表
func ListForensicTasks(ctx *gin.Context) {
	var tasks []model.ForensicTask
	// 按时间倒序，顺便把关联的文件也查出来统计数量
	database.GetDB().Preload("Files").Order("created_at desc").Find(&tasks)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": tasks})
}

// GetForensicTask 获取单个案件详情及其下的所有文件（前端轮询解析进度用）
func GetForensicTask(ctx *gin.Context) {
	id := ctx.Param("id")
	var task model.ForensicTask
	
	err := database.GetDB().Preload("Files").First(&task, id).Error
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Task not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": task})
}

// DeleteForensicFile 删除单个证据文件（连同磁盘文件一起删）
func DeleteForensicFile(ctx *gin.Context) {
	id := ctx.Param("id")
	db := database.GetDB()
	var file model.ForensicFile

	if err := db.First(&file, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "File not found"})
		return
	}

	// 1. 删除磁盘物理文件
	_ = os.Remove(file.FilePath)
	// 2. 删除数据库记录
	db.Delete(&file)
	// (可选) 3. 如果需要，可以调用 VictoriaLogs API 删除对应的日志，但通常保留即可或依赖 TTL

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "File deleted"})
}

// DeleteForensicTask 删除整个取证案件
func DeleteForensicTask(ctx *gin.Context) {
	id := ctx.Param("id")
	db := database.GetDB()
	var task model.ForensicTask

	if err := db.Preload("Files").First(&task, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Task not found"})
		return
	}

	// 1. 遍历删除磁盘上的物理文件
	for _, file := range task.Files {
		_ = os.Remove(file.FilePath)
		db.Delete(&file)
	}

	// 2. 删除案件本身
	db.Delete(&task)

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Task and related files deleted"})
}