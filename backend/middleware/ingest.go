package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

func IngestMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			ctx.JSON(401, gin.H{"code": 401, "msg": "Token 验证失败"})
			ctx.Abort()
			return
		}
		token = token[7:]

		// 1. 优先查 Badger
		cache, err := database.GetTokenCache(token)
		if err == nil {
			ctx.Set("ingest_config", cache)
			ctx.Next()
			return
		}

		// 2. 缓存未命中，查数据库
		db := database.GetDB()
		var auth model.IngestAuth
		if err := db.Where("secret_key = ?", token).First(&auth).Error; err != nil {
			ctx.JSON(401, gin.H{"code": 401, "msg": "Token 无效"})
			ctx.Abort()
			return
		}

		var target model.Ingest
		if err := db.First(&target, auth.IngestID).Error; err != nil {
			ctx.JSON(404, gin.H{"code": 404, "msg": "配置不存在"})
			ctx.Abort()
			return
		}

		// 3. 存入 Badger 供下次使用
		config := database.IngestCache{
			ID:           target.ID,
			Endpoint:     target.Endpoint,
			StreamFields: target.StreamFields,
		}
		_ = database.SetTokenCache(token, config)

		ctx.Set("ingest_config", &config)
		ctx.Next()
	}
}
