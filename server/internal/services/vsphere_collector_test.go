package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestVSphereCollector(t *testing.T) {
	db, teardown := setupTestDB()
	defer teardown()

	// 创建测试配置
	config := &VSphereConfig{
		Host:        "localhost",
		Port:        443,
		Username:    "testuser",
		Password:    "testpass",
		Insecure:    true,
		CollectInterval: 30 * time.Second,
		BatchSize:   10,
	}

	collector := NewVSphereCollector(db, config)

	t.Run("NewVSphereCollector", func(t *testing.T) {
		assert.NotNil(t, collector)
		assert.Equal(t, 30*time.Second, collector.config.CollectInterval)
		assert.Equal(t, 10, collector.config.BatchSize)
	})

	t.Run("StartStop", func(t *testing.T) {
		err := collector.Start()
		assert.NoError(t, err)
		assert.True(t, collector.isRunning)

		collector.Stop()
		assert.False(t, collector.isRunning)
	})

	t.Run("MetricValue", func(t *testing.T) {
		metric := MetricValue{
			VMID:      "vm-001",
			Timestamp: time.Now(),
			Metric:    "cpu_usage",
			Value:     75.5,
		}

		assert.Equal(t, "vm-001", metric.VMID)
		assert.Equal(t, "cpu_usage", metric.Metric)
		assert.Equal(t, 75.5, metric.Value)
	})

	t.Run("VMStatus", func(t *testing.T) {
		status := &VMStatus{
			VMID:       "vm-001",
			Name:       "Test VM",
			PowerState: "poweredOn",
			IPAddress:  "192.168.1.100",
			CPUUsage:   new(int32),
			MemoryUsage: new(int32),
			DiskUsage:  new(int32),
			UpdatedAt:  time.Now(),
		}
		*status.CPUUsage = 50
		*status.MemoryUsage = 2048
		*status.DiskUsage = 100

		assert.Equal(t, "vm-001", status.VMID)
		assert.Equal(t, "Test VM", status.Name)
		assert.Equal(t, "poweredOn", status.PowerState)
		assert.Equal(t, "192.168.1.100", status.IPAddress)
		assert.Equal(t, int32(50), *status.CPUUsage)
	})
}

// setupTestDB 创建测试数据库
func setupTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	// 迁移表结构
	db.AutoMigrate(
		&models.VM{},
		&models.MetricRecord{},
	)

	return db, func() {
		db.Close()
	}
}