package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)



// InitPermissions 初始化权限和角色数据
func InitPermissions(db *gorm.DB) error {
	// 检查是否已有数据
	var count int64
	if err := db.Model(&Permission{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // 已有数据，跳过初始化
	}

	// 创建权限
	permissions := []Permission{
		// VM管理权限
		{ID: "vm:read", Name: "查看VM", Description: strPtr("查看虚拟机信息"), Resource: "vm", Action: "read", Level: "read"},
		{ID: "vm:write", Name: "编辑VM", Description: strPtr("编辑虚拟机信息"), Resource: "vm", Action: "write", Level: "write"},
		{ID: "vm:admin", Name: "管理VM", Description: strPtr("管理虚拟机（包括删除）"), Resource: "vm", Action: "admin", Level: "admin"},

		// 告警管理权限
		{ID: "alert:read", Name: "查看告警", Description: strPtr("查看告警规则和记录"), Resource: "alert", Action: "read", Level: "read"},
		{ID: "alert:write", Name: "编辑告警", Description: strPtr("编辑告警规则"), Resource: "alert", Action: "write", Level: "write"},
		{ID: "alert:admin", Name: "管理告警", Description: strPtr("管理告警（包括删除）"), Resource: "alert", Action: "admin", Level: "admin"},

		// 历史数据权限
		{ID: "history:read", Name: "查看历史数据", Description: strPtr("查询历史监控数据"), Resource: "history", Action: "read", Level: "read"},
		{ID: "history:export", Name: "导出数据", Description: strPtr("导出历史数据"), Resource: "history", Action: "write", Level: "write"},

		// 用户管理权限
		{ID: "user:read", Name: "查看用户", Description: strPtr("查看用户信息"), Resource: "user", Action: "read", Level: "read"},
		{ID: "user:write", Name: "编辑用户", Description: strPtr("编辑用户信息"), Resource: "user", Action: "write", Level: "write"},
		{ID: "user:admin", Name: "管理用户", Description: strPtr("管理用户（包括删除）"), Resource: "user", Action: "admin", Level: "admin"},

		// 系统权限
		{ID: "system:read", Name: "查看系统信息", Description: strPtr("查看系统健康状态"), Resource: "system", Action: "read", Level: "read"},
		{ID: "system:admin", Name: "系统管理", Description: strPtr("系统配置和管理"), Resource: "system", Action: "admin", Level: "admin"},
	}

	for _, perm := range permissions {
		if err := db.Create(&perm).Error; err != nil {
			return err
		}
	}

	// 创建角色
	roles := []Role{
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Name:        "系统管理员",
			Description: strPtr("拥有所有权限"),
			Level:       1,
			Path:        "/admin",
			IsSystem:    true,
		},
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Name:        "运维工程师",
			Description: strPtr("日常运维操作权限"),
			Level:       1,
			Path:        "/operator",
			IsSystem:    true,
		},
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
			Name:        "只读用户",
			Description: strPtr("仅查看权限"),
			ParentID:    uuidPtr("role_operator"),
			Level:       2,
			Path:        "/operator/viewer",
			IsSystem:    true,
		},
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"),
			Name:        "IT经理",
			Description: strPtr("查看和报表权限"),
			Level:       1,
			Path:        "/manager",
			IsSystem:    true,
		},
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"),
			Name:        "安全工程师",
			Description: strPtr("安全监控和审计权限"),
			Level:       1,
			Path:        "/security",
			IsSystem:    true,
		},
	}

	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}

	// 创建角色权限关联
	rolePermissions := []RolePermission{
		// 系统管理员拥有所有权限
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "vm:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "vm:write"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "vm:admin"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "alert:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "alert:write"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "alert:admin"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "history:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "history:export"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "user:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "user:write"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "user:admin"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "system:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), PermissionID: "system:admin"},

		// 运维工程师权限
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), PermissionID: "vm:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), PermissionID: "vm:write"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), PermissionID: "alert:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), PermissionID: "alert:write"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), PermissionID: "history:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), PermissionID: "history:export"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), PermissionID: "user:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"), PermissionID: "system:read"},

		// IT经理权限
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"), PermissionID: "vm:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"), PermissionID: "alert:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"), PermissionID: "history:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"), PermissionID: "history:export"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440004"), PermissionID: "system:read"},

		// 安全工程师权限
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"), PermissionID: "vm:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"), PermissionID: "alert:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"), PermissionID: "alert:admin"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"), PermissionID: "history:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"), PermissionID: "history:export"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"), PermissionID: "user:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"), PermissionID: "system:read"},
		{RoleID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440005"), PermissionID: "system:admin"},
	}

	for _, rp := range rolePermissions {
		if err := db.Create(&rp).Error; err != nil {
			return err
		}
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}

func uuidPtr(s string) *uuid.UUID {
	id := uuid.MustParse(s)
	return &id
}
