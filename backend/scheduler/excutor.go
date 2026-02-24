package scheduler

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/laenix/vsentry/automation"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

// ExecuteRule 执行规则查询
func ExecuteRule(rule model.Rule) {
	// 1. 从配置获取 VictoriaLogs 地址
	vLogsAddr := viper.GetString("victorialogs.url")
	if vLogsAddr == "" {
		vLogsAddr = "http://127.0.0.1:9428" // 默认地址
	}

	// 构造 VictoriaLogs 查询请求
	resp, err := http.PostForm(vLogsAddr+"/select/logsql/query", url.Values{"query": {rule.Query}})
	if err != nil {
		log.Printf("[Rule:%d] Request failed: %v", rule.ID, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		return
	}

	// 触发告警入库，包含你喜欢的 Label 和 Assignee
	saveAlert(rule, string(body))
}

func saveAlert(rule model.Rule, evidence string) {
	db := database.GetDB()
	now := time.Now()

	// 1. 查找是否存在该规则的活跃事件 (未解决的)
	var incident model.Incident
	err := db.Where("rule_id = ? AND status != ?", rule.ID, "resolved").
		Order("last_seen desc").First(&incident).Error

	if err != nil {
		// 创建新 Incident
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

	// 2. 检查指纹去重（防止同一条原始日志重复生成证据）
	fp := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d-%s", rule.ID, evidence))))
	var alert model.Alert
	if db.Where("fingerprint = ?", fp).First(&alert).Error != nil {
		// 存储新证据并关联到 Incident
		alert = model.Alert{
			IncidentID:  incident.ID,
			RuleID:      rule.ID,
			Content:     evidence,
			Fingerprint: fp,
		}
		db.Create(&alert)

		// 3. 更新事件的统计信息
		db.Model(&incident).Updates(map[string]interface{}{
			"alert_count": incident.AlertCount + 1,
			"last_seen":   now,
		})
	}
	go automation.DispatchByIncident(incident)
}
