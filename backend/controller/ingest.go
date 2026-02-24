package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/ingest"
)

// CollectIngest 接收日志请求并投递到异步队列
func CollectIngest(ctx *gin.Context) {
	// 1. 从 IngestMiddleware 中获取已经查询好的配置对象
	val, exists := ctx.Get("ingest_config")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "未找到配置信息"})
		return
	}
	config := val.(*database.IngestCache)

	// 2. 解析日志 Payload
	// 这里使用 interface{} 以支持任意格式的 JSON 日志
	var payload interface{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的 JSON 格式"})
		return
	}

	// 3. 投递到全局 LogQueue
	// 如果队列 LogQueue 满了（10000个堆积），这里会发生阻塞（Backpressure）
	// 这可以防止内存溢出，并确保日志在 VictoriaLogs 恢复前不被丢弃
	ingest.LogQueue <- ingest.LogPayload{
		Config: *config,
		Data:   payload,
	}

	// 4. 立即返回 202 Accepted，表示日志已进入处理管道
	ctx.JSON(http.StatusAccepted, gin.H{
		"code": 202,
		"msg":  "Log accepted",
	})
}
