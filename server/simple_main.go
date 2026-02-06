package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vm-monitoring-system/internal/api"
	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/logger"
	"vm-monitoring-system/internal/models"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("🔍 简化启动测试...")
	
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 配置: %+v\n", cfg.Server)
	
	// 初始化数据库
	db, err := models.InitDB(cfg.Database)
	if err != nil {
		fmt.Printf("❌ 数据库失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ 数据库OK")
	
	// 创建服务器
	server := api.NewServer(cfg, db)
	fmt.Println("✅ 服务器创建")
	
	// 启动服务器
	go func() {
		fmt.Println("🔄 正在启动HTTP服务器...")
		if err := server.Start(); err != nil {
			fmt.Printf("❌ 启动失败: %v\n", err)
		} else {
			fmt.Println("✅ 服务器启动成功")
		}
	}()
	
	// 等待一下看看
	time.Sleep(3 * time.Second)
	fmt.Printf("🚀 服务器应该在: http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
	
	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case sig := <-quit:
		fmt.Printf("收到信号: %v\n", sig)
	case <-time.After(10 * time.Second):
		fmt.Println("超时退出")
	}
}