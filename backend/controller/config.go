package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// GetConfig GetSystem配置（公开Info）
func GetConfig(ctx *gin.Context) {
	externalURL := viper.GetString("server.external_url")
	if externalURL == "" {
		externalURL = "http://localhost:8088"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"external_url": externalURL,
		},
	})
}