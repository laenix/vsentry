package scheduler

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/laenix/vsentry/automation"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

// ExecuteRule 执行规则查询
func ExecuteRule(rule model.Rule) {
	vLogsAddr := viper.GetString("victorialogs.url")
	if vLogsAddr == "" {
		vLogsAddr = "http://127.0.0.1:9428"
	}

	// ✅ 核心修正：绝对信任用户的规则！不强加任何额外的时间窗口拼接
	finalQuery := strings.TrimSpace(rule.Query)
	log.Printf("[Rule:%d] Executing: %s", rule.ID, finalQuery)

	// 发送给 VictoriaLogs
	resp, err := http.PostForm(vLogsAddr+"/select/logsql/query", url.Values{
		"query": {finalQuery},
		"limit": {"1000"}, // 在 HTTP API 层面设置兜底 limit，不影响用户的 LogSQL
	})
	if err != nil {
		log.Printf("[Rule:%d] Request failed: %v", rule.ID, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		return
	}

	// ✅ 致命错误拦截：如果 LogSQL 写错了 (如拼写错误)，阻断执行，防止污染数据库
	if resp.StatusCode >= 400 {
		log.Printf("[Rule:%d] Query Syntax Error: %s | Query: %s", rule.ID, string(body), finalQuery)
		return
	}

	saveAlert(rule, string(body))
}

func saveAlert(rule model.Rule, evidence string) {
	db := database.GetDB()
	now := time.Now().UTC()

	var incident model.Incident
	err := db.Where("rule_id = ? AND status != ?", rule.ID, "resolved").
		Order("last_seen desc").First(&incident).Error

	if err != nil {
		incident = model.Incident{
			RuleID:    rule.ID,
			Name:      rule.Name,
			Severity:  rule.Severity,
			Status:    "new",
			FirstSeen: now,
			LastSeen:  now,
		}
		db.Create(&incident)
	}

	// 解析 NDJSON，针对每一条原始日志计算独立指纹
	lines := strings.Split(strings.TrimSpace(evidence), "\n")
	newAlertsCount := 0

	for _, line := range lines {
		if line == "" {
			continue
		}
		// 指纹计算基于规则ID和该单条日志内容
		fp := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d-%s", rule.ID, line))))
		
		var count int64
		db.Model(&model.Alert{}).Where("fingerprint = ?", fp).Count(&count)
		
		if count == 0 {
			alert := model.Alert{
				IncidentID:  incident.ID,
				RuleID:      rule.ID,
				Content:     line,
				Fingerprint: fp,
			}
			db.Create(&alert)
			newAlertsCount++
		}
	}

	// 仅当产生新告警证据时，才更新 Incident 计数并触发 SOAR 剧本
	if newAlertsCount > 0 {
		db.Model(&incident).Updates(map[string]interface{}{
			"alert_count": incident.AlertCount + newAlertsCount,
			"last_seen":   now,
		})
		go automation.DispatchByIncident(incident)
	}
}