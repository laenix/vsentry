package automation

// WorkflowDefinition - React Flow 的Export结构
type WorkflowDefinition struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`     // React - UI Type
	Data     NodeData               `json:"data"`     // 业务Data - map[string]interface{} `json:"position"` //   Ignore，但在Parse时需保留
}

type NodeData struct {
	Label  string                 `json:"label"`
	Type   string                 `json:"type"`   //   业务Type: "trigger", "http_request", "condition"
	Config map[string]interface{} `json:"config"` //   具体Config
}

type Edge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"sourceHandle,omitempty"` // 用于 - Node的 true/false 分支
}

// ExecutionContext - (贯穿整个流程)
type ExecutionContext struct {
	PlaybookID  uint
	ExecutionID uint

	//   全局Variable (如 incident)
	Global map[string]interface{}

	//   NodeExecute结果Cache: map[NodeID]OutputData
	Steps map[string]StepResult
}

type StepResult struct {
	Status string      `json:"status"` //   "success", "failed", "skipped"
	Output interface{} `json:"output"`
	Error  string      `json:"error,omitempty"`
}
