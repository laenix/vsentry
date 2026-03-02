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

	// VictoriaLogs 接受 ISO8601 格式，需去除毫秒
	twelveHoursAgo := time.Now().UTC().Add(-12 * time.Hour).Format("2006-01-02T15:04:05Z")
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// 【修复1】：强制组合查询语句，注入严格的时间边界
	finalQuery := fmt.Sprintf("(%s) AND _time:[%s, %s]", rule.Query, twelveHoursAgo, now)
	log.Printf("[Rule:%d] Executing: %s", rule.ID, finalQuery)

	resp, err := http.PostForm(vLogsAddr+"/select/logsql/query", url.Values{
		"query": {finalQuery},
		"limit": {"1000"},
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

	saveAlert(rule, string(body))
}

func saveAlert(rule model.Rule, evidence string) {
	db := database.GetDB()
	now := time.Now()

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

	// 【修复2】：解析 NDJSON，针对每一条原始日志计算独立指纹
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