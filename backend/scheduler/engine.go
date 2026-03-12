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
		// 开启 WithSeconds 以支持更细粒度的检测（如every 10 seconds检测一次暴力破解）
		scheduler: cron.New(cron.WithSeconds()),
		entryIDs:  make(map[uint]cron.EntryID),
	}
	GlobalEngine.scheduler.Start()
	log.Println("Scheduler Engine initialized with Cron format support")
}

func (e *CronEngine) ReloadRules() {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. 清理旧Task：防止 Reload 时Task堆积
	for _, entryID := range e.entryIDs {
		e.scheduler.Remove(entryID)
	}
	e.entryIDs = make(map[uint]cron.EntryID)

	var rules []model.Rule
	db := database.GetDB()
	// 只加载Enable的Rule
	if err := db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		log.Printf("Scheduler load error: %v", err)
		return
	}

	// 2. RegisterTask
	for _, r := range rules {
		rule := r

		// SkipForensicsRule（ForensicsRule由EvidenceUpload触发）
		if rule.Type == "forensic" {
			continue
		}

		// SkipInvestigationRule（InvestigationRule由User手动触发）
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

// TriggerBacktrace 触发Rule回溯
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

	// Parse回溯开始Time
	startTime := parseBacktraceStart(rule.BacktraceStart)
	now := time.Now()

	// 逐days回溯：模拟过去的every一daysExecute一次
	for d := startTime; d.Before(now); d = d.AddDate(0, 0, 1) {
		dayStart := d
		dayEnd := d.AddDate(0, 0, 1)

		log.Printf("[Backtrace] Simulating: %s -> %s", dayStart.Format("2006-01-02"), dayEnd.Format("2006-01-02"))

		// 构建带Time范围的Query
		// 注意：这里Need根据RuleQuery的特性AddTimeFilter
		query := buildQueryWithTimeRange(rule.Query, dayStart, dayEnd)
		ExecuteRuleWithQuery(rule, query)
	}

	log.Printf("[Backtrace] Completed for rule: %s", rule.Name)
}

// parseBacktraceStart Parse回溯开始Time
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
		// 尝试ParseDate格式
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			return t
		}
		// Default回溯1year
		return now.AddDate(-1, 0, 0)
	}
}

// buildQueryWithTimeRange 构建带Time范围的Query
func buildQueryWithTimeRange(query string, start, end time.Time) string {
	startStr := start.Format("2006-01-02") + "T00:00:00Z"
	endStr := end.Format("2006-01-02") + "T00:00:00Z"
	
	// 在QueryMediumAddTimeFilter（如果QueryMedium没有TimeFilter的话）
	// 这里简单Handle，实际可能Need根据Query内容智能Add
	return fmt.Sprintf("_time:[%s TO %s] %s", startStr, endStr, query)
}

func (e *CronEngine) Stop() {
	if e.scheduler != nil {
		e.scheduler.Stop()
	}
}
