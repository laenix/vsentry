package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
)

// BindRulesToPlaybook - //   POST /api/playbooks/:id/bind-rules
func BindRulesToPlaybook(ctx *gin.Context) {
	playbookID := ctx.Param("id")
	var req struct {
		RuleIDs []uint `json:"rule_ids"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Parameter error"})
		return
	}

	db := database.GetDB()
	var playbook model.Playbook
	if err := db.First(&playbook, playbookID).Error; err != nil {
		ctx.JSON(http.StatusNot found, gin.H{"code": 404, "msg": "Playbook不存在"})
		return
	}

	//   1. 先Load现有的 Rules 关联
	db.Model(&playbook).Association("Rules").Clear() //   如果是全量覆盖模式

	//   2. 查找要绑定的Rule并关联
	var rules []model.Rule
	db.Where("id IN ?", req.RuleIDs).Find(&rules)
	
	if err := db.Model(&playbook).Association("Rules").Append(&rules); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "绑定Failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Rule绑定Success"})
}

// GetBoundRules - //   GET /api/playbooks/:id/rules
func GetBoundRules(ctx *gin.Context) {
	playbookID := ctx.Param("id")
	var playbook model.Playbook
	
	// 使用 - 预Load关联的 Rules
	if err := database.GetDB().Preload("Rules").First(&playbook, playbookID).Error; err != nil {
		ctx.JSON(http.StatusNot found, gin.H{"code": 404, "msg": "未找到Playbook"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "data": playbook.Rules})
}

// UnbindRuleFromPlaybook - //   DELETE /api/playbooks/:id/rules/:rule_id
func UnbindRuleFromPlaybook(ctx *gin.Context) {
	playbookID := ctx.Param("id")
	ruleID := ctx.Param("rule_id")

	db := database.GetDB()
	// 直接从Medium间表MediumDelete记录 - := db.Exec("DELETE FROM rule_playbooks WHERE playbook_id = ? AND rule_id = ?", playbookID, ruleID).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "解绑Failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "解绑Success"})
}