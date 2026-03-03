package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/ingest"
)

// CollectIngest 接收 JSONL 日志请求并投递到异步队列
func CollectIngest(ctx *gin.Context) {
	val, exists := ctx.Get("ingest_config")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "未找到配置信息"})
		return
	}
	config := val.(*database.IngestCache)

	// 【核心改造】：放弃 ShouldBindJSON，使用流式解码器逐行读取 JSONL
	decoder := json.NewDecoder(ctx.Request.Body)

	validCount := 0
	for {
		// 这里用 map[string]interface{} 来兜住 OCSF 事件
		var event map[string]interface{}

		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break // 读到请求体末尾，正常结束
			}
			// 如果某一行 JSON 损坏，跳过继续读下一行，而不是直接阻断整个批次
			continue
		}

		// 读出一个事件，就立刻投递到全局 LogQueue
		ingest.LogQueue <- ingest.LogPayload{
			Config: *config,
			Data:   event,
		}
		validCount++
	}

	if validCount == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的 JSONL 数据或数据为空"})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"code": 202,
		"msg":  "Logs accepted",
	})
}
