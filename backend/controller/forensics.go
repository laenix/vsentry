package controller

import (
	"bufio"
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

// CreateForensicTask Create一个New的Forensics沙箱Task
func CreateForensicTask(ctx *gin.Context) {
	var req model.ForensicTask
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid parameters"})
		return
	}
	database.GetDB().Create(&req)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": req})
}

// UploadForensicFile UploadEvidenceFile并触发AsyncParse
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

	// 生成Security的FilePath
	ext := strings.ToLower(filepath.Ext(file.Filename))
	safeFileName := fmt.Sprintf("task_%s_%d%s", taskID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(ForensicUploadDir, safeFileName)

	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to save file"})
		return
	}

	// 记录到Data库
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

	// 🔥 核心：StartAsyncParse协程，不阻塞ago端Response
	go processForensicFile(forensicFile)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "File uploaded successfully. Parsing started in background.",
		"data": forensicFile,
	})
}

// processForensicFile AsyncFileParse分发器
func processForensicFile(f model.ForensicFile) {
	db := database.GetDB()
	db.Model(&f).Update("parse_status", "parsing")

	// 1. 调用同级包的工厂Method
	p, err := forensic.GetParser(f.FileType)
	if err != nil {
		db.Model(&f).Updates(map[string]interface{}{
			"parse_status":  "failed",
			"parse_message": err.Error(),
		})
		return
	}

	// 2. Execute真正的硬核Parse
	parsedEvents, err := p.Parse(f.FilePath)
	if err != nil {
		db.Model(&f).Updates(map[string]interface{}{
			"parse_status":  "failed",
			"parse_message": err.Error(),
		})
		return
	}

	// 3. 将Parse出的Data打入 VictoriaLogs
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

	// 4. 触发ForensicsRule
	go triggerForensicRules(f.TaskID, f.ID)
}

// triggerForensicRules 触发ForensicsRule
func triggerForensicRules(caseID, fileID uint) {
	db := database.GetDB()

	// Get所有Enable的ForensicsRule
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
		// 构建Query：Limit为当agoCase和File，不LimitTime
		query := fmt.Sprintf("env:forensics task_id:%d forensic_file_id:%d | %s", caseID, fileID, rule.Query)

		queryURL := fmt.Sprintf("%s/select/logsql/query?query=%s&limit=1000", vlURL, url.QueryEscape(query))

		resp, err := http.Get(queryURL)
		if err != nil || resp.StatusCode >= 400 {
			log.Printf("[Forensic] Rule %d query failed: %v", rule.ID, err)
			continue
		}
		defer resp.Body.Close()

		// VictoriaLogs Return NDJSON 格式，Need逐行Parse
		var matchedData []map[string]interface{}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			var record map[string]interface{}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}
			matchedData = append(matchedData, record)
		}

		if len(matchedData) > 0 {
			// 生成Alert
			saveForensicAlert(rule, caseID, fileID, matchedData)
		}
	}
}

// saveForensicAlert SaveForensicsAlert
func saveForensicAlert(rule model.Rule, caseID, fileID uint, matchedData []map[string]interface{}) {
	db := database.GetDB()
	now := time.Now().UTC()

	// Create Incident
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

	// Create Alert
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

// ListForensicTasks Get所有ForensicsCaseList
func ListForensicTasks(ctx *gin.Context) {
	var tasks []model.ForensicTask
	// 按Time倒序，顺便把关联的File也查出来统计数量
	database.GetDB().Preload("Files").Order("created_at desc").Find(&tasks)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": tasks})
}

// GetForensicTask Get单个CaseDetail及其下的所有File（ago端轮询Parse进度用）
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

// DeleteForensicFile Delete单个EvidenceFile（连同磁盘File一起删）
func DeleteForensicFile(ctx *gin.Context) {
	id := ctx.Param("id")
	db := database.GetDB()
	var file model.ForensicFile

	if err := db.First(&file, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "File not found"})
		return
	}

	// 1. Delete磁盘物理File
	_ = os.Remove(file.FilePath)
	// 2. DeleteData库记录
	db.Delete(&file)
	// (可选) 3. 如果Need，可以调用 VictoriaLogs API Delete对应的Log，但通常保留即可或依赖 TTL

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "File deleted"})
}

// DeleteForensicTask Delete整个ForensicsCase
func DeleteForensicTask(ctx *gin.Context) {
	id := ctx.Param("id")
	db := database.GetDB()
	var task model.ForensicTask

	if err := db.Preload("Files").First(&task, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Task not found"})
		return
	}

	// 1. 遍历Delete磁盘上的物理File
	for _, file := range task.Files {
		_ = os.Remove(file.FilePath)
		db.Delete(&file)
	}

	// 2. DeleteCase本身
	db.Delete(&task)

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Task and related files deleted"})
}

// ExecuteForensicRules ExecuteForensicsRule
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

	// Get选Medium的Rule
	var rules []model.Rule
	if err := db.Where("id IN ? AND type = ? AND enabled = ?", req.RuleIDs, "forensic", true).Find(&rules).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to fetch rules"})
		return
	}

	results := make([]ForensicRuleResult, 0)

	for _, rule := range rules {
		// 构建Query：Limit为当agoCase和File，不LimitTime（ForensicsData可能来自任意Time）
		query := fmt.Sprintf("env:forensics task_id:%d forensic_file_id:%d | %s", req.CaseID, req.FileID, rule.Query)
		
		// 调用 VictoriaLogs Query
		queryURL := fmt.Sprintf("%s/select/logsql/query?query=%s&limit=100", vlURL, url.QueryEscape(query))
		
		resp, err := http.Get(queryURL)
		if err != nil || resp.StatusCode >= 400 {
			log.Printf("VictoriaLogs query failed: %v, status: %d", err, resp.StatusCode)
			continue
		}
		defer resp.Body.Close()

		// VictoriaLogs Return NDJSON 格式，Need逐行Parse
		var matchedData []map[string]interface{}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			var record map[string]interface{}
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}
			matchedData = append(matchedData, record)
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