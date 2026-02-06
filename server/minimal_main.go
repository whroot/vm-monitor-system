package main

import (
	"fmt"
	"net/http"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
)

func main() {
	fmt.Println("🔍 最简服务器启动测试...")
	
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ 配置失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 配置加载: %+v\n", cfg.Server)
	
	// 初始化数据库
	_, err = models.InitDB(cfg.Database)
	if err != nil {
		fmt.Printf("❌ 数据库失败: %v\n", err)
		return
	}
	fmt.Println("✅ 数据库OK")
	
	// 创建简单HTTP服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "VM监控系统后端运行正常！\n")
		fmt.Fprintf(w, "时间: %s\n", cfg.Server)
	})
	
	// 启动HTTP服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("🚀 启动HTTP服务器: http://%s\n", addr)
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("❌ 服务器启动失败: %v\n", err)
		return
	}
}