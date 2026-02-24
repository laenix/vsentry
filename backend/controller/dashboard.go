package controller

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

type DashboardStats struct {
	TotalAlerts    int64            `json:"total_alerts"`
	NewAlerts      int64            `json:"new_alerts"`
	SeverityCounts map[string]int64 `json:"severity_counts"`
	// 从 /metrics 提取的实时指标
	VLogsMetrics map[string]float64 `json:"vlogs_metrics"`
}

func GetDashboard(ctx *gin.Context) {
	db := database.GetDB()
	var stats DashboardStats
	stats.SeverityCounts = make(map[string]int64)

	// 1. 内部告警统计
	db.Model(&model.Alert{}).Count(&stats.TotalAlerts)
	db.Model(&model.Alert{}).Where("status = ?", "New").Count(&stats.NewAlerts)

	// 2. 获取并解析 VictoriaLogs 原始 Metrics
	stats.VLogsMetrics = fetchAndParseVLogsMetrics()

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": stats,
		"msg":  "success",
	})
}

func fetchAndParseVLogsMetrics() map[string]float64 {
	results := make(map[string]float64)

	vLogsAddr := viper.GetString("victorialogs.url")
	if vLogsAddr == "" {
		vLogsAddr = "http://localhost:9428"
	}

	resp, err := http.Get(vLogsAddr + "/metrics")
	if err != nil {
		return results
	}
	defer resp.Body.Close()

	// 我们想要关注的几个核心指标 Key
	targets := map[string]bool{
		"vm_rows_inserted_total":    true, // 总插入行数
		"vl_active_streams":         true, // 当前活跃的日志流
		"vm_free_disk_space_bytes":  true, // 剩余磁盘空间
		"vl_log_storage_size_bytes": true, // 日志占用的磁盘空间
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// 略过注释行
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		// 处理带 Label 的 Key，如 vm_rows_inserted_total{type="log"}
		cleanKey := key
		if strings.Contains(key, "{") {
			cleanKey = strings.Split(key, "{")[0]
		}

		if targets[cleanKey] {
			val, _ := strconv.ParseFloat(parts[1], 64)
			results[cleanKey] = val
		}
	}

	return results
}
