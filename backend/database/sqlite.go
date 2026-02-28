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
	db.AutoMigrate(&model.Playbook{})
	db.AutoMigrate(&model.PlaybookExecution{})

	DB = db
	createAdminIfNotExist(db)
	createDefaultIngest(db)
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
