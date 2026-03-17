package crd

// Playbook 是 VSentry SOAR 剧本的 Kubernetes CRD 定义
// 支持声明式配置，可通过 kubectl apply -f playbook.yaml 部署
//
// 示例用法:
//
// apiVersion: vsentry.io/v1
// kind: Playbook
// metadata:
//   name: detect-and-isolate-threat
// spec:
//   trigger:
//     source: falco
//     conditions:
//       - severity = critical
//       - category = container_escape
//   actions:
//     - name: capture-memory
//       type: forensics
//       config:
//         capture: memory
//         timeout: 30s
//     - name: isolate-pod
//       type: kubernetes
//       config:
//         action: patch
//         patch:
//           spec:
//             tolerations: []
//     - name: notify
//       type: webhook
//       config:
//         url: https://hooks.slack.com/xxx
//         method: POST

type Playbook struct {
	APIVersion string `json:"apiVersion"` // vsentry.io/v1
	Kind       string `json:"kind"`       // Playbook
	Metadata   Metadata `json:"metadata"`
	Spec       PlaybookSpec `json:"spec"`
}

type Metadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace,omitempty"`   // 默认 default
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type PlaybookSpec struct {
	// Trigger 定义触发条件
	Trigger Trigger `json:"trigger"`

	// Actions 定义响应动作链
	Actions []Action `json:"actions"`

	// 可选：启用状态
	Enabled bool `json:"enabled,omitempty"`

	// 可选：描述
	Description string `json:"description,omitempty"`
}

// Trigger 定义告警触发条件
type Trigger struct {
	// Source 定义触发源类型
	// 支持: falco, tetragon, manual, webhook
	Source string `json:"source"`

	// Conditions 使用 LogSQL 表达式定义触发条件
	// 示例: "severity = critical" 或 "process.exec = /bin/bash"
	Conditions []string `json:"conditions"`

	// 可选：严重程度过滤
	Severity string `json:"severity,omitempty"` // critical, high, medium, low
}

// Action 定义响应动作
type Action struct {
	// Name 动作名称
	Name string `json:"name"`

	// Type 动作类型
	// 支持:
	//   - webhook: HTTP 请求
	//   - email: 发送邮件
	//   - kubernetes: K8s 操作 (evict, delete, patch)
	//   - forensics: 取证采集 (memory, filesystem)
	//   - expression: 表达式计算
	//   - condition: 条件分支
	Type string `json:"type"`

	// Config 动作配置
	Config ActionConfig `json:"config"`

	// 可选：条件分支配置 (仅当 Type = condition 时有效)
	Branchs *ActionBranchs `json:"branchs,omitempty"`
}

// ActionConfig 动作配置
type ActionConfig struct {
	// Webhook 配置
	URL         string            `json:"url,omitempty"`
	Method      string            `json:"method,omitempty"` // GET, POST, PUT, DELETE
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	Template    string            `json:"template,omitempty"` // 模板名称

	// Email 配置
	SMTPHost     string `json:"smtpHost,omitempty"`
	SMTPPort     int    `json:"smtpPort,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	To           string `json:"to,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Content      string `json:"content,omitempty"`

	// Kubernetes 配置
	K8sAction   string      `json:"action,omitempty"` // evict, delete, patch
	Kind         string      `json:"kind,omitempty"`  // Pod, Service, NetworkPolicy
	Namespace    string      `json:"namespace,omitempty"`
	Selector     string      `json:"selector,omitempty"` // 支持模板变量 {{ .incident.pod }}
	Patch        interface{} `json:"patch,omitempty"`

	// Forensics 配置
	Capture   string `json:"capture,omitempty"`   // memory, filesystem, all
	Timeout   string `json:"timeout,omitempty"`   // 30s, 1m
	Storage   string `json:"storage,omitempty"`   // s3://bucket/path

	// Expression 配置
	Expression string `json:"expression,omitempty"`
}

// ActionBranchs 条件分支配置
type ActionBranchs struct {
	// TrueBranch 当条件为 true 时执行的子动作
	TrueBranch []Action `json:"trueBranch,omitempty"`

	// FalseBranch 当条件为 false 时执行的子动作
	FalseBranch []Action `json:"falseBranch,omitempty"`
}
