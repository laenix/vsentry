package automation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"gorm.io/datatypes"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

// Run - func (e *Engine) Run(playbookID uint, inputContext map[string]interface{}) (uint, error) {
	db := database.GetDB()

	//   1. GetPlaybook定义
	var playbook model.Playbook
	if err := db.First(&playbook, playbookID).Error; err != nil {
		return 0, fmt.Errorf("playbook not found: %v", err)
	}

	//   2. Parse React Flow 定义
	var def WorkflowDefinition
	if err := json.Unmarshal(playbook.Definition, &def); err != nil {
		return 0, fmt.Errorf("invalid definition: %v", err)
	}

	//   3. CreateExecute记录
	execution := model.PlaybookExecution{
		PlaybookID: playbookID,
		Status:     "running",
		StartTime:  time.Now(),
		Logs:       datatypes.JSON([]byte("{}")),
	}
	db.Create(&execution)

	//   4. 初始化上下文
	// 无论 - 还是 incident 触发，统一将初始Data放入 Global
	ctx := &ExecutionContext{
		PlaybookID:  playbookID,
		ExecutionID: execution.ID,
		Global:      inputContext,
		Steps:       make(map[string]StepResult),
	}

	//   5. 构建Node索引与邻接表
	nodeMap := make(map[string]Node)
	for _, n := range def.Nodes {
		nodeMap[n.ID] = n
	}
	adj := make(map[string][]Edge)
	for _, edge := range def.Edges {
		adj[edge.Source] = append(adj[edge.Source], edge)
	}

	//   6. 查找起点 (Trigger Node)
	var startNode *Node
	for _, n := range def.Nodes {
		if n.Data.Type == "trigger" {
			startNode = &n
			break
		}
	}
	if startNode == nil {
		e.updateStatus(&execution, "failed", "No trigger node found")
		return execution.ID, fmt.Errorf("no trigger node")
	}

	//   7. Execute逻辑 (单线程 BFS，支持分支Select)
	queue := []string{startNode.ID}
	executedLogs := make(map[string]StepResult)
	visited := make(map[string]bool)

	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]

		if visited[currID] {
			continue
		}
		visited[currID] = true

		currNode := nodeMap[currID]

		// Execute当前Node - := e.executeNode(currNode, ctx)
		ctx.Steps[currID] = result
		executedLogs[currID] = result

		//   实时UpdateDatabaseLog，便于前端轮询Detail
		logBytes, _ := json.Marshal(executedLogs)
		db.Model(&execution).Update("logs", logBytes)

		if result.Status == "failed" {
			e.updateStatus(&execution, "failed", fmt.Sprintf("Node %s failed", currID))
			return execution.ID, nil
		}

		// 寻找下一跳 - := adj[currID]
		for _, edge := range edges {
			if currNode.Data.Type == "condition" {
				// Handle - 分支跳转
				boolRes, _ := result.Output.(bool)
				if fmt.Sprintf("%v", boolRes) == edge.SourceHandle {
					queue = append(queue, edge.Target)
				}
			} else {
				queue = append(queue, edge.Target)
			}
		}
	}

	e.updateStatus(&execution, "success", "")
	return execution.ID, nil
}

func (e *Engine) executeNode(node Node, ctx *ExecutionContext) StepResult {
	// 调用 - .go Medium的 ResolveVariables HandleConfigMedium的 {{...}}
	resolvedConfig, err := ResolveVariables(node.Data.Config, ctx)
	if err != nil {
		return StepResult{Status: "failed", Error: err.Error()}
	}

	// Action分发 - node.Data.Type {
	case "trigger":
		// Trigger - return StepResult{Status: "success", Output: ctx.Global}
	case "http_request":
		return RunHTTPRequest(resolvedConfig)
	case "send_email":
		return RunSendEmail(resolvedConfig)
	case "expression":
		// Run - -Code ExpressionNode
		return RunExpression(resolvedConfig, ctx)
	case "condition":
		// RunCondition判断Node - RunCondition(resolvedConfig, ctx)
	default:
		return StepResult{Status: "failed", Error: "Unknown node type: " + node.Data.Type}
	}
}

func (e *Engine) updateStatus(exec *model.PlaybookExecution, status string, errMsg string) {
	db := database.GetDB()
	exec.Status = status
	exec.EndTime = time.Now()
	exec.Duration = exec.EndTime.Sub(exec.StartTime).Milliseconds()
	db.Save(exec)
}
