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
	// 1. 初始化基础配置
	config.InitConfig()

	// 2. 初始化持久化存储 (SQLite)
	database.InitDB()

	// 3. 初始化本地高速缓存 (BadgerDB)
	database.InitBadger()

	// 4. 启动异步日志分发调度器 (消费者)
	// 该协程负责根据 IngestID 分发日志并管理 VictoriaLogs 实例的生命周期
	go ingest.StartDispatcher()

	scheduler.InitScheduler()            // 启动引擎
	scheduler.GlobalEngine.ReloadRules() // 首次加载规则
	// 5. 设置 Gin 引擎
	r := gin.Default()
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

	// 7. 在协程中启动 Web 服务，避免阻塞主线程
	go func() {
		log.Printf("VSentry Server is running on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 8. 监听系统中断信号以实现优雅关机
	// SIGINT: Ctrl+C, SIGTERM: 容器或系统停止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-quit
	log.Println("Shutting down VSentry server...")

	// 9. 设置关机超时时间 (例如 5 秒)，确保未完成的任务有时间处理
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// A. 首先停止接收新的 HTTP 请求
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// B. 停止所有活跃的 Ingest Workers
	// 这会触发每个实例的 Final Flush，确保缓冲区日志全部发出
	ingest.StopAllWorkers()

	// C. 关闭数据库和缓存连接
	if database.Cache != nil {
		log.Println("Closing BadgerDB...")
		database.Cache.Close()
	}

	log.Println("Server exiting gracefully")
}
