package automation

import (
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// DispatchByIncident 由 scheduler/executor.go 调用
// 当有新的 Incident 或 Count 增加时触发
func DispatchByIncident(incident model.Incident) {
	db := database.GetDB()
	var playbooks []model.Playbook

	// 核心逻辑：通过 JOIN 中间表 rule_playbooks 筛选受关联的剧本
	db.Joins("JOIN rule_playbooks ON rule_playbooks.playbook_id = playbooks.id").
		Where("rule_playbooks.rule_id = ? AND playbooks.is_active = ? AND playbooks.trigger_type = ?",
			incident.RuleID, true, "incident").
		Find(&playbooks)

	engine := NewEngine()
	// 将事件对象注入 Global 上下文，供 expr 引擎解析变量
	inputContext := map[string]interface{}{
		"incident": incident,
	}

	for _, pb := range playbooks {
		// 异步执行，不阻塞告警入库主流程
		go engine.Run(pb.ID, inputContext)
	}
}

// DispatchManual 手动触发调用入口
func DispatchManual(playbookID uint, mockContext map[string]interface{}) (uint, error) {
	engine := NewEngine()
	// 手动触发时，Global 环境直接透传 mockContext
	return engine.Run(playbookID, mockContext)
}
