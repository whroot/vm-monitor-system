package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"vm-monitoring-system/internal/api"
	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/logger"
	"vm-monitoring-system/internal/models"
	"go.uber.org/zap"
)

func main() {
	// 打印启动横幅
	printBanner()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 验证配置安全性
	fmt.Println("\n🔐 验证配置安全性...")
	api.ValidateConfig(cfg)

	db, err := models.InitDB(cfg.Database)
	if err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}

	if err := models.AutoMigrate(db); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	if err := models.InitCache(cfg.Redis); err != nil {
		logger.Fatal("初始化缓存失败", zap.Error(err))
	}

	if err := models.InitPermissions(db); err != nil {
		logger.Fatal("初始化权限数据失败", zap.Error(err))
	}

	// 打印安全提醒
	printSecurityReminder()

	server := api.NewServer(cfg, db)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\n🚀 启动VM监控系统...")
	fmt.Printf("📡 服务地址: http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Println("✅ 服务启动成功!\n")

	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("启动服务器失败", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("正在关闭服务器...")
	if err := server.Stop(); err != nil {
		logger.Error("关闭服务器出错", zap.Error(err))
	}
	logger.Info("服务器已关闭")
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║    ███╗   ██╗███████╗ ██████╗ ██████╗ ███████╗               ║
║    ████╗  ██║██╔════╝██╔════╝ ██╔══██╗██╔════╝               ║
║    ██╔██╗ ██║█████╗  ██║  ███╗██████╔╝█████╗                 ║
║    ██║╚██╗██║██╔══╝  ██║   ██║██╔══██╗██╔══╝                 ║
║    ██║ ╚████║███████╗╚██████╔╝██████╔╝███████╗               ║
║    ╚═╝  ╚═══╝╚══════╝ ╚═════╝ ╚═════╝ ╚══════╝               ║
║                                                                  ║
║                    虚拟机监控系统 v1.0                           ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

func printSecurityReminder() {
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║ ⚠️  安全提醒                                                    ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  1. 默认管理员账户: admin / admin123                            ║")
	fmt.Println("║  2. ⚠️  首次登录必须修改密码                                    ║")
	fmt.Println("║  3. 生产环境请设置以下环境变量:                                 ║")
	fmt.Println("║     - JWT_SECRET: JWT签名密钥 (必需)                            ║")
	fmt.Println("║     - DB_PASSWORD: 数据库密码 (必需)                            ║")
	fmt.Println("║     - REDIS_PASSWORD: Redis密码 (建议)                          ║")
	fmt.Println("║     - APP_MODE: 设为 production                                ║")
	fmt.Println("║                                                                  ║")
	fmt.Println("║  生成强JWT密钥:                                                 ║")
	fmt.Println("║     export JWT_SECRET=\"$(openssl rand -base64 64)\"             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
}
