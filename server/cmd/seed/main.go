package main

import (
	"fmt"
	"log"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
)

func main() {
	fmt.Println("🌱 初始化权限数据...")
	cfg, _ := config.Load()
	db, err := models.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	if err := models.SeedDefaultData(db); err != nil {
		log.Fatalf("初始化数据失败: %v", err)
	}

	fmt.Println("✅ 权限数据初始化完成!")

	// 为测试用户分配角色
	testUserID := "3a2e28e4-759f-49b0-b4fe-f90d2769416f"
	adminRoleID := "11111111-1111-1111-1111-111111111111"

	if err := models.AssignRoleToUser(db,
		models.MustParseUUID(testUserID),
		models.MustParseUUID(adminRoleID),
	); err != nil {
		fmt.Printf("用户角色分配失败(可能已存在): %v\n", err)
	} else {
		fmt.Println("✅ 已为测试用户分配系统管理员角色")
	}
}

func init() {
	// 添加 MustParseUUID 辅助函数
}
