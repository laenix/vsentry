package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/config"
	"github.com/laenix/vsentry/database"
	"github.com/laenix/vsentry/ingest"
	"github.com/laenix/vsentry/routers"
	"github.com/laenix/vsentry/scheduler"
	"github.com/spf13/viper"
)

func main() {
	// 1. Initialize基础配置
	config.InitConfig()

	// 2. Initialize持久化Storage (SQLite)
	database.InitDB()

	// 3. Initialize本地High速缓存 (BadgerDB)
	database.InitBadger()

	// 4. StartAsyncLog分发Schedule器 (消费者)
	// 该协程负责根据 IngestID 分发Log并Manage VictoriaLogs 实例的生命周期
	go ingest.StartDispatcher()

	scheduler.InitScheduler()            // StartEngine
	scheduler.GlobalEngine.ReloadRules() // 首次加载Rule
	// 5. Settings Gin Engine
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// 支持大FileUpload (100MB)
	r.MaxMultipartMemory = 100 << 20
	r = routers.CollectRouter(r)

	// 6. 配置 HTTP Server 以支持优雅关机
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 7. 在协程MediumStart Web Service，避免阻塞主线程
	go func() {
		log.Printf("VSentry Server is running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 8. 监听SystemMedium断信号以实现优雅关机
	// SIGINT: Ctrl+C, SIGTERM: 容器或SystemStop信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-quit
	log.Println("Shutting down VSentry server...")

	// 9. Settings关机TimeoutTime (例如 5 seconds)，确保未完成的Task有TimeHandle
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// A. 首先StopReceiveNew的 HTTP Request
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// B. Stop所有活跃的 Ingest Workers
	// 这会触发every个实例的 Final Flush，确保缓冲区Log全部发出
	ingest.StopAllWorkers()

	// C. 关闭Data库和缓存Connection
	if database.Cache != nil {
		log.Println("Closing BadgerDB...")
		database.Cache.Close()
	}

	log.Println("Server exiting gracefully")
}
