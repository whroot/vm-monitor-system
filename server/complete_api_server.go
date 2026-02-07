package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"vm-monitoring-system/internal/config"
	"vm-monitoring-system/internal/models"
)

const (
	TokenExpireDuration   = time.Hour * 24
	RefreshExpireDuration = time.Hour * 168
)

var jwtKey = []byte("your-super-secret-jwt-key")

type Claims struct {
	UserID   uuid.UUID `json:"userId"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

type AuthHandler struct{ db *gorm.DB }
type VMHandler struct{ db *gorm.DB }
type AlertHandler struct{ db *gorm.DB }
type PermissionHandler struct{ db *gorm.DB }
type RoleHandler struct{ db *gorm.DB }

type AlertRule struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key"`
	Name               string    `gorm:"type:varchar(200);not null"`
	Description        *string   `gorm:"type:text"`
	Scope              string    `gorm:"type:varchar(20);not null"`
	Severity           string    `gorm:"type:varchar(20);not null"`
	Enabled            bool      `gorm:"not null;default:true"`
	Cooldown           int       `gorm:"not null;default:300"`
	NotificationConfig string    `gorm:"type:jsonb"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

type AlertRecord struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	RuleID         uuid.UUID  `gorm:"type:uuid;not null"`
	RuleName       string     `gorm:"type:varchar(200);not null"`
	VMID           *uuid.UUID `gorm:"type:uuid"`
	VMName         *string    `gorm:"type:varchar(200)"`
	Metric         string     `gorm:"type:varchar(50);not null"`
	Severity       string     `gorm:"type:varchar(20);not null"`
	TriggerValue   float64    `gorm:"not null"`
	Threshold      float64    `gorm:"not null"`
	TriggeredAt    time.Time  `gorm:"not null"`
	Status         string     `gorm:"type:varchar(20);not null;default:'active'"`
	AcknowledgedBy *uuid.UUID `gorm:"type:uuid"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
	ResolvedBy     *uuid.UUID `gorm:"type:uuid"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	Resolution     *string    `gorm:"type:text"`
	Duration       *int       `json:"duration,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
}

func NewAuthHandler(db *gorm.DB) *AuthHandler             { return &AuthHandler{db: db} }
func NewVMHandler(db *gorm.DB) *VMHandler                 { return &VMHandler{db: db} }
func NewAlertHandler(db *gorm.DB) *AlertHandler           { return &AlertHandler{db: db} }
func NewPermissionHandler(db *gorm.DB) *PermissionHandler { return &PermissionHandler{db: db} }
func NewRoleHandler(db *gorm.DB) *RoleHandler             { return &RoleHandler{db: db} }

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	var existing models.User
	if err := h.db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		c.JSON(409, gin.H{"code": 409, "message": "用户名已存在"})
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := models.User{
		ID: uuid.New(), Username: req.Username, Email: req.Email,
		PasswordHash: string(hashedPassword), Name: req.Name, Status: "active",
		Preferences: models.UserPreferences{Language: "zh-CN", Theme: "light", Timezone: "Asia/Shanghai", DateFormat: "YYYY-MM-DD"},
	}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(201, gin.H{"code": 201, "message": "注册成功", "data": gin.H{"userId": user.ID, "username": user.Username}})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct{ Username, Password string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	var user models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"code": 401, "message": "用户名或密码错误"})
		return
	}
	if user.Status != "active" {
		c.JSON(401, gin.H{"code": 401, "message": "账户已被禁用"})
		return
	}
	accessToken, _ := h.generateToken(&user)
	refreshToken, _ := h.generateRefreshToken(&user)
	c.JSON(200, gin.H{
		"code": 200, "message": "登录成功",
		"data": gin.H{
			"user": gin.H{
				"id":          user.ID,
				"username":    user.Username,
				"email":       user.Email,
				"name":        user.Name,
				"status":      user.Status,
				"preferences": user.Preferences,
			},
			"accessToken": accessToken, "refreshToken": refreshToken, "tokenType": "Bearer", "expiresIn": int(TokenExpireDuration.Seconds()),
		},
	})
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := &Claims{UserID: user.ID, Username: user.Username, Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpireDuration)), IssuedAt: jwt.NewNumericDate(time.Now()), Issuer: "vm-monitor"}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func (h *AuthHandler) generateRefreshToken(user *models.User) (string, error) {
	claims := &Claims{UserID: user.ID, Username: user.Username, Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshExpireDuration)), IssuedAt: jwt.NewNumericDate(time.Now()), Issuer: "vm-monitor"}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(401, gin.H{"code": 401, "message": "未提供认证Token"})
			return
		}
		claims, err := (&AuthHandler{}).parseToken(authHeader[7:])
		if err != nil || claims == nil {
			c.AbortWithStatusJSON(401, gin.H{"code": 401, "message": "无效或已过期的Token"})
			return
		}
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (h *AuthHandler) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) { return jwtKey, nil })
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (h *VMHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	var vms []models.VM
	var total int64

	// 构建查询条件
	query := h.db.Model(&models.VM{}).Where("is_deleted = ?", false)

	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 操作系统筛选
	if osType := c.Query("osType"); osType != "" {
		query = query.Where("os_type = ?", osType)
	}

	// 关键字搜索
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("(name LIKE ? OR ip::text LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 先获取总数
	query.Count(&total)

	// 获取数据
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&vms)
	if result.Error != nil {
		vms = []models.VM{}
		total = 0
	}

	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": gin.H{"vms": vms, "total": total, "page": page, "pageSize": pageSize}})
}

func (h *VMHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var vm models.VM
	if err := h.db.Where("is_deleted = ? AND id = ?", false, id).First(&vm).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "VM不存在"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": vm})
}

func (h *VMHandler) Create(c *gin.Context) {
	var req struct{ Name, IP, OSType *string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	vm := models.VM{ID: uuid.New(), Name: *req.Name, IP: req.IP, OSType: req.OSType, Status: "unknown", Tags: map[string]interface{}{}, Metadata: map[string]interface{}{}}
	if err := h.db.Create(&vm).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建失败: " + err.Error()})
		return
	}
	c.JSON(201, gin.H{"code": 201, "message": "创建成功", "data": vm})
}

func (h *VMHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var vm models.VM
	if err := h.db.Where("is_deleted = ? AND id = ?", false, id).First(&vm).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "VM不存在"})
		return
	}
	var req struct{ Name, IP, Status *string }
	c.ShouldBindJSON(&req)
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.IP != nil {
		updates["ip"] = *req.IP
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	updates["updated_at"] = time.Now()
	h.db.Model(&vm).Updates(updates)
	c.JSON(200, gin.H{"code": 200, "message": "更新成功", "data": vm})
}

func (h *VMHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var vm models.VM
	if err := h.db.Where("is_deleted = ? AND id = ?", false, id).First(&vm).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "VM不存在"})
		return
	}
	h.db.Model(&vm).Updates(map[string]interface{}{"is_deleted": true, "deleted_at": time.Now()})
	c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
}

func (h *VMHandler) GetStats(c *gin.Context) {
	var stats struct{ Total, Running, Stopped, Warning int64 }
	h.db.Model(&models.VM{}).Where("is_deleted = ?", false).Count(&stats.Total)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "running").Count(&stats.Running)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "poweredOff").Count(&stats.Stopped)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "warning").Count(&stats.Warning)
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": stats})
}

func (h *VMHandler) GetRealtimeMetrics(c *gin.Context) {
	id := c.Param("id")
	var vm models.VM
	if err := h.db.Where("is_deleted = ? AND id = ?", false, id).First(&vm).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "VM不存在"})
		return
	}
	metrics := gin.H{
		"vmId": vm.ID, "vmName": vm.Name, "timestamp": time.Now(),
		"cpu":     gin.H{"usage": randFloat(10, 90), "cores": 4, "mhz": randFloat(1000, 3500)},
		"memory":  gin.H{"usage": randFloat(30, 90), "totalGB": 16, "usedGB": randFloat(5, 14), "swapUsage": randFloat(0, 20)},
		"disk":    gin.H{"usage": randFloat(40, 80), "totalGB": 500, "usedGB": randFloat(200, 400), "readIOPS": randFloat(0, 100), "writeIOPS": randFloat(0, 80)},
		"network": gin.H{"usageMbps": randFloat(0, 500), "inMBps": randFloat(0, 200), "outMBps": randFloat(0, 300)},
	}
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": metrics})
}

func randFloat(min, max float64) float64 {
	return min + (max-min)*float64(time.Now().UnixNano()%10000000)/10000000
}

func (h *AlertHandler) ListRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize
	var total int64
	h.db.Model(&AlertRule{}).Count(&total)
	var rules []AlertRule
	h.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&rules)
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": gin.H{"rules": rules, "total": total, "page": page, "pageSize": pageSize}})
}

func (h *AlertHandler) GetRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的规则ID"})
		return
	}
	var rule AlertRule
	if err := h.db.First(&rule, "id = ?", ruleID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "规则不存在"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": rule})
}

func (h *AlertHandler) CreateRule(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Scope    string `json:"scope" binding:"required"`
		Severity string `json:"severity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	rule := AlertRule{ID: uuid.New(), Name: req.Name, Scope: req.Scope, Severity: req.Severity, Enabled: true, Cooldown: 300, NotificationConfig: "{}"}
	if err := h.db.Create(&rule).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(201, gin.H{"code": 201, "message": "创建成功", "data": gin.H{"ruleId": rule.ID}})
}

func (h *AlertHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的规则ID"})
		return
	}
	var rule AlertRule
	if err := h.db.First(&rule, "id = ?", ruleID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "规则不存在"})
		return
	}
	var req struct {
		Name, Severity *string
		Enabled        *bool
	}
	c.ShouldBindJSON(&req)
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	updates["updated_at"] = time.Now()
	h.db.Model(&rule).Updates(updates)
	c.JSON(200, gin.H{"code": 200, "message": "更新成功"})
}

func (h *AlertHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	ruleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的规则ID"})
		return
	}
	var rule AlertRule
	if err := h.db.First(&rule, "id = ?", ruleID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "规则不存在"})
		return
	}
	h.db.Delete(&rule)
	c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
}

func (h *AlertHandler) ListRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize
	var total int64
	h.db.Model(&AlertRecord{}).Count(&total)
	var records []AlertRecord
	h.db.Order("triggered_at DESC").Offset(offset).Limit(pageSize).Find(&records)
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": gin.H{"records": records, "total": total, "page": page, "pageSize": pageSize}})
}

func (h *AlertHandler) GetStats(c *gin.Context) {
	var stats struct{ Total, Active, Critical, Warning int64 }
	h.db.Model(&AlertRecord{}).Count(&stats.Total)
	h.db.Model(&AlertRecord{}).Where("status = ?", "active").Count(&stats.Active)
	h.db.Model(&AlertRecord{}).Where("severity = ?", "critical").Count(&stats.Critical)
	h.db.Model(&AlertRecord{}).Where("severity = ?", "warning").Count(&stats.Warning)
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": stats})
}

func (h *AlertHandler) Acknowledge(c *gin.Context) {
	id := c.Param("id")
	recordID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的告警ID"})
		return
	}
	var record AlertRecord
	if err := h.db.First(&record, "id = ?", recordID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "告警不存在"})
		return
	}
	now := time.Now()
	userID, _ := c.Get("userId")
	username, _ := c.Get("username")
	h.db.Model(&record).Updates(map[string]interface{}{
		"status": "acknowledged", "acknowledged_by": userID, "acknowledged_by_name": username,
		"acknowledged_at": &now, "updated_at": now,
	})
	c.JSON(200, gin.H{"code": 200, "message": "确认成功"})
}

func (h *AlertHandler) Resolve(c *gin.Context) {
	id := c.Param("id")
	recordID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的告警ID"})
		return
	}
	var record AlertRecord
	if err := h.db.First(&record, "id = ?", recordID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "告警不存在"})
		return
	}
	var req struct{ Resolution string }
	c.ShouldBindJSON(&req)
	now := time.Now()
	duration := int(now.Sub(record.TriggeredAt).Seconds())
	userID, _ := c.Get("userId")
	username, _ := c.Get("username")
	h.db.Model(&record).Updates(map[string]interface{}{
		"status": "resolved", "resolved_by": userID, "resolved_by_name": username,
		"resolved_at": &now, "duration": duration, "resolution": req.Resolution, "updated_at": now,
	})
	c.JSON(200, gin.H{"code": 200, "message": "解决成功"})
}

func (h *AlertHandler) CreateTestAlert(c *gin.Context) {
	var req struct {
		VMName, Severity, Metric string
		Value                    float64
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	record := AlertRecord{ID: uuid.New(), RuleID: uuid.New(), RuleName: "测试告警规则", VMName: &req.VMName, Metric: req.Metric, Severity: req.Severity, TriggerValue: req.Value, Threshold: 80, TriggeredAt: time.Now(), Status: "active"}
	if err := h.db.Create(&record).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(201, gin.H{"code": 201, "message": "测试告警创建成功", "data": gin.H{"alertId": record.ID}})
}

// Permission Handler
func (h *PermissionHandler) List(c *gin.Context) {
	var permissions []models.Permission
	h.db.Order("resource, action").Find(&permissions)
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": permissions})
}

func (h *PermissionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var permission models.Permission
	if err := h.db.First(&permission, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "权限不存在"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": permission})
}

func (h *PermissionHandler) Create(c *gin.Context) {
	var req struct {
		ID          string `json:"id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Resource    string `json:"resource" binding:"required"`
		Action      string `json:"action" binding:"required"`
		Level       string `json:"level"`
		Scope       string `json:"scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	desc := req.Description
	permission := models.Permission{
		ID:          req.ID,
		Name:        req.Name,
		Description: &desc,
		Resource:    req.Resource,
		Action:      req.Action,
		Level:       req.Level,
		Scope:       req.Scope,
	}
	if err := h.db.Create(&permission).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(201, gin.H{"code": 201, "message": "创建成功", "data": permission})
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&models.Permission{}, "id = ?", id).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
}

func (h *PermissionHandler) ListByRole(c *gin.Context) {
	roleID := c.Param("roleId")
	parsedRoleID, err := uuid.Parse(roleID)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}
	var role models.Role
	if err := h.db.Preload("Permissions").First(&role, parsedRoleID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "角色不存在"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": role.Permissions})
}

func (h *PermissionHandler) AssignToRole(c *gin.Context) {
	roleID := c.Param("roleId")
	parsedRoleID, err := uuid.Parse(roleID)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}
	var req struct {
		PermissionIDs []string `json:"permissionIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var role models.Role
	if err := h.db.First(&role, parsedRoleID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "角色不存在"})
		return
	}

	var permissions []models.Permission
	h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions)

	if err := h.db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "分配失败"})
		return
	}

	c.JSON(200, gin.H{"code": 200, "message": "分配成功"})
}

// Role Handler
func (h *RoleHandler) List(c *gin.Context) {
	var roles []models.Role
	h.db.Preload("Permissions").Order("level ASC").Find(&roles)

	for i := range roles {
		var count int64
		h.db.Model(&models.UserRole{}).Where("role_id = ?", roles[i].ID).Count(&count)
		roles[i].UserCount = int(count)
	}

	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": roles})
}

func (h *RoleHandler) Get(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}
	var role models.Role
	if err := h.db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "角色不存在"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": role})
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=2,max=50"`
		Description string `json:"description"`
		Level       int    `json:"level"`
		Path        string `json:"path" binding:"required"`
		IsSystem    bool   `json:"isSystem"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	desc := req.Description
	role := models.Role{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: &desc,
		Level:       req.Level,
		Path:        req.Path,
		IsSystem:    req.IsSystem,
	}

	if err := h.db.Create(&role).Error; err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(201, gin.H{"code": 201, "message": "创建成功", "data": gin.H{"roleId": role.ID}})
}

func (h *RoleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}
	var role models.Role
	if err := h.db.First(&role, roleID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "角色不存在"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Level       *int    `json:"level"`
		Path        *string `json:"path"`
	}
	c.ShouldBindJSON(&req)

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Level != nil {
		updates["level"] = *req.Level
	}
	if req.Path != nil {
		updates["path"] = *req.Path
	}

	h.db.Model(&role).Updates(updates)
	c.JSON(200, gin.H{"code": 200, "message": "更新成功"})
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}
	var role models.Role
	if err := h.db.First(&role, roleID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "角色不存在"})
		return
	}

	if role.IsSystem {
		c.JSON(400, gin.H{"code": 400, "message": "系统角色无法删除"})
		return
	}

	h.db.Delete(&role)
	c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
}

func (h *RoleHandler) GetUsers(c *gin.Context) {
	id := c.Param("id")
	roleID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的角色ID"})
		return
	}
	var users []models.User
	h.db.Joins("JOIN user_roles ON user_roles.user_id = users.id").Where("user_roles.role_id = ?", roleID).Find(&users)
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": users})
}

func (h *RoleHandler) AssignUser(c *gin.Context) {
	var req struct {
		UserID  uuid.UUID   `json:"userId" binding:"required"`
		RoleIDs []uuid.UUID `json:"roleIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	h.db.Where("user_id = ?", req.UserID).Delete(&models.UserRole{})

	for _, roleID := range req.RoleIDs {
		userRole := models.UserRole{
			ID:     uuid.New(),
			UserID: req.UserID,
			RoleID: roleID,
		}
		h.db.Create(&userRole)
	}

	c.JSON(200, gin.H{"code": 200, "message": "分配成功"})
}

func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("userId")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}
	var roles []models.Role
	h.db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").Where("user_roles.user_id = ?", parsedUserID).Find(&roles)
	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": roles})
}

func (h *RoleHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("userId")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	var user models.User
	if err := h.db.Preload("Roles.Permissions").First(&user, parsedUserID).Error; err != nil {
		c.JSON(404, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	permissionMap := make(map[string]models.Permission)
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			permissionMap[perm.ID] = perm
		}
	}

	permissions := make([]models.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": permissions})
}

func main() {
	fmt.Println("🔧 启动完整API服务器（含权限管理）...")
	cfg, _ := config.Load()
	db, _ := models.InitDB(cfg.Database)

	fmt.Println("🌱 初始化权限数据...")
	if err := models.SeedDefaultData(db); err != nil {
		fmt.Printf("⚠️  初始化权限数据失败: %v\n", err)
	} else {
		fmt.Println("✅ 权限数据初始化完成")
	}

	// 自动迁移数据库结构
	db.AutoMigrate(&models.User{})
	fmt.Println("✅ 数据库迁移完成")

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	authHandler := NewAuthHandler(db)
	vmHandler := NewVMHandler(db)
	alertHandler := NewAlertHandler(db)
	permissionHandler := NewPermissionHandler(db)
	roleHandler := NewRoleHandler(db)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "VM监控系统API运行正常", "version": "1.0.0"})
	})

	// 创建uploads目录
	os.MkdirAll("uploads", 0755)
	// 静态文件服务（头像）
	router.Static("/uploads", "./uploads")

	router.POST("/api/v1/auth/register", func(c *gin.Context) { authHandler.Register(c) })
	router.POST("/api/v1/auth/login", func(c *gin.Context) { authHandler.Login(c) })

	api := router.Group("/api/v1")
	api.Use(JWTMiddleware())
	{
		api.GET("/users", func(c *gin.Context) {
			var users []models.User
			db.Preload("Roles").Find(&users)
			baseURL := strings.TrimRight(c.Request.Host, "/")
			for i := range users {
				if users[i].Avatar != nil && strings.HasPrefix(*users[i].Avatar, "/uploads/") {
					avatarURL := fmt.Sprintf("http://%s%s", baseURL, *users[i].Avatar)
					users[i].Avatar = &avatarURL
				}
			}
			c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": users})
		})
		api.DELETE("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			var user models.User
			if err := db.First(&user, "id = ?", id).Error; err != nil {
				c.JSON(404, gin.H{"code": 404, "message": "用户不存在"})
				return
			}
			if err := db.Unscoped().Delete(&user).Error; err != nil {
				c.JSON(500, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
				return
			}
			c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
		})
		api.GET("/auth/profile", func(c *gin.Context) {
			userID, _ := c.Get("userId")
			var user models.User
			db.Preload("Roles").First(&user, userID)
			baseURL := strings.TrimRight(c.Request.Host, "/")
			if user.Avatar != nil && strings.HasPrefix(*user.Avatar, "/uploads/") {
				avatarURL := fmt.Sprintf("http://%s%s", baseURL, *user.Avatar)
				user.Avatar = &avatarURL
			}
			c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": user})
		})
		api.PUT("/auth/profile", func(c *gin.Context) {
			userID, _ := c.Get("userId")
			var user models.User
			if err := db.First(&user, userID).Error; err != nil {
				c.JSON(404, gin.H{"code": 404, "message": "用户不存在"})
				return
			}
			var req struct {
				Name       *string `json:"name"`
				Email      *string `json:"email"`
				Phone      *string `json:"phone"`
				Department *string `json:"department"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
				return
			}
			updates := make(map[string]interface{})
			if req.Name != nil {
				updates["name"] = *req.Name
			}
			if req.Email != nil {
				updates["email"] = *req.Email
			}
			if req.Phone != nil {
				updates["phone"] = *req.Phone
			}
			if req.Department != nil {
				updates["department"] = *req.Department
			}
			if len(updates) > 0 {
				updates["updated_at"] = time.Now()
				if err := db.Model(&user).Updates(updates).Error; err != nil {
					c.JSON(500, gin.H{"code": 500, "message": "更新失败: " + err.Error()})
					return
				}
			}
			db.Preload("Roles").First(&user, userID)
			baseURL := strings.TrimRight(c.Request.Host, "/")
			if user.Avatar != nil && strings.HasPrefix(*user.Avatar, "/uploads/") {
				avatarURL := fmt.Sprintf("http://%s%s", baseURL, *user.Avatar)
				user.Avatar = &avatarURL
			}
			c.JSON(200, gin.H{"code": 200, "message": "更新成功", "data": user})
		})
		api.POST("/auth/avatar", func(c *gin.Context) {
			userID, _ := c.Get("userId")
			var user models.User
			if err := db.First(&user, userID).Error; err != nil {
				c.JSON(404, gin.H{"code": 404, "message": "用户不存在"})
				return
			}
			file, err := c.FormFile("avatar")
			if err != nil {
				c.JSON(400, gin.H{"code": 400, "message": "请上传头像文件"})
				return
			}
			if file.Size > 2*1024*1024 {
				c.JSON(400, gin.H{"code": 400, "message": "头像大小不能超过2MB"})
				return
			}
			allowedTypes := map[string]bool{"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true}
			if !allowedTypes[file.Header.Get("Content-Type")] {
				c.JSON(400, gin.H{"code": 400, "message": "只支持 JPG/PNG/GIF/WEBP 格式"})
				return
			}
			ext := ".jpg"
			switch file.Header.Get("Content-Type") {
			case "image/png":
				ext = ".png"
			case "image/gif":
				ext = ".gif"
			case "image/webp":
				ext = ".webp"
			}
			filename := fmt.Sprintf("avatars/%s%s", userID, ext)
			if err := c.SaveUploadedFile(file, "uploads/"+filename); err != nil {
				os.MkdirAll("uploads", 0755)
				if err := c.SaveUploadedFile(file, "uploads/"+filename); err != nil {
					c.JSON(500, gin.H{"code": 500, "message": "保存文件失败: " + err.Error()})
					return
				}
			}
			avatarURL := "/uploads/" + filename
			db.Model(&user).Update("avatar", avatarURL)
			baseURL := strings.TrimRight(c.Request.Host, "/")
			fullAvatarURL := fmt.Sprintf("http://%s%s", baseURL, avatarURL)
			c.JSON(200, gin.H{"code": 200, "message": "上传成功", "data": gin.H{"avatarUrl": fullAvatarURL}})
		})
		api.POST("/auth/logout", func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 200, "message": "登出成功"})
		})

		api.GET("/vms", func(c *gin.Context) { vmHandler.List(c) })
		api.GET("/vms/stats", func(c *gin.Context) { vmHandler.GetStats(c) })
		api.GET("/vms/:id", func(c *gin.Context) { vmHandler.Get(c) })
		api.POST("/vms", func(c *gin.Context) { vmHandler.Create(c) })
		api.PUT("/vms/:id", func(c *gin.Context) { vmHandler.Update(c) })
		api.DELETE("/vms/:id", func(c *gin.Context) { vmHandler.Delete(c) })
		api.GET("/vms/:id/metrics", func(c *gin.Context) { vmHandler.GetRealtimeMetrics(c) })

		// 获取所有VM的实时指标
		api.GET("/vms/metrics/all", func(c *gin.Context) {
			type VMMetrics struct {
				VMID        string    `json:"vmId"`
				VMName      string    `json:"vmName"`
				CPUUsage    float64   `json:"cpuUsage"`
				MemoryUsage float64   `json:"memoryUsage"`
				DiskUsage   float64   `json:"diskUsage"`
				DiskRead    float64   `json:"diskReadMbps"`
				DiskWrite   float64   `json:"diskWriteMbps"`
				NetworkIn   float64   `json:"networkInMbps"`
				NetworkOut  float64   `json:"networkOutMbps"`
				Temperature float64   `json:"temperature"`
				UpdatedAt   time.Time `json:"updatedAt"`
			}

			var metricsList []VMMetrics
			rows, err := db.Raw(`
				SELECT 
					v.id as vm_id,
					v.name as vm_name,
					m.cpu_usage,
					m.memory_usage,
					m.disk_usage,
					m.disk_read_mbps,
					m.disk_write_mbps,
					m.network_in_mbps,
					m.network_out_mbps,
					m.temperature,
					m.recorded_at
				FROM vms v
				LEFT JOIN vm_metrics m ON v.id = m.vm_id
				WHERE v.is_deleted = false
				ORDER BY m.recorded_at DESC
			`).Rows()

			if err != nil {
				c.JSON(500, gin.H{"code": 500, "message": "查询失败: " + err.Error()})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var m VMMetrics
				rows.Scan(&m.VMID, &m.VMName, &m.CPUUsage, &m.MemoryUsage, &m.DiskUsage, &m.DiskRead, &m.DiskWrite, &m.NetworkIn, &m.NetworkOut, &m.Temperature, &m.UpdatedAt)
				metricsList = append(metricsList, m)
			}

			c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": metricsList})
		})

		// 获取VM历史指标
		api.GET("/vms/:id/metrics/history", func(c *gin.Context) {
			id := c.Param("id")
			period := c.DefaultQuery("period", "24h")

			var hours int
			switch period {
			case "1h":
				hours = 1
			case "6h":
				hours = 6
			case "24h":
				hours = 24
			case "7d":
				hours = 168
			case "30d":
				hours = 720
			default:
				hours = 24
			}

			var history []struct {
				Timestamp   time.Time `json:"timestamp"`
				CPUUsage    float64   `json:"cpuUsage"`
				MemoryUsage float64   `json:"memoryUsage"`
				DiskUsage   float64   `json:"diskUsage"`
				DiskRead    float64   `json:"diskReadMbps"`
				DiskWrite   float64   `json:"diskWriteMbps"`
				NetworkIn   float64   `json:"networkInMbps"`
				NetworkOut  float64   `json:"networkOutMbps"`
				Temperature float64   `json:"temperature"`
			}

			startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

			rows, err := db.Table("vm_metrics_history").
				Select("recorded_at as timestamp, cpu_usage, memory_usage, disk_usage, disk_read_mbps, disk_write_mbps, network_in_mbps, network_out_mbps, temperature").
				Where("vm_id = ? AND recorded_at >= ?", id, startTime).
				Order("recorded_at ASC").
				Rows()

			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var h struct {
						Timestamp   time.Time `json:"timestamp"`
						CPUUsage    float64   `json:"cpuUsage"`
						MemoryUsage float64   `json:"memoryUsage"`
						DiskUsage   float64   `json:"diskUsage"`
						DiskRead    float64   `json:"diskReadMbps"`
						DiskWrite   float64   `json:"diskWriteMbps"`
						NetworkIn   float64   `json:"networkInMbps"`
						NetworkOut  float64   `json:"networkOutMbps"`
						Temperature float64   `json:"temperature"`
					}
					rows.Scan(&h.Timestamp, &h.CPUUsage, &h.MemoryUsage, &h.DiskUsage, &h.DiskRead, &h.DiskWrite, &h.NetworkIn, &h.NetworkOut, &h.Temperature)
					history = append(history, h)
				}
			}

			c.JSON(200, gin.H{"code": 200, "message": "获取成功", "data": gin.H{
				"period":  period,
				"metrics": history,
			}})
		})

		api.GET("/alerts/rules", func(c *gin.Context) { alertHandler.ListRules(c) })
		api.GET("/alerts/rules/:id", func(c *gin.Context) { alertHandler.GetRule(c) })
		api.POST("/alerts/rules", func(c *gin.Context) { alertHandler.CreateRule(c) })
		api.PUT("/alerts/rules/:id", func(c *gin.Context) { alertHandler.UpdateRule(c) })
		api.DELETE("/alerts/rules/:id", func(c *gin.Context) { alertHandler.DeleteRule(c) })
		api.GET("/alerts/records", func(c *gin.Context) { alertHandler.ListRecords(c) })
		api.GET("/alerts/stats", func(c *gin.Context) { alertHandler.GetStats(c) })
		api.POST("/alerts/:id/acknowledge", func(c *gin.Context) { alertHandler.Acknowledge(c) })
		api.POST("/alerts/:id/resolve", func(c *gin.Context) { alertHandler.Resolve(c) })
		api.POST("/alerts/test", func(c *gin.Context) { alertHandler.CreateTestAlert(c) })

		// 权限管理
		api.GET("/permissions", func(c *gin.Context) { permissionHandler.List(c) })
		api.GET("/permissions/:id", func(c *gin.Context) { permissionHandler.Get(c) })
		api.POST("/permissions", func(c *gin.Context) { permissionHandler.Create(c) })
		api.DELETE("/permissions/:id", func(c *gin.Context) { permissionHandler.Delete(c) })
		api.GET("/permissions/role/:roleId", func(c *gin.Context) { permissionHandler.ListByRole(c) })
		api.POST("/permissions/role/:roleId", func(c *gin.Context) { permissionHandler.AssignToRole(c) })

		// 角色管理
		api.GET("/roles", func(c *gin.Context) { roleHandler.List(c) })
		api.GET("/roles/:id", func(c *gin.Context) { roleHandler.Get(c) })
		api.POST("/roles", func(c *gin.Context) { roleHandler.Create(c) })
		api.PUT("/roles/:id", func(c *gin.Context) { roleHandler.Update(c) })
		api.DELETE("/roles/:id", func(c *gin.Context) { roleHandler.Delete(c) })
		api.GET("/roles/:id/users", func(c *gin.Context) { roleHandler.GetUsers(c) })
		api.POST("/roles/users", func(c *gin.Context) { roleHandler.AssignUser(c) })
		api.GET("/users/:userId/roles", func(c *gin.Context) { roleHandler.GetUserRoles(c) })
		api.GET("/users/:userId/permissions", func(c *gin.Context) { roleHandler.GetUserPermissions(c) })
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("🚀 API服务器启动: http://localhost:%d\n", cfg.Server.Port)

	router.Run(addr)
}
