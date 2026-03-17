package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/crd"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// CRDController 处理 Playbook CRD 的导入导出
// 支持将 VSentry 内部的 Playbook 模型与 Kubernetes CRD 格式互相转换

// ImportPlaybookCRD 将 CRD YAML/JSON 导入为内部 Playbook 模型
// POST /api/playbooks/import
func ImportPlaybookCRD(ctx *gin.Context) {
	var req struct {
		Content string `json:"content"` // CRD YAML 或 JSON 内容
		Format  string `json:"format"`  // "yaml" 或 "json"
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误", "error": err.Error()})
		return
	}

	// 解析 CRD
	var playbookCRD crd.Playbook
	var err error

	if req.Format == "yaml" || req.Format == "" {
		// TODO: 使用 yaml 库解析 (需要添加依赖 gopkg.in/yaml.v3)
		// 目前先用 JSON 测试
		err = json.Unmarshal([]byte(req.Content), &playbookCRD)
	} else {
		err = json.Unmarshal([]byte(req.Content), &playbookCRD)
	}

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "CRD 解析失败", "error": err.Error()})
		return
	}

	// 转换为内部模型
	playbook := convertCRDToModel(&playbookCRD)

	// 保存到数据库
	database.GetDB().Create(&playbook)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "Playbook 导入成功",
		"data": playbook,
	})
}

// ExportPlaybookCRD 将内部 Playbook 模型导出为 CRD YAML/JSON
// GET /api/playbooks/:id/export?format=yaml|json
func ExportPlaybookCRD(ctx *gin.Context) {
	id := ctx.Param("id")
	playbookID, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid playbook ID"})
		return
	}

	format := ctx.DefaultQuery("format", "yaml")

	// 查询 Playbook
	var playbook model.Playbook
	if err := database.GetDB().First(&playbook, playbookID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "Playbook not found"})
		return
	}

	// 转换为 CRD
	playbookCRD := convertModelToCRD(&playbook)

	var result string
	if format == "json" {
		data, err := json.MarshalIndent(playbookCRD, "", "  ")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "JSON 序列化失败", "error": err.Error()})
			return
		}
		result = string(data)
	} else {
		// YAML 输出需要 yaml 库
		// TODO: 使用 yaml 库序列化
		data, err := json.MarshalIndent(playbookCRD, "", "  ")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "YAML 序列化失败", "error": err.Error()})
			return
		}
		// 简单模拟 YAML 格式 (实际项目中建议使用 gopkg.in/yaml.v3)
		result = "---\n" + string(data)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"name":    playbookCRD.Metadata.Name,
			"format":  format,
			"content": result,
		},
	})
}

// ListPlaybookCRDs 列出示例 CRD
// GET /api/playbooks/crd/examples
func ListPlaybookCRDs(ctx *gin.Context) {
	examples := []gin.H{
		{
			"name":        "detect-and-isolate-threat",
			"description": "检测可疑进程执行时，自动采集内存快照、隔离 Pod 并发送告警",
			"trigger":     "falco",
			"actions":      4,
		},
		{
			"name":        "conditional-response",
			"description": "根据告警严重程度执行不同的响应动作",
			"trigger":     "falco",
			"actions":      1,
		},
		{
			"name":        "tetragon-sensitive-file-access",
			"description": "检测容器内敏感文件访问并自动隔离",
			"trigger":     "tetragon",
			"actions":      4,
		},
		{
			"name":        "manual-investigation",
			"description": "手动触发的调查剧本，用于批量查询和报告生成",
			"trigger":     "manual",
			"actions":      2,
		},
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": examples})
}

// GetPlaybookCRDExample 获取示例 CRD 内容
// GET /api/playbooks/crd/examples/:name
func GetPlaybookCRDExample(ctx *gin.Context) {
	name := ctx.Param("name")

	// 实际项目中这里应该读取 examples/ 目录下的文件
	// 这里返回预定义的示例
	example := getCRDExample(name)
	if example == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "示例不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"name":    name,
			"content": example,
		},
	})
}

// ===== 内部转换函数 =====

// convertCRDToModel 将 CRD 转换为内部 Playbook 模型
func convertCRDToModel(crd *crd.Playbook) model.Playbook {
	playbook := model.Playbook{
		Name:        crd.Metadata.Name,
		Description: crd.Spec.Description,
		IsActive:    crd.Spec.Enabled,
		TriggerType: crd.Spec.Trigger.Source,
	}

	// 将 Trigger 和 Actions 序列化为 Definition JSON
	definition := map[string]interface{}{
		"trigger": crd.Spec.Trigger,
		"actions": crd.Spec.Actions,
	}

	defBytes, _ := json.Marshal(definition)
	playbook.Definition = defBytes

	return playbook
}

// convertModelToCRD 将内部 Playbook 模型转换为 CRD
func convertModelToCRD(playbook *model.Playbook) crd.Playbook {
	crdPlaybook := crd.Playbook{
		APIVersion: "vsentry.io/v1",
		Kind:       "Playbook",
		Metadata: crd.Metadata{
			Name: playbook.Name,
		},
		Spec: crd.PlaybookSpec{
			Description: playbook.Description,
			Enabled:    playbook.IsActive,
		},
	}

	// 从 Definition JSON 解析 Trigger 和 Actions
	if playbook.Definition != nil {
		var def map[string]interface{}
		json.Unmarshal(playbook.Definition, &def)

		if trigger, ok := def["trigger"].(map[string]interface{}); ok {
			crdPlaybook.Spec.Trigger = parseTrigger(trigger)
		}
		if actions, ok := def["actions"].([]interface{}); ok {
			crdPlaybook.Spec.Actions = parseActions(actions)
		}
	}

	// 设置触发类型
	crdPlaybook.Spec.Trigger.Source = playbook.TriggerType

	return crdPlaybook
}

func parseTrigger(m map[string]interface{}) crd.Trigger {
	trigger := crd.Trigger{
		Source:     getString(m, "source"),
		Severity:   getString(m, "severity"),
		Conditions: getStringSlice(m, "conditions"),
	}
	return trigger
}

func parseActions(arr []interface{}) []crd.Action {
	actions := make([]crd.Action, len(arr))
	for i, a := range arr {
		if m, ok := a.(map[string]interface{}); ok {
			action := crd.Action{
				Name: getString(m, "name"),
				Type: getString(m, "type"),
			}
			if config, ok := m["config"].(map[string]interface{}); ok {
				action.Config = parseActionConfig(config)
			}
			actions[i] = action
		}
	}
	return actions
}

func parseActionConfig(m map[string]interface{}) crd.ActionConfig {
	config := crd.ActionConfig{
		URL:      getString(m, "url"),
		Method:   getString(m, "method"),
		Body:     getString(m, "body"),
		Capture:  getString(m, "capture"),
		Timeout:  getString(m, "timeout"),
		Storage:  getString(m, "storage"),
		// K8s
		K8sAction: getString(m, "action"),
		Kind:      getString(m, "kind"),
		Namespace: getString(m, "namespace"),
		Selector:  getString(m, "selector"),
		// Email
		SMTPHost: getString(m, "smtpHost"),
		SMTPPort: getInt(m, "smtpPort"),
		To:       getString(m, "to"),
		Subject:  getString(m, "subject"),
		Content:  getString(m, "content"),
		// Expression
		Expression: getString(m, "expression"),
	}
	return config
}

// ===== 辅助函数 =====

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if arr, ok := m[key].([]interface{}); ok {
		result := make([]string, len(arr))
		for i, v := range arr {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

// getCRDExample 返回预定义的 CRD 示例
// TODO: 实际应该读取 examples/ 目录
func getCRDExample(name string) string {
	examples := map[string]string{
		"detect-and-isolate-threat": `apiVersion: vsentry.io/v1
kind: Playbook
metadata:
  name: detect-and-isolate-threat
  labels:
    severity: critical
    category: incident-response
spec:
  enabled: true
  description: |
    检测到可疑进程执行时，自动采集内存快照、隔离 Pod 并发送告警
  trigger:
    source: falco
    severity: critical
    conditions:
      - "severity = critical"
      - "category = container_escape"
  actions:
    - name: capture-memory
      type: forensics
      config:
        capture: memory
        timeout: 30s
    - name: isolate-pod
      type: kubernetes
      config:
        action: patch
        kind: Pod
        selector: {{ .incident.pod }}
    - name: notify
      type: webhook
      config:
        url: https://hooks.slack.com/xxx
        method: POST`,
		"conditional-response": `apiVersion: vsentry.io/v1
kind: Playbook
metadata:
  name: conditional-response
spec:
  trigger:
    source: falco
  actions:
    - name: check-severity
      type: condition
      config:
        expression: incident.severity == "critical"
      branchs:
        trueBranch:
          - name: critical-actions
            type: expression
        falseBranch:
          - name: low-severity-actions
            type: webhook`,
		"tetragon-sensitive-file-access": `apiVersion: vsentry.io/v1
kind: Playbook
metadata:
  name: tetragon-sensitive-file-access
spec:
  trigger:
    source: tetragon
    conditions:
      - "event_type = process_exec"
  actions:
    - name: alert
      type: email
      config:
        to: soc@company.com`,
		"manual-investigation": `apiVersion: vsentry.io/v1
kind: Playbook
metadata:
  name: manual-investigation
spec:
  trigger:
    source: manual
  actions:
    - name: query-alerts
      type: expression
      config:
        expression: query_logs("severity >= medium", last_24h)`,
	}
	return examples[name]
}
