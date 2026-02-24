package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// ListIncidents 获取事件列表（不带详情，用于大屏展示）
func ListIncidents(ctx *gin.Context) {
	var incidents []model.Incident
	database.GetDB().Order("last_seen desc").Find(&incidents)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": incidents})
}

// GetIncidentDetail 获取事件及其关联的所有证据 (Alerts)
// GetIncidentDetail 获取事件详情及其包含的所有告警证据
func GetIncidentDetail(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少事件ID"})
		return
	}

	var incident model.Incident
	// 使用 Preload("Alerts") 自动执行关联查询，获取该事件下的所有证据
	db := database.GetDB()
	err := db.Preload("Alerts").First(&incident, id).Error

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

// AcknowledgeIncident 认领事件
func AcknowledgeIncident(ctx *gin.Context) {
	id := ctx.Query("id")
	userID, _ := ctx.Get("userid") // 从 AuthMiddleware 获取

	database.GetDB().Model(&model.Incident{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   "acknowledged",
		"assignee": userID,
	})
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已受理该事件"})
}

// ResolveIncident 解决事件
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
