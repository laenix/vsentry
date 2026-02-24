package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

// ListCustomTables 获取自定义表列表
func ListCustomTables(ctx *gin.Context) {
	db := database.GetDB()
	var tables []model.CustomTable
	db.Where("is_active = ?", true).Find(&tables)
	ctx.JSON(200, gin.H{"code": 200, "data": tables})
}

// AddCustomTable 添加自定义表
func AddCustomTable(ctx *gin.Context) {
	var req model.CustomTable
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	if req.Name == "" || req.StreamFields == "" {
		ctx.JSON(400, gin.H{"msg": "Name and StreamFields are required"})
		return
	}

	req.IsActive = true
	database.GetDB().Create(&req)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Created successfully", "data": req})
}

// UpdateCustomTable 更新自定义表
func UpdateCustomTable(ctx *gin.Context) {
	var req model.CustomTable
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	if req.ID == 0 {
		ctx.JSON(400, gin.H{"msg": "ID is required"})
		return
	}

	db := database.GetDB()
	var existing model.CustomTable
	if err := db.First(&existing, req.ID).Error; err != nil {
		ctx.JSON(404, gin.H{"msg": "Not found"})
		return
	}

	db.Model(&existing).Updates(req)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Updated successfully"})
}

// DeleteCustomTable 删除自定义表（软删除）
func DeleteCustomTable(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(400, gin.H{"msg": "ID is required"})
		return
	}

	db := database.GetDB()
	var table model.CustomTable
	if err := db.First(&table, id).Error; err != nil {
		ctx.JSON(404, gin.H{"msg": "Not found"})
		return
	}

	db.Model(&table).Update("is_active", false)
	ctx.JSON(200, gin.H{"code": 200, "msg": "Deleted successfully"})
}

// GetExternalURL 获取外部访问URL（给前端用）
func GetExternalURL(ctx *gin.Context) {
	externalURL := viper.GetString("server.external_url")
	if externalURL == "" {
		externalURL = "http://localhost:8088"
	}

	// 构造完整的 Ingest URL
	data := gin.H{
		"external_url":       externalURL,
		"ingest_endpoint":    externalURL + "/api/ingest/collect",
		"query_endpoint":     externalURL + "/select/logsql/query",
		"metrics_endpoint":   externalURL + "/metrics",
	}

	ctx.JSON(200, gin.H{"code": 200, "data": data})
}