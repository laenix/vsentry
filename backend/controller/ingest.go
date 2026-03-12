package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/ingest"
)

// CollectIngest - JSONL LogRequest并投递到AsyncQueue
func CollectIngest(ctx *gin.Context) {
	val, exists := ctx.Get("ingest_config")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "未找到ConfigInfo"})
		return
	}
	config := val.(*database.IngestCache)

	//   【核心改造】：放弃 ShouldBindJSON，使用流式解码器逐行读取 JSONL
	decoder := json.NewDecoder(ctx.Request.Body)

	validCount := 0
	for {
		// 这里用 - [string]interface{} 来兜住 OCSF Event
		var event map[string]interface{}

		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break //   读到Request体末尾，正常结束
			}
			// 如果某一行 - 损坏，SkipContinue读下一行，而不是直接阻断整个批次
			continue
		}

		//   读出一个Event，就立刻投递到全局 LogQueue
		ingest.LogQueue <- ingest.LogPayload{
			Config: *config,
			Data:   event,
		}
		validCount++
	}

	if validCount == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid JSONL Data或Data为空"})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"code": 202,
		"msg":  "Logs accepted",
	})
}
