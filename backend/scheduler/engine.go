package scheduler

import (
	"log"
	"sync"

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

func (e *CronEngine) Stop() {
	if e.scheduler != nil {
		e.scheduler.Stop()
	}
}
