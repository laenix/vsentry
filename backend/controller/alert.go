package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// ListAlerts - Get alert list
func ListAlerts(ctx *gin.Context) {
	var alerts []model.Alert
	database.GetDB().Order("id desc").Find(&alerts)
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": alerts, "msg": "success"})
}

// Acknowledge - Acknowledge an alert
func Acknowledge(ctx *gin.Context) {
	id := ctx.Query("id")
	userID, _ := ctx.Get("userid")

	db := database.GetDB()
	err := db.Model(&model.Alert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":   "acknowledged",
		"assignee": userID,
	}).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to acknowledge"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Alert acknowledged"})
}

// Resolve - Resolve and close an alert
func Resolve(ctx *gin.Context) {
	var req struct {
		ID             uint   `json:"id"`
		Classification string `json:"classification"`
		Comment        string `json:"comment"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameter"})
		return
	}

	db := database.GetDB()
	err := db.Model(&model.Alert{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"status":                 "resolved",
		"closing_classification": req.Classification,
		"closing_comment":        req.Comment,
	}).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to close"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Alert resolved"})
}

// Assign - Assign or transfer an alert
func Assign(ctx *gin.Context) {
	var req struct {
		ID     uint `json:"id"`
		UserID uint `json:"user_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid parameter"})
		return
	}

	db := database.GetDB()
	err := db.Model(&model.Alert{}).Where("id = ?", req.ID).Updates(map[string]interface{}{
		"assignee": req.UserID,
		"status":   "acknowledged",
	}).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to assign"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Alert assigned"})
}
