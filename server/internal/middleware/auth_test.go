package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"vm-monitoring-system/internal/models"
)

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("MissingAuthorizationHeader", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)

		JWTAuth("test-secret")(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "未提供认证信息")
	})

	t.Run("InvalidAuthorizationFormat", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
		c.Request.Header.Set("Authorization", "InvalidFormat token123")

		JWTAuth("test-secret")(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "认证格式错误")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid.token.here")

		JWTAuth("test-secret")(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token无效或已过期")
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		secret := "test-secret"
		claims := jwt.MapClaims{
			"sub":   "user-123",
			"type":  "access",
			"exp":   time.Now().Add(-time.Hour).Unix(),
			"iat":   time.Now().Add(-2 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(secret))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokenString)

		JWTAuth(secret)(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("WrongTokenType", func(t *testing.T) {
		secret := "test-secret"
		claims := jwt.MapClaims{
			"sub":   "user-123",
			"type":  "refresh",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(secret))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokenString)

		JWTAuth(secret)(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Token类型错误")
	})

	t.Run("ValidToken", func(t *testing.T) {
		secret := "test-secret"
		claims := jwt.MapClaims{
			"sub":   "user-123",
			"type":  "access",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(secret))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokenString)

		JWTAuth(secret)(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CORSPreflightRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("OPTIONS", "/api/v1/test", nil)

		CORS()(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("CORSNormalRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)

		CORS()(c)
		c.Next()

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestRequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GenerateRequestID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test?query=value", nil)

		RequestLogger()(c)
		c.Next()

		requestID := c.GetHeader("X-Request-ID")
		assert.NotEmpty(t, requestID)
		assert.Contains(t, requestID, "req_")
	})

	t.Run("UseExistingRequestID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
		c.Request.Header.Set("X-Request-ID", "existing-request-id")

		RequestLogger()(c)
		c.Next()

		assert.Equal(t, "existing-request-id", c.GetHeader("X-Request-ID"))
	})
}

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("NoErrors", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)

		ErrorHandler()(c)
		c.Next()

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("WithError", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)

		ErrorHandler()(c)
		c.Error(gin.Error{
			Err:  assert.AnError,
			Type: gin.ErrorTypePrivate,
		})
		c.Next()

		assert.Equal(t, -1, w.Code)
	})
}

func TestUniqueStrings(t *testing.T) {
	t.Run("RemoveDuplicates", func(t *testing.T) {
		input := []string{"a", "b", "a", "c", "b", "d"}
		result := uniqueStrings(input)

		assert.Len(t, result, 4)
		assert.Contains(t, result, "a")
		assert.Contains(t, result, "b")
		assert.Contains(t, result, "c")
		assert.Contains(t, result, "d")
	})

	t.Run("EmptySlice", func(t *testing.T) {
		result := uniqueStrings([]string{})
		assert.Empty(t, result)
	})
}

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	assert.NotEmpty(t, id1)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "req_")
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("UserIDExists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(string(contextKeyUserID), "user-123")

		userID := GetUserID(c)

		assert.Equal(t, "user-123", userID)
	})

	t.Run("UserIDNotExists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userID := GetUserID(c)

		assert.Empty(t, userID)
	})
}

func TestHasPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("HasPermission", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(string(contextKeyPermissions), []string{"vm:read", "vm:write", "alert:manage"})

		assert.True(t, HasPermission(c, "vm:read"))
		assert.True(t, HasPermission(c, "vm:write"))
		assert.False(t, HasPermission(c, "user:manage"))
	})

	t.Run("WildcardPermission", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(string(contextKeyPermissions), []string{"*"})

		assert.True(t, HasPermission(c, "vm:read"))
		assert.True(t, HasPermission(c, "user:manage"))
		assert.True(t, HasPermission(c, "any:permission"))
	})

	t.Run("NoPermissions", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(string(contextKeyPermissions), []string{})

		assert.False(t, HasPermission(c, "vm:read"))
	})
}

func TestGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("UserExists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		testUser := models.User{
			ID:       "user-123",
			Username: "testuser",
		}
		c.Set(string(contextKeyUser), testUser)

		user := GetUser(c)

		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
	})

	t.Run("UserNotExists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		user := GetUser(c)

		assert.Nil(t, user)
	})
}

func TestGetPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("PermissionsExist", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(string(contextKeyPermissions), []string{"vm:read", "vm:write"})

		perms := GetPermissions(c)

		assert.Len(t, perms, 2)
		assert.Contains(t, perms, "vm:read")
	})

	t.Run("PermissionsNotExist", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		perms := GetPermissions(c)

		assert.Nil(t, perms)
	})
}
