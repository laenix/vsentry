package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/model"
	"github.com/robfig/cron/v3"
)

type CronEngine struct {
	scheduler *cron.Cron
	entryIDs  map[uint]cron.EntryID
	mu        sync.Mutex
}

var GlobalEngine *CronEngine

func InitScheduler() {
	GlobalEngine = &CronEngine{
		// 开启 WithSeconds 以支持更细粒度的检测（如每 10 秒检测一次暴力破解）
		scheduler: cron.New(cron.WithSeconds()),
		entryIDs:  make(map[uint]cron.EntryID),
	}
	GlobalEngine.scheduler.Start()
	log.Println("Scheduler Engine initialized with Cron format support")
}

func (e *CronEngine) ReloadRules() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. 清理旧任务：防止 Reload 时任务堆积
	for _, entryID := range e.entryIDs {
		e.scheduler.Remove(entryID)
	}
	e.entryIDs = make(map[uint]cron.EntryID)

	var rules []model.Rule
	db := database.GetDB()
	// 只加载启用的规则
	if err := db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		log.Printf("Scheduler load error: %v", err)
		return
	}

	// 2. 注册任务
	for _, r := range rules {
		rule := r

		// 跳过取证规则（取证规则由证据上传触发）
		if rule.Type == "forensic" {
			continue
		}

		// 跳过调查规则（调查规则由用户手动触发）
		if rule.Type == "investigation" {
			continue
		}

		// 此时 rule.Interval 已经是 "@every 5m" 或 "0 */10 * * * *"
		entryID, err := e.scheduler.AddFunc(rule.Interval, func() {
			ExecuteRule(rule)
		})

		if err != nil {
			log.Printf("Failed to schedule rule [%s]: %v", rule.Name, err)
			continue
		}
		e.entryIDs[rule.ID] = entryID
	}
	log.Printf("Scheduler: Successfully reloaded %d rules", len(rules))
}

// TriggerBacktrace 触发规则回溯
func TriggerBacktrace(ruleID uint) {
	db := database.GetDB()
	var rule model.Rule
	if err := db.First(&rule, ruleID).Error; err != nil {
		log.Printf("[Backtrace] Rule not found: %d", ruleID)
		return
	}

	if !rule.EnableBacktrace {
		log.Printf("[Backtrace] Backtrace not enabled for rule: %s", rule.Name)
		return
	}

	log.Printf("[Backtrace] Starting backtrace for rule: %s (start: %s)", rule.Name, rule.BacktraceStart)

	// 解析回溯开始时间
	startTime := parseBacktraceStart(rule.BacktraceStart)
	now := time.Now()

	// 逐天回溯：模拟过去的每一天执行一次
	for d := startTime; d.Before(now); d = d.AddDate(0, 0, 1) {
		dayStart := d
		dayEnd := d.AddDate(0, 0, 1)

		log.Printf("[Backtrace] Simulating: %s -> %s", dayStart.Format("2006-01-02"), dayEnd.Format("2006-01-02"))

		// 构建带时间范围的查询
		// 注意：这里需要根据规则查询的特性添加时间过滤
		query := buildQueryWithTimeRange(rule.Query, dayStart, dayEnd)
		ExecuteRuleWithQuery(rule, query)
	}

	log.Printf("[Backtrace] Completed for rule: %s", rule.Name)
}

// parseBacktraceStart 解析回溯开始时间
func parseBacktraceStart(startStr string) time.Time {
	now := time.Now()
	
	switch startStr {
	case "1y":
		return now.AddDate(-1, 0, 0)
	case "180d":
		return now.AddDate(0, -6, 0)
	case "90d":
		return now.AddDate(0, -3, 0)
	case "30d":
		return now.AddDate(0, 0, -30)
	case "7d":
		return now.AddDate(0, 0, -7)
	default:
		// 尝试解析日期格式
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			return t
		}
		// 默认回溯1年
		return now.AddDate(-1, 0, 0)
	}
}

// buildQueryWithTimeRange 构建带时间范围的查询
func buildQueryWithTimeRange(query string, start, end time.Time) string {
	startStr := start.Format("2006-01-02") + "T00:00:00Z"
	endStr := end.Format("2006-01-02") + "T00:00:00Z"
	
	// 在查询中添加时间过滤（如果查询中没有时间过滤的话）
	// 这里简单处理，实际可能需要根据查询内容智能添加
	return fmt.Sprintf("_time:[%s TO %s] %s", startStr, endStr, query)
}

func (e *CronEngine) Stop() {
	if e.scheduler != nil {
		e.scheduler.Stop()
	}
}
