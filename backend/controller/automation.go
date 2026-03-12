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

// RunPlaybook Execute/DebugPlaybook
func RunPlaybook(ctx *gin.Context) {
	// 1. Parse URL 参数 (Playbook ID)
	idStr := ctx.Param("id")
	playbookID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid playbook ID"})
		return
	}

	// 2. Parse Body 参数 (Incident Context)
	var req struct {
		IncidentID uint `json:"incident_id"`
		DryRun     bool `json:"dry_run"` // 目ago先预留，暂不Handle
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid request body"})
		return
	}

	// 3. 构建上下文 (Context Injection)
	// 我们NeedQuery真实的 Incident Data注入到 global VariableMedium
	inputContext := make(map[string]interface{})

	if req.IncidentID > 0 {
		var incident model.Incident
		// 预加载 Alerts，这样在Playbook里可以用 {{incident.Alerts}} 访问Evidence
		result := database.GetDB().Preload("Alerts").First(&incident, req.IncidentID)
		if result.Error == nil {
			// 将 Struct 转为 map[string]interface{} 以便 expr Engine使用
			// 这里用了一个小技巧：JSON Marshal -> Unmarshal
			// 也可以直接传 Struct，expr 支持 Struct 字段访问，但 map 更灵活
			inBytes, _ := json.Marshal(incident)
			var inMap map[string]interface{}
			json.Unmarshal(inBytes, &inMap)

			inputContext["incident"] = inMap
		}
	}

	// 4. StartEngine
	engine := automation.NewEngine()
	executionID, err := engine.Run(uint(playbookID), inputContext)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	// 5. ReturnExecute ID
	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "Playbook execution started",
		"data": gin.H{"execution_id": executionID},
	})
}

// ListPlaybooks GetPlaybookList
func ListPlaybooks(ctx *gin.Context) {
	var playbooks []model.Playbook
	// List页通常不Need加载巨大的 Definition JSON，可以节省带宽
	database.GetDB().Omit("definition").Order("updated_at desc").Find(&playbooks)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbooks})
}

// GetPlaybook GetPlaybookDetail (包含画布定义)
func GetPlaybook(ctx *gin.Context) {
	id := ctx.Param("id")
	var playbook model.Playbook
	if err := database.GetDB().First(&playbook, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"msg": "Playbook not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbook})
}

// CreatePlaybook Create (Already exists，补充完整性)
func CreatePlaybook(ctx *gin.Context) {
	var req model.Playbook
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Args error"})
		return
	}
	req.IsActive = false // Default关闭
	database.GetDB().Create(&req)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": req})
}

// UpdatePlaybook UpdatePlaybook (Save画布)
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

	// Update字段
	playbook.Name = req.Name
	playbook.Description = req.Description
	playbook.Definition = req.Definition // 核心：Update React Flow JSON
	playbook.TriggerType = req.TriggerType
	playbook.IsActive = req.IsActive

	database.GetDB().Save(&playbook)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbook})
}

// DeletePlaybook DeletePlaybook
func DeletePlaybook(ctx *gin.Context) {
	id := ctx.Param("id")
	database.GetDB().Delete(&model.Playbook{}, id)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Deleted"})
}

// GetExecutionHistory GetExecute历史
func GetExecutionHistory(ctx *gin.Context) {
	id := ctx.Param("id") // Playbook ID
	var history []model.PlaybookExecution
	database.GetDB().Where("playbook_id = ?", id).Order("id desc").Limit(20).Find(&history)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": history})
}

// GetExecutionDetail Get单次ExecuteDetail (用于ago端轮询Status和Log)
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
