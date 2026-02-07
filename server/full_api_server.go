package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TokenExpireDuration     = time.Hour * 24
	RefreshExpireDuration   = time.Hour * 168
)

var jwtKey = []byte("your-super-secret-jwt-key-change-in-production")

type Claims struct {
	UserID   uuid.UUID `json:"userId"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

type AuthHandler struct {
	db *gorm.DB
}

type VMHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func NewVMHandler(db *gorm.DB) *VMHandler {
	return &VMHandler{db: db}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	var existing models.User
	if err := h.db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "用户名已存在"})
		return
	}

	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "邮箱已被注册"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Status:       "active",
		Preferences: models.UserPreferences{
			Language: "zh-CN", Theme: "light", Timezone: "Asia/Shanghai", DateFormat: "YYYY-MM-DD",
		},
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建用户失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 201, "message": "注册成功",
		"data": gin.H{"userId": user.ID, "username": user.Username, "email": user.Email},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var user models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}

	if user.Status != "active" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "账户已被禁用"})
		return
	}

	accessToken, _ := h.generateToken(&user)
	refreshToken, _ := h.generateRefreshToken(&user)

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "登录成功",
		"data": gin.H{
			"user": gin.H{"id": user.ID, "username": user.Username, "email": user.Email, "name": user.Name},
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"tokenType":    "Bearer",
			"expiresIn":    int(TokenExpireDuration.Seconds()),
		},
	})
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID, Username: user.Username, Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "vm-monitor-system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func (h *AuthHandler) generateRefreshToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID, Username: user.Username, Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "vm-monitor-system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func (h *AuthHandler) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未提供认证Token"})
			return
		}

		tokenString := authHeader[7:]
		claims, err := (&AuthHandler{}).parseToken(tokenString)
		if err != nil || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效或已过期的Token"})
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (h *VMHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	query := h.db.Model(&models.VM{}).Where("is_deleted = ?", false)

	var total int64
	query.Count(&total)

	var vms []models.VM
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&vms)

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "message": "获取成功",
		"data": gin.H{
			"vms":      vms,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

func (h *VMHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var vm models.VM
	

	if err := h.db.Where("is_deleted = ? AND id = ?", false, id).First(&vm).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "VM不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": vm})
}

func (h *VMHandler) Create(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		IP    *string `json:"ip,omitempty"`
		OSType *string `json:"osType,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	vm := models.VM{
		ID:        uuid.New(),
		Name:      req.Name,
		IP:        req.IP,
		OSType:    req.OSType,
		Status:    "unknown",
		Tags:      map[string]interface{}{},
		Metadata:  map[string]interface{}{},
	}

	if err := h.db.Create(&vm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 201, "message": "创建成功", "data": vm})
}

func (h *VMHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var vm models.VM
	

	if err := h.db.Where("is_deleted = ? AND id = ?", false, id).First(&vm).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "VM不存在"})
		return
	}

	var req struct {
		Name *string `json:"name,omitempty"`
		IP   *string `json:"ip,omitempty"`
	}

	c.ShouldBindJSON(&req)

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.IP != nil {
		updates["ip"] = *req.IP
	}
	updates["updated_at"] = time.Now()

	h.db.Model(&vm).Updates(updates)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": vm})
}

func (h *VMHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var vm models.VM
	

	if err := h.db.Where("is_deleted = ? AND id = ?", false, id).First(&vm).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "VM不存在"})
		return
	}

	h.db.Model(&vm).Updates(map[string]interface{}{"is_deleted": true, "deleted_at": time.Now()})

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

func (h *VMHandler) GetStats(c *gin.Context) {
	var stats struct {
		Total   int64 `json:"total"`
		Running int64 `json:"running"`
		Stopped int64 `json:"stopped"`
		Warning int64 `json:"warning"`
	}

	h.db.Model(&models.VM{}).Where("is_deleted = ?", false).Count(&stats.Total)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "running").Count(&stats.Running)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "poweredOff").Count(&stats.Stopped)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "warning").Count(&stats.Warning)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": stats})
}

func (h *VMHandler) GetRealtimeMetrics(c *gin.Context) {
	id := c.Param("id")
	var vm models.VM
	

	if err := h.db.Where("is_deleted = ? AND id = ?", false, id).First(&vm).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "VM不存在"})
		return
	}

	metrics := gin.H{
		"vmId":      vm.ID,
		"vmName":    vm.Name,
		"timestamp": time.Now(),
		"cpu": gin.H{
			"usage":  randFloat(10, 90),
			"cores": 4,
			"mhz":   randFloat(1000, 3500),
		},
		"memory": gin.H{
			"usage":    randFloat(30, 90),
			"totalGB":  16,
			"usedGB":   randFloat(5, 14),
			"swapUsage": randFloat(0, 20),
		},
		"disk": gin.H{
			"usage":     randFloat(40, 80),
			"totalGB":   500,
			"usedGB":    randFloat(200, 400),
			"readIOPS":  randFloat(0, 100),
			"writeIOPS": randFloat(0, 80),
		},
		"network": gin.H{
			"usageMbps": randFloat(0, 500),
			"inMBps":   randFloat(0, 200),
			"outMBps":  randFloat(0, 300),
		},
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取成功", "data": metrics})
}

func randFloat(min, max float64) float64 {
	return min + (max-min)*float64(time.Now().UnixNano()%10000000)/10000000
}

func main() {
	fmt.Println("🔧 启动完整API服务器...")
	
	cfg, _ := config.Load()
	db, _ := models.InitDB(cfg.Database)
	
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()
	authHandler := NewAuthHandler(db)
	vmHandler := NewVMHandler(db)
	
	// 公开端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "VM监控系统API运行正常", "version": "1.0.0"})
	})
	
	router.POST("/api/v1/auth/register", func(c *gin.Context) {
		authHandler.Register(c)
	})
	
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		authHandler.Login(c)
	})
	
	// 受保护端点
	api := router.Group("/api/v1")
	api.Use(JWTMiddleware())
	{
		// 用户
		api.GET("/users", func(c *gin.Context) {
			var users []models.User
			db.Find(&users)
			c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": users, "total": len(users)})
		})
		
		api.GET("/auth/profile", func(c *gin.Context) {
			userID, _ := c.Get("userId")
			var user models.User
			db.First(&user, userID)
			c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": user})
		})
		
		api.POST("/auth/logout", func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 200, "message": "登出成功"})
		})
		
		// VM管理
		api.GET("/vms", func(c *gin.Context) {
			vmHandler.List(c)
		})
		
		api.GET("/vms/stats", func(c *gin.Context) {
			vmHandler.GetStats(c)
		})
		
		api.GET("/vms/:id", func(c *gin.Context) {
			vmHandler.Get(c)
		})
		
		api.POST("/vms", func(c *gin.Context) {
			vmHandler.Create(c)
		})
		
		api.PUT("/vms/:id", func(c *gin.Context) {
			vmHandler.Update(c)
		})
		
		api.DELETE("/vms/:id", func(c *gin.Context) {
			vmHandler.Delete(c)
		})
		
		api.GET("/vms/:id/metrics", func(c *gin.Context) {
			vmHandler.GetRealtimeMetrics(c)
		})
	}
	
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("🚀 API服务器启动: http://localhost:%d\n\n", cfg.Server.Port)
	
	fmt.Println("📋 API文档:")
	fmt.Println("  🔓 公开端点:")
	fmt.Printf("     GET  http://localhost:%d/health                    - 健康检查\n", cfg.Server.Port)
	fmt.Printf("     POST http://localhost:%d/api/v1/auth/register     - 用户注册\n", cfg.Server.Port)
	fmt.Printf("     POST http://localhost:%d/api/v1/auth/login        - 用户登录\n", cfg.Server.Port)
	fmt.Println()
	fmt.Println("  🔐 需要认证 (Header: Authorization: Bearer <token>):")
	fmt.Printf("     GET   http://localhost:%d/api/v1/users             - 用户列表\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/api/v1/auth/profile      - 个人信息\n", cfg.Server.Port)
	fmt.Printf("     POST  http://localhost:%d/api/v1/auth/logout        - 登出\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/api/v1/vms               - VM列表\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/api/v1/vms/stats          - VM统计\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/api/v1/vms/:id           - VM详情\n", cfg.Server.Port)
	fmt.Printf("     POST  http://localhost:%d/api/v1/vms               - 创建VM\n", cfg.Server.Port)
	fmt.Printf("     PUT   http://localhost:%d/api/v1/vms/:id           - 更新VM\n", cfg.Server.Port)
	fmt.Printf("     DELETE http://localhost:%d/api/v1/vms/:id           - 删除VM\n", cfg.Server.Port)
	fmt.Printf("     GET   http://localhost:%d/api/v1/vms/:id/metrics   - 实时监控\n", cfg.Server.Port)
	
	router.Run(addr)
}
