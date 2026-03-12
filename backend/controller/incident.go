package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// ListIncidents GetEventList（不带Detail，用于大屏展示）
func ListIncidents(ctx *gin.Context) {
	var incidents []model.Incident
	database.GetDB().Order("last_seen desc").Find(&incidents)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": incidents})
}

// GetIncidentDetail GetEvent及其关联的所有Evidence (Alerts)
// GetIncidentDetail GetEventDetail及其包含的所有AlertEvidence
func GetIncidentDetail(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少事件ID"})
		return
	}

	var incident model.Incident
	// 使用 Preload("Alerts") 自动Execute关联Query，Get该Event下的所有Evidence
	// 使用 Preload("Rule") Get关联的RuleInfo（包括RuleType）
	db := database.GetDB()
	err := db.Preload("Alerts").Preload("Rule").First(&incident, id).Error

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到相关事件记录"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": incident,
		"msg":  "success",
	})
}

// AcknowledgeIncident 认领Event
func AcknowledgeIncident(ctx *gin.Context) {
	id := ctx.Query("id")
	userID, _ := ctx.Get("userid") // 从 AuthMiddleware Get

	database.GetDB().Model(&model.Incident{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   "acknowledged",
		"assignee": userID,
	})
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已受理该事件"})
}

// ResolveIncident 解决Event
func ResolveIncident(ctx *gin.Context) {
	var req struct {
		ID             uint   `json:"id"`
		Classification string `json:"classification"`
		Comment        string `json:"comment"`
	}
	if err := ctx.ShouldBindJSON(&req); err == nil {
		database.GetDB().Model(&model.Incident{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"status":                 "resolved",
			"closing_classification": req.Classification,
			"closing_comment":        req.Comment,
		})
		ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "事件已关闭"})
	}
}
