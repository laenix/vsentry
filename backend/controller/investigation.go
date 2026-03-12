package controller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
)

// ExecuteInvestigation - (仅支持 Rule Center 的 rule_id)
func ExecuteInvestigation(ctx *gin.Context) {
	var req struct {
		RuleID     uint              `json:"rule_id" binding:"required"`
		IncidentID uint              `json:"incident_id"`
		Params     map[string]string `json:"params"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid request body"})
		return
	}

	db := database.GetDB()

	// Query - (type = "investigation")
	var rule model.Rule
	if err := db.First(&rule, req.RuleID).Error; err != nil {
		ctx.JSON(http.StatusNot found, gin.H{"msg": "Rule not found"})
		return
	}

	// ValidateRuleType - rule.Type != "investigation" {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "Rule is not an investigation rule"})
		return
	}

	logSQL := rule.Query
	ruleName := rule.Name

	//   初始化Variable池 (Variables Pool)
	vars := make(map[string]string)

	// Settings默认Time范围 - := time.Now().AddDate(-1, 0, 0).UTC().Format(time.RFC3339)
	oneYearLater := time.Now().AddDate(1, 0, 0).UTC().Format(time.RFC3339)
	vars["start_time"] = oneYearAgo
	vars["end_time"] = oneYearLater

	// Load - 及其关联的 Alerts
	if req.IncidentID > 0 {
		var incident model.Incident
		if err := db.Preload("Alerts").First(&incident, req.IncidentID).Error; err == nil {
			vars["incident_id"] = fmt.Sprintf("%d", incident.ID)
			vars["incident_name"] = incident.Name

			vars["start_time"] = incident.FirstSeen.Add(-2 * time.Hour).UTC().Format(time.RFC3339)
			vars["end_time"] = incident.LastSeen.Add(2 * time.Hour).UTC().Format(time.RFC3339)

			if len(incident.Alerts) > 0 {
				firstAlert := incident.Alerts[0]
				if firstAlert.Content != "" {
					var alertContent map[string]interface{}
					if err := json.Unmarshal([]byte(firstAlert.Content), &alertContent); err == nil {
						flattenMap(alertContent, "", vars)

						if val, ok := vars["observer.hostname"]; ok {
							vars["hostname"] = val
						}
						if val, ok := vars["src_endpoint.ip"]; ok {
							vars["src_ip"] = val
						}
						if val, ok := vars["target_user.name"]; ok {
							vars["username"] = val
						}
						if val, ok := vars["actor.user.name"]; ok {
							vars["username"] = val
						}
						if val, ok := vars["process.name"]; ok {
							vars["process_name"] = val
						}
					}
				}
			}
		}
	}

	// 手动Parameter覆盖 - k, v := range req.Params {
		vars[k] = v
	}

	// 动态替换 - Medium的Parameter
	finalLogSQL := logSQL
	for key, val := range vars {
		placeholder := fmt.Sprintf("${%s}", key)
		finalLogSQL = strings.ReplaceAll(finalLogSQL, placeholder, val)
	}

	// 拦截未被替换的Parameter - strings.Contains(finalLogSQL, "${") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg":           "Missing required parameters in rule",
			"query_preview": finalLogSQL,
			"current_vars":  vars,
		})
		return
	}

	// 调用 - vlURL := viper.GetString("victorialogs.url")
	if vlURL == "" {
		vlURL = "http://  localhost:9428"
	}

	params := url.Values{}
	params.Set("query", finalLogSQL)

	limit := vars["_limit"]
	if limit == "" {
		limit = "1000"
	}
	params.Set("limit", limit)

	targetURL := vlURL + "/select/logsql/query?" + params.Encode()
	vlReq, err := http.NewRequest("POST", targetURL, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	vlReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(vlReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "VictoriaLogs unreachable or timeout: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":  "VictoriaLogs Query Failed",
			"detail": string(bodyBytes),
			"logsql": finalLogSQL,
		})
		return
	}

	// Parse - var results []map[string]interface{}
	scanner := bufio.NewScanner(resp.Body)

	const maxCapacity = 10 * 1024 * 1024 // 10MB - := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row map[string]interface{}
		if err := json.Unmarshal([]byte(line), &row); err == nil {
			results = append(results, row)
		} else {
			fmt.Printf("[Investigation] JSON Parse error on line: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("[Investigation] Scanner error: %v\n", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"logsql":        finalLogSQL,
			"rule_name":     ruleName,
			"events":        results,
			"count":         len(results),
			"context_used":  vars,
		},
	})
}

func flattenMap(data map[string]interface{}, prefix string, result map[string]string) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch child := v.(type) {
		case map[string]interface{}:
			flattenMap(child, key, result)
		case string:
			result[key] = child
		case float64, int, bool:
			result[key] = fmt.Sprintf("%v", child)
		}
	}
}
