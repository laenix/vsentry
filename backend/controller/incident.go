package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// ListIncidents - （不带Detail，用于大屏展示）
func ListIncidents(ctx *gin.Context) {
	var incidents []model.Incident
	database.GetDB().Order("last_seen desc").Find(&incidents)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": incidents})
}

// GetIncidentDetail - (Alerts)
// GetIncidentDetail - func GetIncidentDetail(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少EventID"})
		return
	}

	var incident model.Incident
	// 使用 - ("Alerts") 自动Execute关联Query，Get该Event下的所有Evidence
	// 使用 - ("Rule") Get关联的RuleInfo（包括RuleType）
	db := database.GetDB()
	err := db.Preload("Alerts").Preload("Rule").First(&incident, id).Error

	if err != nil {
		ctx.JSON(http.StatusNot found, gin.H{"code": 404, "msg": "未找到相关Event记录"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": incident,
		"msg":  "success",
	})
}

// AcknowledgeIncident - func AcknowledgeIncident(ctx *gin.Context) {
	id := ctx.Query("id")
	userID, _ := ctx.Get("userid") // 从 - Get

	database.GetDB().Model(&model.Incident{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   "acknowledged",
		"assignee": userID,
	})
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已受理该Event"})
}

// ResolveIncident - func ResolveIncident(ctx *gin.Context) {
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
		ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "EventClosed"})
	}
}
