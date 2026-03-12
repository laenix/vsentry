package automation

// WorkflowDefinition 对应ago端 React Flow 的Export结构
type WorkflowDefinition struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`     // React Flow UI Type
	Data     NodeData               `json:"data"`     // 业务Data
	Position map[string]interface{} `json:"position"` // Ignore，但在Parse时需保留
}

type NodeData struct {
	Label  string                 `json:"label"`
	Type   string                 `json:"type"`   // 业务Type: "trigger", "http_request", "condition"
	Config map[string]interface{} `json:"config"` // 具体配置
}

type Edge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"sourceHandle,omitempty"` // 用于 Condition 节点的 true/false 分支
}

// ExecutionContext Execute上下文 (贯穿整个流程)
type ExecutionContext struct {
	PlaybookID  uint
	ExecutionID uint

	// 全局Variable (如 incident)
	Global map[string]interface{}

	// 节点Execute结果缓存: map[NodeID]OutputData
	Steps map[string]StepResult
}

type StepResult struct {
	Status string      `json:"status"` // "success", "failed", "skipped"
	Output interface{} `json:"output"`
	Error  string      `json:"error,omitempty"`
}
