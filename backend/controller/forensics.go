package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

	// 4. 触发取证规则
	go triggerForensicRules(f.TaskID, f.ID)
}

// triggerForensicRules 触发取证规则
func triggerForensicRules(caseID, fileID uint) {
	db := database.GetDB()

	// 获取所有启用的取证规则
	var rules []model.Rule
	if err := db.Where("type = ? AND enabled = ?", "forensic", true).Find(&rules).Error; err != nil {
		log.Printf("[Forensic] Failed to fetch rules: %v", err)
		return
	}

	if len(rules) == 0 {
		log.Printf("[Forensic] No forensic rules enabled")
		return
	}

	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	log.Printf("[Forensic] Triggering %d rules for file %d", len(rules), fileID)

	for _, rule := range rules {
		// 构建查询：限制为当前案件和文件
		query := fmt.Sprintf("env:forensics case_id:%d file_id:%d | %s", caseID, fileID, rule.Query)

		queryURL := fmt.Sprintf("%s/select/logsql/query?query=%s&limit=1000", vlURL, url.QueryEscape(query))

		resp, err := http.Get(queryURL)
		if err != nil || resp.StatusCode >= 400 {
			log.Printf("[Forensic] Rule %d query failed: %v", rule.ID, err)
			continue
		}
		defer resp.Body.Close()

		// 解析响应
		var matchedData []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&matchedData); err != nil {
			log.Printf("[Forensic] Rule %d parse failed: %v", rule.ID, err)
			continue
		}

		if len(matchedData) > 0 {
			// 生成告警
			saveForensicAlert(rule, caseID, fileID, matchedData)
		}
	}
}

// saveForensicAlert 保存取证告警
func saveForensicAlert(rule model.Rule, caseID, fileID uint, matchedData []map[string]interface{}) {
	db := database.GetDB()
	now := time.Now().UTC()

	// 创建 Incident
	incident := model.Incident{
		RuleID:     rule.ID,
		Name:       fmt.Sprintf("[取证] %s - Case %d", rule.Name, caseID),
		Severity:   rule.Severity,
		Status:     "new",
		FirstSeen: now,
		LastSeen:   now,
		AlertCount: len(matchedData),
	}

	if err := db.Create(&incident).Error; err != nil {
		log.Printf("[Forensic] Failed to create incident: %v", err)
		return
	}

	// 创建 Alert
	for _, data := range matchedData {
		content, _ := json.Marshal(data)
		alert := model.Alert{
			IncidentID:  incident.ID,
			RuleID:      rule.ID,
			Content:     string(content),
			Fingerprint: fmt.Sprintf("%d-%d-%s", rule.ID, fileID, string(content[:min(len(content), 100)])),
		}
		db.Create(&alert)
	}

	log.Printf("[Forensic] Created incident %d with %d alerts for rule %d", incident.ID, len(matchedData), rule.ID)
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

// ExecuteForensicRules 执行取证规则
type ExecuteForensicRulesRequest struct {
	CaseID  uint   `json:"case_id" binding:"required"`
	FileID  uint   `json:"file_id" binding:"required"`
	RuleIDs []uint `json:"rule_ids" binding:"required"`
}

type ForensicRuleResult struct {
	RuleID      uint                   `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	Severity    string                 `json:"severity"`
	MatchedData []map[string]interface{} `json:"matched_data"`
	Count       int                    `json:"count"`
}

func ExecuteForensicRules(ctx *gin.Context) {
	var req ExecuteForensicRulesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameters"})
		return
	}

	db := database.GetDB()
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	// 获取选中的规则
	var rules []model.Rule
	if err := db.Where("id IN ? AND type = ? AND enabled = ?", req.RuleIDs, "forensic", true).Find(&rules).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to fetch rules"})
		return
	}

	results := make([]ForensicRuleResult, 0)

	for _, rule := range rules {
		// 构建查询：限制为当前案件和文件
		query := fmt.Sprintf("env:forensics case_id:%d file_id:%d | %s", req.CaseID, req.FileID, rule.Query)
		
		// 调用 VictoriaLogs 查询
		queryURL := fmt.Sprintf("%s/select/logsql/query?query=%s&limit=100", vlURL, url.QueryEscape(query))
		
		resp, err := http.Get(queryURL)
		if err != nil || resp.StatusCode >= 400 {
			continue
		}
		defer resp.Body.Close()

		// 解析响应
		var matchedData []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&matchedData); err != nil {
			continue
		}

		result := ForensicRuleResult{
			RuleID:      rule.ID,
			RuleName:    rule.Name,
			Severity:    rule.Severity,
			MatchedData: matchedData,
			Count:       len(matchedData),
		}
		results = append(results, result)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": results})
}