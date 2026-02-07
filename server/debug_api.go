package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🔧 调试API服务器启动...")
	
	// 创建Gin路由
	router := gin.Default()
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "API服务器运行正常",
		})
	})
	
	// 测试路由
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"test": "success"})
	})
	
	// 启动服务器
	fmt.Println("🚀 启动API服务器: http://localhost:8080")
	fmt.Println("📋 可用端点:")
	fmt.Println("  - GET  http://localhost:8080/health")
	fmt.Println("  - GET  http://localhost:8080/test")
	
	if err := router.Run(":8080"); err != nil {
		fmt.Printf("❌ 服务器启动失败: %v\n", err)
		return
	}
}