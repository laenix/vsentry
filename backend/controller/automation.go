package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/automation"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// RunPlaybook 执行/调试剧本
func RunPlaybook(ctx *gin.Context) {
	// 1. 解析 URL 参数 (Playbook ID)
	idStr := ctx.Param("id")
	playbookID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid playbook ID"})
		return
	}

	// 2. 解析 Body 参数 (Incident Context)
	var req struct {
		IncidentID uint `json:"incident_id"`
		DryRun     bool `json:"dry_run"` // 目前先预留，暂不处理
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid request body"})
		return
	}

	// 3. 构建上下文 (Context Injection)
	// 我们需要查询真实的 Incident 数据注入到 global 变量中
	inputContext := make(map[string]interface{})

	if req.IncidentID > 0 {
		var incident model.Incident
		// 预加载 Alerts，这样在剧本里可以用 {{incident.Alerts}} 访问证据
		result := database.GetDB().Preload("Alerts").First(&incident, req.IncidentID)
		if result.Error == nil {
			// 将 Struct 转为 map[string]interface{} 以便 expr 引擎使用
			// 这里用了一个小技巧：JSON Marshal -> Unmarshal
			// 也可以直接传 Struct，expr 支持 Struct 字段访问，但 map 更灵活
			inBytes, _ := json.Marshal(incident)
			var inMap map[string]interface{}
			json.Unmarshal(inBytes, &inMap)

			inputContext["incident"] = inMap
		}
	}

	// 4. 启动引擎
	engine := automation.NewEngine()
	executionID, err := engine.Run(uint(playbookID), inputContext)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	// 5. 返回执行 ID
	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "Playbook execution started",
		"data": gin.H{"execution_id": executionID},
	})
}

// ListPlaybooks 获取剧本列表
func ListPlaybooks(ctx *gin.Context) {
	var playbooks []model.Playbook
	// 列表页通常不需要加载巨大的 Definition JSON，可以节省带宽
	database.GetDB().Omit("definition").Order("updated_at desc").Find(&playbooks)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbooks})
}

// GetPlaybook 获取剧本详情 (包含画布定义)
func GetPlaybook(ctx *gin.Context) {
	id := ctx.Param("id")
	var playbook model.Playbook
	if err := database.GetDB().First(&playbook, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "Playbook not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbook})
}

// CreatePlaybook 创建 (已存在，补充完整性)
func CreatePlaybook(ctx *gin.Context) {
	var req model.Playbook
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Args error"})
		return
	}
	req.IsActive = false // 默认关闭
	database.GetDB().Create(&req)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": req})
}

// UpdatePlaybook 更新剧本 (保存画布)
func UpdatePlaybook(ctx *gin.Context) {
	id := ctx.Param("id")
	var req model.Playbook
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Args error"})
		return
	}

	var playbook model.Playbook
	if err := database.GetDB().First(&playbook, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "Playbook not found"})
		return
	}

	// 更新字段
	playbook.Name = req.Name
	playbook.Description = req.Description
	playbook.Definition = req.Definition // 核心：更新 React Flow JSON
	playbook.TriggerType = req.TriggerType
	playbook.IsActive = req.IsActive

	database.GetDB().Save(&playbook)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbook})
}

// DeletePlaybook 删除剧本
func DeletePlaybook(ctx *gin.Context) {
	id := ctx.Param("id")
	database.GetDB().Delete(&model.Playbook{}, id)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Deleted"})
}

// GetExecutionHistory 获取执行历史
func GetExecutionHistory(ctx *gin.Context) {
	id := ctx.Param("id") // Playbook ID
	var history []model.PlaybookExecution
	database.GetDB().Where("playbook_id = ?", id).Order("id desc").Limit(20).Find(&history)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": history})
}

// GetExecutionDetail 获取单次执行详情 (用于前端轮询状态和日志)
func GetExecutionDetail(ctx *gin.Context) {
	// 注意：这里的参数名要与路由匹配，建议用 exec_id 区分 playbook_id
	executionID := ctx.Param("exec_id")

	var execution model.PlaybookExecution
	if err := database.GetDB().First(&execution, executionID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "Execution history not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": execution, // 包含 Status 和 Logs (JSON)
	})
}

// ListAllExecutions
func ListAllExecutions(ctx *gin.Context) {
	var executions []model.PlaybookExecution
	database.GetDB().Order("id desc").Limit(100).Find(&executions)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": executions})
}
