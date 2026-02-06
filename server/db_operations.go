package main

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
)

// strPtr 字符串指针辅助函数
func strPtr(s string) *string {
	return &s
}

func main() {
	fmt.Println("🧪 数据库操作测试...")
	
	// 初始化数据库
	dbConfig := config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres", 
		Password:        "postgres",
		Database:        "vm_monitoring",
		SSLMode:         "disable",
	}
	db, err := models.InitDB(dbConfig)
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}
	
	fmt.Println("✅ 数据库连接成功")
	
	// 创建测试用户
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("test123"), bcrypt.DefaultCost)
	user := models.User{
		ID:                 uuid.New(),
		Username:           "testuser",
		Email:              "test@example.com",
		PasswordHash:       string(passwordHash),
		Name:               "测试用户",
		Status:             "active",
		MustChangePassword: false,
		MFAEnabled:         false,
		Preferences:       models.UserPreferences{Language: "zh-CN"},
	}
	
	if err := db.Create(&user).Error; err != nil {
		fmt.Printf("❌ 创建用户失败: %v\n", err)
	} else {
		fmt.Printf("✅ 用户创建成功: %s\n", user.Username)
	}
	
	// 查询用户
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		fmt.Printf("❌ 查询用户失败: %v\n", err)
	} else {
		fmt.Printf("✅ 用户总数: %d\n", len(users))
		for _, u := range users {
			fmt.Printf("  - %s (%s)\n", u.Username, u.Email)
		}
	}
	
	// 创建测试VM
	vm := models.VM{
		ID:        uuid.New(),
		VMwareID: strPtr("vm-test-001"),
		Name:      "测试虚拟机",
		IP:        strPtr("192.168.1.100"),
		OSType:    strPtr("linux"),
		Status:    "running",
	}
	
	if err := db.Create(&vm).Error; err != nil {
		fmt.Printf("❌ 创建VM失败: %v\n", err)
	} else {
		fmt.Printf("✅ VM创建成功: %s\n", vm.Name)
	}
	
	fmt.Println("🎉 数据库操作测试完成！")
}