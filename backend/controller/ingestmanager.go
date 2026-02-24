package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// ListIngest 获取列表
func ListIngest(ctx *gin.Context) {
	db := database.GetDB()
	var ingests []model.Ingest
	db.Find(&ingests)
	ctx.JSON(200, gin.H{"code": 200, "data": ingests})
}

// AddIngest 添加 Ingest 配置
func AddIngest(ctx *gin.Context) {
	var ingest model.Ingest
	if err := ctx.ShouldBindJSON(&ingest); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	database.GetDB().Create(&ingest)
	ctx.JSON(200, gin.H{"code": 200, "msg": "添加成功"})
}

// UpdateIngest 更新配置并失效缓存
func UpdateIngest(ctx *gin.Context) {
	var req model.Ingest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"msg": "参数错误"})
		return
	}

	db := database.GetDB()
	// 1. 获取所有关联的 Token，用于失效缓存
	var auths []model.IngestAuth
	db.Where("ingest_id = ?", req.ID).Find(&auths)

	// 2. 更新数据库
	if err := db.Model(&model.Ingest{}).Where("id = ?", req.ID).Updates(req).Error; err != nil {
		ctx.JSON(500, gin.H{"msg": "更新失败"})
		return
	}

	// 3. 【关键】清理 Badger 中的 Token 缓存
	// 这样下次日志进来时，中间件会重新从 SQLite 加载最新的 StreamFields
	for _, auth := range auths {
		database.DelTokenCache(auth.SecretKey)
	}

	ctx.JSON(200, gin.H{"code": 200, "msg": "更新成功，缓存已同步"})
}

// DeleteIngest 删除配置并失效缓存
func DeleteIngest(ctx *gin.Context) {
	id := ctx.Query("id")
	db := database.GetDB()

	var auths []model.IngestAuth
	db.Where("ingest_id = ?", id).Find(&auths)

	// 清理缓存
	for _, auth := range auths {
		database.DelTokenCache(auth.SecretKey)
	}

	db.Delete(&model.Ingest{}, id)
	ctx.JSON(200, gin.H{"code": 200, "msg": "删除成功"})
}

// GetIngestAuth 获取指定 Ingest 的 Token
func GetIngestAuth(ctx *gin.Context) {
	id := ctx.Param("id")
	db := database.GetDB()

	var auth model.IngestAuth
	if err := db.Where("ingest_id = ?", id).First(&auth).Error; err != nil {
		ctx.JSON(404, gin.H{"code": 404, "msg": "Token not found"})
		return
	}

	ctx.JSON(200, gin.H{"code": 200, "data": gin.H{"token": auth.SecretKey}})
}
