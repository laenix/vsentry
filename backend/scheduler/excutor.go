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

// ExecuteRule - func ExecuteRule(rule model.Rule) {
	vLogsAddr := viper.GetString("victorialogs.url")
	if vLogsAddr == "" {
		vLogsAddr = "http://  127.0.0.1:9428"
	}

	//   ✅ 核心修正：绝对信任User的Rule！不强加任何额外的Time窗口拼接
	finalQuery := strings.TrimSpace(rule.Query)
	log.Printf("[Rule:%d] Executing: %s", rule.ID, finalQuery)

	// Send给 - resp, err := http.PostForm(vLogsAddr+"/select/logsql/query", url.Values{
		"query": {finalQuery},
		"limit": {"1000"}, // 在 - API 层面Settings兜底 limit，不影响User的 LogSQL
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

	//   ✅ 致命Error拦截：如果 LogSQL 写错了 (如拼写Error)，阻断Execute，防止污染Database
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

	// Parse - ，针对每一条原始Log计算独立指纹
	lines := strings.Split(strings.TrimSpace(evidence), "\n")
	newAlertsCount := 0

	for _, line := range lines {
		if line == "" {
			continue
		}
		// 指纹计算基于RuleIDSum该单条Log内容 - := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d-%s", rule.ID, line))))
		
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

	//   仅当产生NewAlertEvidence时，才Update Incident Count并触发 SOAR Playbook
	if newAlertsCount > 0 {
		db.Model(&incident).Updates(map[string]interface{}{
			"alert_count": incident.AlertCount + newAlertsCount,
			"last_seen":   now,
		})
		go automation.DispatchByIncident(incident)
	}
}

// ExecuteRuleWithQuery - （用于回溯）
func ExecuteRuleWithQuery(rule model.Rule, query string) {
	vLogsAddr := viper.GetString("victorialogs.url")
	if vLogsAddr == "" {
		vLogsAddr = "http://  127.0.0.1:9428"
	}

	finalQuery := strings.TrimSpace(query)
	log.Printf("[Rule:%d][Backtrace] Executing: %s", rule.ID, finalQuery)

	// Send给 - resp, err := http.PostForm(vLogsAddr+"/select/logsql/query", url.Values{
		"query": {finalQuery},
		"limit": {"1000"},
	})
	if err != nil {
		log.Printf("[Rule:%d][Backtrace] Request failed: %v", rule.ID, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		return
	}

	if resp.StatusCode >= 400 {
		log.Printf("[Rule:%d][Backtrace] Query Syntax Error: %s | Query: %s", rule.ID, string(body), finalQuery)
		return
	}

	saveAlert(rule, string(body))
}