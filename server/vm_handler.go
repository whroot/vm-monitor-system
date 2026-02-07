package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VMHandler VM管理处理器
type VMHandler struct {
	db *gorm.DB
}

// NewVMHandler 创建VM处理器
func NewVMHandler(db *gorm.DB) *VMHandler {
	return &VMHandler{db: db}
}

// VMResponse VM响应结构
type VMResponse struct {
	ID                 uuid.UUID  `json:"id"`
	VMwareID           *string    `json:"vmwareId,omitempty"`
	Name               string     `json:"name"`
	IP                 *string    `json:"ip,omitempty"`
	OSType             *string    `json:"osType,omitempty"`
	OSVersion          *string    `json:"osVersion,omitempty"`
	CpuCores           *int       `json:"cpuCores,omitempty"`
	MemoryGB           *int       `json:"memoryGb,omitempty"`
	DiskGB             *int       `json:"diskGb,omitempty"`
	NetworkAdapters    *int       `json:"networkAdapters,omitempty"`
	PowerState         *string    `json:"powerState,omitempty"`
	HostID             *string    `json:"hostId,omitempty"`
	HostName           *string    `json:"hostName,omitempty"`
	DatacenterID       *string    `json:"datacenterId,omitempty"`
	DatacenterName     *string    `json:"datacenterName,omitempty"`
	ClusterID          *string    `json:"clusterId,omitempty"`
	ClusterName        *string    `json:"clusterName,omitempty"`
	GroupID            *uuid.UUID `json:"groupId,omitempty"`
	Status             string     `json:"status"`
	LastSeen           *time.Time `json:"lastSeen,omitempty"`
	VMwareToolsStatus  *string    `json:"vmwareToolsStatus,omitempty"`
	VMwareToolsVersion *string    `json:"vmwareToolsVersion,omitempty"`
	Description        *string    `json:"description,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// List 获取VM列表
func (h *VMHandler) List(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")
	groupID := c.Query("groupId")

	// 计算偏移
	offset := (page - 1) * pageSize

	// 构建查询
	query := h.db.Model(&models.VM{}).Where("is_deleted = ?", false)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取列表
	var vms []models.VM
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&vms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询VM列表失败: " + err.Error(),
		})
		return
	}

	// 转换为响应格式
	vmResponses := make([]VMResponse, len(vms))
	for i, vm := range vms {
		vmResponses[i] = VMResponse{
			ID:           vm.ID,
			VMwareID:     vm.VMwareID,
			Name:         vm.Name,
			IP:           vm.IP,
			OSType:       vm.OSType,
			Status:       vm.Status,
			CreatedAt:    vm.CreatedAt,
			UpdatedAt:    vm.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"vms":     vmResponses,
			"total":   total,
			"page":    page,
			"pageSize": pageSize,
		},
	})
}

// Get 获取单个VM详情
func (h *VMHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var vm models.VM
	if err := h.db.Where("is_deleted = ?", false).First(&vm, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "VM不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": VMResponse{
			ID:           vm.ID,
			VMwareID:     vm.VMwareID,
			Name:         vm.Name,
			IP:           vm.IP,
			OSType:       vm.OSType,
			OSVersion:    vm.OSVersion,
			CpuCores:     vm.CpuCores,
			MemoryGB:     vm.MemoryGB,
			DiskGB:       vm.DiskGB,
			PowerState:   vm.PowerState,
			HostName:     vm.HostName,
			Status:       vm.Status,
			LastSeen:     vm.LastSeen,
			Description:  vm.Description,
			CreatedAt:    vm.CreatedAt,
			UpdatedAt:    vm.UpdatedAt,
		},
	})
}

// Create 创建VM
func (h *VMHandler) Create(c *gin.Context) {
	var req struct {
		Name         string  `json:"name" binding:"required"`
		VMwareID     *string `json:"vmwareId,omitempty"`
		IP           *string `json:"ip,omitempty"`
		OSType       *string `json:"osType,omitempty"`
		OSVersion    *string `json:"osVersion,omitempty"`
		CpuCores     *int    `json:"cpuCores,omitempty"`
		MemoryGB     *int    `json:"memoryGb,omitempty"`
		DiskGB       *int    `json:"diskGb,omitempty"`
		HostName     *string `json:"hostName,omitempty"`
		Description  *string `json:"description,omitempty"`
		GroupID      *string `json:"groupId,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 检查名称是否重复
	var existing models.VM
	if err := h.db.Where("name = ? AND is_deleted = ?", req.Name, false).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": "VM名称已存在",
		})
		return
	}

	// 解析groupID
	var groupIDPtr *uuid.UUID
	if req.GroupID != nil && *req.GroupID != "" {
		groupID, err := uuid.Parse(*req.GroupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "无效的groupId格式",
			})
			return
		}
		groupIDPtr = &groupID
	}

	// 创建VM
	vm := models.VM{
		ID:          uuid.New(),
		VMwareID:    req.VMwareID,
		Name:        req.Name,
		IP:          req.IP,
		OSType:      req.OSType,
		OSVersion:   req.OSVersion,
		CpuCores:   req.CpuCores,
		MemoryGB:    req.MemoryGB,
		DiskGB:      req.DiskGB,
		HostName:   req.HostName,
		Description: req.Description,
		GroupID:    groupIDPtr,
		Status:     "unknown",
	}

	if err := h.db.Create(&vm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "创建成功",
		"data": VMResponse{
			ID:          vm.ID,
			Name:        vm.Name,
			IP:          vm.IP,
			OSType:      vm.OSType,
			Status:      vm.Status,
			CreatedAt:   vm.CreatedAt,
			UpdatedAt:   vm.UpdatedAt,
		},
	})
}

// Update 更新VM
func (h *VMHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var vm models.VM
	if err := h.db.Where("is_deleted = ?", false).First(&vm, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "VM不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	var req struct {
		Name         *string `json:"name,omitempty"`
		IP           *string `json:"ip,omitempty"`
		OSType       *string `json:"osType,omitempty"`
		OSVersion    *string `json:"osVersion,omitempty"`
		CpuCores     *int    `json:"cpuCores,omitempty"`
		MemoryGB     *int    `json:"memoryGb,omitempty"`
		DiskGB       *int    `json:"diskGb,omitempty"`
		PowerState   *string `json:"powerState,omitempty"`
		HostName     *string `json:"hostName,omitempty"`
		Status       *string `json:"status,omitempty"`
		Description  *string `json:"description,omitempty"`
		VMwareToolsStatus *string `json:"vmwareToolsStatus,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.IP != nil {
		updates["ip"] = req.IP
	}
	if req.OSType != nil {
		updates["os_type"] = req.OSType
	}
	if req.OSVersion != nil {
		updates["os_version"] = req.OSVersion
	}
	if req.CpuCores != nil {
		updates["cpu_cores"] = req.CpuCores
	}
	if req.MemoryGB != nil {
		updates["memory_gb"] = req.MemoryGB
	}
	if req.DiskGB != nil {
		updates["disk_gb"] = req.DiskGB
	}
	if req.PowerState != nil {
		updates["power_state"] = req.PowerState
	}
	if req.HostName != nil {
		updates["host_name"] = req.HostName
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.VMwareToolsStatus != nil {
		updates["vmware_tools_status"] = req.VMwareToolsStatus
	}

	updates["updated_at"] = time.Now()

	if err := h.db.Model(&vm).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data": VMResponse{
			ID:          vm.ID,
			Name:        vm.Name,
			IP:          vm.IP,
			OSType:      vm.OSType,
			Status:      vm.Status,
			UpdatedAt:   time.Now(),
		},
	})
}

// Delete 删除VM
func (h *VMHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	var vm models.VM
	if err := h.db.Where("is_deleted = ?", false).First(&vm, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "VM不存在",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	// 软删除
	if err := h.db.Model(&vm).Updates(map[string]interface{}{
		"is_deleted": true,
		"deleted_at": time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetStats 获取VM统计
func (h *VMHandler) GetStats(c *gin.Context) {
	var stats struct {
		Total    int64 `json:"total"`
		Running  int64 `json:"running"`
		Stopped  int64 `json:"stopped"`
		Warning  int64 `json:"warning"`
		Unknown  int64 `json:"unknown"`
	}

	// 统计总数
	h.db.Model(&models.VM{}).Where("is_deleted = ?", false).Count(&stats.Total)

	// 统计各状态
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "running").Count(&stats.Running)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "poweredOff").Count(&stats.Stopped)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "warning").Count(&stats.Warning)
	h.db.Model(&models.VM{}).Where("is_deleted = ? AND status = ?", false, "unknown").Count(&stats.Unknown)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": stats,
	})
}

// GetMetrics 获取VM监控指标
func (h *VMHandler) GetMetrics(c *gin.Context) {
	id := c.Param("id")

	var vm models.VM
	if err := h.db.Where("is_deleted = ?", false).First(&vm, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "VM不存在",
		})
		return
	}

	// 解析时间范围
	duration := c.DefaultQuery("duration", "1h")
	var startTime time.Time
	now := time.Now()

	switch duration {
	case "1h":
		startTime = now.Add(-time.Hour)
	case "6h":
		startTime = now.Add(-6 * time.Hour)
	case "24h":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	default:
		startTime = now.Add(-time.Hour)
	}

	// 返回模拟的监控数据（实际应该从InfluxDB查询）
	metrics := gin.H{
		"vmId":      vm.ID,
		"vmName":    vm.Name,
		"duration":  duration,
		"timestamp": now,
		"dataPoints": gin.H{
			"cpuUsage":    generateMockDataPoints(20, 0, 100),
			"memoryUsage": generateMockDataPoints(20, 20, 95),
			"diskUsage":   generateMockDataPoints(20, 30, 90),
			"networkIn":   generateMockDataPoints(20, 0, 500),
			"networkOut":  generateMockDataPoints(20, 0, 300),
		},
		"summary": gin.H{
			"cpuUsageAvg":    45.2,
			"cpuUsageMax":    78.5,
			"cpuUsageMin":    12.3,
			"memoryUsageAvg": 62.8,
			"memoryUsageMax": 89.2,
			"memoryUsageMin": 35.4,
			"diskUsageAvg":   58.3,
			"diskUsageMax":   75.6,
			"diskUsageMin":   42.1,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    metrics,
	})
}

// GetRealtimeMetrics 获取实时监控数据
func (h *VMHandler) GetRealtimeMetrics(c *gin.Context) {
	id := c.Param("id")

	var vm models.VM
	if err := h.db.Where("is_deleted = ?", false).First(&vm, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "VM不存在",
		})
		return
	}

	// 返回实时指标（实际应该从vSphere API获取）
	metrics := gin.H{
		"vmId":      vm.ID,
		"vmName":    vm.Name,
		"timestamp": time.Now(),
		"cpu": gin.H{
			"usage":    randFloat(10, 90),
			"cores":    4,
			"mhz":      randFloat(1000, 3500),
		},
		"memory": gin.H{
			"usage":       randFloat(30, 90),
			"totalGB":     16,
			"usedGB":      randFloat(5, 14),
			"swapUsage":   randFloat(0, 20),
		},
		"disk": gin.H{
			"usage":       randFloat(40, 80),
			"totalGB":     500,
			"usedGB":      randFloat(200, 400),
			"readIOPS":    randFloat(0, 100),
			"writeIOPS":   randFloat(0, 80),
			"readMBps":    randFloat(0, 50),
			"writeMBps":   randFloat(0, 40),
		},
		"network": gin.H{
			"usageMbps": randFloat(0, 500),
			"inMBps":    randFloat(0, 200),
			"outMBps":   randFloat(0, 300),
			"packetsIn": randFloat(0, 10000),
			"packetsOut": randFloat(0, 8000),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    metrics,
	})
}

// BatchOperation 批量操作
func (h *VMHandler) BatchOperation(c *gin.Context) {
	var req struct {
		VMIDs     []string `json:"vmIds" binding:"required"`
		Operation string   `json:"operation" binding:"required"` // powerOn, powerOff, suspend, delete
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 验证操作类型
	validOps := map[string]bool{
		"powerOn":  true,
		"powerOff": true,
		"suspend":  true,
		"delete":   true,
	}

	if !validOps[req.Operation] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的操作类型",
		})
		return
	}

	// 执行批量操作
	success := 0
	failed := 0
	failedVMs := []map[string]string{}

	for _, vmID := range req.VMIDs {
		var vm models.VM
		if err := h.db.Where("is_deleted = ?", false).First(&vm, vmID).Error; err != nil {
			failed++
			failedVMs = append(failedVMs, map[string]string{
				"id":     vmID,
				"reason": "VM不存在",
			})
			continue
		}

		if req.Operation == "delete" {
			// 软删除
			if err := h.db.Model(&vm).Updates(map[string]interface{}{
				"is_deleted": true,
				"deleted_at": time.Now(),
			}).Error; err != nil {
				failed++
				failedVMs = append(failedVMs, map[string]string{
					"id":     vmID,
					"reason": "删除失败",
				})
			} else {
				success++
			}
		} else {
			// 更新电源状态
			stateMap := map[string]string{
				"powerOn":  "poweredOn",
				"powerOff": "poweredOff",
				"suspend":  "suspended",
			}

			if err := h.db.Model(&vm).Updates(map[string]interface{}{
				"power_state": stateMap[req.Operation],
				"status":      stateMap[req.Operation],
				"updated_at":  time.Now(),
			}).Error; err != nil {
				failed++
				failedVMs = append(failedVMs, map[string]string{
					"id":     vmID,
					"reason": "状态更新失败",
				})
			} else {
				success++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "批量操作完成",
		"data": gin.H{
			"total":    len(req.VMIDs),
			"success": success,
			"failed":  failed,
			"failedVMs": failedVMs,
		},
	})
}

// 辅助函数：生成模拟数据点
func generateMockDataPoints(count int, min, max float64) []gin.H {
	points := make([]gin.H, count)
	for i := 0; i < count; i++ {
		points[i] = gin.H{
			"timestamp": time.Now().Add(-time.Duration(count-i) * time.Minute).Format(time.RFC3339),
			"value":     randFloat(min, max),
		}
	}
	return points
}

// 辅助函数：生成随机浮点数
func randFloat(min, max float64) float64 {
	return min + (max-min)*float64(time.Now().Unix()%10000)/10000
}
