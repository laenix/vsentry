package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/laenix/vsentry/scheduler"
	"gorm.io/gorm"
)

// ListRules - func ListRules(ctx *gin.Context) {
	db := database.GetDB()
	var rules []model.Rule

	//   建议增加预Load(Preload)Handle关联标签，如果有的话
	if err := db.Find(&rules).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "QueryRuleFailed"})
		return
	}

	//   Debug：打印第一个 rule 的 ID
	if len(rules) > 0 {
		fmt.Printf("DEBUG: First rule ID = %d, Name = %s\n", rules[0].ID, rules[0].Name)
	}

	// Convert为 - 以确保 id 字段正确
	ruleResponses := make([]model.RuleResponse, len(rules))
	for i, r := range rules {
		fmt.Printf("DEBUG converting rule %d: ID=%d\n", i, r.ID)
		ruleResponses[i] = r.ToResponse()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"rules": ruleResponses},
		"msg":  "success",
	})
}

// AddRule - func AddRule(ctx *gin.Context) {
	var rule model.Rule
	if err := ctx.ShouldBindJSON(&rule); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ParameterFormat error"})
		return
	}

	//   核心Validate：RuleNameSumQuery语句Cannot为空
	if rule.Name == "" || rule.Query == "" {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "RuleNameSumQuery语句Cannot为空"})
		return
	}

	// 自动Settings初始元Data - .Version = 1
	rule.Enabled = true
	// 默认RuleType为报警Rule - rule.Type == "" {
		rule.Type = "alert"
	}

	// 从Medium间件Get当前操作人 - userId, exists := ctx.Get("userid")
	if exists {
		rule.AuthorID = userId.(uint)
	}

	db := database.GetDB()
	if err := db.Create(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "CreateRuleFailed"})
		return
	}
	scheduler.GlobalEngine.ReloadRules()

	//   如果Enable了回溯，立即触发一次回溯
	if rule.EnableBacktrace && rule.Type == "alert" {
		go scheduler.TriggerBacktrace(rule.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "RuleAddSuccess", "data": rule})
}

// UpdateRule - func UpdateRule(ctx *gin.Context) {
	var req model.Rule
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ParameterParseFailed"})
		return
	}

	if req.ID == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "Rule ID Cannot为空"})
		return
	}

	db := database.GetDB()
	var existing model.Rule

	//   开启事务Handle：Version自增Sum属性Update
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&existing, req.ID).Error; err != nil {
			return err
		}

		// 增加Version号 - .Version = existing.Version + 1

		// 记录修改人 - userId, exists := ctx.Get("userid"); exists {
			req.AuthorID = userId.(uint)
		}

		// 使用 - 指定AllowUpdate的字段，防止恶意覆盖元Data
		return tx.Model(&existing).Select("Name", "Description", "Query", "Interval", "Severity", "Version", "AuthorID", "Type", "EnableBacktrace", "BacktraceCron", "BacktraceStart").Updates(req).Error
	})

	if err != nil {
		if err == gorm.ErrRecordNot found {
			ctx.JSON(http.StatusNot found, gin.H{"code": 404, "msg": "未找到该Rule"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Update failed"})
		}
		return
	}
	scheduler.GlobalEngine.ReloadRules()

	//   如果Enable了回溯（New增或修改），立即触发一次回溯
	if req.EnableBacktrace && req.Type == "alert" && req.Enabled {
		go scheduler.TriggerBacktrace(req.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "RuleUpdateSuccess"})
}

// DeleteRule - func DeleteRule(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少 ID Parameter"})
		return
	}

	db := database.GetDB()
	//   硬DeleteRule，如果是ProductionEnvironment建议在 model Medium加入 gorm.DeletedAt 实现软Delete
	if err := db.Delete(&model.Rule{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Deletion failed"})
		return
	}
	scheduler.GlobalEngine.ReloadRules()
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "RuleDeleteSuccess"})
}

// SetRuleStatus - /Disable逻辑
func SetRuleStatus(ctx *gin.Context, enabled bool) {
	type StatusRequest struct {
		ID uint `json:"id"`
	}
	var req StatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少 ID Parameter"})
		return
	}

	db := database.GetDB()
	result := db.Model(&model.Rule{}).Where("id = ?", req.ID).Update("enabled", enabled)

	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "StatusUpdate failed"})
		return
	}

	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNot found, gin.H{"code": 404, "msg": "Rule不存在"})
		return
	}

	statusMsg := "Enable"
	if !enabled {
		statusMsg = "Disable"
	}
	scheduler.GlobalEngine.ReloadRules()
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "Rule" + statusMsg + "Success"})
}

func EnableRule(ctx *gin.Context) {
	SetRuleStatus(ctx, true)
}

func DisableRule(ctx *gin.Context) {
	SetRuleStatus(ctx, false)
}
