package routers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/laenix/vsentry/controller"
	"github.com/laenix/vsentry/middleware"
)

func CollectRouter(r *gin.Engine) *gin.Engine {

	// 托管前端静态File - := os.Getenv("STATIC_PATH")
	if staticPath == "" {
		staticPath = "./dist"
	}

	// 尝试Load静态FileDirectory - _, err := os.Stat(staticPath); err == nil {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())

		//   静态FileService - Vite 构建的Application使用 /assets Path
		r.Static("/assets", staticPath+"/assets")
		r.Static("/static", staticPath)

		// API - /api 前缀
		api := r.Group("/api")
		{
			setupAPIRoutes(api)
		}

		// SPA - : 所有非 API 路由都Return index.html
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// Skip - Path
			if path == "/api" || len(path) > 5 && path[:4] == "/api" {
				c.JSON(404, gin.H{"code": 404, "msg": "API not found"})
				return
			}
			c.File(staticPath + "/index.html")
		})

		return r
	}

	//   如果没有静态File，使用原来的 API 路由方式 (无 /api 前缀)
	setupAPIRoutes(r.Group(""))

	return r
}

func setupAPIRoutes(r *gin.RouterGroup) {
	r.GET("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	r.POST("/login", controller.Login)

	// user - := r.Group("/user", middleware.AuthMiddleware())
	{
		user.POST("/userinfo", controller.Userinfo)
		user.POST("/changepassword", controller.UserChangePassword)
	}
	// dashboard - .GET("/dashboard", middleware.AuthMiddleware(), controller.GetDashboard)
	// users - := r.Group("/users", middleware.AuthMiddleware())
	{
		users.GET("/list", controller.ListUser)
		users.POST("/add", controller.AddUser)
		users.POST("/delete", controller.DeleteUser)
	}
	// ingest - := r.Group("/ingest", middleware.IngestMiddleware())
	{
		ingest.POST("/collect", controller.CollectIngest)
	}
	// ingest - ingestManager := r.Group("/ingestmanager", middleware.AuthMiddleware())
	{
		ingestManager.POST("/add", controller.AddIngest)
		ingestManager.GET("/list", controller.ListIngest)
		ingestManager.POST("/update", controller.UpdateIngest)
		ingestManager.POST("/delete", controller.DeleteIngest)
		ingestManager.GET("/auth/:id", controller.GetIngestAuth)
	}

	//   connectors (third-party integrations)
	connectors := r.Group("/connectors", middleware.AuthMiddleware())
	{
		connectors.GET("/list", controller.ListConnectors)
		connectors.GET("/templates", controller.GetConnectorTemplates)
		connectors.POST("/add", controller.AddConnector)
		connectors.POST("/update", controller.UpdateConnector)
		connectors.POST("/delete", controller.DeleteConnector)
		connectors.POST("/test", controller.TestConnector)
	}

	//   collectors (agent builders)
	collectors := r.Group("/collectors", middleware.AuthMiddleware())
	{
		collectors.GET("/list", controller.ListCollectorConfigs)
		collectors.GET("/templates", controller.GetCollectorTemplates)
		collectors.GET("/channels", controller.GetAvailableChannels)
		collectors.POST("/add", controller.AddCollectorConfig)
		collectors.POST("/update", controller.UpdateCollectorConfig)
		collectors.POST("/delete", controller.DeleteCollectorConfig)
		collectors.POST("/build", controller.BuildCollector)
		collectors.GET("/download", controller.DownloadCollector) //   增加这行，通常Download用 GET
	}

	//   config (public)
	r.GET("/config", controller.GetConfig)

	// VictoriaLogs - endpoints - require authentication
	victorialogs := r.Group("", middleware.AuthMiddleware())
	{
		victorialogs.POST("/select/logsql/query", controller.QueryVictoriaLogs)
		victorialogs.POST("/select/logsql/hits", controller.QueryVictoriaLogsHits)
		victorialogs.GET("/metrics", controller.GetVictoriaLogsMetrics)
		victorialogs.GET("/health", controller.GetVictoriaLogsHealth)
	}

	// custom - customTables := r.Group("/customtables", middleware.AuthMiddleware())
	{
		customTables.GET("/list", controller.ListCustomTables)
		customTables.POST("/add", controller.AddCustomTable)
		customTables.POST("/update", controller.UpdateCustomTable)
		customTables.POST("/delete", controller.DeleteCustomTable)
	}

	// rules - := r.Group("/rules", middleware.AuthMiddleware())
	{
		rules.GET("/list", controller.ListRules)
		rules.POST("/add", controller.AddRule)
		rules.POST("/update", controller.UpdateRule)
		rules.POST("/delete", controller.DeleteRule)
		rules.POST("/enable", controller.EnableRule)
		rules.POST("/disable", controller.DisableRule)
	}
	// alerts - := r.Group("/alerts", middleware.AuthMiddleware())
	{
		alerts.GET("/list", controller.ListAlerts)
		alerts.POST("/acknowledge", controller.Acknowledge)
		alerts.POST("/resolve", controller.Resolve)
		alerts.POST("/assign", controller.Assign)
	}

	// incidents - := r.Group("/incidents", middleware.AuthMiddleware())
	{
		incidentGroup.GET("/list", controller.ListIncidents)
		incidentGroup.GET("/detail", controller.GetIncidentDetail)
		incidentGroup.POST("/acknowledge", controller.AcknowledgeIncident)
		incidentGroup.POST("/resolve", controller.ResolveIncident)
	}

	// investigation - := r.Group("/investigation", middleware.AuthMiddleware())
	{
		investigationGroup.POST("/execute", controller.ExecuteInvestigation) //   核心ExecuteEngine
	}
	// forensics - := r.Group("/forensics", middleware.AuthMiddleware())
	{
		forensicsGroup.GET("/tasks", controller.ListForensicTasks)
		forensicsGroup.POST("/tasks", controller.CreateForensicTask)
		forensicsGroup.GET("/tasks/:id", controller.GetForensicTask)
		forensicsGroup.DELETE("/tasks/:id", controller.DeleteForensicTask)
		
		forensicsGroup.POST("/upload", controller.UploadForensicFile)
		forensicsGroup.DELETE("/files/:id", controller.DeleteForensicFile)
		forensicsGroup.POST("/execute-rules", controller.ExecuteForensicRules)
	}
	// automation - := r.Group("/playbooks", middleware.AuthMiddleware())
	{
		automation.GET("", controller.ListPlaybooks)         // List - .POST("", controller.CreatePlaybook)       // Create - .GET("/:id", controller.GetPlaybook)       // Detail - .PUT("/:id", controller.UpdatePlaybook)    // Update - .DELETE("/:id", controller.DeletePlaybook) // Delete - .POST("/:id/bind-rules", controller.BindRulesToPlaybook)
		automation.GET("/:id/rules", controller.GetBoundRules)
		automation.DELETE("/:id/rules/:rule_id", controller.UnbindRuleFromPlaybook)

		automation.POST("/:id/run", controller.RunPlaybook)               // DebugRun - .GET("/:id/executions", controller.GetExecutionHistory) // 历史记录 - .GET("/executions/:exec_id", controller.GetExecutionDetail)
		automation.GET("/executions", controller.ListAllExecutions) //   Get所有Execute记录

	}
}
