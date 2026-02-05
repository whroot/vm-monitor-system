package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/logger"
	"vm-monitoring-system/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Server HTTP服务器
type Server struct {
	router            *gin.Engine
	config            *config.Config
	db                *gorm.DB
	http              *http.Server
	alertEngine       *services.AlertEngine
	vsphereCollector  *services.VSphereCollector
	permissionMiddleware *PermissionMiddleware
	wsHub            *WebSocketHub
}

// NewServer 创建服务器实例
func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	server := &Server{
		router: router,
		config: cfg,
		db:     db,
	}

	// 创建 WebSocket Hub
	server.wsHub = NewWebSocketHub()
	server.wsHub.Start()

	// 创建权限中间件
	server.permissionMiddleware = NewPermissionMiddleware(db)

	// 注册中间件
	server.setupMiddleware()

	// 注册路由
	server.setupRoutes()

	// 创建HTTP服务器
	server.http = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 初始化vSphere采集器
	server.setupVSphereCollector()

	return server
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	// 基础中间件
	s.router.Use(gin.Recovery())
	s.router.Use(CORS())
	s.router.Use(RequestLogger())
	s.router.Use(ErrorHandler())
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	// API v1版本
	v1 := s.router.Group("/api/v1")
	{
		// 认证相关（不需要认证）
		auth := v1.Group("/auth")
		{
			authHandler := NewAuthHandler(s.db, s.config, nil) // TODO: 传入私钥
			auth.GET("/public-key", authHandler.GetPublicKey)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 需要认证的路由
		authorized := v1.Group("")
		authorized.Use(JWTAuth(s.config.JWT.Secret))
		authorized.Use(PermissionCheck())
		{
			// 认证管理
			auth := authorized.Group("/auth")
			{
				authHandler := NewAuthHandler(s.db, s.config)
				auth.POST("/logout", authHandler.Logout)
				auth.GET("/me", authHandler.GetMe)
				auth.PUT("/password", authHandler.ChangePassword)
				auth.GET("/check", authHandler.CheckPermission)
			}

			// VM管理
			vms := authorized.Group("/vms")
			{
				vmHandler := NewVMHandler(s.db)
				vmHandler.SetCollector(s.vsphereCollector)
				vms.GET("", vmHandler.List)
				vms.GET("/:id", vmHandler.Get)
				vms.POST("", vmHandler.Create)
				vms.PUT("/:id", vmHandler.Update)
				vms.DELETE("/:id", vmHandler.Delete)
				vms.POST("/sync", vmHandler.Sync)
				vms.GET("/statistics", vmHandler.Statistics)

				// 分组管理
				groups := vms.Group("/groups")
				{
					groups.GET("", vmHandler.ListGroups)
					groups.POST("", vmHandler.CreateGroup)
					groups.PUT("/:id", vmHandler.UpdateGroup)
					groups.DELETE("/:id", vmHandler.DeleteGroup)
				}

				// 批量操作
				vms.POST("/batch", vmHandler.Batch)
			}

			// 实时监控
			realtime := authorized.Group("/realtime")
			{
				realtimeHandler := NewRealtimeHandler(s.db, s.wsHub)
				realtime.GET("/vms/:id", realtimeHandler.GetVMMetrics)
				realtime.POST("/vms/batch", realtimeHandler.BatchGetMetrics)
				realtime.GET("/groups/:id", realtimeHandler.GetGroupMetrics)
				realtime.GET("/clusters/:id", realtimeHandler.GetClusters)
				realtime.GET("/overview", realtimeHandler.GetOverview)
			}

			// WebSocket实时推送（需要认证）
			v1 := s.router.Group("/ws/v1")
			v1.Use(JWTAuth(s.config.JWT.Secret))
			{
				v1.GET("/realtime", s.wsHub.HandleConnection)
			}

			// 历史数据
			history := authorized.Group("/history")
			{
				historyHandler := NewHistoryHandler(s.db)
				history.POST("/query", historyHandler.Query)
				history.POST("/aggregate", historyHandler.Aggregate)
				history.POST("/trends", historyHandler.Trends)
				history.POST("/anomalies", historyHandler.Anomalies)
				history.POST("/export", historyHandler.Export)
				history.GET("/export/:id", historyHandler.GetExportTask)
				history.GET("/timeline/:vmId", historyHandler.GetTimeline)
			}

			// 告警管理
			alerts := authorized.Group("/alerts")
			{
				alertHandler := NewAlertHandler(s.db)

				// 告警规则
				rules := alerts.Group("/rules")
				{
					rules.GET("", alertHandler.ListRules)
					rules.GET("/:id", alertHandler.GetRule)
					rules.POST("", alertHandler.CreateRule)
					rules.PUT("/:id", alertHandler.UpdateRule)
					rules.DELETE("/:id", alertHandler.DeleteRule)
					rules.PUT("/batch/status", alertHandler.BatchUpdateRuleStatus)
					rules.POST("/import", alertHandler.ImportRules)
					rules.POST("/export", alertHandler.ExportRules)
				}

				// 告警记录
				records := alerts.Group("/records")
				{
					records.GET("", alertHandler.ListRecords)
					records.GET("/:id", alertHandler.GetRecord)
					records.PUT("/:id/acknowledge", alertHandler.Acknowledge)
					records.PUT("/batch/acknowledge", alertHandler.BatchAcknowledge)
					records.PUT("/:id/resolve", alertHandler.Resolve)
					records.PUT("/:id/ignore", alertHandler.Ignore)
				}

				// 统计
				alerts.GET("/statistics", alertHandler.Statistics)
				alerts.GET("/trends", alertHandler.Trends)
			}

			// 用户权限管理
			users := authorized.Group("/users")
			users.Use(s.permissionMiddleware.RequirePermission("users:manage"))
			{
				userHandler := NewUserHandler(s.db)
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.Get)
				users.POST("", userHandler.Create)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
				users.POST("/:id/reset-password", userHandler.ResetPassword)
				users.PUT("/batch/status", userHandler.BatchUpdateStatus)
				users.GET("/me/permissions", userHandler.GetMyPermissions)
				users.GET("/:id/permissions/detail", userHandler.GetUserPermissions)
			}

			// 角色管理
			roles := authorized.Group("/roles")
			roles.Use(s.permissionMiddleware.RequirePermission("roles:manage"))
			{
				roleHandler := NewRoleHandler(s.db)
				roles.GET("", roleHandler.List)
				roles.GET("/:id", roleHandler.Get)
				roles.POST("", roleHandler.Create)
				roles.PUT("/:id", roleHandler.Update)
				roles.DELETE("/:id", roleHandler.Delete)
				roles.GET("/:id/permissions", roleHandler.GetPermissions)
				roles.PUT("/:id/permissions", roleHandler.UpdatePermissions)
				roles.GET("/:id/users", roleHandler.GetUsers)
			}

			// 权限矩阵
			permissions := authorized.Group("/permissions")
			permissions.Use(s.permissionMiddleware.RequirePermission("permissions:manage"))
			{
				permissionHandler := NewPermissionHandler(s.db)
				permissions.GET("/matrix", permissionHandler.GetMatrix)
				permissions.PUT("/matrix", permissionHandler.UpdateMatrix)
				permissions.POST("/check-conflict", permissionHandler.CheckConflict)
				permissions.GET("/audit", permissionHandler.GetAuditLogs)
				permissions.POST("/report", permissionHandler.GenerateReport)
			}

			// 系统健康
			system := authorized.Group("/system")
			{
				systemHandler := NewSystemHandler(s.db, s.config)
				system.GET("/overview", systemHandler.Overview)
				system.GET("/health-score", systemHandler.HealthScore)
				system.GET("/health-trend", systemHandler.HealthTrend)
				system.GET("/services", systemHandler.Services)
				system.GET("/collectors", systemHandler.Collectors)
				system.GET("/storage", systemHandler.Storage)
				system.GET("/performance", systemHandler.Performance)
				system.GET("/capacity", systemHandler.Capacity)
				system.GET("/config", systemHandler.GetConfig)
				system.PUT("/config", systemHandler.UpdateConfig)
				system.GET("/config/history", systemHandler.GetConfigHistory)
				system.GET("/logs", systemHandler.GetLogs)
				system.GET("/audit-logs", systemHandler.GetAuditLogs)
				system.POST("/logs/export", systemHandler.ExportLogs)
				system.POST("/maintenance/cleanup", systemHandler.Cleanup)
				system.GET("/maintenance/tasks", systemHandler.ListTasks)
				system.GET("/maintenance/tasks/:id", systemHandler.GetTask)
			}
		}
	}
}

// setupVSphereCollector 初始化vSphere采集器
func (s *Server) setupVSphereCollector() {
	// 创建vSphere配置
	vsphereConfig := &services.VSphereConfig{
		Host:        s.config.VSphere.Host,
		Port:        s.config.VSphere.Port,
		Username:    s.config.VSphere.Username,
		Password:    s.config.VSphere.Password,
		Insecure:    s.config.VSphere.Insecure,
		CollectInterval: s.config.VSphere.CollectInterval,
		BatchSize:   s.config.VSphere.BatchSize,
	}

	// 创建vSphere采集器
	s.vsphereCollector = services.NewVSphereCollector(s.db, vsphereConfig)

	// 启动vSphere采集器
	if err := s.vsphereCollector.Start(); err != nil {
		logger.Error("vSphere采集器启动失败", zap.Error(err))
	} else {
		logger.Info("vSphere采集器已启动")
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 启动告警引擎
	s.setupAlertEngine()

	return s.http.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	// 停止 WebSocket Hub
	if s.wsHub != nil {
		s.wsHub.Stop()
		logger.Info("WebSocket Hub 已停止")
	}

	// 停止告警引擎
	if s.alertEngine != nil {
		s.alertEngine.Stop()
		logger.Info("告警引擎已停止")
	}

	// 停止vSphere采集器
	if s.vsphereCollector != nil {
		s.vsphereCollector.Stop()
		logger.Info("vSphere采集器已停止")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.http.Shutdown(ctx)
}
