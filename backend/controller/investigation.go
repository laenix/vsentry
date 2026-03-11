package controller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

// ListInvestigationTemplates 获取所有预置调查模板
func ListInvestigationTemplates(ctx *gin.Context) {
	db := database.GetDB()
	var templates []model.InvestigationTemplate
	if err := db.Find(&templates).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to fetch templates"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": templates})
}

// AddInvestigationTemplate 新增调查模板 (供管理员/高级用户自定义)
func AddInvestigationTemplate(ctx *gin.Context) {
	var req model.InvestigationTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid parameters"})
		return
	}

	database.GetDB().Create(&req)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Created successfully", "data": req})
}

// UpdateInvestigationTemplate 更新调查模板
func UpdateInvestigationTemplate(ctx *gin.Context) {
	var req model.InvestigationTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid parameters"})
		return
	}

	if req.ID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "ID is required for update"})
		return
	}

	db := database.GetDB()
	var existing model.InvestigationTemplate
	if err := db.First(&existing, req.ID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "Template not found"})
		return
	}

	// 仅更新允许修改的字段
	db.Model(&existing).Updates(map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"logsql":      req.LogSQL,
		"parameters":  req.Parameters,
	})

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Updated successfully"})
}

// DeleteInvestigationTemplate 删除调查模板
func DeleteInvestigationTemplate(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "ID is required"})
		return
	}

	db := database.GetDB()
	if err := db.Delete(&model.InvestigationTemplate{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to delete"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Deleted successfully"})
}

// ExecuteInvestigation 核心！执行带参调查查询 (支持 Incident 与 Alert 上下文注入)
func ExecuteInvestigation(ctx *gin.Context) {
	var req struct {
		TemplateID uint              `json:"template_id" binding:"required"`
		IncidentID uint              `json:"incident_id"`
		Params     map[string]string `json:"params"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid request body"})
		return
	}

	db := database.GetDB()

	// 1. 查询模板
	var template model.InvestigationTemplate
	if err := db.First(&template, req.TemplateID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "Template not found"})
		return
	}

	// 2. 初始化变量池 (Variables Pool)
	vars := make(map[string]string)

	// 设置默认时间范围（如果模板需要时间参数但未提供）
	// 使用过去1年到未来1年，覆盖绝大多数取证场景
	oneYearAgo := time.Now().AddDate(-1, 0, 0).UTC().Format(time.RFC3339)
	oneYearLater := time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)
	vars["start_time"] = oneYearAgo
	vars["end_time"] = oneYearLater

	// 3. 加载 Incident 及其关联的 Alerts
	if req.IncidentID > 0 {
		var incident model.Incident
		if err := db.Preload("Alerts").First(&incident, req.IncidentID).Error; err == nil {
			vars["incident_id"] = fmt.Sprintf("%d", incident.ID)
			vars["incident_name"] = incident.Name

			// ✅ 修复时间格式：VictoriaLogs 更喜欢这种格式 2026-03-03T15:04:05Z
			// 我们去掉毫秒，保留 Z
			vars["start_time"] = incident.FirstSeen.Add(-2 * time.Hour).UTC().Format(time.RFC3339)
			vars["end_time"] = incident.LastSeen.Add(2 * time.Hour).UTC().Format(time.RFC3339)

			if len(incident.Alerts) > 0 {
				firstAlert := incident.Alerts[0]
				if firstAlert.Content != "" {
					var alertContent map[string]interface{}
					if err := json.Unmarshal([]byte(firstAlert.Content), &alertContent); err == nil {
						flattenMap(alertContent, "", vars)

						if val, ok := vars["observer.hostname"]; ok {
							vars["hostname"] = val
						}
						if val, ok := vars["src_endpoint.ip"]; ok {
							vars["src_ip"] = val
						}
						if val, ok := vars["target_user.name"]; ok {
							vars["username"] = val
						}
						if val, ok := vars["actor.user.name"]; ok {
							vars["username"] = val
						}
						if val, ok := vars["process.name"]; ok {
							vars["process_name"] = val
						}
					}
				}
			}
		}
	}

	// 4. 将前端手动传过来的参数覆盖进去
	for k, v := range req.Params {
		vars[k] = v
	}

	// 5. 动态替换 LogSQL 中的参数
	finalLogSQL := template.LogSQL
	for key, val := range vars {
		placeholder := fmt.Sprintf("${%s}", key)
		finalLogSQL = strings.ReplaceAll(finalLogSQL, placeholder, val)
	}

	// 6. 拦截未被替换的参数
	if strings.Contains(finalLogSQL, "${") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":           "Missing required parameters in template",
			"query_preview": finalLogSQL,
			"current_vars":  vars,
		})
		return
	}

	// 7. 调用底层的 VictoriaLogs 进行查询
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	params := url.Values{}
	params.Set("query", finalLogSQL)

	limit := vars["_limit"]
	if limit == "" {
		limit = "1000"
	}
	params.Set("limit", limit)

	targetURL := vlURL + "/select/logsql/query?" + params.Encode()
	vlReq, err := http.NewRequest("POST", targetURL, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	vlReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// ✅ 关键修复 1：强制设置 15 秒超时，绝不允许挂起死锁
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(vlReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "VictoriaLogs unreachable or timeout: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// ✅ 关键修复 2：检查 VictoriaLogs 是否返回报错！如果查出 400，把原样错误抛给前端
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "VictoriaLogs Query Failed",
			"detail": string(bodyBytes),
			"logsql": finalLogSQL,
		})
		return
	}

	// ✅ 关键修复 3：使用 Scanner 逐行安全解析 JSONLine，彻底根除死循环
	var results []map[string]interface{}
	scanner := bufio.NewScanner(resp.Body)

	// 扩大 Scanner 的 buffer，防止单条日志内容过长导致 bufio 溢出报错
	const maxCapacity = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // 忽略空行
		}
		var row map[string]interface{}
		if err := json.Unmarshal([]byte(line), &row); err == nil {
			results = append(results, row)
		} else {
			// 某一行解析失败，打个日志，继续解析下一行，绝对不会死循环
			fmt.Printf("[Investigation] JSON Parse error on line: %v\n", err)
		}
	}

	// 检查 Scanner 是否发生底层错误
	if err := scanner.Err(); err != nil {
		fmt.Printf("[Investigation] Scanner error: %v\n", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"logsql":       finalLogSQL,
			"events":       results,
			"count":        len(results),
			"context_used": vars,
		},
	})
}

// 辅助函数保持不变
func flattenMap(data map[string]interface{}, prefix string, result map[string]string) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch child := v.(type) {
		case map[string]interface{}:
			flattenMap(child, key, result)
		case string:
			result[key] = child
		case float64, int, bool:
			result[key] = fmt.Sprintf("%v", child)
		}
	}
}
