package services

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MetricData 指标数据
type MetricData struct {
	ID        uuid.UUID `json:"id" db:"id"`
	VMID      string    `json:"vmId" db:"vm_id"`
	Metric    string    `json:"metric" db:"metric"`
	Value     float64   `json:"value" db:"value"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Tags      map[string]string `json:"tags,omitempty" db:"tags"`
}

// TimeSeriesService 时序数据服务
type TimeSeriesService struct {
	db *gorm.DB
}

// NewTimeSeriesService 创建时序数据服务
func NewTimeSeriesService(db *gorm.DB) *TimeSeriesService {
	return &TimeSeriesService{
		db: db,
	}
}

// InsertMetrics 批量插入指标数据
func (s *TimeSeriesService) InsertMetrics(metrics []MetricData) error {
	if len(metrics) == 0 {
		return nil
	}

	// 准备批量插入数据
	records := make([]MetricRecord, 0, len(metrics))
	for _, m := range metrics {
		records = append(records, MetricRecord{
			ID:        uuid.New(),
			VMID:      m.VMID,
			Metric:    m.Metric,
			Value:     m.Value,
			Timestamp: m.Timestamp,
			Tags:      m.Tags,
			CreatedAt: time.Now(),
		})
	}

	// 批量插入到数据库
	if err := s.db.CreateInBatches(records, 100).Error; err != nil {
		return fmt.Errorf("批量插入指标数据失败: %w", err)
	}

	log.Printf("已保存 %d 条指标数据", len(records))
	return nil
}

// QueryMetrics 查询指标数据
func (s *TimeSeriesService) QueryMetrics(vmIDs []string, metrics []string, startTime, endTime time.Time) ([]MetricData, error) {
	var records []MetricRecord

	query := s.db.Model(&MetricRecord{}).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime)

	if len(vmIDs) > 0 {
		query = query.Where("vm_id IN ?", vmIDs)
	}

	if len(metrics) > 0 {
		query = query.Where("metric IN ?", metrics)
	}

	if err := query.Order("timestamp ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("查询指标数据失败: %w", err)
	}

	// 转换为MetricData
	result := make([]MetricData, len(records))
	for i, r := range records {
		result[i] = MetricData{
			ID:        r.ID,
			VMID:      r.VMID,
			Metric:    r.Metric,
			Value:     r.Value,
			Timestamp: r.Timestamp,
			Tags:      r.Tags,
		}
	}

	return result, nil
}

// AggregateMetrics 聚合指标数据
func (s *TimeSeriesService) AggregateMetrics(vmIDs []string, metrics []string, startTime, endTime time.Time, interval time.Duration, aggregation string) ([]MetricAggregate, error) {
	// 构建时间桶查询
	buckets := s.generateTimeBuckets(startTime, endTime, interval)

	results := make([]MetricAggregate, 0, len(buckets))

	for i, bucket := range buckets {
		var records []MetricRecord

		query := s.db.Model(&MetricRecord{}).
			Where("timestamp BETWEEN ? AND ?", bucket.StartTime, bucket.EndTime)

		if len(vmIDs) > 0 {
			query = query.Where("vm_id IN ?", vmIDs)
		}

		if len(metrics) > 0 {
			query = query.Where("metric IN ?", metrics)
		}

		if err := query.Find(&records).Error; err != nil {
			log.Printf("聚合查询失败: %v", err)
			continue
		}

		if len(records) == 0 {
			continue
		}

		// 计算聚合值
		aggregate := s.calculateAggregate(records, aggregation)
		aggregate.Timestamp = bucket.StartTime
		aggregate.Interval = interval

		results = append(results, aggregate)
		i++
	}

	return results, nil
}

// GetLatestMetrics 获取最新指标数据
func (s *TimeSeriesService) GetLatestMetrics(vmIDs []string, metrics []string) (map[string]map[string]float64, error) {
	var records []MetricRecord

	subQuery := s.db.Model(&MetricRecord{}).
		Select("vm_id, metric, MAX(timestamp) as timestamp")

	if len(vmIDs) > 0 {
		subQuery = subQuery.Where("vm_id IN ?", vmIDs)
	}

	if len(metrics) > 0 {
		subQuery = subQuery.Where("metric IN ?", metrics)
	}

	subQuery = subQuery.Group("vm_id, metric")

	query := s.db.Table("metric_records m").
		Select("m.*").
		Joins("INNER JOIN (?) sub ON m.vm_id = sub.vm_id AND m.metric = sub.metric AND m.timestamp = sub.timestamp", subQuery)

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("查询最新指标数据失败: %w", err)
	}

	// 组织结果
	result := make(map[string]map[string]float64)
	for _, r := range records {
		if _, ok := result[r.VMID]; !ok {
			result[r.VMID] = make(map[string]float64)
		}
		result[r.VMID][r.Metric] = r.Value
	}

	return result, nil
}

// generateTimeBuckets 生成时间桶
func (s *TimeSeriesService) generateTimeBuckets(startTime, endTime time.Time, interval time.Duration) []TimeBucket {
	buckets := []TimeBucket{}

	for t := startTime; t.Before(endTime) || t.Equal(endTime); t = t.Add(interval) {
		bucket := TimeBucket{
			StartTime: t,
			EndTime:   t.Add(interval),
		}
		buckets = append(buckets, bucket)

		t = bucket.EndTime
	}

	return buckets
}

// calculateAggregate 计算聚合值
func (s *TimeSeriesService) calculateAggregate(records []MetricRecord, aggregation string) MetricAggregate {
	result := MetricAggregate{
		Metric: records[0].Metric,
		VMID:   records[0].VMID,
		Count:  len(records),
	}

	values := make([]float64, len(records))
	for i, r := range records {
		values[i] = r.Value
	}

	switch aggregation {
	case "avg":
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		result.Value = sum / float64(len(values))
	case "max":
		max := values[0]
		for _, v := range values {
			if v > max {
				max = v
			}
		}
		result.Value = max
	case "min":
		min := values[0]
		for _, v := range values {
			if v < min {
				min = v
			}
		}
		result.Value = min
	case "sum":
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		result.Value = sum
	default:
		// 默认使用平均值
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		result.Value = sum / float64(len(values))
	}

	return result
}

// CleanOldData 清理旧数据
func (s *TimeSeriesService) CleanOldData(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result := s.db.Where("timestamp < ?", cutoffTime).Delete(&MetricRecord{})
	if result.Error != nil {
		return fmt.Errorf("清理旧数据失败: %w", result.Error)
	}

	log.Printf("已清理 %d 条旧数据 (保留 %d 天)", result.RowsAffected, retentionDays)
	return nil
}

// GetStorageStats 获取存储统计
func (s *TimeSeriesService) GetStorageStats() (*StorageStats, error) {
	stats := &StorageStats{}

	// 总记录数
	if err := s.db.Model(&MetricRecord{}).Count(&stats.TotalRecords).Error; err != nil {
		return nil, err
	}

	// 最早记录时间
	var minTime time.Time
	if err := s.db.Model(&MetricRecord{}).Select("MIN(timestamp)").Scan(&minTime).Error; err != nil {
		return nil, err
	}
	stats.OldestRecord = minTime

	// 最新记录时间
	var maxTime time.Time
	if err := s.db.Model(&MetricRecord{}).Select("MAX(timestamp)").Scan(&maxTime).Error; err != nil {
		return nil, err
	}
	stats.LatestRecord = maxTime

	// 按指标类型统计
	var byMetric []struct {
		Metric string
		Count  int64
	}
	if err := s.db.Model(&MetricRecord{}).
		Select("metric, count(*) as count").
		Group("metric").
		Scan(&byMetric).Error; err != nil {
		return nil, err
	}

	stats.ByMetric = make(map[string]int64)
	for _, m := range byMetric {
		stats.ByMetric[m.Metric] = m.Count
	}

	return stats, nil
}

// MetricRecord 指标记录 (数据库模型)
type MetricRecord struct {
	ID        uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VMID      string            `gorm:"type:varchar(100);not null;index:idx_vm_metric_time,priority:3" json:"vmId"`
	Metric    string            `gorm:"type:varchar(50);not null;index:idx_vm_metric_time,priority:2" json:"metric"`
	Value     float64           `gorm:"type:double precision;not null" json:"value"`
	Timestamp time.Time         `gorm:"not null;index:idx_vm_metric_time,priority:1" json:"timestamp"`
	Tags      map[string]string `gorm:"type:jsonb" json:"tags,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
}

// TableName 指定表名
func (MetricRecord) TableName() string {
	return "metric_records"
}

// MetricAggregate 聚合指标
type MetricAggregate struct {
	VMID      string        `json:"vmId"`
	Metric    string        `json:"metric"`
	Value     float64       `json:"value"`
	Count     int           `json:"count"`
	Timestamp time.Time     `json:"timestamp"`
	Interval  time.Duration `json:"interval"`
}

// TimeBucket 时间桶
type TimeBucket struct {
	StartTime time.Time
	EndTime   time.Time
}

// StorageStats 存储统计
type StorageStats struct {
	TotalRecords int64            `json:"totalRecords"`
	OldestRecord time.Time        `json:"oldestRecord"`
	LatestRecord time.Time        `json:"latestRecord"`
	ByMetric     map[string]int64 `json:"byMetric"`
}
