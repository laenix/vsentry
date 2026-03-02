package controller

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// QueryVictoriaLogs 代理查询请求到 VictoriaLogs
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

	// 构建查询参数
	params := url.Values{}
	if query != "" {
		params.Set("query", query)
	}
	if limit != "" {
		params.Set("limit", limit)
	}
	// VictoriaLogs 需要 ISO 格式时间（不带 Z 后缀或毫秒）或 Unix 时间戳
	if start != "" {
		// 移除 UTC 时区的 Z 后缀和毫秒
		start = strings.TrimSuffix(start, "Z")
		if idx := strings.Index(start, "."); idx != -1 {
			start = start[:idx] // 移除毫秒部分
		}
		params.Set("start", start)
	}
	if end != "" {
		end = strings.TrimSuffix(end, "Z")
		if idx := strings.Index(end, "."); idx != -1 {
			end = end[:idx] // 移除毫秒部分
		}
		params.Set("end", end)
	}

	targetURL := vlURL + "/select/logsql/query?" + params.Encode()

	// 创建请求
	req, err := http.NewRequest("POST", targetURL, ctx.Request.Body)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 透传响应
	ctx.Header("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(ctx.Writer, resp.Body)
}

// QueryVictoriaLogsHits 查询命中数
func QueryVictoriaLogsHits(ctx *gin.Context) {
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	query := ctx.PostForm("query")

	targetURL := vlURL + "/select/logsql/hits?" + url.Values{
		"query": {query},
	}.Encode()

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

// GetVictoriaLogsMetrics 获取 VictoriaLogs 指标
func GetVictoriaLogsMetrics(ctx *gin.Context) {
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	// 转发 /metrics 请求
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

// ProxyVictoriaLogsSelect 通用代理（透传所有 /select/* 请求）
func ProxyVictoriaLogsSelect(ctx *gin.Context) {
	path := ctx.Param("path")
	
	vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://localhost:9428"
	}

	// 转发原始请求
	targetURL := vlURL + "/select/" + path
	
	// 如果有查询参数，追加上去
	if ctx.Request.URL.RawQuery != "" {
		targetURL += "?" + ctx.Request.URL.RawQuery
	}

	req, err := http.NewRequest(ctx.Request.Method, targetURL, ctx.Request.Body)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 复制请求头
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

	// 透传响应头
	for k, v := range resp.Header {
		for _, vv := range v {
			ctx.Header(k, vv)
		}
	}
	io.Copy(ctx.Writer, resp.Body)
}