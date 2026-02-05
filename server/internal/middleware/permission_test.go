package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"vm-monitoring-system/internal/models"
)

func setupPermissionTestDB() (*gorm.DB, func()) {
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

	return db, func() {
		db.Close()
	}
}

func TestPermissionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, teardown := setupPermissionTestDB()
	defer teardown()

	t.Run("RequirePermission_UserHasPermission", func(t *testing.T) {
		middleware := NewPermissionMiddleware(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "user-123")

		handler := middleware.RequirePermission("vm:read")
		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("RequirePermission_UserNotAuthenticated", func(t *testing.T) {
		middleware := NewPermissionMiddleware(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)

		handler := middleware.RequirePermission("vm:read")
		handler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "未认证")
	})

	t.Run("RequireRole_UserHasRole", func(t *testing.T) {
		middleware := NewPermissionMiddleware(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "user-123")

		handler := middleware.RequireRole("admin")
		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("RequireRole_UserNotAuthenticated", func(t *testing.T) {
		middleware := NewPermissionMiddleware(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)

		handler := middleware.RequireRole("admin")
		handler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("LoadUserPermissions", func(t *testing.T) {
		middleware := NewPermissionMiddleware(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "user-123")

		handler := middleware.LoadUserPermissions()
		handler(c)

		permissions, exists := c.Get("permissions")
		assert.True(t, exists)
		assert.NotNil(t, permissions)
	})

	t.Run("RequireRole_MultipleRoles", func(t *testing.T) {
		middleware := NewPermissionMiddleware(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "user-123")

		handler := middleware.RequireRole("admin", "moderator", "viewer")
		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("RequirePermission_UserLacksPermission", func(t *testing.T) {
		middleware := NewPermissionMiddleware(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "user-123")

		handler := middleware.RequirePermission("system:config")
		handler(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "权限不足")
	})
}
