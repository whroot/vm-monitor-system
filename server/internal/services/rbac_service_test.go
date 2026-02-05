package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"vm-monitoring-system/internal/models"
)

func setupRBACTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
	)

	testData(db)

	return db, func() {
		db.Close()
	}
}

func testData(db *gorm.DB) {
	permissions := []models.Permission{
		{ID: "vm:read", Name: "查看虚拟机", Description: "可以查看虚拟机列表和详情"},
		{ID: "vm:write", Name: "管理虚拟机", Description: "可以创建、修改、删除虚拟机"},
		{ID: "alert:read", Name: "查看告警", Description: "可以查看告警列表"},
		{ID: "alert:manage", Name: "管理告警", Description: "可以创建、修改、删除告警规则"},
		{ID: "user:read", Name: "查看用户", Description: "可以查看用户列表"},
		{ID: "user:manage", Name: "管理用户", Description: "可以创建、修改、删除用户"},
	}
	db.Create(&permissions)

	roles := []models.Role{
		{ID: "role-admin", Name: "管理员", Description: "系统管理员"},
		{ID: "role-user", Name: "普通用户", Description: "普通用户"},
		{ID: "role-viewer", Name: "只读用户", Description: "只读用户"},
	}
	db.Create(&roles)

	rolePermissions := []models.RolePermission{
		{RoleID: "role-admin", PermissionID: "*"},
		{RoleID: "role-user", PermissionID: "vm:read"},
		{RoleID: "role-user", PermissionID: "vm:write"},
		{RoleID: "role-user", PermissionID: "alert:read"},
		{RoleID: "role-viewer", PermissionID: "vm:read"},
		{RoleID: "role-viewer", PermissionID: "alert:read"},
	}
	db.Create(&rolePermissions)

	users := []models.User{
		{ID: "user-admin", Username: "admin", Password: "hashed_password", Status: "active"},
		{ID: "user-normal", Username: "normal_user", Password: "hashed_password", Status: "active"},
		{ID: "user-viewer", Username: "viewer", Password: "hashed_password", Status: "active"},
	}
	db.Create(&users)

	userRoles := []models.UserRole{
		{UserID: "user-admin", RoleID: "role-admin"},
		{UserID: "user-normal", RoleID: "role-user"},
		{UserID: "user-viewer", RoleID: "role-viewer"},
	}
	db.Create(&userRoles)
}

func TestRBACService_CheckPermission(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("AdminHasAllPermissions", func(t *testing.T) {
		hasPermission, err := service.CheckPermission("user-admin", "vm:read")
		assert.NoError(t, err)
		assert.True(t, hasPermission)

		hasPermission, err = service.CheckPermission("user-admin", "user:manage")
		assert.NoError(t, err)
		assert.True(t, hasPermission)

		hasPermission, err = service.CheckPermission("user-admin", "system:config")
		assert.NoError(t, err)
		assert.True(t, hasPermission)
	})

	t.Run("NormalUserHasLimitedPermissions", func(t *testing.T) {
		hasPermission, err := service.CheckPermission("user-normal", "vm:read")
		assert.NoError(t, err)
		assert.True(t, hasPermission)

		hasPermission, err = service.CheckPermission("user-normal", "vm:write")
		assert.NoError(t, err)
		assert.True(t, hasPermission)

		hasPermission, err = service.CheckPermission("user-normal", "user:manage")
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})

	t.Run("ViewerHasReadOnlyPermissions", func(t *testing.T) {
		hasPermission, err := service.CheckPermission("user-viewer", "vm:read")
		assert.NoError(t, err)
		assert.True(t, hasPermission)

		hasPermission, err = service.CheckPermission("user-viewer", "vm:write")
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		hasPermission, err := service.CheckPermission("non-existent", "vm:read")
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})
}

func TestRBACService_GetUserPermissions(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("AdminPermissions", func(t *testing.T) {
		permissions, err := service.GetUserPermissions("user-admin")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(permissions), 6)
	})

	t.Run("NormalUserPermissions", func(t *testing.T) {
		permissions, err := service.GetUserPermissions("user-normal")
		assert.NoError(t, err)
		assert.Len(t, permissions, 3)
	})

	t.Run("ViewerPermissions", func(t *testing.T) {
		permissions, err := service.GetUserPermissions("user-viewer")
		assert.NoError(t, err)
		assert.Len(t, permissions, 2)
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		permissions, err := service.GetUserPermissions("non-existent")
		assert.NoError(t, err)
		assert.Empty(t, permissions)
	})
}

func TestRBACService_GetUserRoles(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("UserRoles", func(t *testing.T) {
		roles, err := service.GetUserRoles("user-admin")
		assert.NoError(t, err)
		assert.Len(t, roles, 1)
		assert.Equal(t, "管理员", roles[0].Name)
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		roles, err := service.GetUserRoles("non-existent")
		assert.NoError(t, err)
		assert.Empty(t, roles)
	})
}

func TestRBACService_HasRole(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("UserHasRole", func(t *testing.T) {
		hasRole, err := service.HasRole("user-admin", "管理员")
		assert.NoError(t, err)
		assert.True(t, hasRole)

		hasRole, err = service.HasRole("user-normal", "普通用户")
		assert.NoError(t, err)
		assert.True(t, hasRole)
	})

	t.Run("UserDoesNotHaveRole", func(t *testing.T) {
		hasRole, err := service.HasRole("user-normal", "管理员")
		assert.NoError(t, err)
		assert.False(t, hasRole)
	})
}

func TestRBACService_GetRolePermissions(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("AdminRolePermissions", func(t *testing.T) {
		permissions, err := service.GetRolePermissions("role-admin")
		assert.NoError(t, err)
		assert.Len(t, permissions, 1)
		assert.Equal(t, "*", permissions[0].ID)
	})

	t.Run("UserRolePermissions", func(t *testing.T) {
		permissions, err := service.GetRolePermissions("role-user")
		assert.NoError(t, err)
		assert.Len(t, permissions, 3)
	})
}

func TestRBACService_AssignRoleToUser(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("AssignNewRole", func(t *testing.T) {
		err := service.AssignRoleToUser("user-viewer", "role-user")
		assert.NoError(t, err)

		hasRole, err := service.HasRole("user-viewer", "普通用户")
		assert.NoError(t, err)
		assert.True(t, hasRole)
	})

	t.Run("AssignDuplicateRole", func(t *testing.T) {
		err := service.AssignRoleToUser("user-normal", "role-user")
		assert.NoError(t, err)
	})
}

func TestRBACService_RemoveRoleFromUser(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("RemoveRole", func(t *testing.T) {
		err := service.RemoveRoleFromUser("user-normal", "role-user")
		assert.NoError(t, err)

		hasRole, err := service.HasRole("user-normal", "普通用户")
		assert.NoError(t, err)
		assert.False(t, hasRole)
	})
}

func TestRBACService_AssignPermissionToRole(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("AssignPermission", func(t *testing.T) {
		err := service.AssignPermissionToRole("role-viewer", "user:read")
		assert.NoError(t, err)

		permissions, err := service.GetRolePermissions("role-viewer")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(permissions), 3)
	})
}

func TestRBACService_RemovePermissionFromRole(t *testing.T) {
	db, teardown := setupRBACTestDB()
	defer teardown()

	service := NewRBACService(db)

	t.Run("RemovePermission", func(t *testing.T) {
		err := service.RemovePermissionFromRole("role-user", "vm:read")
		assert.NoError(t, err)

		hasPermission, err := service.CheckPermission("user-normal", "vm:read")
		assert.NoError(t, err)
		assert.False(t, hasPermission)
	})
}
