package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/laenix/vsentry/scheduler"
	"gorm.io/gorm"
)

// ListRules 获取规则列表
func ListRules(ctx *gin.Context) {
	db := database.GetDB()
	var rules []model.Rule

	// 建议增加预加载(Preload)处理关联标签，如果有的话
	if err := db.Find(&rules).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询规则失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"rules": rules},
		"msg":  "success",
	})
}

// AddRule 添加新规则
func AddRule(ctx *gin.Context) {
	var rule model.Rule
	if err := ctx.ShouldBindJSON(&rule); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数格式错误"})
		return
	}

	// 核心验证：规则名称和查询语句不能为空
	if rule.Name == "" || rule.Query == "" {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "规则名称和查询语句不能为空"})
		return
	}

	// 自动设置初始元数据
	rule.Version = 1
	rule.Enabled = true // 默认开启

	// 从中间件获取当前操作人 ID
	userId, exists := ctx.Get("userid")
	if exists {
		rule.AuthorID = userId.(uint)
	}

	db := database.GetDB()
	if err := db.Create(&rule).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建规则失败"})
		return
	}
	scheduler.GlobalEngine.ReloadRules()
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "规则添加成功", "data": rule})
}

// UpdateRule 更新现有规则
func UpdateRule(ctx *gin.Context) {
	var req model.Rule
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数解析失败"})
		return
	}

	if req.ID == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "规则 ID 不能为空"})
		return
	}

	db := database.GetDB()
	var existing model.Rule

	// 开启事务处理：版本自增和属性更新
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&existing, req.ID).Error; err != nil {
			return err
		}

		// 增加版本号
		req.Version = existing.Version + 1

		// 记录修改人
		if userId, exists := ctx.Get("userid"); exists {
			req.AuthorID = userId.(uint)
		}

		// 使用 Select 指定允许更新的字段，防止恶意覆盖元数据
		return tx.Model(&existing).Select("Name", "Description", "Query", "Interval", "Severity", "Version", "AuthorID").Updates(req).Error
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该规则"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		}
		return
	}
	scheduler.GlobalEngine.ReloadRules()
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "规则更新成功"})
}

// DeleteRule 删除规则
func DeleteRule(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少 ID 参数"})
		return
	}

	db := database.GetDB()
	// 硬删除规则，如果是生产环境建议在 model 中加入 gorm.DeletedAt 实现软删除
	if err := db.Delete(&model.Rule{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}
	scheduler.GlobalEngine.ReloadRules()
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "规则删除成功"})
}

// SetRuleStatus 统一处理启用/禁用逻辑
func SetRuleStatus(ctx *gin.Context, enabled bool) {
	type StatusRequest struct {
		ID uint `json:"id"`
	}
	var req StatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少 ID 参数"})
		return
	}

	db := database.GetDB()
	result := db.Model(&model.Rule{}).Where("id = ?", req.ID).Update("enabled", enabled)

	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "状态更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "规则不存在"})
		return
	}

	statusMsg := "启用"
	if !enabled {
		statusMsg = "禁用"
	}
	scheduler.GlobalEngine.ReloadRules()
	ctx.JSON(http.StatusOK, gin.H{"code": 200, "msg": "规则" + statusMsg + "成功"})
}

func EnableRule(ctx *gin.Context) {
	SetRuleStatus(ctx, true)
}

func DisableRule(ctx *gin.Context) {
	SetRuleStatus(ctx, false)
}
