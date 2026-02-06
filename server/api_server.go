package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🔧 启动基础API服务器...")
	
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
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "VM监控系统运行正常",
		})
	})
	
	// 获取用户列表
	router.GET("/api/v1/users", func(c *gin.Context) {
		var users []models.User
		if err := db.Find(&users).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"code": 200,
			"message": "获取成功",
			"data": users,
		})
	})
	
	// 获取VM列表
	router.GET("/api/v1/vms", func(c *gin.Context) {
		var vms []models.VM
		if err := db.Find(&vms).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"code": 200,
			"message": "获取成功", 
			"data": gin.H{
				"vms": vms,
				"total": len(vms),
				"page": 1,
				"pageSize": 10,
			},
		})
	})
	
	// 创建用户
	router.POST("/api/v1/users", func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		
		if err := db.Create(&user).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(201, gin.H{
			"code": 201,
			"message": "用户创建成功",
			"data": user,
		})
	})
	
	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("🚀 API服务器启动: http://localhost:%d\n", cfg.Server.Port)
	fmt.Printf("📋 可用端点:\n")
	fmt.Printf("  - GET  http://localhost:%d/health\n", cfg.Server.Port)
	fmt.Printf("  - GET  http://localhost:%d/api/v1/users\n", cfg.Server.Port)
	fmt.Printf("  - GET  http://localhost:%d/api/v1/vms\n", cfg.Server.Port)
	fmt.Printf("  - POST http://localhost:%d/api/v1/users\n", cfg.Server.Port)
	
	if err := router.Run(addr); err != nil {
		fmt.Printf("❌ 服务器启动失败: %v\n", err)
	}
}