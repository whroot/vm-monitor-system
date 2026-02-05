package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"vm-monitoring-system/internal/logger"
	"vm-monitoring-system/internal/models"
	"vm-monitoring-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HistoryHandler 历史数据处理器
type HistoryHandler struct {
	db               *gorm.DB
	timeSeriesService *services.TimeSeriesService
}

// NewHistoryHandler 创建历史数据处理器
func NewHistoryHandler(db *gorm.DB) *HistoryHandler {
	return &HistoryHandler{
		db:               db,
		timeSeriesService: services.NewTimeSeriesService(db),
	}
}

// QueryRequest 查询历史数据请求
type QueryRequest struct {
	VMIDs          []string `json:"vmIds" binding:"required"`
	StartTime      string   `json:"startTime" binding:"required"`
	EndTime        string   `json:"endTime" binding:"required"`
	Metrics        []string `json:"metrics"`
	Aggregation    string   `json:"aggregation"`
	AggregationFunc string  `json:"aggregationFunc"`
	Page           int      `json:"page"`
	PageSize       int      `json:"pageSize"`
}

// Query 查询历史数据
func (h *HistoryHandler) Query(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误",
		})
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间不能早于开始时间",
		})
		return
	}

	// 查询历史数据
	metrics, err := h.timeSeriesService.QueryMetrics(req.VMIDs, req.Metrics, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询历史数据失败",
		})
		return
	}

	// 计算总点数
	totalPoints := len(metrics)

	// 分页处理
	var data []gin.H
	if req.Page > 0 && req.PageSize > 0 && totalPoints > 0 {
		start := (req.Page - 1) * req.PageSize
		end := start + req.PageSize
		if start >= totalPoints {
			data = []gin.H{}
		} else {
			if end > totalPoints {
				end = totalPoints
			}
			for _, m := range metrics[start:end] {
				data = append(data, gin.H{
					"vmId":      m.VMID,
					"metric":    m.Metric,
					"value":     m.Value,
					"timestamp": m.Timestamp.Format(time.RFC3339),
					"tags":      m.Tags,
				})
			}
		}
	} else {
		for _, m := range metrics {
			data = append(data, gin.H{
				"vmId":      m.VMID,
				"metric":    m.Metric,
				"value":     m.Value,
				"timestamp": m.Timestamp.Format(time.RFC3339),
				"tags":      m.Tags,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "查询成功",
		"data": gin.H{
			"data": data,
			"meta": gin.H{
				"startTime":      req.StartTime,
				"endTime":        req.EndTime,
				"metrics":        req.Metrics,
				"totalPoints":    totalPoints,
				"vmCount":        len(req.VMIDs),
			},
			"pagination": gin.H{
				"page":       req.Page,
				"pageSize":   req.PageSize,
				"total":      totalPoints,
				"totalPages": int(math.Ceil(float64(totalPoints) / float64(req.PageSize))),
			},
		},
	})
}

// Aggregate 聚合统计
func (h *HistoryHandler) Aggregate(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误",
		})
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间不能早于开始时间",
		})
		return
	}

	// 解析聚合参数
	var interval time.Duration
	if req.Aggregation != "" {
		var err error
		interval, err = time.ParseDuration(req.Aggregation)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "聚合间隔格式错误",
			})
			return
		}
	}

	// 聚合查询
	aggregates, err := h.timeSeriesService.AggregateMetrics(req.VMIDs, req.Metrics, startTime, endTime, interval, req.AggregationFunc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "聚合查询失败",
		})
		return
	}

	// 组织聚合结果
	var overall gin.H
	if len(aggregates) > 0 {
		overall = gin.H{
			"startTime": startTime.Format(time.RFC3339),
			"endTime":   endTime.Format(time.RFC3339),
			"interval":  req.Aggregation,
			"metrics":   gin.H{},
		}

		for _, agg := range aggregates {
			if _, ok := overall["metrics"].(gin.H)[agg.Metric]; !ok {
				overall["metrics"].(gin.H)[agg.Metric] = gin.H{}
			}
			overall["metrics"].(gin.H)[agg.Metric] = gin.H{
				"avg":       0.0,
				"max":       0.0,
				"min":       0.0,
				"sum":       0.0,
				"count":     0,
				"buckets":   []gin.H{},
			}
		}

		for _, agg := range aggregates {
			if _, ok := overall["metrics"].(gin.H)[agg.Metric]; ok {
				overall["metrics"].(gin.H)[agg.Metric].(gin.H)["buckets"] = append(
					overall["metrics"].(gin.H)[agg.Metric].(gin.H)["buckets"].([]gin.H),
				gin.H{
					"startTime": agg.Timestamp.Format(time.RFC3339),
					"endTime":   agg.Timestamp.Add(agg.Interval).Format(time.RFC3339),
					"value":     agg.Value,
					"count":     agg.Count,
				},
				)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "查询成功",
		"data": gin.H{
			"overall": overall,
		},
	})
}

// Trends 趋势分析
func (h *HistoryHandler) Trends(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误",
		})
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间不能早于开始时间",
		})
		return
	}

	// 查询历史数据
	metrics, err := h.timeSeriesService.QueryMetrics(req.VMIDs, req.Metrics, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询历史数据失败",
		})
		return
	}

	// 组织历史数据
	historical := []gin.H{}
	vmMetrics := make(map[string]map[string][]gin.H)
	
	for _, m := range metrics {
		if _, ok := vmMetrics[m.VMID]; !ok {
			vmMetrics[m.VMID] = make(map[string][]gin.H)
		}
		if _, ok := vmMetrics[m.VMID][m.Metric]; !ok {
			vmMetrics[m.VMID][m.Metric] = []gin.H{}
		}
		vmMetrics[m.VMID][m.Metric] = append(vmMetrics[m.VMID][m.Metric], gin.H{
			"timestamp": m.Timestamp.Format(time.RFC3339),
			"value":     m.Value,
			"tags":      m.Tags,
		})
	}

	for vmID, metrics := range vmMetrics {
		for metric, points := range metrics {
			historical = append(historical, gin.H{
				"vmId":      vmID,
				"metric":    metric,
				"points":    points,
				"count":     len(points),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "分析成功",
		"data": gin.H{
			"historical": historical,
		},
	})
}

// Anomalies 异常检测
func (h *HistoryHandler) Anomalies(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误",
		})
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间不能早于开始时间",
		})
		return
	}

	// 查询历史数据
	metrics, err := h.timeSeriesService.QueryMetrics(req.VMIDs, req.Metrics, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询历史数据失败",
		})
		return
	}

	// 简单异常检测（基于Z-score方法）
	anomalies := []gin.H{}
	metricValues := make(map[string][]float64)
	
	// 收集指标值
	for _, m := range metrics {
		key := fmt.Sprintf("%s:%s", m.VMID, m.Metric)
		metricValues[key] = append(metricValues[key], m.Value)
	}

	// 检测异常
	for key, values := range metricValues {
		if len(values) < 2 {
			continue
		}

		// 计算均值和标准差
		mean := 0.0
		for _, v := range values {
			mean += v
		}
		mean /= float64(len(values))

		variance := 0.0
		for _, v := range values {
			variance += math.Pow(v-mean, 2)
		}
		variance /= float64(len(values))
		stdDev := math.Sqrt(variance)

		// 检测Z-score > 3的异常点
		for _, m := range metrics {
			key2 := fmt.Sprintf("%s:%s", m.VMID, m.Metric)
			if key == key2 {
				zScore := (m.Value - mean) / stdDev
				if math.Abs(zScore) > 3 {
					anomalies = append(anomalies, gin.H{
						"vmId":      m.VMID,
						"metric":    m.Metric,
						"value":     m.Value,
						"timestamp": m.Timestamp.Format(time.RFC3339),
						"zScore":    zScore,
						"mean":      mean,
						"stdDev":    stdDev,
					})
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "查询成功",
		"data": gin.H{
			"anomalies": anomalies,
			"total":     len(anomalies),
		},
	})
}

// Export 导出数据
func (h *HistoryHandler) Export(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始时间格式错误",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间格式错误",
		})
		return
	}

	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束时间不能早于开始时间",
		})
		return
	}

	// 创建导出任务
	taskID := "export_" + uuid.New().String()
	
	// 异步处理导出
	go func() {
		metrics, err := h.timeSeriesService.QueryMetrics(req.VMIDs, req.Metrics, startTime, endTime)
		if err != nil {
			logger.Error("导出任务失败", zap.String("task_id", taskID), zap.Error(err))
			return
		}

		// 生成CSV
		csvData := "VM ID,Metric,Timestamp,Value\n"
		for _, m := range metrics {
			csvData += fmt.Sprintf("%s,%s,%s,%.2f\n",
				m.VMID, m.Metric, m.Timestamp.Format(time.RFC3339), m.Value)
		}

		// 保存文件
		filename := fmt.Sprintf("export_%s.csv", taskID)
		filepath := "./exports/" + filename

		if err := os.MkdirAll("./exports", 0755); err != nil {
			logger.Error("创建导出目录失败", zap.String("path", "./exports"), zap.Error(err))
			return
		}

		if err := os.WriteFile(filepath, []byte(csvData), 0644); err != nil {
			logger.Error("保存导出文件失败", zap.String("path", filepath), zap.Error(err))
			return
		}

		logger.Info("导出任务完成", zap.String("task_id", taskID), zap.String("path", filepath))
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"code":    202,
		"message": "导出任务已创建",
		"data": gin.H{
			"id":     taskID,
			"status": "pending",
		},
	})
}

// GetExportTask 获取导出任务状态
func (h *HistoryHandler) GetExportTask(c *gin.Context) {
	taskID := c.Param("id")
	
	// 检查文件是否存在
	filename := fmt.Sprintf("export_%s.csv", taskID)
	filepath := "./exports/" + filename
	
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "导出任务不存在或未完成",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"id":     taskID,
			"status": "completed",
			"file":   filename,
			"path":   filepath,
		},
	})
}

// GetTimeline 获取时间线事件
func (h *HistoryHandler) GetTimeline(c *gin.Context) {
	vmID := c.Param("vmId")
	
	// 查询最近的告警和状态变化
	var events []gin.H
	
	// 查询最近的告警记录
	var recentAlerts []models.AlertRecord
	if err := h.db.Model(&models.AlertRecord{}).
		Where("vm_id = ?", vmID).
		Order("created_at DESC").
		Limit(10).
		Find(&recentAlerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询告警记录失败",
		})
		return
	}

	for _, alert := range recentAlerts {
		events = append(events, gin.H{
			"type":      "alert",
			"title":     alert.RuleName,
			"severity":  alert.Severity,
			"message":   alert.Message,
			"timestamp": alert.CreatedAt.Format(time.RFC3339),
			"resolved":  alert.Resolved,
		})
	}

	// 查询最近的状态变化
	var recentStatusChanges []models.VM
	if err := h.db.Model(&models.VM{}).
		Where("vmware_id = ?", vmID).
		Order("updated_at DESC").
		Limit(10).
		Find(&recentStatusChanges).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询状态变化失败",
		})
		return
	}

	for _, vm := range recentStatusChanges {
		if vm.LastSeen != nil {
			events = append(events, gin.H{
				"type":      "status_change",
				"title":     fmt.Sprintf("状态变更为: %s", vm.Status),
				"severity":  "info",
				"message":   fmt.Sprintf("VM状态从%s变更为%s", vm.Status, vm.Status),
				"timestamp": vm.LastSeen.Format(time.RFC3339),
				"resolved":  true,
			})
		}
	}

	// 按时间排序
	sort.Slice(events, func(i, j int) bool {
		iTime, _ := time.Parse(time.RFC3339, events[i]["timestamp"].(string))
		jTime, _ := time.Parse(time.RFC3339, events[j]["timestamp"].(string))
		return iTime.After(jTime)
	})

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"events": events,
			"total":  len(events),
		},
	})
}
