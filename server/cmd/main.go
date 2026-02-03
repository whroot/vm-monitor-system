package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"vm-monitoring-system/internal/api"
	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	db, err := models.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 自动迁移（仅开发环境使用）
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化缓存
	if err := models.InitCache(cfg.Redis); err != nil {
		log.Fatalf("初始化缓存失败: %v", err)
	}

	// 初始化权限数据
	if err := models.InitPermissions(db); err != nil {
		log.Fatalf("初始化权限数据失败: %v", err)
	}

	// 创建并启动HTTP服务器
	server := api.NewServer(cfg, db)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	<-quit
	log.Println("正在关闭服务器...")
	if err := server.Stop(); err != nil {
		log.Printf("关闭服务器出错: %v", err)
	}
	log.Println("服务器已关闭")
}
