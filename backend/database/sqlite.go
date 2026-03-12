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
	// ConnectionDatabase - , err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	// 自动建立表 - .AutoMigrate(&model.User{})
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
	db.AutoMigrate(&model.ForensicTask{})
	db.AutoMigrate(&model.ForensicFile{})
	db.AutoMigrate(&model.Playbook{})
	db.AutoMigrate(&model.PlaybookExecution{})

	DB = db
	createAdminIfNotExist(db)
	createDefaultIngest(db)
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

		// 使用 - 加密预设Password
		//   建议Password：admin123 (实际ProductionMedium请务必第一次Login后修改)
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

	// 生成随机 - tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	//   Get外部访问Address（用于Client接入）
	//   优先级：EnvironmentVariable EXTERNAL_URL > config.yaml > 默认值
	externalURL := os.Getenv("EXTERNAL_URL")
	if externalURL == "" {
		externalURL = viper.GetString("server.external_url")
	}
	if externalURL == "" {
		externalURL = "http://  localhost:8088"
	}

	log.Printf("Using external URL for ingest endpoint: %s", externalURL)

	// Create默认 - - 使用外部URL（通过后端Transfer）
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

	// Create对应的 - auth := model.IngestAuth{
		IngestID:  ingest.ID,
		SecretKey: token,
	}

	if err := db.Create(&auth).Error; err != nil {
		log.Printf("Failed to create ingest auth: %v", err)
	} else {
		log.Printf("Default ingest created with token: %s", token[:8]+"...")
	}
}

func createDefaultRules(db *gorm.DB) {
	var count int64
	db.Model(&model.Rule{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("Creating default forensic and investigation rules...")

	defaultRules := []model.Rule{
		// Forensic - (type = forensic)
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
		// Investigation - (type = investigation)
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
