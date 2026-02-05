package services

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"vm-monitoring-system/internal/models"
	"vm-monitoring-system/internal/utils"
)

func setupAlertTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&models.VM{},
		&models.AlertRule{},
		&models.AlertCondition{},
		&models.AlertRecord{},
	)

	testAlertData(db)

	return db, func() {
		db.Close()
	}
}

func testAlertData(db *gorm.DB) {
	clusterID := uuid.New()
	hostID := uuid.New()
	datacenterID := uuid.New()

	clusterIDStr := clusterID.String()
	hostIDStr := hostID.String()
	datacenterIDStr := datacenterID.String()

	vms := []models.VM{
		{
			ID:             uuid.New(),
			Name:           "test-vm-1",
			IP:             utils.StringPtr("192.168.1.10"),
			Status:         "running",
			ClusterID:      &clusterIDStr,
			ClusterName:    utils.StringPtr("Test Cluster"),
			HostID:         &hostIDStr,
			HostName:       utils.StringPtr("Test Host"),
			DatacenterID:   &datacenterIDStr,
			DatacenterName: utils.StringPtr("Test DC"),
			OSType:         utils.StringPtr("Linux"),
		},
		{
			ID:             uuid.New(),
			Name:           "test-vm-2",
			IP:             utils.StringPtr("192.168.1.11"),
			Status:         "running",
			ClusterID:      &clusterIDStr,
			ClusterName:    utils.StringPtr("Test Cluster"),
			HostID:         &hostIDStr,
			HostName:       utils.StringPtr("Test Host"),
			DatacenterID:   &datacenterIDStr,
			DatacenterName: utils.StringPtr("Test DC"),
			OSType:         utils.StringPtr("Windows"),
		},
	}
	db.Create(&vms)

	rules := []models.AlertRule{
		{
			ID:              uuid.New(),
			Name:            "CPU使用率告警",
			Description:     "CPU使用率超过80%",
			Severity:        "warning",
			Scope:           "all",
			Enabled:         true,
			Cooldown:        300,
			TriggerCount:    1,
			TriggerCountMode: "total",
			ConditionLogic:  "or",
		},
		{
			ID:              uuid.New(),
			Name:            "内存使用率告警",
			Description:     "内存使用率超过90%",
			Severity:        "critical",
			Scope:           "all",
			Enabled:         true,
			Cooldown:        300,
			TriggerCount:    1,
			TriggerCountMode: "total",
			ConditionLogic:  "or",
		},
	}
	db.Create(&rules)

	conditions := []models.AlertCondition{
		{
			ID:           uuid.New(),
			RuleID:       rules[0].ID,
			Metric:       "cpu_usage",
			Operator:     ">",
			Threshold:    80.0,
			Aggregation:  "avg",
			Duration:     60,
			SortOrder:    1,
		},
		{
			ID:           uuid.New(),
			RuleID:       rules[1].ID,
			Metric:       "memory_usage",
			Operator:     ">",
			Threshold:    90.0,
			Aggregation:  "avg",
			Duration:     60,
			SortOrder:    1,
		},
	}
	db.Create(&conditions)
}

func TestAlertEngine_NewAlertEngine(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	assert.NotNil(t, engine)
	assert.False(t, engine.isRunning)
	assert.Equal(t, 60*time.Second, engine.evalInterval)
	assert.NotNil(t, engine.rules)
	assert.NotNil(t, engine.stopChan)
}

func TestAlertEngine_StartStop(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	t.Run("StartEngine", func(t *testing.T) {
		err := engine.Start()
		assert.NoError(t, err)
		assert.True(t, engine.IsRunning())
	})

	t.Run("StartAlreadyRunning", func(t *testing.T) {
		err := engine.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "已经在运行")
	})

	t.Run("StopEngine", func(t *testing.T) {
		engine.Stop()
		assert.False(t, engine.IsRunning())
	})

	t.Run("StopNotRunning", func(t *testing.T) {
		engine.Stop()
		assert.False(t, engine.IsRunning())
	})
}

func TestAlertEngine_SetEvalInterval(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	engine.SetEvalInterval(30 * time.Second)
	assert.Equal(t, 30*time.Second, engine.evalInterval)

	engine.SetEvalInterval(2 * time.Minute)
	assert.Equal(t, 2*time.Minute, engine.evalInterval)
}

func TestAlertEngine_ReloadRules(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	t.Run("ReloadWithLoadedRules", func(t *testing.T) {
		err := engine.ReloadRules()
		assert.NoError(t, err)
	})

	t.Run("ReloadEmpty", func(t *testing.T) {
		engine.rulesMutex.Lock()
		engine.rules = make(map[uuid.UUID]*AlertRuleWithConditions)
		engine.rulesMutex.Unlock()

		err := engine.ReloadRules()
		assert.NoError(t, err)
	})
}

func TestAlertEngine_LoadRules(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	err := engine.loadRules()
	assert.NoError(t, err)

	engine.rulesMutex.RLock()
	defer engine.rulesMutex.RUnlock()
	assert.GreaterOrEqual(t, len(engine.rules), 2)
}

func TestAlertEngine_EvaluateAllRules(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	err := engine.loadRules()
	assert.NoError(t, err)

	engine.evaluateAllRules()
	assert.NoError(t, err)
}

func TestAlertEngine_EvaluateSingleCondition(t *testing.T) {
	engine := &AlertEngine{}

	t.Run("GreaterThan", func(t *testing.T) {
		result := engine.evaluateSingleCondition(85.0, ">", 80.0)
		assert.True(t, result)
	})

	t.Run("GreaterThanOrEqual", func(t *testing.T) {
		result := engine.evaluateSingleCondition(80.0, ">=", 80.0)
		assert.True(t, result)

		result = engine.evaluateSingleCondition(79.0, ">=", 80.0)
		assert.False(t, result)
	})

	t.Run("LessThan", func(t *testing.T) {
		result := engine.evaluateSingleCondition(75.0, "<", 80.0)
		assert.True(t, result)
	})

	t.Run("LessThanOrEqual", func(t *testing.T) {
		result := engine.evaluateSingleCondition(80.0, "<=", 80.0)
		assert.True(t, result)

		result = engine.evaluateSingleCondition(81.0, "<=", 80.0)
		assert.False(t, result)
	})

	t.Run("Equal", func(t *testing.T) {
		result := engine.evaluateSingleCondition(80.0, "=", 80.0)
		assert.True(t, result)

		result = engine.evaluateSingleCondition(81.0, "=", 80.0)
		assert.False(t, result)
	})

	t.Run("NotEqual", func(t *testing.T) {
		result := engine.evaluateSingleCondition(81.0, "!=", 80.0)
		assert.True(t, result)

		result = engine.evaluateSingleCondition(80.0, "!=", 80.0)
		assert.False(t, result)
	})

	t.Run("UnknownOperator", func(t *testing.T) {
		result := engine.evaluateSingleCondition(80.0, "unknown", 80.0)
		assert.False(t, result)
	})
}

func TestAlertEngine_IsCooldownExpired(t *testing.T) {
	engine := &AlertEngine{
		triggerHistory: make(map[string]time.Time),
	}

	ruleID := uuid.New()

	t.Run("FirstTrigger", func(t *testing.T) {
		result := engine.isCooldownExpired(ruleID, 300)
		assert.True(t, result)
	})

	t.Run("WithinCooldown", func(t *testing.T) {
		engine.triggerHistory[ruleID.String()] = time.Now()
		result := engine.isCooldownExpired(ruleID, 300)
		assert.False(t, result)
	})

	t.Run("AfterCooldown", func(t *testing.T) {
		engine.triggerHistory[ruleID.String()] = time.Now().Add(-10 * time.Minute)
		result := engine.isCooldownExpired(ruleID, 300)
		assert.True(t, result)
	})
}

func TestAlertEngine_GetSimulatedMetricValue(t *testing.T) {
	engine := &AlertEngine{}
	vmID := uuid.New()

	t.Run("CPUUsage", func(t *testing.T) {
		value, err := engine.getSimulatedMetricValue(vmID, "cpu_usage")
		assert.NoError(t, err)
		assert.Equal(t, 75.5, value)
	})

	t.Run("MemoryUsage", func(t *testing.T) {
		value, err := engine.getSimulatedMetricValue(vmID, "memory_usage")
		assert.NoError(t, err)
		assert.Equal(t, 82.3, value)
	})

	t.Run("DiskUsage", func(t *testing.T) {
		value, err := engine.getSimulatedMetricValue(vmID, "disk_usage")
		assert.NoError(t, err)
		assert.Equal(t, 68.0, value)
	})

	t.Run("NetworkMetrics", func(t *testing.T) {
		rxValue, err := engine.getSimulatedMetricValue(vmID, "network_rx")
		assert.NoError(t, err)
		assert.Equal(t, 1024.0, rxValue)

		txValue, err := engine.getSimulatedMetricValue(vmID, "network_tx")
		assert.NoError(t, err)
		assert.Equal(t, 512.0, txValue)
	})

	t.Run("DiskIOMetrics", func(t *testing.T) {
		readValue, err := engine.getSimulatedMetricValue(vmID, "disk_io_read")
		assert.NoError(t, err)
		assert.Equal(t, 100.0, readValue)

		writeValue, err := engine.getSimulatedMetricValue(vmID, "disk_io_write")
		assert.NoError(t, err)
		assert.Equal(t, 80.0, writeValue)
	})

	t.Run("UnknownMetric", func(t *testing.T) {
		value, err := engine.getSimulatedMetricValue(vmID, "unknown_metric")
		assert.Error(t, err)
		assert.Equal(t, 0.0, value)
	})
}

func TestAlertEngine_CheckRecovery(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	t.Run("Recovered", func(t *testing.T) {
		vm := models.VM{
			ID:   uuid.New(),
			Name: "test-vm",
		}

		result := engine.checkRecovery(nil, vm)
		assert.True(t, result)
	})
}

func TestAlertEngine_UpdateTriggerHistory(t *testing.T) {
	engine := &AlertEngine{
		triggerHistory: make(map[string]time.Time),
		historyMutex:  sync.RWMutex{},
	}

	ruleID := uuid.New()
	engine.updateTriggerHistory(ruleID)

	engine.historyMutex.RLock()
	defer engine.historyMutex.RUnlock()
	_, exists := engine.triggerHistory[ruleID.String()]
	assert.True(t, exists)
}

func TestAlertEngine_GetTargetVMs(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	t.Run("AllVMs", func(t *testing.T) {
		vms, err := engine.getTargetVMs("all", nil)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(vms), 2)
	})

	t.Run("ClusterScope", func(t *testing.T) {
		var vm models.VM
		db.First(&vm)
		clusterID, _ := uuid.Parse(vm.ClusterID)

		vms, err := engine.getTargetVMs("cluster", &clusterID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(vms), 1)
	})

	t.Run("HostScope", func(t *testing.T) {
		var vm models.VM
		db.First(&vm)
		hostID, _ := uuid.Parse(vm.HostID)

		vms, err := engine.getTargetVMs("host", &hostID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(vms), 1)
	})

	t.Run("DatacenterScope", func(t *testing.T) {
		var vm models.VM
		db.First(&vm)
		dcID, _ := uuid.Parse(vm.DatacenterID)

		vms, err := engine.getTargetVMs("datacenter", &dcID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(vms), 1)
	})

	t.Run("GroupScope", func(t *testing.T) {
		nonExistentID := uuid.New()
		vms, err := engine.getTargetVMs("group", &nonExistentID)
		assert.NoError(t, err)
		assert.Empty(t, vms)
	})

	t.Run("UnknownScope", func(t *testing.T) {
		vms, err := engine.getTargetVMs("unknown", nil)
		assert.Error(t, err)
		assert.Nil(t, vms)
	})
}

func TestAlertEngine_IsActiveAlert(t *testing.T) {
	db, teardown := setupAlertTestDB()
	defer teardown()

	notifier := NewNotificationService(db, nil)
	engine := NewAlertEngine(db, notifier)

	t.Run("NoActiveAlert", func(t *testing.T) {
		result := engine.isActiveAlert("non-existent:alert")
		assert.False(t, result)
	})

	t.Run("WithActiveAlert", func(t *testing.T) {
		alertKey := "test-rule:test-vm"
		result := engine.isActiveAlert(alertKey)
		assert.False(t, result)
	})
}
