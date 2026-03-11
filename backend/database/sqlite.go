package database

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

	"github.com/laenix/vsentry/model"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	dbPath := viper.GetString("database.path")
	if dbPath == "" {
		dbPath = "vsentry.db"
	}
	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	//自动建立表
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.UserLoginLogs{})
	db.AutoMigrate(&model.UserActionLogs{})
	db.AutoMigrate(&model.Ingest{})
	db.AutoMigrate(&model.IngestAuth{})
	db.AutoMigrate(&model.CustomTable{})
	db.AutoMigrate(&model.Connector{})
	db.AutoMigrate(&model.CollectorConfig{})
	db.AutoMigrate(&model.Rule{})
	db.AutoMigrate(&model.Alert{})
	db.AutoMigrate(&model.Incident{})
	db.AutoMigrate(&model.InvestigationTemplate{})
	db.AutoMigrate(&model.ForensicTask{})
	db.AutoMigrate(&model.ForensicFile{})
	db.AutoMigrate(&model.Playbook{})
	db.AutoMigrate(&model.PlaybookExecution{})

	DB = db
	createAdminIfNotExist(db)
	createDefaultIngest(db)
	creatrDefaultInvestigationTemplates(db)
	createDefaultRules(db)
	return db
}

func GetDB() *gorm.DB {
	return DB
}

func createAdminIfNotExist(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count == 0 {
		log.Println("No users found, creating default admin...")

		// 使用 bcrypt 加密预设密码
		// 建议密码：admin123 (实际生产中请务必第一次登录后修改)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

		admin := model.User{
			UserName: "admin",
			Password: string(hashedPassword),
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Failed to create default admin: %v", err)
		} else {
			log.Println("Default admin 'admin' created with password 'admin123'")
		}
	}
}

func createDefaultIngest(db *gorm.DB) {
	var count int64
	db.Model(&model.Ingest{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("No ingest found, creating default local VictoriaLogs...")

	// 生成随机 token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	// 获取外部访问地址（用于客户端接入）
	// 优先级：环境变量 EXTERNAL_URL > config.yaml > 默认值
	externalURL := os.Getenv("EXTERNAL_URL")
	if externalURL == "" {
		externalURL = viper.GetString("server.external_url")
	}
	if externalURL == "" {
		externalURL = "http://localhost:8088"
	}

	log.Printf("Using external URL for ingest endpoint: %s", externalURL)

	// 创建默认 Ingest - 使用外部URL（通过后端转发）
	ingest := model.Ingest{
		Name:         "VictoriaLogs Ingest",
		Endpoint:     externalURL + "/api/ingest/collect",
		Type:         "victorialogs",
		Source:       "build-in",
		StreamFields: "_stream_fields=channel,source,host",
	}

	if err := db.Create(&ingest).Error; err != nil {
		log.Printf("Failed to create default ingest: %v", err)
		return
	}

	// 创建对应的 Auth
	auth := model.IngestAuth{
		IngestID:  ingest.ID,
		SecretKey: token,
	}

	if err := db.Create(&auth).Error; err != nil {
		log.Printf("Failed to create ingest auth: %v", err)
	} else {
		log.Printf("Default ingest created with token: %s", token[:8]+"...")
	}
}

func creatrDefaultInvestigationTemplates(db *gorm.DB) {
	var count int64
	db.Model(&model.InvestigationTemplate{}).Count(&count)

	// 如果表里已经有数据了，就不重复插入
	if count > 0 {
		return
	}

	defaultTemplates := []model.InvestigationTemplate{
		{
			Name:        "同主机历史事件",
			Description: "调查特定主机在事件发生前后的所有行为日志，用于时间线还原。",
			// ✅ 注意这里把等号改成了冒号
			LogSQL:     `observer.hostname:"${hostname}" AND _time:[${start_time}, ${end_time}]`,
			Parameters: `["hostname", "start_time", "end_time"]`,
		},
		{
			Name:        "同用户活动轨迹",
			Description: "追踪特定用户（如被盗用的域账号）在全网的登录和操作记录。",
			LogSQL:      `(target_user.name:"${username}" OR actor.user.name:"${username}") AND _time:[${start_time}, ${end_time}]`,
			Parameters:  `["username", "start_time", "end_time"]`,
		},
		{
			Name:        "横向移动检测 (同源IP)",
			Description: "检测同一个攻击源 IP 在短时间内对内网其他主机的扫描或登录尝试。",
			LogSQL:      `src_endpoint.ip:"${src_ip}" AND _time:[${start_time}, ${end_time}]`,
			Parameters:  `["src_ip", "start_time", "end_time"]`,
		},
		{
			Name:        "进程链与子进程回溯",
			Description: "输入可疑进程名，查询是谁启动了它（父进程），以及它又启动了什么子进程。",
			LogSQL:      `(process.name:"${process_name}" OR process.parent.name:"${process_name}") AND observer.hostname:"${hostname}"`,
			Parameters:  `["process_name", "hostname"]`,
		},
		{
			Name:        "暴力破解历史溯源",
			Description: "统计该攻击源 IP 在过去一段时间内的所有认证失败记录。",
			LogSQL:      `src_endpoint.ip:"${src_ip}" AND activity_name:"Logon Failed" AND _time:[${start_time}, ${end_time}]`,
			Parameters:  `["src_ip", "start_time", "end_time"]`,
		},
	}

	db.Create(&defaultTemplates)
}

func createDefaultRules(db *gorm.DB) {
	var count int64
	db.Model(&model.Rule{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("Creating default forensic and investigation rules...")

	defaultRules := []model.Rule{
		// Forensic rules (type = forensic)
		{
			Name:        "Suspicious PowerShell Execution",
			Description: "Detect suspicious PowerShell commands often used in attacks",
			Query:       `process.name:"powershell.exe" AND (process.cmd_line contains "-enc" OR process.cmd_line contains "DownloadString" OR process.cmd_line contains "Invoke-Expression")`,
			Type:        "forensic",
			Severity:    "high",
			Enabled:     true,
			Version:     1,
		},
		{
			Name:        "Malicious DNS Queries",
			Description: "Detect DNS queries to known malicious domains",
			Query:       `dns.query_name contains ".xyz" OR dns.query_name contains ".top" OR dns.query_name contains ".club"`,
			Type:        "forensic",
			Severity:    "medium",
			Enabled:     true,
			Version:     1,
		},
		{
			Name:        "Failed Login Attempts",
			Description: "Detect multiple failed login attempts indicating brute force",
			Query:       `activity_name:"Logon Failed"`,
			Type:        "forensic",
			Severity:    "medium",
			Enabled:     true,
			Version:     1,
		},
		{
			Name:        "Privilege Escalation Detection",
			Description: "Detect potential privilege escalation via new service creation",
			Query:       `activity_name:"Service Created" AND (process.name contains "sc.exe" OR process.name contains "New-Service")`,
			Type:        "forensic",
			Severity:    "high",
			Enabled:     true,
			Version:     1,
		},
		{
			Name:        "WebLogic Detection",
			Description: "Detect WebLogic traffic on port 7001",
			Query:       `protocol:"HTTP" | dst_port:"7001" or src_port:"7001"`,
			Type:        "forensic",
			Severity:    "medium",
			Enabled:     true,
			Version:     1,
		},
		// Investigation rules (type = investigation)
		{
			Name:        "Host Timeline Investigation",
			Description: "Query all events for a specific host within time range",
			Query:       `observer.hostname:"${hostname}" AND _time:[${start_time}, ${end_time}]`,
			Type:        "investigation",
			Severity:    "low",
			Enabled:     true,
			Version:     1,
		},
		{
			Name:        "User Activity Investigation",
			Description: "Query all activities for a specific user",
			Query:       `(target_user.name:"${username}" OR actor.user.name:"${username}") AND _time:[${start_time}, ${end_time}]`,
			Type:        "investigation",
			Severity:    "low",
			Enabled:     true,
			Version:     1,
		},
		{
			Name:        "IP Reputation Check",
			Description: "Check all events associated with a specific IP",
			Query:       `(src_endpoint.ip:"${ip}" OR dst_endpoint.ip:"${ip}") AND _time:[${start_time}, ${end_time}]`,
			Type:        "investigation",
			Severity:    "low",
			Enabled:     true,
			Version:     1,
		},
	}

	if err := db.Create(&defaultRules).Error; err != nil {
		log.Printf("Failed to create default rules: %v", err)
	} else {
		log.Printf("Created %d default rules", len(defaultRules))
	}
}
