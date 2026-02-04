package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTimeSeriesService(t *testing.T) {
	db, teardown := setupTestDB()
	defer teardown()

	service := NewTimeSeriesService(db)

	t.Run("InsertMetrics", func(t *testing.T) {
		metrics := []MetricData{
			{
				ID:        uuid.New(),
				VMID:      "vm-001",
				Metric:    "cpu_usage",
				Value:     75.5,
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New(),
				VMID:      "vm-001",
				Metric:    "memory_usage",
				Value:     60.2,
				Timestamp: time.Now(),
			},
		}

		err := service.InsertMetrics(metrics)
		assert.NoError(t, err)
	})

	t.Run("QueryMetrics", func(t *testing.T) {
		startTime := time.Now().Add(-time.Hour)
		endTime := time.Now()

		metrics, err := service.QueryMetrics([]string{"vm-001"}, []string{"cpu_usage", "memory_usage"}, startTime, endTime)
		assert.NoError(t, err)
		assert.Greater(t, len(metrics), 0)
	})

	t.Run("AggregateMetrics", func(t *testing.T) {
		startTime := time.Now().Add(-time.Hour)
		endTime := time.Now()

		aggregates, err := service.AggregateMetrics([]string{"vm-001"}, []string{"cpu_usage"}, startTime, endTime, 5*time.Minute, "avg")
		assert.NoError(t, err)
		assert.Greater(t, len(aggregates), 0)
	})

	t.Run("GetLatestMetrics", func(t *testing.T) {
		latest, err := service.GetLatestMetrics([]string{"vm-001"}, []string{"cpu_usage"})
		assert.NoError(t, err)
		assert.NotEmpty(t, latest)
	})

	t.Run("GetStorageStats", func(t *testing.T) {
		stats, err := service.GetStorageStats()
		assert.NoError(t, err)
		assert.NotNil(t, stats)
	})

	t.Run("CleanOldData", func(t *testing.T) {
		err := service.CleanOldData(1) // 清理1天前的数据
		assert.NoError(t, err)
	})
}