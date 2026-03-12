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

// RunPlaybook - /DebugPlaybook
func RunPlaybook(ctx *gin.Context) {
	//   1. Parse URL Parameter (Playbook ID)
	idStr := ctx.Param("id")
	playbookID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid playbook ID"})
		return
	}

	//   2. Parse Body Parameter (Incident Context)
	var req struct {
		IncidentID uint `json:"incident_id"`
		DryRun     bool `json:"dry_run"` //   目前先预留，暂不Handle
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid request body"})
		return
	}

	//   3. Build context (Context Injection)
	// 我们NeedQuery真实的 - Data注入到 global VariableMedium
	inputContext := make(map[string]interface{})

	if req.IncidentID > 0 {
		var incident model.Incident
		// 预Load - ，这样在Playbook里Can用 {{incident.Alerts}} 访问Evidence
		result := database.GetDB().Preload("Alerts").First(&incident, req.IncidentID)
		if result.Error == nil {
			// Convert - to map[string]interface{} 以便 expr Engine使用
			//   这里用了一个小技巧：JSON Marshal -> Unmarshal
			// 也Can直接传 - ，expr supports Struct field access, but map is more flexible
			inBytes, _ := json.Marshal(incident)
			var inMap map[string]interface{}
			json.Unmarshal(inBytes, &inMap)

			inputContext["incident"] = inMap
		}
	}

	//   4. StartEngine
	engine := automation.NewEngine()
	executionID, err := engine.Run(uint(playbookID), inputContext)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	//   5. ReturnExecute ID
	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "Playbook execution started",
		"data": gin.H{"execution_id": executionID},
	})
}

// ListPlaybooks - func ListPlaybooks(ctx *gin.Context) {
	var playbooks []model.Playbook
	// List页通常不NeedLoad巨大的 - JSON，Can节省带宽
	database.GetDB().Omit("definition").Order("updated_at desc").Find(&playbooks)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbooks})
}

// GetPlaybook - (包含画布定义)
func GetPlaybook(ctx *gin.Context) {
	id := ctx.Param("id")
	var playbook model.Playbook
	if err := database.GetDB().First(&playbook, id).Error; err != nil {
		ctx.JSON(http.StatusNot found, gin.H{"msg": "Playbook not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbook})
}

// CreatePlaybook - (Already exists，补充完整性)
func CreatePlaybook(ctx *gin.Context) {
	var req model.Playbook
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Args error"})
		return
	}
	req.IsActive = false // 默认Close - .GetDB().Create(&req)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": req})
}

// UpdatePlaybook - (Save画布)
func UpdatePlaybook(ctx *gin.Context) {
	id := ctx.Param("id")
	var req model.Playbook
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Args error"})
		return
	}

	var playbook model.Playbook
	if err := database.GetDB().First(&playbook, id).Error; err != nil {
		ctx.JSON(http.StatusNot found, gin.H{"msg": "Playbook not found"})
		return
	}

	// Update字段 - .Name = req.Name
	playbook.Description = req.Description
	playbook.Definition = req.Definition //   核心：Update React Flow JSON
	playbook.TriggerType = req.TriggerType
	playbook.IsActive = req.IsActive

	database.GetDB().Save(&playbook)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbook})
}

// DeletePlaybook - func DeletePlaybook(ctx *gin.Context) {
	id := ctx.Param("id")
	database.GetDB().Delete(&model.Playbook{}, id)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Deleted"})
}

// GetExecutionHistory - func GetExecutionHistory(ctx *gin.Context) {
	id := ctx.Param("id") // Playbook - var history []model.PlaybookExecution
	database.GetDB().Where("playbook_id = ?", id).Order("id desc").Limit(20).Find(&history)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": history})
}

// GetExecutionDetail - (用于前端轮询StatusSumLog)
func GetExecutionDetail(ctx *gin.Context) {
	//   注意：这里的Parameter名要与路由匹配，建议用 exec_id 区分 playbook_id
	executionID := ctx.Param("exec_id")

	var execution model.PlaybookExecution
	if err := database.GetDB().First(&execution, executionID).Error; err != nil {
		ctx.JSON(http.StatusNot found, gin.H{"msg": "Execution history not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": execution, // 包含 - Sum Logs (JSON)
	})
}

// ListAllExecutions - ListAllExecutions(ctx *gin.Context) {
	var executions []model.PlaybookExecution
	database.GetDB().Order("id desc").Limit(100).Find(&executions)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": executions})
}
