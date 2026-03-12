package controller

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// QueryVictoriaLogs 代理QueryRequest到 VictoriaLogs
func QueryVictoriaLogs(ctx *gin.Context) {
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	// 构建目标URL
	query := ctx.PostForm("query")
	limit := ctx.PostForm("limit")
	start := ctx.PostForm("start")
	end := ctx.PostForm("end")

	// 构建Query参数
	params := url.Values{}
	if query != "" {
		params.Set("query", query)
	}
	if limit != "" {
		params.Set("limit", limit)
	}
	// VictoriaLogs Need ISO 格式Time（不带 Z 后缀或毫seconds）或 Unix Time戳
	if start != "" {
		// 移除 UTC 时区的 Z 后缀和毫seconds
		start = strings.TrimSuffix(start, "Z")
		if idx := strings.Index(start, "."); idx != -1 {
			start = start[:idx] // 移除毫seconds部分
		}
		params.Set("start", start)
	}
	if end != "" {
		end = strings.TrimSuffix(end, "Z")
		if idx := strings.Index(end, "."); idx != -1 {
			end = end[:idx] // 移除毫seconds部分
		}
		params.Set("end", end)
	}

	targetURL := vlURL + "/select/logsql/query?" + params.Encode()

	// CreateRequest
	req, err := http.NewRequest("POST", targetURL, ctx.Request.Body)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// SendRequest
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 透传Response
	ctx.Header("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(ctx.Writer, resp.Body)
}

// QueryVictoriaLogsHits Query命Medium数
func QueryVictoriaLogsHits(ctx *gin.Context) {
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	query := ctx.PostForm("query")
	step := ctx.PostForm("step")

	// 构建Query参数
	params := url.Values{}
	if query != "" {
		params.Set("query", query)
	}
	if step != "" {
		params.Set("step", step)
	}

	targetURL := vlURL + "/select/logsql/hits?" + params.Encode()

	req, err := http.NewRequest("POST", targetURL, ctx.Request.Body)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	ctx.Header("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(ctx.Writer, resp.Body)
}

// GetVictoriaLogsMetrics Get VictoriaLogs 指标
func GetVictoriaLogsMetrics(ctx *gin.Context) {
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	// Transfer /metrics Request
	req, err := http.NewRequest("GET", vlURL+"/metrics", nil)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	ctx.Header("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(ctx.Writer, resp.Body)
}

// GetVictoriaLogsHealth 健康检查
func GetVictoriaLogsHealth(ctx *gin.Context) {
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	// 尝试访问健康检查端点
	resp, err := http.Get(vlURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		ctx.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	defer resp.Body.Close()

	ctx.JSON(200, gin.H{"status": "healthy"})
}

// ProxyVictoriaLogsSelect 通用代理（透传所有 /select/* Request）
func ProxyVictoriaLogsSelect(ctx *gin.Context) {
	path := ctx.Param("path")
	
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	// Transfer原始Request
	targetURL := vlURL + "/select/" + path
	
	// 如果有Query参数，追加上去
	if ctx.Request.URL.RawQuery != "" {
		targetURL += "?" + ctx.Request.URL.RawQuery
	}

	req, err := http.NewRequest(ctx.Request.Method, targetURL, ctx.Request.Body)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 复制Request头
	for k, v := range ctx.Request.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 透传Response头
	for k, v := range resp.Header {
		for _, vv := range v {
			ctx.Header(k, vv)
		}
	}
	io.Copy(ctx.Writer, resp.Body)
}