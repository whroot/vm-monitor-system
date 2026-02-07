package main

import (
	"fmt"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🔧 启动完整API服务器...")
	
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 配置失败: %v\n", err)
		return
	}
	
	// 初始化数据库
	db, err := models.InitDB(cfg.Database)
	if err != nil {
		fmt.Printf("❌ 数据库失败: %v\n", err)
		return
	}
	
	// 创建Gin路由
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()
	
	// 创建认证处理器
	authHandler := NewAuthHandler(db)
	
	// ====== 公开端点（无需认证） ======
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "VM监控系统API运行正常",
			"version": "1.0.0",
		})
	})
	
	// 用户注册
	router.POST("/api/v1/auth/register", func(c *gin.Context) {
		authHandler.Register(c)
	})
	
	// 用户登录
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		authHandler.Login(c)
	})
	
	// 刷新Token
	router.POST("/api/v1/auth/refresh", func(c *gin.Context) {
		authHandler.RefreshToken(c)
	})
	
	// ====== 受保护端点（需要认证） ======
	
	api := router.Group("/api/v1")
	api.Use(JWTMiddleware())
	{
		// 用户相关
		api.GET("/users", func(c *gin.Context) {
			var users []models.User
			if err := db.Find(&users).Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取成功",
				"data":    users,
				"total":   len(users),
			})
		})
		
		// 获取当前用户信息
		api.GET("/auth/profile", func(c *gin.Context) {
			authHandler.GetProfile(c)
		})
		
		// 用户登出
		api.POST("/auth/logout", func(c *gin.Context) {
			authHandler.Logout(c)
		})
		
		// 修改密码
		api.POST("/auth/change-password", func(c *gin.Context) {
			authHandler.ChangePassword(c)
		})
		
		// 获取VM列表
		api.GET("/vms", func(c *gin.Context) {
			var vms []models.VM
			if err := db.Find(&vms).Error; err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取成功",
				"data": gin.H{
					"vms":    vms,
					"total":  len(vms),
					"page":   1,
					"pageSize": 10,
				},
			})
		})
		
		// 获取VM统计
		api.GET("/vms/stats", func(c *gin.Context) {
			var total, running, stopped, warning int64
			db.Model(&models.VM{}).Count(&total)
			
			c.JSON(200, gin.H{
				"code":    200,
				"message": "获取成功",
				"data": gin.H{
					"total":    total,
					"running":  running,
					"stopped":  stopped,
					"warning":  warning,
				},
			})
		})
	}
	
	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("🚀 API服务器启动: http://localhost:%d\n", cfg.Server.Port)
	fmt.Printf("\n📋 API文档:\n")
	fmt.Printf("  🔓 公开端点:\n")
	fmt.Printf("     POST  http://localhost:%d/api/v1/auth/register  - 用户注册\n", cfg.Server.Port)
	fmt.Printf("     POST  http://localhost:%d/api/v1/auth/login     - 用户登录\n", cfg.Server.Port)
	fmt.Printf("     POST  http://localhost:%d/api/v1/auth/refresh   - 刷新Token\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/health               - 健康检查\n", cfg.Server.Port)
	fmt.Printf("\n  🔐 需要认证 (Header: Authorization: Bearer <token>):\n")
	fmt.Printf("     GET   http://localhost:%d/api/v1/users          - 用户列表\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/api/v1/auth/profile   - 当前用户信息\n", cfg.Server.Port)
	fmt.Printf("     POST  http://localhost:%d/api/v1/auth/logout    - 登出\n", cfg.Server.Port)
	fmt.Printf("     POST  http://localhost:%d/api/v1/auth/change-password - 修改密码\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/api/v1/vms            - VM列表\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/api/v1/vms/stats       - VM统计\n", cfg.Server.Port)
	
	if err := router.Run(addr); err != nil {
		fmt.Printf("❌ 服务器启动失败: %v\n", err)
	}
}