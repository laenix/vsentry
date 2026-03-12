package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// ListAlerts GetAlertList
func ListAlerts(ctx *gin.Context) {
	var alerts []model.Alert
	database.GetDB().Order("id desc").Find(&alerts)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": alerts, "msg": "success"})
}

// Acknowledge 认领Alert
func Acknowledge(ctx *gin.Context) {
	id := ctx.Query("id")
	userID, _ := ctx.Get("userid") // 从 AuthMiddleware Get

	db := database.GetDB()
	err := db.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   "acknowledged",
		"assignee": userID,
	}).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "认领失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已成功认领"})
}

// Resolve 解决并关闭Alert
func Resolve(ctx *gin.Context) {
	var req struct {
		ID             uint   `json:"id"`
		Classification string `json:"classification"`
		Comment        string `json:"comment"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	db := database.GetDB()
	err := db.Model(&model.Alert{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"status":                 "resolved",
		"closing_classification": req.Classification,
		"closing_comment":        req.Comment,
	}).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "关闭失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "告警已解决"})
}

// Assign Transfer或指派Alert
func Assign(ctx *gin.Context) {
	var req struct {
		ID     uint `json:"id"`
		UserID uint `json:"user_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	db := database.GetDB()
	// Update逻辑：Settings受理人，如果原Status是 new 则自动转为 acknowledged
	err := db.Model(&model.Alert{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"assignee": req.UserID,
		"status":   "acknowledged",
	}).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "指派失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "指派成功"})
}
