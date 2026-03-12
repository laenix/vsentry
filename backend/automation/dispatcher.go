package automation

import (
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// DispatchByIncident - scheduler/executor.go 调用
// 当有New的 - 或 Count 增加时触发
func DispatchByIncident(incident model.Incident) {
	db := database.GetDB()
	var playbooks []model.Playbook

	//   核心逻辑：通过 JOIN Medium间表 rule_playbooks 筛选受关联的Playbook
	db.Joins("JOIN rule_playbooks ON rule_playbooks.playbook_id = playbooks.id").
		Where("rule_playbooks.rule_id = ? AND playbooks.is_active = ? AND playbooks.trigger_type = ?",
			incident.RuleID, true, "incident").
		Find(&playbooks)

	engine := NewEngine()
	// 将Event对象注入 - 上下文，供 expr EngineParseVariable
	inputContext := map[string]interface{}{
		"incident": incident,
	}

	for _, pb := range playbooks {
		//   AsyncExecute，不BlockAlert入库主流程
		go engine.Run(pb.ID, inputContext)
	}
}

// DispatchManual - func DispatchManual(playbookID uint, mockContext map[string]interface{}) (uint, error) {
	engine := NewEngine()
	//   手动触发时，Global Environment直接透传 mockContext
	return engine.Run(playbookID, mockContext)
}
