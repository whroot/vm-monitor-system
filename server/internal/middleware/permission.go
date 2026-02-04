package middleware

import (
	"net/http"
	"strings"

	"vm-monitoring-system/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionMiddleware struct {
	db *gorm.DB
}

func NewPermissionMiddleware(db *gorm.DB) *PermissionMiddleware {
	return &PermissionMiddleware{db: db}
}

// RequirePermission 需要指定权限
func (m *PermissionMiddleware) RequirePermission(permissionID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
			})
			c.Abort()
			return
		}

		// 检查权限
		var count int64
		m.db.Table("users").
			Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Joins("JOIN role_permissions ON user_roles.role_id = role_permissions.role_id").
			Where("users.id = ? AND role_permissions.permission_id = ?", userID, permissionID).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
				"required": permissionID,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 需要指定角色
func (m *PermissionMiddleware) RequireRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未认证",
			})
			c.Abort()
			return
		}

		// 检查角色
		var count int64
		placeholders := make([]string, len(roleNames))
		args := make([]interface{}, len(roleNames)+1)
		args[0] = userID

		for i, roleName := range roleNames {
			placeholders[i] = "?"
			args[i+1] = roleName
		}

		m.db.Table("users").
			Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Joins("JOIN roles ON user_roles.role_id = roles.id").
			Where("users.id = ? AND roles.name IN ("+strings.Join(placeholders, ",")+")", args...).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "角色不符",
				"required": roleNames,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoadUserPermissions 加载用户权限到上下文
func (m *PermissionMiddleware) LoadUserPermissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		// 获取用户权限
		var permissions []models.Permission
		m.db.Table("permissions").
			Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
			Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
			Where("user_roles.user_id = ?", userID).
			Find(&permissions)

		// 提取权限ID
		permissionIDs := make([]string, len(permissions))
		for i, p := range permissions {
			permissionIDs[i] = p.ID
		}

		c.Set("permissions", permissionIDs)
		c.Next()
	}
}
